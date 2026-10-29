package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

// Embedded subtitle support: uses external ffprobe/ffmpeg to detect and extract
// subtitle tracks inside MKV/MP4 containers. If ffmpeg/ffprobe are not installed,
// it automatically falls back to "sidecar subtitle files only".

var (
	ffOnce                sync.Once
	ffmpegBin, ffprobeBin string
)

func detectFFmpeg() {
	ffOnce.Do(func() {
		if p, err := exec.LookPath("ffmpeg"); err == nil {
			ffmpegBin = p
		}
		if p, err := exec.LookPath("ffprobe"); err == nil {
			ffprobeBin = p
		}
	})
}

func haveFFprobe() bool { detectFFmpeg(); return ffprobeBin != "" }
func haveFFmpeg() bool  { detectFFmpeg(); return ffmpegBin != "" }

// Containers that may carry embedded subtitles
func isEmbeddableSubExt(ext string) bool {
	switch ext {
	case ".mkv", ".webm", ".mp4", ".m4v", ".mov", ".avi":
		return true
	}
	return false
}

type probeKey struct {
	path  string
	size  int64
	mtime int64
}

var (
	probeCache = map[probeKey][]SubtitleTrack{}
	probeMu    sync.Mutex
)

// probeEmbeddedSubs lists subtitle tracks inside the container via ffprobe
// (returns global stream indices).
func probeEmbeddedSubs(absPath string) []SubtitleTrack {
	if !haveFFprobe() {
		return nil
	}
	fi, err := os.Stat(absPath)
	if err != nil {
		return nil
	}
	key := probeKey{absPath, fi.Size(), fi.ModTime().UnixNano()}
	probeMu.Lock()
	if v, ok := probeCache[key]; ok {
		probeMu.Unlock()
		return v
	}
	probeMu.Unlock()

	cmd := exec.Command(ffprobeBin, "-v", "error",
		"-select_streams", "s",
		"-show_entries", "stream=index:stream_tags=language,title",
		"-of", "json", absPath)
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	var parsed struct {
		Streams []struct {
			Index int `json:"index"`
			Tags  struct {
				Language string `json:"language"`
				Title    string `json:"title"`
			} `json:"tags"`
		} `json:"streams"`
	}
	if json.Unmarshal(out, &parsed) != nil {
		return nil
	}

	var subs []SubtitleTrack
	for i, s := range parsed.Streams {
		lang := s.Tags.Language
		if lang == "" {
			lang = s.Tags.Title
		}
		if lang == "" {
			lang = fmt.Sprintf("Track %d", i+1)
		}
		subs = append(subs, SubtitleTrack{Lang: lang, Embedded: true, Index: s.Index})
	}

	probeMu.Lock()
	if len(probeCache) > 512 {
		probeCache = map[probeKey][]SubtitleTrack{}
	}
	probeCache[key] = subs
	probeMu.Unlock()
	return subs
}

// subtitlesListHandler: GET /_actions/subtitles/{vol}/{path}
// Returns the available subtitle list for the video (sidecar + embedded).
func subtitlesListHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(strings.TrimPrefix(r.URL.Path, "/_actions/subtitles/"), "/")
	_, info := getVolAndPath(path)
	absPath := getFullAbsPath(path)
	if info == nil || absPath == "" {
		http.NotFound(w, r)
		return
	}
	if !hasPermission(getUser(r), info, 'r') {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	fi, err := os.Stat(absPath)
	if err != nil || fi.IsDir() {
		http.NotFound(w, r)
		return
	}

	ext := strings.ToLower(filepath.Ext(absPath))
	base := strings.TrimSuffix(filepath.Base(absPath), filepath.Ext(absPath))
	webPrefix := ""
	if i := strings.LastIndex(path, "/"); i != -1 {
		webPrefix = path[:i+1]
	}

	subs := findSubtitles(filepath.Dir(absPath), base, webPrefix)
	if isEmbeddableSubExt(ext) {
		subs = append(subs, probeEmbeddedSubs(absPath)...)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")
	json.NewEncoder(w).Encode(subs)
}

// subtitleExtractHandler: GET /_actions/subtitle/{vol}/{path}?stream=N
// Uses ffmpeg to extract the N-th (global index) subtitle track to WebVTT and streams it back.
func subtitleExtractHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(strings.TrimPrefix(r.URL.Path, "/_actions/subtitle/"), "/")
	_, info := getVolAndPath(path)
	absPath := getFullAbsPath(path)
	if info == nil || absPath == "" {
		http.NotFound(w, r)
		return
	}
	if !hasPermission(getUser(r), info, 'r') {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if !haveFFmpeg() {
		http.Error(w, "ffmpeg not available", http.StatusNotImplemented)
		return
	}
	idx, err := strconv.Atoi(r.URL.Query().Get("stream"))
	if err != nil || idx < 0 {
		http.Error(w, "invalid stream", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/vtt; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")

	cmd := exec.Command(ffmpegBin, "-v", "error", "-i", absPath,
		"-map", fmt.Sprintf("0:%d", idx), "-f", "webvtt", "pipe:1")
	cmd.Stdout = w
	var errBuf strings.Builder
	cmd.Stderr = &errBuf
	if runErr := cmd.Run(); runErr != nil {
		writeLog("subtitle", fmt.Sprintf("(%s) EXTRACT-FAILED %s stream=%d: %v %s",
			clientIP(r), path, idx, runErr, strings.TrimSpace(errBuf.String())))
	}
}
