# netdisk - Engineering White Paper

## Overview

netdisk is a feature-rich lightweight file server written in Go. It provides a web file-management UI, a WebDAV service, a built-in media player, file share links, a WebClip tool, a file editor, search, a recycle bin, and more. Single-binary deployment with zero external dependencies.

## Recent Highlights

> 2026-10-18

- **Streaming large-file uploads**: `uploadHandler` uses `MultipartReader` to write each part straight to the target file, never through `/tmp`, eliminating the "drops halfway" failures caused by tmpfs/quota limits; on failure it automatically removes the half-written file.
- **Connection diagnostics log (`diag.go`)**: middleware records request/response bytes and duration for uploads/downloads; client disconnects are logged as `ctxErr`, failed response writes as `writeErr`; handler panics are captured with stack traces to help pinpoint disconnects.
- **Redesigned upload progress panel**: pinned to the top-center; the first line shows `(done/total)` and the total size, the second line shows speed + progress bar + percent, and a dropdown arrow expands the per-file list. Uploads run concurrently (up to 3 files at a time).
- **Subtitle styling**: white text with a black outline and no background box; the font scales with the video's rendered size, so windowed and fullscreen stay proportional.
- **Recycle bin web UI**: browse deleted items from the `+` menu with checkboxes for multi-select, and footer actions **Select All / Invert / Restore / Delete / Empty Bin**; restore returns items to their original paths.
- **Name-conflict resolution**: when uploading or restoring, if a destination already exists the UI asks how to proceed — for files **Overwrite / Auto rename / Skip / Cancel**, and for an uploaded folder that already exists **Merge / Rename folder / Skip folder / Cancel**. The chosen policy is sent in the request URL, so it is applied reliably regardless of multipart field ordering; choosing **Skip** is handled entirely on the client, so conflicting files are not transferred at all. HTML responses carry `no-cache` headers so a browser never keeps running stale client logic after an upgrade. Folder uploads land in the correct nested path (fixed double-nesting), and drag-and-drop folders keep their directory structure (`webkitRelativePath` is read-only, so it is now set with `Object.defineProperty`).
- **Sticky top bar on top**: raised the top bar's stacking order so the `+` button is no longer covered by the row action (⋯) menus while scrolling the list.
- **Streaming ZIP + ZIP64 (`zipwriter.go`)**: switched to `archive/zip` streaming writes, no longer reading whole files into memory, fixing OOM and the `uint32` overflow for ≥4GB files during batch downloads.
- **Streaming WebClip uploads**: likewise avoids caching large files to `/tmp`.
- **Embedded subtitle support (optional)**: when `ffmpeg`/`ffprobe` are installed, subtitle tracks inside MKV/MP4 containers are detected and extracted (converted to WebVTT); otherwise it falls back to sidecar subtitle files with the same name.
- **Audio codec hint**: after a video starts playing, if the browser cannot decode its audio track (AC-3 / DTS / TrueHD, etc.), an overlay shows a hint.
- **Log-delete confirmation**: replaced the native browser `confirm` with a web-style modal.
- **Build script**: at the end, `build.sh` asks whether to move artifacts to `~/bn/netdisk` (a single artifact is renamed) and remove `dist/`; default is N.

## Tech Stack

| Layer | Technology |
|------|------|
| Language | Go 1.22 |
| WebDAV | `golang.org/x/net/webdav` |
| Frontend | Vanilla HTML/CSS/JS (embedded via Go embed) |
| Icons | Lucide Icons (embedded) |
| Markdown | marked.min.js (embed) |
| Syntax highlighting | highlight.min.js (embed) |
| Build | `-trimpath -ldflags="-s -w"` static linking |

## Project Layout

```
netdisk/
├── main.go           # entry point, routing, CLI flags, WebDAV setup
├── handlers.go       # core handlers, auth, sessions, file ops, recycle bin
├── diag.go           # connection diagnostics log (upload/download drops, panic capture)
├── subtitles.go      # embedded subtitle extraction (optional ffmpeg/ffprobe, sidecar fallback)
├── features.go       # static files, search, share links, file preview
├── pathutil.go       # portable filenames, URL escaping, path sandboxing
├── move.go           # rename with cross-filesystem copy fallback
├── webclip.go        # WebClip tool
├── zipwriter.go      # streaming ZIP (archive/zip, no temp files, ZIP64)
├── templates.go      # Go HTML templates (full frontend embedded)
├── embed.go          # go:embed static assets
├── static/           # embedded JS/CSS (marked, highlight.js, Lucide)
│   ├── marked.min.js
│   ├── highlight.min.js
│   ├── hljs-dark.min.css
│   └── lucide.min.js
├── build.sh          # build script (Linux/macOS)
├── build.cmd         # build script (Windows)
├── .github/workflows/release.yml  # CI: build all platforms + publish Release
├── LICENSE           # MIT
├── THIRD_PARTY_NOTICES.md
├── go.mod
└── go.sum
```

## Core Architecture

### Multi-Volume System

netdisk supports multiple independent mount points via the `-v` flag; each volume has its own path, permissions, and user list:

```
-v <local path>:<volume name>:<perms>,<users>
```

Permission characters:
- `r` - read (browse, download)
- `w` - write (upload, edit, rename, create)
- `d` - delete (delete, move to recycle bin)
- `m` - move (cut & paste)

Example: `-v ~/documents:docs:rwd,admin -v ~/public:public:r,all`

Windows drive letters and UNC paths work too; quote the whole argument when the path contains spaces:

```bat
netdisk -v "C:\Users\me\Documents:docs:rwdm,admin" ^
         -v "D:\Media:media:r,user" ^
         -v "\\server\share:net:r,user"
```

The parser splits from the right at the last two colons, so the drive-letter colon is not mistaken for a separator.

### Authentication and Sessions

- **HTTP Basic Auth**: all requests require Basic auth by default
- **Session management**: a session cookie (`cd_session`) is created after successful auth, with a configurable expiry (default 168 hours)
- **Password safety**: uses `crypto/subtle.ConstantTimeCompare` to prevent timing attacks
- **Share links**: `/share/<token>` requires no auth and supports an expiry time

### File Management API

| Endpoint | Method | Description |
|------|------|------|
| `/_actions/upload/{path}` | POST | File upload (multiple files + folders, streamed) |
| `/_actions/new_file` | POST | Create file |
| `/_actions/new_directory` | POST | Create directory |
| `/_actions/rename` | POST | Rename |
| `/_actions/delete` | POST | Delete (move to recycle bin) |
| `/_actions/batch_delete` | POST | Batch delete |
| `/_actions/set_clipboard` | POST | Set clipboard (copy/cut) |
| `/_actions/paste` | POST | Paste |
| `/_actions/zip_download` | POST | Batch ZIP download |
| `/_actions/properties` | POST | File properties (recursive) |
| `/_actions/upload_check` | POST | Check which upload target files/folders already exist (conflict prompt) |
| `/api/search?q=&dir=` | GET | File search (3 levels deep) |
| `/api/share` | GET | Get the current user's active share for `?path=` |
| `/api/share` | POST | Create share link |
| `/api/share/{token}` | DELETE | Delete share link |
| `/api/shares` | GET | List all share links |
| `/api/logs` | GET | View activity log (optional `?category=`) |
| `/api/logs` | DELETE | Clear log |
| `/api/logs/delete` | POST | Delete the log file |
| `/api/recycle` | GET | List restorable items in the recycle bin |
| `/api/recycle/restore` | POST | Restore items to their original paths |
| `/api/recycle/delete` | POST | Permanently delete items |
| `/api/recycle/clear` | POST | Empty the recycle bin |
| `/preview/{path}` | GET | File preview (Markdown/code) |
| `/edit/{path}` | GET/POST | Built-in editor (read/save text & code) |
| `/_actions/subtitles/{path}` | GET | List video subtitles (sidecar + embedded) |
| `/_actions/subtitle/{path}?stream=N` | GET | Extract an embedded subtitle track as WebVTT (requires ffmpeg) |
| `/login` | GET/POST | Login page / authentication |
| `/logout` | GET | Clear session |
| `/share/{token}` | GET | Public share view (no auth; file, code/Markdown preview, or ZIP for a folder) |
| `/static/{file}` | GET | Embedded static assets (JS/CSS) |

### WebClip API

| Endpoint | Method | Description |
|------|------|------|
| `/wc/` | GET | WebClip page |
| `/wc/api/cards` | GET | List cards |
| `/wc/api/cards` | POST | Add cards (files/images and/or plain text) |
| `/wc/api/cards/{id}` | DELETE | Delete a card (and its file) |
| `/wc/api/cards/clear` | POST | Clear all cards |
| `/wc/file/{filename}` | GET | Serve a stored WebClip file |

### WebDAV Service

- Based on `golang.org/x/net/webdav`
- Each volume has its own WebDAV endpoint: `/wd/<volume>/`
- Supports all WebDAV methods: PROPFIND, PROPPATCH, MKCOL, PUT, DELETE, COPY, MOVE, LOCK, UNLOCK
- Basic auth + volume-level permission control
- Media access is logged automatically

### Built-in Media Player

The frontend embeds a complete audio/video player:
- **Audio playback**: bottom dock bar with play/pause, previous/next, progress bar, speed, shuffle
- **Video playback**: fullscreen overlay with subtitle tracks (SRT→VTT auto conversion), filename shown while paused, fullscreen
- **Subtitles**: white text with a black outline and no background box, automatically scaled to the video size. Prefers sidecar subtitles with the same name (`.srt`/`.vtt`/`.ass`/`.ssa`/`.sub`; SRT/VTT are converted in the browser); if `ffmpeg`/`ffprobe` are installed, it also detects and extracts **embedded subtitle tracks** from MKV/MP4 containers (converted to WebVTT). Without ffmpeg, only sidecar subtitles are used.
- **Playlist**: generated automatically from the current directory, including audio and video files
- **Keyboard shortcuts**: case-separated design — lowercase controls files, uppercase controls media; see below

### WebClip

Path: `/wc/`
- Upload files/images to the WebClip space
- Plain-text clips supported
- Card-based display with delete and clear
- Files are served at `/wc/file/{filename}`

### File Preview

- **Markdown**: rendered by marked.js with syntax highlighting
- **Code files**: highlight.js syntax highlighting (supports .txt/.md/.json/.xml/.html/.css/.js/.py/.sh/.conf/.ini/.log/.cfg)
- **Images**: Windows-style lightbox with a large top image + bottom thumbnail strip, slideshow, keyboard delete
- **Video**: video player overlay with subtitles, filename while paused, h/l seek, H/L large-step seek, f fullscreen

### Streaming ZIP (zipwriter.go)

Streaming writes based on the standard library `archive/zip`:
- File contents are streamed to the HTTP response, **never read into memory as a whole**, with no disk temp files
- CRC32 and ZIP64 (≥4GB files / offsets) are handled automatically by the standard library
- `AddFile` supports files and directories (directories expanded recursively)
- Used for batch downloads and share-link directory downloads

### Connection Diagnostics Log (diag.go)

Requests for uploads/downloads/WebDAV pass through `transferLogMiddleware`, to help diagnose "large file drops halfway":
- Records status code, request-body bytes received, response bytes sent, and duration
- Client disconnects are logged as `ctxErr=context canceled`; failed response writes as `writeErr=...`
- Captures handler panics with a stack trace (then re-throws as-is)
- Large-file uploads are streamed (`MultipartReader`), never through `/tmp`
- On upload failure, the half-written file is removed automatically

## Frontend Features

- **Responsive design**: desktop table view / mobile card view, adaptive
- **Theme**: CSS-variable based (dark by default)
- **File sorting**: by name/size/mtime, ascending/descending
- **Batch operations**: checkbox selection with batch delete/download/copy/cut
- **Drag-and-drop upload**: drag files and folders onto the page
- **Upload progress panel**: top-center panel with overall count/total size, live speed, progress bar and percent, plus an expandable per-file list (up to 3 concurrent uploads)
- **Name-conflict prompt**: before uploading (and when restoring) the UI detects existing destinations and asks Overwrite / Auto rename / Skip / Cancel; an uploaded folder whose name already exists asks Merge / Rename folder / Skip folder / Cancel
- **Per-row actions**: quick preview/play button plus a menu (edit, share, rename, delete)
- **File editor**: in-browser editor for text/code/Markdown files (`.txt/.md/.json/.xml/.html/.css/.js/.py/.sh/.conf/.ini/.log/.cfg`)
- **Share links**: create a public link (with an optional expiry and optional password) from any file's or folder's menu, with an inline URL field, QR code, and a Cancel Share button; links persist across restarts in `share.json`
- **Page URL / QR**: the QR button left of the breadcrumbs shows the current page URL, with copy and QR code
- **Search**: filename search within the current directory, results ranked (exact > prefix > alphabetical)
- **Hidden files**: toggle dotfiles
- **Grid view**: switchable grid layout with keyboard navigation and checkbox selection
- **Activity log viewer**: category tabs (Web / WebDAV / Media) with clear, delete-file and export
- **Recycle bin**: browse deleted items from the `+` menu with checkboxes for multi-select and footer actions Select All / Invert / Restore / Delete / Empty Bin
- **Keyboard shortcut reference**: in-app help panel listing all shortcuts
- **Toast notifications**: operation feedback, with red error messages
- **Sticky header**: the desktop file-list header stays fixed while scrolling
- **Permission-aware**: delete is disabled without delete permission, with a red error message

## Build and Deploy

```bash
# Build for the current platform -> dist/netdisk-<os>_<arch>
./build.sh

# Cross-compile all platforms -> dist/
./build.sh all

# Windows (amd64)
build.cmd

# Version defaults to the date (YY.MM.DD), or set it manually:
VERSION=26.10.18 ./build.sh all

# Run example
netdisk -p 8880 \
  -a admin:secret123 \
  -a user:userpass \
  -v ~/documents:docs:rwdm,admin \
  -v ~/public:public:r,user \
  -t 168

# Command-line flags
-p <port>        Port (default 8880)
-a <user:pass>   Add user (repeatable)
-v <path:name:perms,users>  Add volume (repeatable)
-t <hours>       Session timeout in hours (default 168, 0 = no expiry)
-m <size>        Max upload size (default 10G)
-c <path>        Config directory (default ~/.config/netdisk)
```

Note: at the end, `build.sh` asks whether to move artifacts to `~/bn/netdisk` (a single artifact is renamed to `netdisk`) and remove `dist/`; the default is N.

### Platform Notes (Windows / macOS)

- **Windows paths**: drive letters (`C:\data`), UNC paths (`\\server\share`), `~`, and `%USERPROFILE%` are supported. Paths with spaces must be quoted.
- **Cross-volume operations**: deleting to the recycle bin, restoring, and cut/paste work across Windows drives and macOS external disks. If `rename` is not possible, netdisk falls back to a recursive copy followed by removal of the source.
- **Portable filenames**: uploaded/created names are normalized for Windows rules. Invalid characters and control characters are replaced with `_`; trailing dots/spaces are removed; reserved device names such as `CON`, `NUL`, `COM1`, and `LPT1` (including `CON.txt`) are prefixed with `_`.
- **URL-safe names**: file and directory links escape each path segment, so names containing `#`, `?`, `%`, spaces, or non-ASCII characters work.
- **Path isolation**: paths are checked with `filepath.Rel`, not a string prefix. Symlinks/junctions that point outside a volume are rejected for file operations.
- **Windows hidden files**: files with the Windows hidden or system attribute are hidden by default, in addition to dotfiles.
- **macOS**: paths use the same POSIX rules as Linux. The default config directory is `~/.config/netdisk`; external disks are handled by the cross-volume fallback.
- **Unsigned binaries**: macOS Gatekeeper and Windows SmartScreen may warn about downloaded binaries. macOS users may need to right-click → Open once or remove the quarantine attribute; production deployments should sign/notarize the binaries.
- **External tools**: embedded subtitle extraction still requires `ffmpeg`/`ffprobe` on `PATH`.
- **Known platform limits**: Windows ACLs are not mapped to netdisk's `rwdm` permissions; very long paths depend on Go/Windows long-path support; case-insensitive filesystems (Windows, default APFS) and macOS Unicode normalization can make differently-spelled names refer to the same file.

## Configuration and Data

- **Config directory**: `~/.config/netdisk/`
- **Log file**: `~/.config/netdisk/netdisk.log`
- **Recycle bin**: `~/.config/netdisk/recycle/` — the original path is reconstructed from the directory layout (the recycle dir mirrors the volume-relative paths), so no separate path index/database is needed
- **WebClip data**: `~/.config/netdisk/webclip/`
- **Share links**: `~/.config/netdisk/share.json` — tokens, paths, expiry, and optional password hashes (PBKDF2-HMAC-SHA256 with a per-link random salt; plaintext passwords are never stored)

## Design Decisions

1. **Single-binary deployment**: go:embed embeds all frontend assets; no external files
2. **Multi-volume isolation**: different directories can have independent permissions and user lists for multi-tenant isolation
3. **Recycle bin**: deletes move files to the recycle directory instead of removing them, so they can be restored
4. **Streaming ZIP**: streaming writes based on `archive/zip`, avoiding whole-file buffering and temp files; large files automatically use ZIP64
5. **Sessions instead of JWT**: simple in-memory sessions, suitable for intranet use
6. **Embedded templates**: all HTML/CSS/JS is embedded in the Go source, zero dependencies after build
7. **ConstantTimeCompare**: password comparison uses a constant-time algorithm to prevent timing attacks
8. **Case-separated shortcuts**: lowercase letters control file navigation, uppercase control media playback — no conflicts, no mode switching

## Keyboard Shortcuts

### File Navigation (always active)

| Key | Action |
|---|---|
| j / k | Next / previous row |
| h | Go to parent directory |
| l / Enter | Enter directory / smart preview (dispatch by file type) |
| Space | Select / deselect current row |
| a | Select all / deselect all |
| d | Delete selected (requires delete permission) |
| s | Focus search box |
| ? | Show shortcut help |

### Media Controls (during audio/video playback)

| Key | Action |
|---|---|
| P / Space | Play / pause |
| J / K | Next / previous track |
| < / > | Slower / faster (0.5x–3.0x) |
| S | Cycle playback speed |
| - / + | Volume down / up (10% steps) |
| f | Toggle fullscreen (video) |

### Video Player Overlay

| Key | Action |
|---|---|
| h / l | Seek back 5s / forward 5s |
| H / L | Seek back 60s / forward 60s |
| ← / → | Seek back 5s / forward 5s |
| f | Toggle fullscreen |
| Escape | Close video |

### Image Viewer (Lightbox)

| Key | Action |
|---|---|
| j / k | Previous / next image |
| ← / → | Previous / next image |
| P | Slideshow play / pause (fullscreen, 3s interval) |
| d | Delete current image (requires delete permission) |
| L | Close |
| Escape | Close |

## Limitations

- No HTTPS/TLS support (intranet use only)
- Sessions are stored in memory and lost on restart
- No user registration; users must be predefined via command-line flags
- File search is limited to 3 levels deep
- WebDAV locks are in-memory (`webdav.NewMemLS`) and lost on restart
- No file versioning
- No virus scanning on upload
- No automatic log rotation
- Theme is dark-only (no light-mode toggle UI)
- No in-app audio-track switching (HTML5 `<video>` does not expose alternate audio tracks in Chrome/Firefox/Edge)
- Embedded subtitle extraction requires `ffmpeg`/`ffprobe` on the system (without them, only sidecar subtitle files are supported; ASS/SSA need ffmpeg for correct conversion)
- Windows ACLs are not used; permissions are enforced by netdisk itself
- Case-insensitive filesystems and macOS Unicode normalization can cause names that differ only by case/normalization to refer to the same file
- Windows long paths are handled by Go where possible, but extremely long paths are not tested end-to-end

## License

This project is licensed under the [MIT License](LICENSE).

Third-party components bundled with this project (marked, highlight.js, Lucide, golang.org/x/net) and their full license texts are available in [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md).
