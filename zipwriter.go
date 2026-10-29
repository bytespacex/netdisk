package main

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
)

// ZipStreamingWriter writes files to a ZIP stream:
//   - never reads whole file contents into memory (avoids OOM on large files)
//   - never uses disk temp files
//   - uses archive/zip, so ZIP64 (>=4GB files / offsets) is handled automatically
type ZipStreamingWriter struct {
	zw *zip.Writer
}

func NewZipStreamingWriter(w io.Writer) *ZipStreamingWriter {
	return &ZipStreamingWriter{zw: zip.NewWriter(w)}
}

// AddFile adds absPath (a file or directory) to the archive. Directories are expanded recursively.
// Errors are returned so callers can log them; callers must not start writing the response body
// before an HTTP error can still be returned.
func (z *ZipStreamingWriter) AddFile(absPath, arcname string) error {
	fi, err := os.Stat(absPath)
	if err != nil {
		return err
	}
	if fi.IsDir() {
		return filepath.Walk(absPath, func(p string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
				// Do not follow symlinks out of the shared directory.
				return nil
			}
			rel, err := filepath.Rel(absPath, p)
			if err != nil {
				return nil
			}
			return z.addFile(p, filepath.Join(arcname, rel), info)
		})
	}
	return z.addFile(absPath, arcname, fi)
}

func (z *ZipStreamingWriter) addFile(absPath, arcname string, fi os.FileInfo) error {
	f, err := os.Open(absPath)
	if err != nil {
		return err
	}
	defer f.Close()

	hdr := &zip.FileHeader{
		Name:     filepath.ToSlash(arcname),
		Method:   zip.Store, // same as the old implementation: no compression (faster for media / pre-compressed files)
		Modified: fi.ModTime(),
	}
	hdr.SetMode(fi.Mode())

	fw, err := z.zw.CreateHeader(hdr)
	if err != nil {
		return err
	}
	_, err = io.Copy(fw, f)
	return err
}

func (z *ZipStreamingWriter) Close() error {
	return z.zw.Close()
}
