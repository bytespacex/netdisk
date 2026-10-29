package main

import (
	"fmt"
	"io"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
)

// transferLogThreshold is the minimum byte count for a transfer to be recorded in
// the connection diagnostics log. Requests below this threshold without errors are
// not logged, to avoid flooding the log.
const transferLogThreshold = 8 << 20 // 8 MiB

// clientIP extracts the client address (without port) from RemoteAddr.
func clientIP(r *http.Request) string {
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

// countingReader wraps the request body and records how many bytes were actually
// read from the connection. On an interrupted upload this is how much the server had received.
type countingReader struct {
	rc io.ReadCloser
	n  int64
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.rc.Read(p)
	c.n += int64(n)
	return n, err
}

func (c *countingReader) Close() error { return c.rc.Close() }

// transferRecorder wraps ResponseWriter and records the status code, bytes written,
// and write errors. On an interrupted download, err is usually non-nil
// (broken pipe / connection reset).
type transferRecorder struct {
	http.ResponseWriter
	status  int
	written int64
	err     error
}

func (t *transferRecorder) WriteHeader(code int) {
	if t.status == 0 {
		t.status = code
	}
	t.ResponseWriter.WriteHeader(code)
}

func (t *transferRecorder) Write(p []byte) (int, error) {
	if t.status == 0 {
		t.status = http.StatusOK
	}
	n, err := t.ResponseWriter.Write(p)
	t.written += int64(n)
	if err != nil {
		t.err = err
	}
	return n, err
}

// ReadFrom passes through to the underlying ReaderFrom to preserve the sendfile
// zero-copy path for large files.
func (t *transferRecorder) ReadFrom(src io.Reader) (int64, error) {
	if t.status == 0 {
		t.status = http.StatusOK
	}
	var (
		n   int64
		err error
	)
	if rf, ok := t.ResponseWriter.(io.ReaderFrom); ok {
		n, err = rf.ReadFrom(src)
	} else {
		n, err = io.Copy(struct{ io.Writer }{t}, src)
	}
	t.written += n
	if err != nil {
		t.err = err
	}
	return n, err
}

func (t *transferRecorder) Flush() {
	if f, ok := t.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// transferLogMiddleware records connection-level diagnostics for uploads/downloads
// (and any response above the threshold), and catches handler panics. It only writes
// logs and does not change existing behavior.
func transferLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &transferRecorder{ResponseWriter: w}
		cr := &countingReader{rc: r.Body}
		r.Body = cr

		defer func() {
			if p := recover(); p != nil {
				writeLog("net", fmt.Sprintf(
					"(%s) PANIC %s %s: %v | reqBytes=%d respBytes=%d | %s",
					clientIP(r), r.Method, r.URL.Path, p, cr.n, rec.written,
					oneLine(string(debug.Stack()))))
				panic(p) // hand back to net/http default handling (closes the connection)
			}
		}()

		next.ServeHTTP(rec, r)

		ctxErr := r.Context().Err()
		if rec.written < transferLogThreshold && cr.n < transferLogThreshold &&
			ctxErr == nil && rec.err == nil {
			return
		}

		status := rec.status
		if status == 0 {
			status = http.StatusOK
		}
		detail := ""
		switch {
		case rec.err != nil:
			detail = " writeErr=" + rec.err.Error()
		case ctxErr != nil:
			detail = " ctxErr=" + ctxErr.Error()
		}
		writeLog("net", fmt.Sprintf(
			"(%s) %s %s status=%d reqBytes=%d respBytes=%d dur=%s%s",
			clientIP(r), r.Method, r.URL.Path, status, cr.n, rec.written,
			time.Since(start).Round(time.Millisecond), detail))
	})
}

func oneLine(s string) string {
	s = strings.ReplaceAll(s, "\n", " | ")
	s = strings.ReplaceAll(s, "\t", " ")
	if len(s) > 2000 {
		s = s[:2000]
	}
	return s
}
