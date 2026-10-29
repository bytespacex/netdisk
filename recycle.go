package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ==============================================================================
// ===== Recycle Bin (web restore UI) ===========================================
// ==============================================================================

// RecycleItem is a single restorable entry in the recycle bin.
type RecycleItem struct {
	Path  string `json:"path"` // original web path, e.g. "vol/sub/file.txt"
	Name  string `json:"name"`
	IsDir bool   `json:"is_dir"`
	Size  int64  `json:"size"`
	Mtime string `json:"mtime"`
}

// recycleItems walks the recycle dir and returns restorable entries.
// The recycle dir mirrors the original web paths, so the relative path under
// the recycle dir is exactly the original web path.
//
// Only "leaf" directories (empty) and files are listed: parent directories that
// were created just to hold a nested deletion are skipped, which keeps every
// listed item unambiguous to restore.
func recycleItems() []RecycleItem {
	items := []RecycleItem{}
	if recycleBin == "" {
		return items
	}
	filepath.Walk(recycleBin, func(p string, fi os.FileInfo, err error) error {
		if err != nil || p == recycleBin {
			return nil
		}
		if fi.IsDir() {
			if entries, e := os.ReadDir(p); e == nil && len(entries) > 0 {
				return nil // container dir, not a deleted entry itself
			}
		}
		rel, e := filepath.Rel(recycleBin, p)
		if e != nil {
			return nil
		}
		items = append(items, RecycleItem{
			Path:  filepath.ToSlash(rel),
			Name:  filepath.Base(p),
			IsDir: fi.IsDir(),
			Size:  fi.Size(),
			Mtime: fi.ModTime().Format("2006-01-02 15:04"),
		})
		return nil
	})
	sort.Slice(items, func(i, j int) bool { return items[i].Path < items[j].Path })
	return items
}

// recyclePath maps an original web path to its path inside the recycle dir,
// rejecting anything that would escape the recycle dir.
func recyclePath(webPath string) string {
	clean := strings.TrimPrefix(filepath.ToSlash(webPath), "/")
	abs := filepath.Join(recycleBin, filepath.FromSlash(clean))
	if !pathInside(recycleBin, abs) {
		return ""
	}
	return abs
}

// pathInside reports whether p is inside (or equal to) root.
func pathInside(root, p string) bool {
	rel, err := filepath.Rel(root, p)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

// uniquePath returns dst if it does not exist, otherwise a sibling name that
// does not collide, so an existing file is never overwritten.
//
// With an empty tag the name becomes "name (N).ext"; otherwise
// "name (tag N).ext" (e.g. "name (restored 1).ext").
func uniquePath(dst, tag string) string {
	if _, err := os.Lstat(dst); os.IsNotExist(err) {
		return dst
	}
	dir := filepath.Dir(dst)
	base := filepath.Base(dst)
	ext := filepath.Ext(base)
	stem := strings.TrimSuffix(base, ext)
	name := func(n int) string {
		if tag == "" {
			return fmt.Sprintf("%s (%d)%s", stem, n, ext)
		}
		return fmt.Sprintf("%s (%s %d)%s", stem, tag, n, ext)
	}
	for i := 1; i < 1000; i++ {
		cand := filepath.Join(dir, name(i))
		if _, err := os.Lstat(cand); os.IsNotExist(err) {
			return cand
		}
	}
	return filepath.Join(dir, name(int(time.Now().UnixNano())))
}

// uniqueDest returns dst if it does not exist, otherwise a sibling name with a
// " (restored N)" suffix so an existing file is never overwritten.
func uniqueDest(dst string) string {
	return uniquePath(dst, "restored")
}

// cleanupEmptyDirs removes empty directories left behind under the recycle dir.
func cleanupEmptyDirs() {
	if recycleBin == "" {
		return
	}
	var clean func(dir string) bool
	clean = func(dir string) bool {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return false
		}
		empty := true
		for _, e := range entries {
			if e.IsDir() {
				sub := filepath.Join(dir, e.Name())
				if clean(sub) {
					os.Remove(sub)
				} else {
					empty = false
				}
			} else {
				empty = false
			}
		}
		return empty
	}
	clean(recycleBin)
}

func recycleListHandler(w http.ResponseWriter, r *http.Request) {
	if recycleBin == "" {
		json.NewEncoder(w).Encode([]RecycleItem{})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(recycleItems())
}

// recycleRestoreHandler moves items from the recycle dir back to their original
// location. Requires write permission on the target volume.
//
// If a destination already exists and no conflict policy was supplied, the
// handler replies 409 with the conflicting paths so the UI can ask the user.
// Supported policies: "overwrite", "rename", "skip".
func recycleRestoreHandler(w http.ResponseWriter, r *http.Request) {
	var d struct {
		Paths    []string `json:"paths"`
		Conflict string   `json:"conflict"`
	}
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Ask the caller how to handle existing destinations before touching anything.
	if d.Conflict == "" {
		conflicts := []string{}
		for _, p := range d.Paths {
			dst := getFullAbsPath(p)
			if dst == "" {
				continue
			}
			if _, err := os.Lstat(dst); err == nil {
				conflicts = append(conflicts, p)
			}
		}
		if len(conflicts) > 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":   false,
				"conflicts": conflicts,
				"total":     len(d.Paths),
			})
			return
		}
	}

	restored, skipped := 0, 0
	for _, p := range d.Paths {
		_, info := getVolAndPath(p)
		if info == nil || !hasPermission(getUser(r), info, 'w') {
			skipped++
			continue
		}
		src := recyclePath(p)
		dst := getFullAbsPath(p)
		if src == "" || dst == "" {
			skipped++
			continue
		}
		if _, err := os.Stat(src); err != nil {
			skipped++
			continue
		}
		if _, err := os.Lstat(dst); err == nil {
			switch d.Conflict {
			case "skip":
				skipped++
				continue
			case "overwrite":
				if err := os.RemoveAll(dst); err != nil {
					skipped++
					continue
				}
			default: // "rename" (or any unknown value)
				dst = uniqueDest(dst)
			}
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			writeLog("recycle", fmt.Sprintf("(%s) RESTORE-FAILED %s: %v", clientIP(r), p, err))
			skipped++
			continue
		}
		if err := movePath(src, dst); err != nil {
			writeLog("recycle", fmt.Sprintf("(%s) RESTORE-FAILED %s: %v", clientIP(r), p, err))
			skipped++
			continue
		}
		restored++
	}
	cleanupEmptyDirs()
	if restored > 0 {
		logActivity("web", "RESTORED_FROM_RECYCLE", strings.Join(d.Paths, ", "), r)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"restored": restored, "skipped": skipped})
}

// recycleDeleteHandler permanently removes items from the recycle dir.
// Requires delete permission on the target volume.
func recycleDeleteHandler(w http.ResponseWriter, r *http.Request) {
	var d struct {
		Paths []string `json:"paths"`
	}
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}
	for _, p := range d.Paths {
		_, info := getVolAndPath(p)
		if info == nil || !hasPermission(getUser(r), info, 'd') {
			continue
		}
		if src := recyclePath(p); src != "" {
			os.RemoveAll(src)
		}
	}
	cleanupEmptyDirs()
	jsonOK(w)
}

// recycleClearHandler empties the whole recycle bin.
func recycleClearHandler(w http.ResponseWriter, r *http.Request) {
	os.RemoveAll(recycleBin)
	os.MkdirAll(recycleBin, 0755)
	logActivity("web", "RECYCLE_CLEARED", "", r)
	jsonOK(w)
}
