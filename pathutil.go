package main

import (
	"mime"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"
)

// unsafeNameChars covers characters that are invalid (or problematic) on
// Windows and control characters that are invalid on every supported platform.
var unsafeNameChars = regexp.MustCompile(`[\\/:*?"<>|\x00-\x1f\x7f]`)

var windowsReservedNames = map[string]bool{
	"CON": true, "PRN": true, "AUX": true, "NUL": true,
	"COM1": true, "COM2": true, "COM3": true, "COM4": true, "COM5": true,
	"COM6": true, "COM7": true, "COM8": true, "COM9": true,
	"LPT1": true, "LPT2": true, "LPT3": true, "LPT4": true, "LPT5": true,
	"LPT6": true, "LPT7": true, "LPT8": true, "LPT9": true,
	"CONIN$": true, "CONOUT$": true, "CLOCK$": true,
}

// getSafeName turns a browser-supplied filename into a single, portable path
// component. It is deliberately conservative: names that are legal on Unix but
// invalid on Windows are normalized too, so a volume can be moved between
// platforms without creating names that cannot be accessed.
func getSafeName(filename string) string {
	if filename == "" {
		return "unnamed_file"
	}

	// On Unix filepath.Base does not treat backslashes as separators, so the
	// regex below also replaces them. On Windows Base handles both separators.
	filename = filepath.Base(filename)
	filename = unsafeNameChars.ReplaceAllString(filename, "_")
	filename = strings.TrimRight(filename, ". ")
	if filename == "" {
		return "unnamed_file"
	}

	// Windows reserves device names even when an extension is present, e.g.
	// "CON.txt" is still the console device.
	base := filename
	if i := strings.IndexByte(base, '.'); i >= 0 {
		base = base[:i]
	}
	if windowsReservedNames[strings.ToUpper(base)] {
		filename = "_" + filename
	}

	filename = strings.TrimRight(truncateName(filename, 220), ". ")
	if filename == "" || filename == "." || filename == ".." {
		return "unnamed_file"
	}
	return filename
}

// validNameComponent reports whether name is a single, already-existing path
// component that can safely be appended to a directory. Unlike getSafeName it
// does not rewrite the name, so it can be used to address existing files.
func validNameComponent(name string) bool {
	if name == "" || name == "." || name == ".." {
		return false
	}
	if strings.ContainsAny(name, `/\`) {
		return false
	}
	return !strings.ContainsRune(name, 0)
}

// sanitizeRelDir validates and normalizes the directory part of a relative
// upload path (for example "folder/sub/file.txt"). Every component is passed
// through getSafeName and ".." components are rejected.
func sanitizeRelDir(rel string) (string, bool) {
	rel = strings.ReplaceAll(rel, "\\", "/")
	parts := strings.Split(rel, "/")
	if len(parts) <= 1 {
		return "", true
	}

	dirs := make([]string, 0, len(parts)-1)
	for _, part := range parts[:len(parts)-1] {
		if part == "" || part == "." {
			continue
		}
		if part == ".." {
			return "", false
		}
		safe := getSafeName(part)
		if safe == "" || safe == "." || safe == ".." {
			return "", false
		}
		dirs = append(dirs, safe)
	}
	return filepath.Join(dirs...), true
}

func truncateName(name string, max int) string {
	if len(name) <= max {
		return name
	}
	ext := filepath.Ext(name)
	if len(ext) >= max {
		return truncateUTF8(name, max)
	}
	stem := strings.TrimSuffix(name, ext)
	return truncateUTF8(stem, max-len(ext)) + ext
}

func truncateUTF8(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if len(s) <= max {
		return s
	}
	for max > 0 && !utf8.RuneStart(s[max]) {
		max--
	}
	return s[:max]
}

// urlPathEscape escapes each segment of a slash-separated web path, leaving the
// separators intact. This is needed for names containing '#', '?' or '%'.
func urlPathEscape(p string) string {
	parts := strings.Split(p, "/")
	for i := range parts {
		parts[i] = url.PathEscape(parts[i])
	}
	return strings.Join(parts, "/")
}

// attachmentDisposition builds a correctly quoted/encoded Content-Disposition
// value for arbitrary (including non-ASCII) filenames.
func attachmentDisposition(filename string) string {
	if v := mime.FormatMediaType("attachment", map[string]string{"filename": filename}); v != "" {
		return v
	}
	return "attachment"
}

// pathInsideVolume is pathInside plus symlink checks. A path is accepted only
// if every existing component resolves to a location inside root. This blocks
// both lexical traversal and symlinks/junctions that point outside the volume.
func pathInsideVolume(root, p string) bool {
	root = filepath.Clean(root)
	p = filepath.Clean(p)
	if !pathInside(root, p) {
		return false
	}

	realRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return false
	}

	rel, err := filepath.Rel(root, p)
	if err != nil || rel == "." {
		return err == nil
	}

	cur := realRoot
	for _, part := range strings.Split(rel, string(filepath.Separator)) {
		if part == "" || part == "." {
			continue
		}
		if part == ".." {
			return false
		}

		next := filepath.Join(cur, part)
		fi, err := os.Lstat(next)
		if err != nil {
			// The remainder does not exist yet. Since ".." components were
			// rejected, no later component can be an existing symlink.
			return os.IsNotExist(err)
		}
		if fi.Mode()&os.ModeSymlink != 0 {
			resolved, err := filepath.EvalSymlinks(next)
			if err != nil || !pathInside(realRoot, resolved) {
				return false
			}
			cur = resolved
			continue
		}
		cur = next
	}
	return pathInside(realRoot, cur)
}

// relToUploadBase returns the path of p relative to the upload base directory.
func relToUploadBase(base, p string) (string, bool) {
	base = strings.Trim(base, "/")
	p = strings.Trim(p, "/")
	if base == "" {
		return p, true
	}
	if p == base {
		return "", true
	}
	if !strings.HasPrefix(p, base+"/") {
		return "", false
	}
	return p[len(base)+1:], true
}

// sanitizedUploadFile computes the actual target for one uploaded file. Only
// the path relative to the upload base is normalized; the base directory itself
// is used as-is so existing names containing characters such as '?' still work.
func sanitizedUploadFile(baseAbs, rel string) string {
	rel = strings.ReplaceAll(rel, "\\", "/")
	if rel == "" {
		return ""
	}
	dir, ok := sanitizeRelDir(rel)
	if !ok {
		return ""
	}
	name := rel
	if i := strings.LastIndexByte(rel, '/'); i >= 0 {
		name = rel[i+1:]
	}
	name = getSafeName(name)
	if name == "" {
		return ""
	}
	abs := filepath.Join(baseAbs, filepath.FromSlash(dir), name)
	if !pathInsideVolume(baseAbs, abs) {
		return ""
	}
	return abs
}

// sanitizedUploadDir computes the actual target for an uploaded top-level
// folder, normalizing only the components below baseAbs.
func sanitizedUploadDir(baseAbs, rel string) string {
	rel = strings.ReplaceAll(rel, "\\", "/")
	parts := strings.Split(rel, "/")
	cleaned := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" || part == "." {
			continue
		}
		if part == ".." {
			return ""
		}
		safe := getSafeName(part)
		if safe == "" || safe == "." || safe == ".." {
			return ""
		}
		cleaned = append(cleaned, safe)
	}
	abs := filepath.Join(baseAbs, filepath.Join(cleaned...))
	if !pathInsideVolume(baseAbs, abs) {
		return ""
	}
	return abs
}

func validVolumeName(name string) bool {
	if name == "" || name == "." || name == ".." {
		return false
	}
	if strings.TrimSpace(name) != name {
		return false
	}
	if strings.ContainsAny(name, `/\:*?"<>|`) {
		return false
	}
	for _, r := range name {
		if r < 0x20 || r == 0x7f {
			return false
		}
	}
	return true
}
