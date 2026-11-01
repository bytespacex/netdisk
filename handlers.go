package main

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var activityLogger = slog.Default()
var logFile *os.File
var logMu sync.Mutex
var configDir string

func initLogger() {
	os.MkdirAll(configDir, 0755)
	logPath := filepath.Join(configDir, "netdisk.log")
	var err error
	logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("Failed to open log file: %v", err)
	}
}

func writeLog(category, msg string) {
	logMu.Lock()
	defer logMu.Unlock()
	if logFile == nil {
		return
	}
	ts := time.Now().Format("2006-01-02 15:04:05")
	fmt.Fprintf(logFile, "[%s] [%s] %s\n", ts, category, msg)
	logFile.Sync()
}

func getLogFile() string {
	return filepath.Join(configDir, "netdisk.log")
}

// ==============================================================================
// ===== Types ==================================================================
// ==============================================================================

type Volume struct {
	Path  string
	Perms string
	Users []string
}

type Session struct {
	Username         string
	ClipboardSources []string
	ClipboardAction  string
	CreatedAt        time.Time
}

type FileItem struct {
	Name           string
	Path           string
	IsDir          bool
	Size           int64
	Mtime          time.Time
	FormattedMtime string
	CanDelete      bool
	CanEdit        bool
	CanRename      bool
}

type Breadcrumb struct {
	Name string
	Path string
}

type IndexData struct {
	Items       []FileItem
	Breadcrumbs []Breadcrumb
	CurrentPath string
	Playlist    []PlaylistItem
	AllVolumes  []string
	CurrentUser string
	TotalSize   int64
	FileCount   int
	UserPerms   string
	SortBy      string
	Order       string
	NameOrder   string
	SizeOrder   string
	MtimeOrder  string
	CanPaste    bool
	Referrer    string
	ShowHidden  bool
	Version     string
}

type EditData struct {
	Path     string
	Content  string
	Referrer string
}

type SubtitleTrack struct {
	Lang     string `json:"lang"`
	Path     string `json:"path,omitempty"`
	Embedded bool   `json:"embedded,omitempty"`
	Index    int    `json:"index"`
}

type PlaylistItem struct {
	Name      string          `json:"name"`
	Path      string          `json:"path"`
	IsVideo   bool            `json:"is_video,omitempty"`
	Subtitles []SubtitleTrack `json:"subtitles,omitempty"`
}

// ==============================================================================
// ===== Global State ===========================================================
// ==============================================================================

var (
	users   = make(map[string]string) // username -> password
	volumes = make(map[string]*Volume)

	sessions   = make(map[string]*Session)
	sessionMu  sync.Mutex
	sessionTTL time.Duration

	recycleBin   string
	webclipDir   string
	webclipFiles string
)

// ==============================================================================
// ===== Session Management =====================================================
// ==============================================================================

func getSession(r *http.Request) *Session {
	cookie, err := r.Cookie("cd_session")
	if err != nil {
		return nil
	}
	sessionMu.Lock()
	defer sessionMu.Unlock()
	sess, ok := sessions[cookie.Value]
	if !ok {
		return nil
	}
	if sessionTTL > 0 && time.Since(sess.CreatedAt) > sessionTTL {
		delete(sessions, cookie.Value)
		return nil
	}
	return sess
}

func createSession(w http.ResponseWriter, username string) *Session {
	b := make([]byte, 16)
	rand.Read(b)
	sid := hex.EncodeToString(b)

	sess := &Session{
		Username:  username,
		CreatedAt: time.Now(),
	}
	sessionMu.Lock()
	sessions[sid] = sess
	sessionMu.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     "cd_session",
		Value:    sid,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   int(sessionTTL.Seconds()),
	})
	return sess
}

func clearSession(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("cd_session")
	if err == nil {
		sessionMu.Lock()
		delete(sessions, cookie.Value)
		sessionMu.Unlock()
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "cd_session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
}

func cleanExpiredSessions() {
	for {
		time.Sleep(5 * time.Minute)
		if sessionTTL <= 0 {
			continue
		}
		sessionMu.Lock()
		now := time.Now()
		for k, v := range sessions {
			if now.Sub(v.CreatedAt) > sessionTTL {
				delete(sessions, k)
			}
		}
		sessionMu.Unlock()
	}
}

// ==============================================================================
// ===== Auth Middleware =========================================================
// ==============================================================================

func authRequired(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")

		volName, volInfo := getVolAndPath(path)

		if sess := getSession(r); sess != nil {
			if _, ok := users[sess.Username]; ok {
				if volName != "" && volInfo != nil && !hasPermission(sess.Username, volInfo, 'r') {
					clearSession(w, r)
				} else {
					r.Header.Set("X-User", sess.Username)
					next(w, r)
					return
				}
			}
		}

		username, password, ok := r.BasicAuth()
		if ok && checkCredentials(username, password) {
			if volName != "" && volInfo != nil && !hasPermission(username, volInfo, 'r') {
				http.Error(w, fmt.Sprintf("User '%s' does not have permission for this volume.", username), http.StatusForbidden)
				return
			}
			if sessionTTL > 0 {
				createSession(w, username)
			}
			r.Header.Set("X-User", username)
			next(w, r)
			return
		}

		showLoginPage(w, r, "")
	}
}

type LoginData struct {
	Error string
	Next  string
}

func showLoginPage(w http.ResponseWriter, r *http.Request, errMsg string) {
	next := r.URL.Query().Get("next")
	if next == "" {
		next = r.URL.Path
	}
	data := LoginData{Error: errMsg, Next: next}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	loginTmpl.Execute(w, data)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		showLoginPage(w, r, "")
		return
	}

	r.ParseForm()
	username := strings.TrimSpace(r.FormValue("username"))
	password := r.FormValue("password")
	next := r.FormValue("next")

	if !checkCredentials(username, password) {
		showLoginPage(w, r, "Invalid username or password")
		return
	}

	createSession(w, username)
	logActivity("web", "LOGIN", username, r)

	if next == "" || next == "/login" {
		next = "/"
	}
	http.Redirect(w, r, next, http.StatusFound)
}

func checkCredentials(username, password string) bool {
	p, ok := users[username]
	if !ok {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(p), []byte(password)) == 1
}

// ==============================================================================
// ===== Helper Functions =======================================================
// ==============================================================================

func parseSize(sizeStr string) (int64, error) {
	if sizeStr == "" {
		return 1 << 30, nil
	}
	sizeStr = strings.ToUpper(strings.TrimSpace(sizeStr))
	re := regexp.MustCompile(`^(\d+(?:\.\d+)?)\s*([KMGT]?)B?$`)
	m := re.FindStringSubmatch(sizeStr)
	if m == nil {
		return 0, fmt.Errorf("invalid size format: %s", sizeStr)
	}
	num, _ := strconv.ParseFloat(m[1], 64)
	switch m[2] {
	case "K":
		return int64(num * 1024), nil
	case "M":
		return int64(num * 1024 * 1024), nil
	case "G":
		return int64(num * 1024 * 1024 * 1024), nil
	case "T":
		return int64(num * 1024 * 1024 * 1024 * 1024), nil
	default:
		return int64(num), nil
	}
}

func getVolAndPath(path string) (string, *Volume) {
	path = strings.Trim(path, "/")
	if path == "" {
		return "", nil
	}
	parts := strings.SplitN(path, "/", 2)
	volName := parts[0]
	vol, ok := volumes[volName]
	if !ok {
		return "", nil
	}
	return volName, vol
}

func getFullAbsPath(path string) string {
	path = strings.Trim(path, "/")
	if path == "" {
		return ""
	}
	parts := strings.SplitN(path, "/", 2)
	volName := parts[0]
	subPath := ""
	if len(parts) > 1 {
		subPath = parts[1]
	}
	vol, ok := volumes[volName]
	if !ok {
		return ""
	}
	absPath := filepath.Clean(filepath.Join(vol.Path, filepath.FromSlash(subPath)))
	if !pathInsideVolume(vol.Path, absPath) {
		return ""
	}
	return absPath
}

func hasPermission(user string, vol *Volume, perm rune) bool {
	if vol == nil {
		return false
	}
	for _, u := range vol.Users {
		if u == user {
			return strings.ContainsRune(vol.Perms, perm)
		}
	}
	return false
}

func logActivity(source, action, path string, r *http.Request) {
	ip := "unknown-ip"
	user := "anonymous"
	if r != nil {
		ip = r.RemoteAddr
		if idx := strings.LastIndex(ip, ":"); idx != -1 {
			ip = ip[:idx]
		}
		if u := r.Header.Get("X-User"); u != "" {
			user = u
		}
	}
	msg := fmt.Sprintf("(%s) User '%s' %s '%s'", ip, user, action, path)
	writeLog(source, msg)
	if source == "web" || source == "webdav" {
		if isAudioExt(path) || isVideoExt(path) || isImageExt(path) {
			writeLog("media", msg)
		}
	}
}

func getUser(r *http.Request) string {
	if u := r.Header.Get("X-User"); u != "" {
		return u
	}
	return "anonymous"
}

func formatFileSize(size int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)
	switch {
	case size >= GB:
		return fmt.Sprintf("%.1f GB", float64(size)/float64(GB))
	case size >= MB:
		return fmt.Sprintf("%.1f MB", float64(size)/float64(MB))
	case size >= KB:
		return fmt.Sprintf("%.1f KB", float64(size)/float64(KB))
	default:
		return fmt.Sprintf("%d B", size)
	}
}

// ==============================================================================
// ===== Recycle Bin ============================================================
// ==============================================================================

func performMoveToRecycle(paths []string, r *http.Request) []string {
	var success []string
	for _, p := range paths {
		_, info := getVolAndPath(p)
		if info == nil || !hasPermission(getUser(r), info, 'd') {
			continue
		}
		absP := getFullAbsPath(p)
		if absP == "" {
			continue
		}
		if _, err := os.Stat(absP); os.IsNotExist(err) {
			continue
		}

		targetPath := recyclePath(p)
		if targetPath == "" {
			continue
		}
		targetDir := filepath.Dir(targetPath)

		if err := os.MkdirAll(targetDir, 0755); err != nil {
			log.Printf("Failed to create recycle directory for %s: %v", p, err)
			continue
		}

		if _, err := os.Lstat(targetPath); err == nil {
			os.RemoveAll(targetPath)
		}

		if err := movePath(absP, targetPath); err != nil {
			log.Printf("Failed to move %s to recycle: %v", p, err)
			continue
		}
		logActivity("web", "DELETED_TO_RECYCLE", p, r)
		success = append(success, p)
	}
	return success
}

// ==============================================================================
// ===== Route Handlers =========================================================
// ==============================================================================

func indexHandler(w http.ResponseWriter, r *http.Request) {
	user := getUser(r)
	var accessible []FileItem
	for name, info := range volumes {
		if hasPermission(user, info, 'r') {
			accessible = append(accessible, FileItem{
				Name:  name,
				IsDir: true,
				Path:  name,
			})
		}
	}
	sort.Slice(accessible, func(i, j int) bool {
		return accessible[i].Name < accessible[j].Name
	})

	data := IndexData{
		Items:       accessible,
		Breadcrumbs: []Breadcrumb{},
		CurrentPath: "",
		Playlist:    []PlaylistItem{},
		CurrentUser: user,
		SortBy:      "name",
		Order:       "asc",
		NameOrder:   "asc",
		SizeOrder:   "asc",
		MtimeOrder:  "asc",
		Version:     Version,
	}
	executeTemplate(w, indexTmpl, data)
}

func handlePathHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(r.URL.Path, "/")

	volName, volInfo := getVolAndPath(path)
	absPath := getFullAbsPath(path)
	if volInfo == nil || absPath == "" {
		http.NotFound(w, r)
		return
	}

	info, err := os.Stat(absPath)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if info.IsDir() {
		handleDir(w, r, path, volName, volInfo, absPath)
	} else {
		logActivity("web", "DOWNLOADED", path, r)
		http.ServeFile(w, r, absPath)
	}
}

func handleDir(w http.ResponseWriter, r *http.Request, path, volName string, volInfo *Volume, absPath string) {
	user := getUser(r)
	isPartial := r.URL.Query().Get("_partial") == "1"
	showHidden := r.URL.Query().Get("show_hidden") == "1"
	sortBy := r.URL.Query().Get("sort")
	order := r.URL.Query().Get("order")
	if sortBy == "" {
		sortBy = "name"
	}
	if order == "" {
		order = "asc"
	}
	isReverse := order == "desc"

	if !isPartial {
		logActivity("web", "LISTED", path, r)
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		http.Error(w, "Cannot read directory", http.StatusInternalServerError)
		return
	}

	var dirsList, filesList []FileItem
	var totalSize int64
	var fileCount int

	for _, entry := range entries {
		fi, err := entry.Info()
		if err != nil {
			continue
		}
		if !showHidden && isHiddenFile(entry.Name(), fi) {
			continue
		}
		itemPath := path + "/" + entry.Name()
		if path == "" {
			itemPath = volName + "/" + entry.Name()
		}
		item := FileItem{
			Name:           entry.Name(),
			Path:           itemPath,
			IsDir:          entry.IsDir(),
			Size:           fi.Size(),
			Mtime:          fi.ModTime(),
			FormattedMtime: fi.ModTime().Format("06-01-02"),
		}

		if entry.IsDir() {
			dirsList = append(dirsList, item)
		} else {
			filesList = append(filesList, item)
			if !isPartial {
				totalSize += fi.Size()
				fileCount++
			}
		}
	}

	canWrite := hasPermission(user, volInfo, 'w')
	sortFiles(dirsList, sortBy, isReverse)
	sortFiles(filesList, sortBy, isReverse)

	allItems := append(dirsList, filesList...)
	for i := range allItems {
		allItems[i].CanDelete = hasPermission(user, volInfo, 'd')
		allItems[i].CanEdit = !allItems[i].IsDir && canWrite && isEditableExt(allItems[i].Name)
		allItems[i].CanRename = canWrite
	}

	if isPartial {
		nameOrder := "asc"
		sizeOrder := "asc"
		mtimeOrder := "asc"
		if sortBy == "name" && order == "asc" {
			nameOrder = "desc"
		} else if sortBy == "size" && order == "asc" {
			sizeOrder = "desc"
		} else if sortBy == "mtime" && order == "asc" {
			mtimeOrder = "desc"
		}
		partialData := IndexData{
			Items:      allItems,
			SortBy:     sortBy,
			Order:      order,
			NameOrder:  nameOrder,
			SizeOrder:  sizeOrder,
			MtimeOrder: mtimeOrder,
		}
		executeTemplate(w, partialTmpl, partialData)
		return
	}

	playlist := []PlaylistItem{}
	if !isPartial {
		for _, item := range allItems {
			if !item.IsDir && (isAudioExt(item.Name) || isVideoExt(item.Name)) {
				pi := PlaylistItem{Name: item.Name, Path: item.Path, IsVideo: isVideoExt(item.Name)}
				if pi.IsVideo {
					baseName := strings.TrimSuffix(item.Name, filepath.Ext(item.Name))
					webPrefix := path + "/"
					if path == "" {
						webPrefix = volName + "/"
					}
					pi.Subtitles = findSubtitles(absPath, baseName, webPrefix)
				}
				playlist = append(playlist, pi)
			}
		}
	}

	subPath := ""
	if idx := strings.Index(path, "/"); idx != -1 {
		subPath = path[idx+1:]
	}

	breadcrumbs := []Breadcrumb{{Name: volName, Path: volName}}
	if subPath != "" {
		parts := strings.Split(subPath, "/")
		for i, p := range parts {
			bcPath := volName + "/" + strings.Join(parts[:i+1], "/")
			breadcrumbs = append(breadcrumbs, Breadcrumb{Name: p, Path: bcPath})
		}
	}

	_, sess := getUserAndSession(r)
	canPaste := sess != nil && len(sess.ClipboardSources) > 0 && canWrite

	nameOrder := "asc"
	sizeOrder := "asc"
	mtimeOrder := "asc"
	if sortBy == "name" && order == "asc" {
		nameOrder = "desc"
	} else if sortBy == "size" && order == "asc" {
		sizeOrder = "desc"
	} else if sortBy == "mtime" && order == "asc" {
		mtimeOrder = "desc"
	}

	data := IndexData{
		Items:       allItems,
		Breadcrumbs: breadcrumbs,
		CurrentPath: path,
		Playlist:    playlist,
		AllVolumes:  getVolumeNames(),
		CurrentUser: user,
		TotalSize:   totalSize,
		FileCount:   fileCount,
		UserPerms:   volInfo.Perms,
		SortBy:      sortBy,
		Order:       order,
		NameOrder:   nameOrder,
		SizeOrder:   sizeOrder,
		MtimeOrder:  mtimeOrder,
		CanPaste:    canPaste,
		ShowHidden:  showHidden,
		Version:     Version,
	}
	executeTemplate(w, indexTmpl, data)
}

func sortFiles(items []FileItem, sortBy string, reverse bool) {
	sort.Slice(items, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "size":
			less = items[i].Size < items[j].Size
		case "mtime":
			less = items[i].Mtime.Before(items[j].Mtime)
		default:
			less = strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
		}
		if reverse {
			return !less
		}
		return less
	})
}

func isAudioExt(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".mp3", ".m4a", ".ogg", ".flac", ".wav", ".opus":
		return true
	}
	return false
}

func isEditableExt(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".txt", ".md", ".json", ".xml", ".html", ".css", ".js", ".py", ".sh", ".conf", ".ini", ".log", ".cfg":
		return true
	}
	return false
}

func isVideoExt(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".mp4", ".mkv", ".webm", ".avi", ".mov", ".flv", ".wmv", ".m4v":
		return true
	}
	return false
}

func isImageExt(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg", ".bmp", ".ico":
		return true
	}
	return false
}

func findSubtitles(dirPath, baseName, webPrefix string) []SubtitleTrack {
	var subs []SubtitleTrack
	exts := []string{".srt", ".vtt", ".ass", ".ssa", ".sub"}
	for _, ext := range exts {
		subFile := filepath.Join(dirPath, baseName+ext)
		if _, err := os.Stat(subFile); err == nil {
			webPath := webPrefix + baseName + ext
			subs = append(subs, SubtitleTrack{Lang: ext[1:], Path: webPath})
		}
	}
	return subs
}

func getUserAndSession(r *http.Request) (string, *Session) {
	user := getUser(r)
	return user, getSession(r)
}

func getVolumeNames() []string {
	var names []string
	for n := range volumes {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// --- Edit ---

func editHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/edit/")
	path = strings.TrimPrefix(path, "edit/")

	_, volInfo := getVolAndPath(path)
	absPath := getFullAbsPath(path)
	if absPath == "" || volInfo == nil {
		http.NotFound(w, r)
		return
	}

	info, err := os.Stat(absPath)
	if err != nil || info.IsDir() || !isEditableExt(path) {
		http.NotFound(w, r)
		return
	}

	if r.Method == http.MethodPost {
		if !hasPermission(getUser(r), volInfo, 'w') {
			http.Error(w, "Write denied", http.StatusForbidden)
			return
		}
		r.ParseForm()
		content := r.FormValue("content")
		if err := os.WriteFile(absPath, []byte(content), 0644); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		logActivity("web", "EDITED", path, r)
		pdir := ""
		if idx := strings.LastIndex(path, "/"); idx != -1 {
			pdir = path[:idx]
		}
		if pdir != "" {
			http.Redirect(w, r, "/"+urlPathEscape(pdir), http.StatusFound)
		} else {
			http.Redirect(w, r, "/", http.StatusFound)
		}
		return
	}

	if !hasPermission(getUser(r), volInfo, 'r') {
		http.Error(w, "Read denied", http.StatusForbidden)
		return
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		content = []byte(fmt.Sprintf("Error: %v", err))
	}

	data := EditData{
		Path:     path,
		Content:  string(content),
		Referrer: r.Header.Get("Referer"),
	}
	executeTemplate(w, editTmpl, data)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	clearSession(w, r)
	http.Redirect(w, r, "/login", http.StatusFound)
}

// --- Actions ---

func renameHandler(w http.ResponseWriter, r *http.Request) {
	user := getUser(r)
	var d struct {
		Path    string `json:"path"`
		OldName string `json:"old_name"`
		NewName string `json:"new_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if !validNameComponent(d.OldName) || strings.TrimSpace(d.NewName) == "" {
		jsonError(w, "Invalid name", http.StatusBadRequest)
		return
	}

	newName := getSafeName(d.NewName)
	if !validNameComponent(newName) {
		jsonError(w, "Invalid name", http.StatusBadRequest)
		return
	}

	rel := d.Path + "/" + d.OldName
	_, info := getVolAndPath(rel)
	oldAbs := getFullAbsPath(rel)
	if info == nil || oldAbs == "" || !hasPermission(user, info, 'w') {
		jsonError(w, "Denied", http.StatusForbidden)
		return
	}

	newAbs := filepath.Join(filepath.Dir(oldAbs), newName)
	if !pathInsideVolume(info.Path, newAbs) {
		jsonError(w, "Invalid name", http.StatusBadRequest)
		return
	}

	// Do not silently overwrite an existing entry. A case-only rename on a
	// case-insensitive filesystem still refers to the same file and is allowed.
	if oldInfo, err := os.Lstat(oldAbs); err == nil {
		if newInfo, err := os.Lstat(newAbs); err == nil && !os.SameFile(oldInfo, newInfo) {
			jsonError(w, "An item with that name already exists", http.StatusConflict)
			return
		}
	}

	if err := os.Rename(oldAbs, newAbs); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	logActivity("web", "RENAMED", d.OldName+" -> "+newName, r)
	jsonOK(w)
}

func setClipboardHandler(w http.ResponseWriter, r *http.Request) {
	var d struct {
		Paths  []string `json:"paths"`
		Action string   `json:"action"`
	}
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	sess := getSession(r)
	if sess == nil {
		sess = createSession(w, getUser(r))
	}
	sess.ClipboardSources = d.Paths
	sess.ClipboardAction = d.Action
	jsonOK(w)
}

func pasteHandler(w http.ResponseWriter, r *http.Request) {
	user := getUser(r)
	var d struct {
		Destination string `json:"destination"`
	}
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	_, dInfo := getVolAndPath(d.Destination)
	dAbs := getFullAbsPath(d.Destination)
	if dInfo == nil || dAbs == "" || !hasPermission(user, dInfo, 'w') {
		jsonError(w, "Denied", http.StatusForbidden)
		return
	}

	sess := getSession(r)
	if sess == nil {
		jsonOK(w)
		return
	}

	for _, sRel := range sess.ClipboardSources {
		sAbs := getFullAbsPath(sRel)
		if sAbs == "" {
			continue
		}
		baseName := filepath.Base(sAbs)
		final := filepath.Join(dAbs, baseName)

		if sess.ClipboardAction == "copy" {
			if _, err := os.Lstat(final); err == nil {
				ext := filepath.Ext(baseName)
				nameNoExt := strings.TrimSuffix(baseName, ext)
				for i := 1; ; i++ {
					final = filepath.Join(dAbs, fmt.Sprintf("%s-copy%d%s", nameNoExt, i, ext))
					if _, err := os.Lstat(final); err != nil {
						break
					}
				}
			}
			if !pathInsideVolume(dAbs, final) {
				writeLog("paste", fmt.Sprintf("(%s) SKIPPED-ESCAPE %s -> %s", clientIP(r), sRel, final))
				continue
			}
			if err := copyPath(sAbs, final); err != nil {
				writeLog("paste", fmt.Sprintf("(%s) COPY-FAILED %s -> %s: %v", clientIP(r), sRel, final, err))
			}
		} else if sess.ClipboardAction == "move" {
			if _, err := os.Lstat(final); err == nil {
				ext := filepath.Ext(baseName)
				nameNoExt := strings.TrimSuffix(baseName, ext)
				for i := 1; ; i++ {
					final = filepath.Join(dAbs, fmt.Sprintf("%s-copy%d%s", nameNoExt, i, ext))
					if _, err := os.Lstat(final); err != nil {
						break
					}
				}
			}
			if !pathInsideVolume(dAbs, final) {
				writeLog("paste", fmt.Sprintf("(%s) SKIPPED-ESCAPE %s -> %s", clientIP(r), sRel, final))
				continue
			}
			if err := movePath(sAbs, final); err != nil {
				writeLog("paste", fmt.Sprintf("(%s) MOVE-FAILED %s -> %s: %v", clientIP(r), sRel, final, err))
			}
		}
	}

	sessionMu.Lock()
	if sess != nil {
		sess.ClipboardSources = nil
	}
	sessionMu.Unlock()

	jsonOK(w)
}

func copyPath(src, dst string) error {
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(src)
		if err != nil {
			return err
		}
		return os.Symlink(target, dst)
	}
	if info.IsDir() {
		return copyDir(src, dst)
	}
	return copyFile(src, dst)
}

func copyFile(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}

	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}
	// Best effort: preserve the modification time on the copy.
	_ = os.Chtimes(dst, info.ModTime(), info.ModTime())
	return nil
}

func copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dst, srcInfo.Mode().Perm()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if err := copyPath(srcPath, dstPath); err != nil {
			return err
		}
	}
	_ = os.Chtimes(dst, srcInfo.ModTime(), srcInfo.ModTime())
	return nil
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/_actions/upload/")
	path = strings.Trim(path, "/")
	user := getUser(r)

	_, info := getVolAndPath(path)
	absPath := getFullAbsPath(path)
	if info == nil || absPath == "" || !hasPermission(user, info, 'w') {
		http.Error(w, "Write denied", http.StatusForbidden)
		return
	}

	// Streaming read: write each part straight to the target file, never through /tmp,
	// so ParseMultipartForm never buffers the whole request body in the temp dir
	// (which is limited by the /tmp quota).
	mr, err := r.MultipartReader()
	if err != nil {
		writeLog("upload", fmt.Sprintf(
			"(%s) NO-MULTIPART user=%s path=%s err=%v", clientIP(r), user, path, err))
		http.Error(w, "Upload failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	writeLog("upload", fmt.Sprintf(
		"(%s) BEGIN user=%s path=%s contentLength=%d", clientIP(r), user, path, r.ContentLength))

	// The frontend sends the "paths" field first (a JSON array), then each "file" part
	// in order, so relPaths[i] corresponds to the i-th file.
	var relPaths []string
	// The conflict policy comes from the URL query (always available before the
	// body is read); the multipart field is a fallback and cannot override it.
	queryConflict := strings.TrimSpace(r.URL.Query().Get("conflict"))
	conflict := queryConflict
	if conflict == "" {
		conflict = "overwrite"
	}
	fileIndex := 0
	skipped := 0
	failMsg := ""
	partial := "" // path of the file currently being written (may be incomplete)

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			writeLog("upload", fmt.Sprintf(
				"(%s) STREAM-ERROR user=%s path=%s ctxErr=%v err=%v",
				clientIP(r), user, path, r.Context().Err(), err))
			if partial != "" {
				os.Remove(partial)
				writeLog("upload", fmt.Sprintf("(%s) CLEANUP removed partial file %s", clientIP(r), partial))
			}
			http.Error(w, "Upload interrupted: "+err.Error(), http.StatusBadRequest)
			return
		}

		if part.FormName() != "file" {
			switch part.FormName() {
			case "paths":
				b, _ := io.ReadAll(io.LimitReader(part, 1<<20))
				json.Unmarshal(b, &relPaths)
			case "conflict":
				b, _ := io.ReadAll(io.LimitReader(part, 1<<10))
				if v := strings.TrimSpace(string(b)); v != "" && queryConflict == "" {
					conflict = v
				}
			}
			part.Close()
			continue
		}

		filename := part.FileName()
		if filename == "" {
			part.Close()
			fileIndex++
			continue
		}

		targetDir := absPath
		if fileIndex < len(relPaths) && relPaths[fileIndex] != "" {
			dir, ok := sanitizeRelDir(relPaths[fileIndex])
			if !ok {
				writeLog("upload", fmt.Sprintf("(%s) SKIPPED-INVALID-PATH user=%s file=%q", clientIP(r), user, relPaths[fileIndex]))
				part.Close()
				fileIndex++
				skipped++
				continue
			}
			if dir != "" {
				targetDir = filepath.Join(absPath, dir)
				if !pathInsideVolume(absPath, targetDir) {
					writeLog("upload", fmt.Sprintf("(%s) SKIPPED-ESCAPE user=%s file=%q", clientIP(r), user, relPaths[fileIndex]))
					part.Close()
					fileIndex++
					skipped++
					continue
				}
				if err := os.MkdirAll(targetDir, 0755); err != nil {
					writeLog("upload", fmt.Sprintf("(%s) MKDIR-FAILED user=%s dir=%s: %v", clientIP(r), user, targetDir, err))
					part.Close()
					fileIndex++
					skipped++
					continue
				}
			}
		}

		safeName := getSafeName(filename)
		saveTo := filepath.Join(targetDir, safeName)
		if !pathInsideVolume(absPath, saveTo) {
			writeLog("upload", fmt.Sprintf("(%s) SKIPPED-ESCAPE user=%s file=%q", clientIP(r), user, filename))
			part.Close()
			fileIndex++
			skipped++
			continue
		}
		rel := filepath.Join(path, safeName)

		// Conflict policy when a file with the same name already exists.
		if _, err := os.Stat(saveTo); err == nil {
			switch conflict {
			case "skip":
				writeLog("upload", fmt.Sprintf("(%s) SKIPPED-EXISTS user=%s file=%s", clientIP(r), user, rel))
				part.Close()
				fileIndex++
				skipped++
				continue
			case "rename":
				saveTo = uniquePath(saveTo, "")
			default: // "overwrite"
				os.RemoveAll(saveTo)
			}
		}

		dst, err := os.Create(saveTo)
		if err != nil {
			writeLog("upload", fmt.Sprintf("(%s) CREATE-FAILED user=%s %s: %v", clientIP(r), user, saveTo, err))
			if failMsg == "" {
				failMsg = err.Error()
			}
			part.Close()
			fileIndex++
			continue
		}

		partial = saveTo
		n, copyErr := io.Copy(dst, part)
		closeErr := dst.Close()
		part.Close()
		if copyErr != nil || closeErr != nil {
			writeLog("upload", fmt.Sprintf(
				"(%s) COPY-ERROR user=%s file=%s copied=%d copyErr=%v closeErr=%v ctxErr=%v",
				clientIP(r), user, rel, n, copyErr, closeErr, r.Context().Err()))
			// Remove the half-written file to avoid leaving a corrupted file behind
			os.Remove(saveTo)
			writeLog("upload", fmt.Sprintf("(%s) CLEANUP removed partial file %s", clientIP(r), saveTo))
			if failMsg == "" {
				if copyErr != nil {
					failMsg = copyErr.Error()
				} else {
					failMsg = closeErr.Error()
				}
			}
		} else {
			writeLog("upload", fmt.Sprintf("(%s) DONE user=%s file=%s size=%d", clientIP(r), user, rel, n))
			logActivity("web", "UPLOAD", rel, r)
		}
		partial = ""
		fileIndex++
	}

	if failMsg != "" {
		http.Error(w, "Upload failed: "+failMsg, http.StatusInternalServerError)
		return
	}

	// The web UI uploads via XHR and expects JSON so it can report skipped files.
	if r.Header.Get("X-Requested-With") == "XMLHttpRequest" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "skipped": skipped})
		return
	}

	referrer := r.Header.Get("Referer")
	if referrer == "" {
		referrer = "/" + urlPathEscape(path)
	}
	http.Redirect(w, r, referrer, http.StatusFound)
}

// uploadCheckHandler reports which of the given web paths already exist, so the
// upload UI can ask the user how to resolve name conflicts before uploading.
// "paths" are target files; "dirs" are the top-level folders being uploaded;
// "base" is the current directory, which is not filename-sanitized.
func uploadCheckHandler(w http.ResponseWriter, r *http.Request) {
	var d struct {
		Paths []string `json:"paths"`
		Dirs  []string `json:"dirs"`
		Base  string   `json:"base"`
	}
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}
	user := getUser(r)
	baseAbs := getFullAbsPath(d.Base)
	conflicts := []string{}
	if baseAbs == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"conflicts":     conflicts,
			"dir_conflicts": []string{},
		})
		return
	}
	for _, p := range d.Paths {
		_, info := getVolAndPath(p)
		if !hasPermission(user, info, 'w') {
			continue
		}
		rel, ok := relToUploadBase(d.Base, p)
		if !ok {
			continue
		}
		abs := sanitizedUploadFile(baseAbs, rel)
		if abs == "" {
			continue
		}
		if _, err := os.Lstat(abs); err == nil {
			conflicts = append(conflicts, p)
		}
	}

	type dirConflict struct {
		Path string `json:"path"`
		Name string `json:"name"` // suggested non-existing folder name
	}
	dirConflicts := []dirConflict{}
	for _, p := range d.Dirs {
		_, info := getVolAndPath(p)
		if !hasPermission(user, info, 'w') {
			continue
		}
		rel, ok := relToUploadBase(d.Base, p)
		if !ok {
			continue
		}
		abs := sanitizedUploadDir(baseAbs, rel)
		if abs == "" {
			continue
		}
		if fi, err := os.Stat(abs); err == nil && fi.IsDir() {
			dirConflicts = append(dirConflicts, dirConflict{
				Path: p,
				Name: filepath.Base(uniquePath(abs, "")),
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"conflicts":     conflicts,
		"dir_conflicts": dirConflicts,
	})
}

func newFileHandler(w http.ResponseWriter, r *http.Request) {
	user := getUser(r)
	var d struct {
		Path     string `json:"path"`
		Filename string `json:"filename"`
	}
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}
	_, info := getVolAndPath(d.Path)
	absPath := getFullAbsPath(d.Path)
	if info == nil || absPath == "" || !hasPermission(user, info, 'w') {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	name := getSafeName(d.Filename)
	saveTo := filepath.Join(absPath, name)
	if !pathInsideVolume(info.Path, saveTo) {
		jsonError(w, "Invalid filename", http.StatusBadRequest)
		return
	}
	f, err := os.OpenFile(saveTo, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	f.Close()
	jsonOK(w)
}

func newDirectoryHandler(w http.ResponseWriter, r *http.Request) {
	user := getUser(r)
	var d struct {
		Path string `json:"path"`
		Name string `json:"dirname"`
	}
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}
	_, info := getVolAndPath(d.Path)
	absPath := getFullAbsPath(d.Path)
	if info == nil || absPath == "" || !hasPermission(user, info, 'w') {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	name := getSafeName(d.Name)
	dirPath := filepath.Join(absPath, name)
	if !pathInsideVolume(info.Path, dirPath) {
		jsonError(w, "Invalid directory name", http.StatusBadRequest)
		return
	}
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	var d struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil || d.Path == "" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	success := performMoveToRecycle([]string{d.Path}, r)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      len(success) > 0,
		"deleted_path": d.Path,
	})
}

func batchDeleteHandler(w http.ResponseWriter, r *http.Request) {
	var d struct {
		Paths []string `json:"paths"`
	}
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}
	success := performMoveToRecycle(d.Paths, r)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"count":   len(success),
	})
}

func zipDownloadHandler(w http.ResponseWriter, r *http.Request) {
	var d struct {
		Paths []string `json:"paths"`
	}
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", attachmentDisposition("download.zip"))

	zipWriter := NewZipStreamingWriter(w)
	for _, p := range d.Paths {
		absPath := getFullAbsPath(p)
		if absPath == "" {
			continue
		}
		if err := zipWriter.AddFile(absPath, filepath.Base(absPath)); err != nil {
			writeLog("zip", fmt.Sprintf("(%s) ADD-FAILED %s: %v", clientIP(r), p, err))
		}
	}
	if err := zipWriter.Close(); err != nil {
		writeLog("zip", fmt.Sprintf("(%s) CLOSE-FAILED: %v", clientIP(r), err))
	}
}

// --- JSON helpers ---

func propertiesHandler(w http.ResponseWriter, r *http.Request) {
	var d struct {
		Paths []string `json:"paths"`
	}
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	var files, dirs int
	var totalSize int64

	var walkDir func(absPath string)
	walkDir = func(absPath string) {
		entries, err := os.ReadDir(absPath)
		if err != nil {
			return
		}
		for _, entry := range entries {
			fi, err := entry.Info()
			if err != nil {
				continue
			}
			if isHiddenFile(entry.Name(), fi) {
				continue
			}
			fullPath := filepath.Join(absPath, entry.Name())
			if entry.IsDir() {
				dirs++
				walkDir(fullPath)
			} else {
				files++
				totalSize += fi.Size()
			}
		}
	}

	for _, p := range d.Paths {
		absPath := getFullAbsPath(p)
		if absPath == "" {
			continue
		}
		info, err := os.Stat(absPath)
		if err != nil {
			continue
		}
		if info.IsDir() {
			dirs++
			walkDir(absPath)
		} else {
			files++
			totalSize += info.Size()
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"files":     files,
		"dirs":      dirs,
		"totalSize": totalSize,
	})
}

func jsonOK(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": msg})
}

func logViewHandler(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	logPath := getLogFile()

	data, err := os.ReadFile(logPath)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]string{})
		return
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	var filtered []string = []string{}
	catLower := strings.ToLower(category)
	search := "[" + catLower + "]"
	for _, line := range lines {
		if line == "" {
			continue
		}
		if catLower != "" && catLower != "all" {
			lower := strings.ToLower(line)
			idx := strings.Index(lower, search)
			if idx == -1 {
				continue
			}
			after := idx + len(search)
			if after < len(lower) && lower[after] != ' ' {
				continue
			}
		}
		filtered = append(filtered, line)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	json.NewEncoder(w).Encode(filtered)
}

func logClearHandler(w http.ResponseWriter, r *http.Request) {
	logPath := getLogFile()
	os.Truncate(logPath, 0)
	jsonOK(w)
}

func logDeleteHandler(w http.ResponseWriter, r *http.Request) {
	logPath := getLogFile()
	os.Remove(logPath)
	logMu.Lock()
	if logFile != nil {
		logFile.Close()
		logFile = nil
	}
	logMu.Unlock()
	initLogger()
	jsonOK(w)
}

func executeTemplate(w http.ResponseWriter, tmpl *template.Template, data interface{}) {
	// Never let the browser cache HTML: the inline JS/CSS is versioned with the
	// binary, and a stale page would keep running old client logic.
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	tmpl.Execute(w, data)
}
