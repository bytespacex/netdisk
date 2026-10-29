package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/net/webdav"
)

// Version is set at build time via -ldflags "-X main.Version=...".
// It must remain a var (not a const) for the linker -X flag to work.
var Version = "26.07.11"

type stringSlice []string

func (s *stringSlice) String() string { return strings.Join(*s, ", ") }
func (s *stringSlice) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func main() {
	var (
		port       int
		auths      stringSlice
		vols       stringSlice
		timeout    int
		maxUpload  string
		configPath string
	)

	flag.IntVar(&port, "p", 8880, "port")
	flag.Var(&auths, "a", "auth user:pass (repeatable)")
	flag.Var(&vols, "v", "vol path:name:perms,users (repeatable)")
	flag.IntVar(&timeout, "t", 168, "session timeout in hours, 0=no expiry")
	flag.StringVar(&maxUpload, "m", "10G", "max upload size")
	flag.StringVar(&configPath, "c", "", "config directory (default: ~/.config/netdisk)")
	flag.Usage = func() { fmt.Print(SHOWHELP) }
	flag.Parse()

	// Show help if -h or --help
	if len(os.Args) > 1 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
		fmt.Print(SHOWHELP)
		return
	}

	maxBody, err := parseSize(maxUpload)
	if err != nil {
		maxBody = 1 << 30
	}

	sessionTTL = time.Duration(timeout) * time.Hour

	// Parse auth
	for _, a := range auths {
		parts := strings.SplitN(a, ":", 2)
		if len(parts) == 2 {
			users[parts[0]] = parts[1]
		}
	}

	// Parse volumes. Split from the right so Windows drive letters
	// (C:\data:docs:rwdm,admin) work as well as Unix paths containing colons.
	for _, v := range vols {
		volPath, volName, permsUsers, ok := parseVolumeSpec(v)
		if !ok {
			fmt.Fprintf(os.Stderr, "ignoring volume %q: expected <path>:<name>:<perms>,<users>\n", v)
			continue
		}
		if !validVolumeName(volName) {
			fmt.Fprintf(os.Stderr, "ignoring volume %q: invalid volume name %q\n", v, volName)
			continue
		}
		if _, exists := volumes[volName]; exists {
			fmt.Fprintf(os.Stderr, "ignoring volume %q: duplicate volume name %q\n", v, volName)
			continue
		}

		var perms, usersAllowed string
		if idx := strings.Index(permsUsers, ","); idx != -1 {
			perms = permsUsers[:idx]
			usersAllowed = permsUsers[idx+1:]
		} else {
			perms = permsUsers
		}
		volPath = expandUser(volPath)
		if abs, err := filepath.Abs(volPath); err == nil {
			volPath = abs
		}

		var allowedUsers []string
		if usersAllowed != "" {
			allowedUsers = strings.Split(usersAllowed, ",")
		} else {
			for u := range users {
				allowedUsers = append(allowedUsers, u)
			}
		}

		info, err := os.Stat(volPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ignoring volume %q: %v\n", v, err)
			continue
		}
		if !info.IsDir() {
			fmt.Fprintf(os.Stderr, "ignoring volume %q: %s is not a directory\n", v, volPath)
			continue
		}
		volumes[volName] = &Volume{
			Path:  volPath,
			Perms: perms,
			Users: allowedUsers,
		}
	}

	// Setup config and data dirs
	if configPath == "" {
		homeDir, _ := os.UserHomeDir()
		configPath = filepath.Join(homeDir, ".config", "netdisk")
	}
	configPath = expandUser(configPath)
	if abs, err := filepath.Abs(configPath); err == nil {
		configPath = abs
	}
	configDir = configPath
	recycleBin = filepath.Join(configDir, "recycle")
	webclipDir = filepath.Join(configDir, "webclip")
	webclipFiles = filepath.Join(webclipDir, "files")

	for _, d := range []string{configDir, recycleBin, webclipDir, webclipFiles} {
		os.MkdirAll(d, 0755)
	}

	initLogger()
	loadShareLinks()

	// Start session cleaner
	go cleanExpiredSessions()

	// Setup routes
	mux := http.NewServeMux()

	// Static files (embedded JS/CSS)
	mux.HandleFunc("GET /static/{file}", staticHandler)

	// Login
	mux.HandleFunc("GET /login", loginHandler)
	mux.HandleFunc("POST /login", loginHandler)

	// Edit
	mux.HandleFunc("GET /edit/{path...}", authRequired(editHandler))
	mux.HandleFunc("POST /edit/{path...}", authRequired(editHandler))

	// Logout
	mux.HandleFunc("GET /logout", logoutHandler)

	// Share links - handled in catchAllHandler to avoid route conflict

	// API: search
	mux.HandleFunc("GET /api/search", authRequired(searchHandler))

	// API: share create/delete
	mux.HandleFunc("POST /api/share", authRequired(shareCreateHandler))
	mux.HandleFunc("GET /api/share", authRequired(shareGetHandler))
	mux.HandleFunc("DELETE /api/share/{token}", authRequired(shareDeleteHandler))
	mux.HandleFunc("GET /api/shares", authRequired(shareListHandler))

	// Preview
	mux.HandleFunc("GET /preview/{path...}", authRequired(previewHandler))

	// Subtitles (sidecar + embedded via ffmpeg)
	mux.HandleFunc("GET /_actions/subtitles/{path...}", authRequired(subtitlesListHandler))
	mux.HandleFunc("GET /_actions/subtitle/{path...}", authRequired(subtitleExtractHandler))

	// Actions
	mux.HandleFunc("POST /_actions/rename", authRequired(renameHandler))
	mux.HandleFunc("POST /_actions/set_clipboard", authRequired(setClipboardHandler))
	mux.HandleFunc("POST /_actions/paste", authRequired(pasteHandler))
	mux.HandleFunc("POST /_actions/upload/{path...}", authRequired(uploadHandler))
	mux.HandleFunc("POST /_actions/upload_check", authRequired(uploadCheckHandler))
	mux.HandleFunc("POST /_actions/new_file", authRequired(newFileHandler))
	mux.HandleFunc("POST /_actions/new_directory", authRequired(newDirectoryHandler))
	mux.HandleFunc("POST /_actions/zip_download", authRequired(zipDownloadHandler))
	mux.HandleFunc("POST /_actions/delete", authRequired(deleteHandler))
	mux.HandleFunc("POST /_actions/batch_delete", authRequired(batchDeleteHandler))
	mux.HandleFunc("POST /_actions/properties", authRequired(propertiesHandler))
	mux.HandleFunc("GET /api/logs", authRequired(logViewHandler))
	mux.HandleFunc("DELETE /api/logs", authRequired(logClearHandler))
	mux.HandleFunc("POST /api/logs/delete", authRequired(logDeleteHandler))

	// Recycle bin
	mux.HandleFunc("GET /api/recycle", authRequired(recycleListHandler))
	mux.HandleFunc("POST /api/recycle/restore", authRequired(recycleRestoreHandler))
	mux.HandleFunc("POST /api/recycle/delete", authRequired(recycleDeleteHandler))
	mux.HandleFunc("POST /api/recycle/clear", authRequired(recycleClearHandler))

	// WebClip
	mux.HandleFunc("GET /wc/", authRequired(webclipHomeHandler))
	mux.HandleFunc("GET /wc/file/{filename}", authRequired(webclipGetFileHandler))
	mux.HandleFunc("GET /wc/api/cards", authRequired(webclipCardsAPIHandler))
	mux.HandleFunc("POST /wc/api/cards", authRequired(webclipCardsAPIHandler))
	mux.HandleFunc("DELETE /wc/api/cards/{card_id}", authRequired(webclipDeleteCardHandler))
	mux.HandleFunc("POST /wc/api/cards/clear", authRequired(webclipClearAllHandler))

	// WebDAV
	setupWebDAV(mux)

	// Catch-all: handle auth internally to allow unauthenticated /share/ routes
	mux.HandleFunc("/{path...}", catchAllHandler)

	fmt.Printf("netdisk %s running on http://0.0.0.0:%d\n", Version, port)
	fmt.Println("WebDAV endpoint: http://0.0.0.0:" + fmt.Sprintf("%d", port) + "/wd/")

	// Apply max request body size
	handler := maxBodyMiddleware(mux, maxBody)
	// Log connection-level diagnostics (disconnects, panics) for uploads/downloads
	handler = transferLogMiddleware(handler)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), handler))
}

func setupWebDAV(mux *http.ServeMux) {
	for vname, info := range volumes {
		handler := &webdav.Handler{
			Prefix:     "/wd/" + vname,
			FileSystem: webdav.Dir(info.Path),
			LockSystem: webdav.NewMemLS(),
		}
		authed := webdavAuth(vname, handler)
		// Register all WebDAV methods explicitly to avoid pattern conflicts
		for _, method := range []string{"GET", "HEAD", "OPTIONS", "PROPFIND", "PROPPATCH", "MKCOL", "PUT", "DELETE", "COPY", "MOVE", "LOCK", "UNLOCK"} {
			pattern := method + " /wd/" + vname + "/{path...}"
			mux.Handle(pattern, authed)
		}
		pattern := "GET /wd/" + vname
		mux.Handle(pattern, authed)
	}
}

func webdavAuth(vname string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || !checkCredentials(username, password) {
			w.Header().Set("WWW-Authenticate", `Basic realm="netdisk WebDAV"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		info := volumes[vname]
		if info == nil {
			http.Error(w, "Volume not found", http.StatusNotFound)
			return
		}

		// Check permissions based on method
		perm := 'r'
		switch r.Method {
		case "PUT", "DELETE", "MKCOL", "PROPPATCH", "COPY", "MOVE", "LOCK", "UNLOCK":
			perm = 'w'
		}

		if !hasPermission(username, info, perm) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		ip := r.RemoteAddr
		if idx := strings.LastIndex(ip, ":"); idx != -1 {
			ip = ip[:idx]
		}
		writeLog("webdav", fmt.Sprintf("(%s) User '%s' %s /wd/%s%s", ip, username, r.Method, vname, r.URL.Path))

		if r.Method == "GET" {
			filePath := vname + r.URL.Path
			if isAudioExt(filePath) || isVideoExt(filePath) || isImageExt(filePath) {
				writeLog("media", fmt.Sprintf("(%s) User '%s' accessed '%s'", ip, username, filePath))
			}
		}

		next.ServeHTTP(w, r)
	})
}

// parseVolumeSpec parses <path>:<name>:<perms>,<users> from the right. The
// path itself may contain colons (Windows drive letters, Unix filenames).
func parseVolumeSpec(v string) (path, name, permsUsers string, ok bool) {
	i := strings.LastIndex(v, ":")
	if i <= 0 || i == len(v)-1 {
		return "", "", "", false
	}
	j := strings.LastIndex(v[:i], ":")
	if j <= 0 {
		return "", "", "", false
	}
	path = v[:j]
	name = v[j+1 : i]
	permsUsers = v[i+1:]
	if path == "" || name == "" || permsUsers == "" {
		return "", "", "", false
	}
	return path, name, permsUsers, true
}

func expandUser(path string) string {
	if strings.HasPrefix(path, "~") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[1:])
	}
	// Windows: expand %USERPROFILE% or %HOME% if present
	if strings.Contains(path, "%USERPROFILE%") || strings.Contains(path, "%HOME%") {
		home, _ := os.UserHomeDir()
		path = strings.ReplaceAll(path, "%USERPROFILE%", home)
		path = strings.ReplaceAll(path, "%HOME%", home)
	}
	return path
}

func catchAllHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		authRequired(indexHandler)(w, r)
	} else if strings.HasPrefix(path, "share/") {
		shareViewHandler(w, r)
	} else {
		authRequired(handlePathHandler)(w, r)
	}
}

func maxBodyMiddleware(next http.Handler, maxBody int64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxBody)
		next.ServeHTTP(w, r)
	})
}

var SHOWHELP = `netdisk ` + Version + `

Usage:
  netdisk [options]

Options:
  -p <port>        Port (default: 8880)
  -a <user:pass>   Add user (repeatable)
  -v <path:name:perms,users>  Add volume (repeatable)
  -t <hours>       Session timeout in hours (default: 168, 0=no expiry)
  -m <size>      Max upload size (default: 10G)
  -c <path>        Config directory (default: ~/.config/netdisk)

Windows volume examples (quote paths with spaces):
  -v C:\data:docs:rwdm,admin
  -v "D:\My Media:media:r,user"
  -v \\server\share:net:r,user

Keyboard shortcuts (in player):
  j/↓: next  k/↑: pre  <: speed down  >: speed up
  space: play/pause  s: speed  d: delete

netdisk -p 8880 -a bs:bs -a us:us -a fm:fm -v ~/:nd:rwdm,bs -v ~/:ud:r,us -v ~/dl/:dl:rwdm,us -v ~/fm/:fm:r,fm -v ~/md/:md:r -v ~/md/mv/:md:r ~/md/ok:md:r -t 168
`
