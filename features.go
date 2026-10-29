package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// ==============================================================================
// ===== Static Files ===========================================================
// ==============================================================================

func staticHandler(w http.ResponseWriter, r *http.Request) {
	filename := strings.TrimPrefix(r.URL.Path, "/static/")
	data, err := staticFiles.ReadFile("static/" + filename)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if strings.HasSuffix(filename, ".js") {
		w.Header().Set("Content-Type", "application/javascript")
	} else if strings.HasSuffix(filename, ".css") {
		w.Header().Set("Content-Type", "text/css")
	}
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Write(data)
}

// ==============================================================================
// ===== File Search ============================================================
// ==============================================================================

type SearchResult struct {
	Name           string `json:"name"`
	Path           string `json:"path"`
	IsDir          bool   `json:"is_dir"`
	Size           int64  `json:"size"`
	FormattedMtime string `json:"formatted_mtime"`
	CanDelete      bool   `json:"can_delete"`
	CanEdit        bool   `json:"can_edit"`
	CanRename      bool   `json:"can_rename"`
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	query := strings.ToLower(r.URL.Query().Get("q"))
	dirPath := r.URL.Query().Get("dir")
	user := getUser(r)

	absDir := getFullAbsPath(dirPath)
	if absDir == "" {
		json.NewEncoder(w).Encode([]SearchResult{})
		return
	}

	_, volInfo := getVolAndPath(dirPath)
	if volInfo == nil || !hasPermission(user, volInfo, 'r') {
		json.NewEncoder(w).Encode([]SearchResult{})
		return
	}

	canWrite := hasPermission(user, volInfo, 'w')
	canDelete := hasPermission(user, volInfo, 'd')
	var results []SearchResult

	var searchDir func(dir string, relPrefix string, depth int)
	searchDir = func(dir string, relPrefix string, depth int) {
		if depth > 3 {
			return
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			return
		}
		for _, entry := range entries {
			name := entry.Name()
			fi, err := entry.Info()
			if err != nil {
				continue
			}
			if isHiddenFile(name, fi) {
				continue
			}
			if query != "" && !strings.Contains(strings.ToLower(name), query) {
				continue
			}
			itemPath := relPrefix + name
			if entry.IsDir() {
				itemPath += "/"
			}
			results = append(results, SearchResult{
				Name:           name,
				Path:           itemPath,
				IsDir:          entry.IsDir(),
				Size:           fi.Size(),
				FormattedMtime: fi.ModTime().Format("06-01-02"),
				CanDelete:      canDelete,
				CanEdit:        !entry.IsDir() && canWrite && isEditableExt(name),
				CanRename:      canWrite,
			})
			if entry.IsDir() {
				searchDir(filepath.Join(dir, name), relPrefix+name+"/", depth+1)
			}
		}
	}

	searchDir(absDir, dirPath+"/", 0)

	sort.Slice(results, func(i, j int) bool {
		ni, nj := strings.ToLower(results[i].Name), strings.ToLower(results[j].Name)
		q := strings.ToLower(query)
		if query == "" {
			return ni < nj
		}
		ii, ij := ni == q, nj == q
		if ii != ij {
			return ii
		}
		pi, pj := strings.HasPrefix(ni, q), strings.HasPrefix(nj, q)
		if pi != pj {
			return pi
		}
		return ni < nj
	})

	if results == nil {
		results = []SearchResult{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// ==============================================================================
// ===== Share Links ============================================================
// ==============================================================================

type ShareLink struct {
	Token        string    `json:"token"`
	Path         string    `json:"path"`
	CreatedBy    string    `json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
	Filename     string    `json:"filename"`
	PasswordHash string    `json:"password_hash,omitempty"`
	PasswordSalt string    `json:"password_salt,omitempty"`
}

type SharePasswordData struct {
	Token    string
	Filename string
	Error    string
}

var (
	shareLinks   = make(map[string]*ShareLink)
	shareLinksMu sync.Mutex
)

const (
	shareStoreFile        = "share.json"
	sharePasswordRounds   = 100000
	sharePasswordKeyBytes = 32
)

type shareStore struct {
	Shares []*ShareLink `json:"shares"`
}

func shareStorePath() string {
	return filepath.Join(configDir, shareStoreFile)
}

// saveShareLinksLocked writes the in-memory shares to <config>/share.json.
// The caller must hold shareLinksMu.
func saveShareLinksLocked() {
	list := make([]*ShareLink, 0, len(shareLinks))
	for _, s := range shareLinks {
		list = append(list, s)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].CreatedAt.Before(list[j].CreatedAt)
	})

	data, err := json.MarshalIndent(shareStore{Shares: list}, "", "  ")
	if err != nil {
		writeLog("web", "SHARE save failed: "+err.Error())
		return
	}

	path := shareStorePath()
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		writeLog("web", "SHARE save failed: "+err.Error())
		return
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		writeLog("web", "SHARE save failed: "+err.Error())
	}
}

// loadShareLinks restores persisted share links at startup.
func loadShareLinks() {
	data, err := os.ReadFile(shareStorePath())
	if err != nil {
		if !os.IsNotExist(err) {
			writeLog("web", "SHARE load failed: "+err.Error())
		}
		return
	}

	var store shareStore
	if err := json.Unmarshal(data, &store); err != nil {
		writeLog("web", "SHARE load failed: "+err.Error())
		return
	}

	now := time.Now()
	removed := false
	shareLinksMu.Lock()
	for _, s := range store.Shares {
		if s == nil || s.Token == "" {
			continue
		}
		if !s.ExpiresAt.IsZero() && now.After(s.ExpiresAt) {
			removed = true
			continue
		}
		shareLinks[s.Token] = s
	}
	if removed {
		saveShareLinksLocked()
	}
	shareLinksMu.Unlock()
}

// pbkdf2SHA256 is a small dependency-free PBKDF2-HMAC-SHA256 implementation.
func pbkdf2SHA256(password, salt []byte, iterations, keyLen int) []byte {
	if iterations < 1 {
		iterations = 1
	}
	hLen := sha256.Size
	numBlocks := (keyLen + hLen - 1) / hLen
	dk := make([]byte, 0, numBlocks*hLen)
	var blockIndex [4]byte

	for block := 1; block <= numBlocks; block++ {
		binary.BigEndian.PutUint32(blockIndex[:], uint32(block))
		mac := hmac.New(sha256.New, password)
		mac.Write(salt)
		mac.Write(blockIndex[:])
		u := mac.Sum(nil)
		t := make([]byte, len(u))
		copy(t, u)

		for n := 1; n < iterations; n++ {
			mac = hmac.New(sha256.New, password)
			mac.Write(u)
			u = mac.Sum(u[:0])
			for i := range t {
				t[i] ^= u[i]
			}
		}
		dk = append(dk, t...)
	}
	return dk[:keyLen]
}

func setSharePassword(share *ShareLink, password string) error {
	if password == "" {
		share.PasswordHash = ""
		share.PasswordSalt = ""
		return nil
	}
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return err
	}
	hash := pbkdf2SHA256([]byte(password), salt, sharePasswordRounds, sharePasswordKeyBytes)
	share.PasswordSalt = base64.StdEncoding.EncodeToString(salt)
	share.PasswordHash = base64.StdEncoding.EncodeToString(hash)
	return nil
}

func verifySharePassword(share *ShareLink, password string) bool {
	if share.PasswordHash == "" {
		return true
	}
	salt, err := base64.StdEncoding.DecodeString(share.PasswordSalt)
	if err != nil {
		return false
	}
	want, err := base64.StdEncoding.DecodeString(share.PasswordHash)
	if err != nil || len(want) == 0 {
		return false
	}
	got := pbkdf2SHA256([]byte(password), salt, sharePasswordRounds, len(want))
	return subtle.ConstantTimeCompare(got, want) == 1
}

func shareCookieName(token string) string {
	return "share_auth_" + token
}

func shareCookieValue(share *ShareLink) string {
	mac := hmac.New(sha256.New, []byte(share.PasswordHash))
	mac.Write([]byte("share-auth:" + share.Token))
	return hex.EncodeToString(mac.Sum(nil))
}

func shareAuthorized(r *http.Request, share *ShareLink) bool {
	c, err := r.Cookie(shareCookieName(share.Token))
	if err != nil {
		return false
	}
	want := shareCookieValue(share)
	return subtle.ConstantTimeCompare([]byte(c.Value), []byte(want)) == 1
}

func generateToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func shareResponse(s *ShareLink) map[string]interface{} {
	return map[string]interface{}{
		"token":        s.Token,
		"path":         s.Path,
		"created_by":   s.CreatedBy,
		"created_at":   s.CreatedAt,
		"expires_at":   s.ExpiresAt,
		"filename":     s.Filename,
		"url":          fmt.Sprintf("/share/%s", s.Token),
		"has_password": s.PasswordHash != "",
	}
}

// shareGetHandler returns the current user's active share for a path, if any.
func shareGetHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	user := getUser(r)
	now := time.Now()

	var found *ShareLink
	removed := false
	shareLinksMu.Lock()
	for token, s := range shareLinks {
		if !s.ExpiresAt.IsZero() && now.After(s.ExpiresAt) {
			delete(shareLinks, token)
			removed = true
			continue
		}
		if s.CreatedBy == user && s.Path == path {
			if found == nil || s.CreatedAt.After(found.CreatedAt) {
				found = s
			}
		}
	}
	if removed {
		saveShareLinksLocked()
	}
	shareLinksMu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	if found == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "share": nil})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "share": shareResponse(found)})
}

func shareCreateHandler(w http.ResponseWriter, r *http.Request) {
	var d struct {
		Path     string `json:"path"`
		Hours    int    `json:"hours"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if len(d.Password) > 1024 {
		jsonError(w, "Password is too long", http.StatusBadRequest)
		return
	}
	if d.Hours < 0 {
		d.Hours = 0
	}
	if d.Hours > 876000 {
		d.Hours = 876000
	}

	user := getUser(r)
	absPath := getFullAbsPath(d.Path)
	if absPath == "" {
		jsonError(w, "Invalid path", http.StatusBadRequest)
		return
	}

	token := generateToken()
	share := &ShareLink{
		Token:     token,
		Path:      d.Path,
		CreatedBy: user,
		CreatedAt: time.Now(),
		Filename:  filepath.Base(absPath),
	}
	if d.Hours > 0 {
		share.ExpiresAt = time.Now().Add(time.Duration(d.Hours) * time.Hour)
	}
	if err := setSharePassword(share, d.Password); err != nil {
		jsonError(w, "Failed to secure password", http.StatusInternalServerError)
		return
	}

	shareLinksMu.Lock()
	// Keep at most one active link per user + path. Generating a new link
	// revokes the previous one, and expired links are cleaned up as well.
	for oldToken, old := range shareLinks {
		if (!old.ExpiresAt.IsZero() && time.Now().After(old.ExpiresAt)) ||
			(old.CreatedBy == user && old.Path == d.Path) {
			delete(shareLinks, oldToken)
		}
	}
	shareLinks[token] = share
	saveShareLinksLocked()
	shareLinksMu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"token":        token,
		"url":          fmt.Sprintf("/share/%s", token),
		"expires_at":   share.ExpiresAt,
		"has_password": share.PasswordHash != "",
	})
}

func shareDeleteHandler(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimPrefix(r.URL.Path, "/api/share/")
	user := getUser(r)

	shareLinksMu.Lock()
	if s, ok := shareLinks[token]; ok && s.CreatedBy != user {
		shareLinksMu.Unlock()
		jsonError(w, "Forbidden", http.StatusForbidden)
		return
	}
	delete(shareLinks, token)
	saveShareLinksLocked()
	shareLinksMu.Unlock()
	jsonOK(w)
}

func shareListHandler(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	removed := false
	var list []map[string]interface{}

	shareLinksMu.Lock()
	for token, s := range shareLinks {
		if !s.ExpiresAt.IsZero() && now.After(s.ExpiresAt) {
			delete(shareLinks, token)
			removed = true
			continue
		}
		list = append(list, shareResponse(s))
	}
	if removed {
		saveShareLinksLocked()
	}
	shareLinksMu.Unlock()

	if list == nil {
		list = []map[string]interface{}{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func shareViewHandler(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimPrefix(r.URL.Path, "/share/")
	shareLinksMu.Lock()
	share, ok := shareLinks[token]
	shareLinksMu.Unlock()

	if !ok {
		http.NotFound(w, r)
		return
	}
	if !share.ExpiresAt.IsZero() && time.Now().After(share.ExpiresAt) {
		shareLinksMu.Lock()
		delete(shareLinks, token)
		saveShareLinksLocked()
		shareLinksMu.Unlock()
		http.Error(w, "Link expired", http.StatusGone)
		return
	}

	// Password-protected links require a password before the content is served.
	if share.PasswordHash != "" && !shareAuthorized(r, share) {
		if r.Method == http.MethodPost {
			r.ParseForm()
			if !verifySharePassword(share, r.FormValue("password")) {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.WriteHeader(http.StatusUnauthorized)
				executeTemplate(w, sharePasswordTmpl, SharePasswordData{
					Token:    share.Token,
					Filename: share.Filename,
					Error:    "Incorrect password",
				})
				return
			}
			maxAge := 0
			if !share.ExpiresAt.IsZero() {
				maxAge = int(time.Until(share.ExpiresAt).Seconds())
				if maxAge < 1 {
					maxAge = 1
				}
				if maxAge > 315360000 {
					maxAge = 315360000
				}
			}
			http.SetCookie(w, &http.Cookie{
				Name:     shareCookieName(share.Token),
				Value:    shareCookieValue(share),
				Path:     "/share/" + share.Token,
				MaxAge:   maxAge,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				Secure:   r.TLS != nil,
			})
			logActivity("Share", "PASSWORD_OK", share.Path, r)
			http.Redirect(w, r, "/share/"+share.Token, http.StatusSeeOther)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		executeTemplate(w, sharePasswordTmpl, SharePasswordData{
			Token:    share.Token,
			Filename: share.Filename,
		})
		return
	}

	// A password form POST to an already-unlocked link should not re-serve content.
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Redirect(w, r, "/share/"+share.Token, http.StatusSeeOther)
		return
	}

	absPath := getFullAbsPath(share.Path)
	if absPath == "" {
		http.NotFound(w, r)
		return
	}

	info, err := os.Stat(absPath)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Directory: zip and download
	if info.IsDir() {
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", attachmentDisposition(filepath.Base(absPath)+".zip"))
		zipWriter := NewZipStreamingWriter(w)
		filepath.Walk(absPath, func(path string, fi os.FileInfo, err error) error {
			if err != nil || fi.IsDir() || fi.Mode()&os.ModeSymlink != 0 {
				return nil
			}
			rel, _ := filepath.Rel(absPath, path)
			if zerr := zipWriter.AddFile(path, rel); zerr != nil {
				writeLog("zip", fmt.Sprintf("(%s) ADD-FAILED %s: %v", clientIP(r), path, zerr))
			}
			return nil
		})
		if zerr := zipWriter.Close(); zerr != nil {
			writeLog("zip", fmt.Sprintf("(%s) CLOSE-FAILED: %v", clientIP(r), zerr))
		}
		logActivity("Share", "DOWNLOAD_ZIP", share.Path, r)
		return
	}

	ext := strings.ToLower(filepath.Ext(absPath))

	// Markdown preview
	if ext == ".md" || ext == ".markdown" {
		data, err := os.ReadFile(absPath)
		if err != nil {
			http.Error(w, "Error reading file", http.StatusInternalServerError)
			return
		}
		previewData := PreviewData{
			Path:       share.Path,
			Content:    string(data),
			IsMarkdown: true,
			IsCode:     false,
			Filename:   share.Filename,
			ShareMode:  true,
		}
		executeTemplate(w, previewTmpl, previewData)
		return
	}

	// Code preview
	if isEditableExt(absPath) {
		data, err := os.ReadFile(absPath)
		if err != nil {
			http.Error(w, "Error reading file", http.StatusInternalServerError)
			return
		}
		previewData := PreviewData{
			Path:       share.Path,
			Content:    string(data),
			IsMarkdown: false,
			IsCode:     true,
			Language:   ext[1:],
			Filename:   share.Filename,
			ShareMode:  true,
		}
		executeTemplate(w, previewTmpl, previewData)
		return
	}

	// Serve file directly
	http.ServeFile(w, r, absPath)
}

// ==============================================================================
// ===== Preview ================================================================
// ==============================================================================

type PreviewData struct {
	Path       string
	Content    string
	IsMarkdown bool
	IsCode     bool
	Language   string
	Filename   string
	ShareMode  bool
}

func previewHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/preview/")
	path = strings.TrimPrefix(path, "preview/")

	absPath := getFullAbsPath(path)
	if absPath == "" {
		http.NotFound(w, r)
		return
	}

	info, err := os.Stat(absPath)
	if err != nil || info.IsDir() {
		http.NotFound(w, r)
		return
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}

	ext := strings.ToLower(filepath.Ext(absPath))
	isMd := ext == ".md" || ext == ".markdown"
	isCode := isEditableExt(absPath) && !isMd

	lang := ""
	if isCode && len(ext) > 1 {
		lang = ext[1:]
	}

	previewData := PreviewData{
		Path:       path,
		Content:    string(data),
		IsMarkdown: isMd,
		IsCode:     isCode,
		Language:   lang,
		Filename:   filepath.Base(absPath),
	}
	executeTemplate(w, previewTmpl, previewData)
}
