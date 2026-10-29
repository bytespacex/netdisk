package main

import (
	"encoding/json"
	"html/template"
	"strings"
	"time"
)

var funcMap = template.FuncMap{
	"filesizeformat": func(size int64) string {
		return formatFileSize(size)
	},
	"tojson": func(v interface{}) template.JS {
		b, _ := json.Marshal(v)
		return template.JS(b)
	},
	"contains": func(s, substr string) bool {
		return strings.Contains(s, substr)
	},
	"isAudioExt": func(name string) bool {
		return isAudioExt(name)
	},
	"isVideoExt": func(name string) bool {
		return isVideoExt(name)
	},
	"isImageExt": func(name string) bool {
		return isImageExt(name)
	},
	"isTextExt": func(name string) bool {
		return isEditableExt(name)
	},
	"jsquote": func(s string) string {
		return strings.NewReplacer("\\", "\\\\", "'", "\\'", "\n", "\\n", "\r", "\\r").Replace(s)
	},
	"htmlattr": func(s string) string {
		return strings.NewReplacer("&", "&amp;", "\"", "&quot;", "<", "&lt;", ">", "&gt;").Replace(s)
	},
	"urlpath": func(s string) template.URL {
		// The path is escaped segment by segment, so mark it trusted: otherwise
		// html/template would escape the '%' characters a second time.
		return template.URL(urlPathEscape(s))
	},
	"mtime_ts": func(t interface{}) int64 {
		if tt, ok := t.(time.Time); ok {
			return tt.UnixNano()
		}
		return 0
	},
}

var (
	indexTmpl         *template.Template
	editTmpl          *template.Template
	partialTmpl       *template.Template
	webclipTmpl       *template.Template
	previewTmpl       *template.Template
	loginTmpl         *template.Template
	sharePasswordTmpl *template.Template
)

func init() {
	indexTmpl = template.Must(template.New("index").Funcs(funcMap).Parse(INDEX_HTML))
	template.Must(indexTmpl.New("partial_inner").Funcs(funcMap).Parse(FILE_TABLE_PARTIAL))
	editTmpl = template.Must(template.New("edit").Funcs(funcMap).Parse(EDIT_HTML))
	partialTmpl = template.Must(template.New("partial").Funcs(funcMap).Parse(FILE_TABLE_PARTIAL))
	webclipTmpl = template.Must(template.New("webclip").Funcs(funcMap).Parse(WEBCLIP_HTML))
	previewTmpl = template.Must(template.New("preview").Funcs(funcMap).Parse(PREVIEW_HTML))
	loginTmpl = template.Must(template.New("login").Funcs(funcMap).Parse(LOGIN_HTML))
	sharePasswordTmpl = template.Must(template.New("share-password").Funcs(funcMap).Parse(SHARE_PASSWORD_HTML))
}

const LOGIN_HTML = `<!DOCTYPE html>
<html lang="en" data-theme="dark">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>netdisk - Login</title>
<script src="/static/lucide.min.js"></script>
<style>
:root {
  --main-bg: rgba(239, 241, 245, 0.72);
  --secondary-bg: #eff1f5;
  --text-color: #4c4f69;
  --border-color: #ccd0da;
  --link-color: #1e66f5;
  --glass-bg: rgba(239, 241, 245, 0.55);
  --glass-border: rgba(76, 79, 105, 0.1);
  --glass-shadow: 0 8px 32px rgba(0, 0, 0, 0.08);
  --accent: #1e66f5;
  --text-muted: #6c6f85;
  --danger: #d20f39;
}
[data-theme="dark"] {
  --main-bg: rgba(30, 30, 46, 0.72);
  --secondary-bg: #1e1e2e;
  --text-color: #cdd6f4;
  --border-color: #45475a;
  --link-color: #89b4fa;
  --glass-bg: rgba(30, 30, 46, 0.55);
  --glass-border: rgba(205, 214, 244, 0.08);
  --glass-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
  --accent: #89b4fa;
  --text-muted: #a6adc8;
  --danger: #f38ba8;
}
* { box-sizing: border-box; }
body {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
  margin: 0;
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--secondary-bg);
  color: var(--text-color);
}
.login-card {
  width: 100%;
  max-width: 380px;
  padding: 40px 32px;
  background: var(--glass-bg);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border: 1px solid var(--glass-border);
  border-radius: 16px;
  box-shadow: var(--glass-shadow);
}
.login-logo {
  text-align: center;
  margin-bottom: 32px;
}
.login-logo i {
  width: 48px;
  height: 48px;
  color: var(--accent);
  margin-bottom: 12px;
}
.login-logo h1 {
  font-size: 1.5em;
  font-weight: 600;
  margin: 0;
  color: var(--text-color);
}
.login-logo p {
  font-size: 0.85em;
  color: var(--text-muted);
  margin: 4px 0 0;
}
.form-group {
  margin-bottom: 16px;
}
.form-group label {
  display: block;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-muted);
  margin-bottom: 6px;
}
.form-group input {
  width: 100%;
  height: 42px;
  padding: 0 14px;
  border: 1px solid var(--border-color);
  background: rgba(0, 0, 0, 0.05);
  color: var(--text-color);
  border-radius: 10px;
  font-size: 14px;
  outline: none;
  transition: border-color .15s, box-shadow .15s;
}
[data-theme="dark"] .form-group input {
  background: rgba(0, 0, 0, 0.2);
}
.form-group input:focus {
  border-color: var(--accent);
  box-shadow: 0 0 0 3px rgba(137, 180, 250, 0.15);
}
.form-group input::placeholder {
  color: var(--text-muted);
}
.login-btn {
  width: 100%;
  height: 42px;
  border: none;
  background: var(--accent);
  color: #fff;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: opacity .15s;
  margin-top: 8px;
}
.login-btn:hover { opacity: 0.85; }
.login-btn:active { opacity: 0.7; }
.error-msg {
  background: rgba(var(--danger-rgb, 210, 15, 57), 0.12);
  color: var(--danger);
  padding: 10px 14px;
  border-radius: 8px;
  font-size: 13px;
  margin-bottom: 16px;
  text-align: center;
}
[data-theme="dark"] .error-msg {
  background: rgba(243, 139, 168, 0.12);
}
</style>
</head>
<body>
<div class="login-card">
  <div class="login-logo">
    <i data-lucide="hard-drive"></i>
    <h1>netdisk</h1>
    <p>Sign in to continue</p>
  </div>
  {{if .Error}}
  <div class="error-msg">{{.Error}}</div>
  {{end}}
  <form method="POST" action="/login">
    <input type="hidden" name="next" value="{{.Next}}">
    <div class="form-group">
      <label for="username">Username</label>
      <input type="text" id="username" name="username" placeholder="Enter username" autocomplete="username" required autofocus>
    </div>
    <div class="form-group">
      <label for="password">Password</label>
      <input type="password" id="password" name="password" placeholder="Enter password" autocomplete="current-password" required>
    </div>
    <button type="submit" class="login-btn">Sign In</button>
  </form>
</div>
<script>lucide.createIcons();</script>
</body>
</html>`

const SHARE_PASSWORD_HTML = `<!DOCTYPE html>
<html lang="en" data-theme="dark">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>Password required</title>
<style>
:root{--bg:#181825;--card:#1e1e2e;--text:#cdd6f4;--muted:#a6adc8;--border:#45475a;--accent:#89b4fa;--danger:#f38ba8}
*{box-sizing:border-box}
body{font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,sans-serif;margin:0;min-height:100vh;display:flex;align-items:center;justify-content:center;padding:20px;background:radial-gradient(circle at 50% 0%,#313244 0%,var(--bg) 60%);color:var(--text)}
.card{width:100%;max-width:380px;background:var(--card);border:1px solid var(--border);border-radius:12px;padding:28px;box-shadow:0 20px 60px rgba(0,0,0,.45)}
.lock{width:44px;height:44px;margin:0 auto 16px;display:flex;align-items:center;justify-content:center;border-radius:50%;background:rgba(137,180,250,.12);color:var(--accent)}
.lock svg{width:22px;height:22px}
h1{margin:0 0 6px;font-size:1.15em;text-align:center}
.file{margin:0 0 20px;text-align:center;color:var(--muted);font-size:0.85em;word-break:break-all}
label{display:block;margin-bottom:6px;font-size:0.8em;color:var(--muted)}
input{width:100%;height:42px;padding:0 12px;border:1px solid var(--border);border-radius:8px;background:#181825;color:var(--text);font-size:14px;outline:none;transition:border-color .15s}
input:focus{border-color:var(--accent)}
button{width:100%;height:42px;margin-top:16px;border:none;border-radius:8px;background:var(--accent);color:#11111b;font-size:14px;font-weight:600;cursor:pointer;transition:opacity .15s}
button:hover{opacity:.88}
.error{margin:0 0 14px;padding:9px 12px;border:1px solid var(--danger);border-radius:8px;color:var(--danger);background:rgba(243,139,168,.08);font-size:13px;text-align:center}
</style>
</head>
<body>
<div class="card">
<div class="lock"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="11" width="18" height="11" rx="2"></rect><path d="M7 11V7a5 5 0 0 1 10 0v4"></path></svg></div>
<h1>Password required</h1>
<div class="file">{{.Filename}}</div>
{{if .Error}}<div class="error">{{.Error}}</div>{{end}}
<form method="POST" action="/share/{{.Token}}">
<label for="password">Password</label>
<input type="password" id="password" name="password" placeholder="Enter share password" autocomplete="current-password" required autofocus>
<button type="submit">Unlock</button>
</form>
</div>
</body>
</html>`

const EDIT_HTML = `<!DOCTYPE html>
<html lang="en" data-theme="dark">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>{{.Path}} - Edit</title>
<script src="/static/lucide.min.js"></script>
<style>
:root{--bg:#1e1e2e;--text:#cdd6f4;--border:#45475a;--code-bg:#181825}
body{font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,sans-serif;margin:0;padding:20px;background:var(--bg);color:var(--text)}
.container{width:96%;max-width:1800px;margin:0 auto;transition:max-width .2s}
.container.narrow{max-width:900px}
.header{display:flex;justify-content:space-between;align-items:center;padding-bottom:10px;border-bottom:1px solid var(--border);position:sticky;top:0;z-index:10;background:var(--bg);padding-top:10px}
.header-actions{display:flex;gap:8px;align-items:center}
.header-actions button,.header-actions a{height:28px;padding:0 12px;border:1px solid #89b4fa;color:#89b4fa;background:none;border-radius:6px;cursor:pointer;font-size:0.85em;text-decoration:none;transition:all .15s;display:inline-flex;align-items:center;justify-content:center;gap:5px;line-height:1}
.header-actions button:hover,.header-actions a:hover{background:rgba(137,180,250,0.15)}
.header-actions .btn-save{border-color:#89b4fa;color:#89b4fa}
.header-actions .btn-save:hover{background:rgba(137,180,250,0.15)}
.filename{font-size:1.1em;font-weight:600;display:flex;align-items:center;gap:6px}
.filename i{opacity:.5}
textarea{width:100%;height:calc(100vh - 140px);padding:14px;border:1px solid rgba(255,255,255,0.1);background:rgba(0,0,0,0.2);color:var(--text);border-radius:8px;font-family:'SF Mono',Monaco,Consolas,monospace;font-size:14px;line-height:1.6;resize:none;outline:none;transition:border-color .15s,box-shadow .15s;box-sizing:border-box}
textarea:focus{border-color:#89b4fa;box-shadow:0 0 0 3px rgba(137,180,250,0.15)}
@media(max-width:768px){
.header{flex-direction:column;align-items:flex-start;gap:8px}
.header-actions{width:100%;justify-content:flex-end}
}
</style>
</head>
<body>
<div class="container" id="edit-container">
<div class="header">
<div class="filename"><i data-lucide="file-text" style="width:18px;height:18px"></i>{{.Path}}</div>
<div class="header-actions">
<button class="toggle-btn" onclick="toggleWidth()" title="Toggle width"><i data-lucide="maximize-2" style="width:14px;height:14px"></i></button>
<button class="btn-save" type="submit" form="edit-form"><i data-lucide="save" style="width:14px;height:14px"></i> Save</button>
{{if .Referrer}}<a href="{{.Referrer}}"><i data-lucide="x" style="width:14px;height:14px"></i> Cancel</a>{{else}}<a href="/"><i data-lucide="x" style="width:14px;height:14px"></i> Cancel</a>{{end}}
</div>
</div>
<form id="edit-form" method="post">
<textarea name="content">{{.Content}}</textarea>
</form>
</div>
<script>
lucide.createIcons();
function toggleWidth(){
  var c=document.getElementById('edit-container');
  c.classList.toggle('narrow');
  var isNarrow=c.classList.contains('narrow');
  localStorage.setItem('cd_edit_width',isNarrow?'narrow':'wide');
  var btn=document.querySelector('.toggle-btn i');
  if(btn){btn.setAttribute('data-lucide',isNarrow?'minimize-2':'maximize-2');lucide.createIcons();}
}
if(localStorage.getItem('cd_edit_width')==='narrow'){document.getElementById('edit-container').classList.add('narrow');var btn=document.querySelector('.toggle-btn i');if(btn){btn.setAttribute('data-lucide','minimize-2');lucide.createIcons();}}
</script>
</body>
</html>`

const FILE_TABLE_PARTIAL = `<thead>
<tr class="desktop-row">
<th class="col-check"><input type="checkbox" id="select-all-cb"></th>
<th class="col-name"><a href="javascript:void(0)" onclick="sortBy('name');return false">Name</a></th>
<th class="col-size th-size"><a href="javascript:void(0)" onclick="sortBy('size');return false">Size</a></th>
<th class="col-modified th-modified"><a href="javascript:void(0)" onclick="sortBy('mtime');return false">Modified</a></th>
<th class="col-actions"></th>
</tr>
</thead>
<tbody>
{{range .Items}}
<tr class="desktop-row" data-path="{{.Path | htmlattr}}" data-size="{{.Size}}" data-mtime="{{.Mtime | mtime_ts}}" data-name="{{.Name | htmlattr}}" data-isdir="{{.IsDir}}">
<td class="col-check"><input type="checkbox" class="file-cb" value="{{.Path | htmlattr}}"></td>
<td class="col-name">
<a href="/{{.Path | urlpath}}" class="file-name">
{{if .IsDir}}<i data-lucide="folder" class="file-icon folder-icon"></i>{{else}}<i data-lucide="file" class="file-icon"></i>{{end}}
{{if .IsDir}}[{{.Name}}]{{else}}{{.Name}}{{end}}
</a>
</td>
<td class="col-size"><span class="m-meta-item">
{{if .IsDir}}--{{else}}{{filesizeformat .Size}}{{end}}
</span></td>
<td class="col-modified"><span class="m-meta-item m-date-text">{{.FormattedMtime}}</span></td>
<td class="col-actions">
<div class="view-btn-slot">{{if not .IsDir}}{{if isImageExt .Name}}<button class="action-btn view-btn" onclick="handleSmartView(event, '{{.Path | jsquote}}', '{{.Name | jsquote}}')" title="Preview"><i data-lucide="image"></i></button>{{end}}{{end}}{{if not .IsDir}}{{if or (isAudioExt .Name) (isVideoExt .Name)}}<button class="action-btn view-btn" onclick="handleSmartView(event, '{{.Path | jsquote}}', '{{.Name | jsquote}}')" title="Play"><i data-lucide="play"></i></button>{{end}}{{end}}{{if not .IsDir}}{{if isTextExt .Name}}<button class="action-btn view-btn" onclick="handleSmartView(event, '{{.Path | jsquote}}', '{{.Name | jsquote}}')" title="Preview"><i data-lucide="eye"></i></button>{{end}}{{end}}</div>
<div class="dropdown row-dropdown">
<button class="action-btn more-btn" onclick="event.stopPropagation();toggleDropdown(this)" title="More"><i data-lucide="more-vertical"></i></button>
<div class="dropdown-menu">
              {{if .CanEdit}}<a href="/edit/{{.Path | urlpath}}" class="dropdown-item" title="Edit"><i data-lucide="pencil"></i> Edit</a>{{end}}
              <button class="dropdown-item" onclick="shareFile(event, '{{.Path | jsquote}}')" title="Share"><i data-lucide="share-2"></i> Share</button>
              {{if .CanRename}}<button class="dropdown-item" onclick="renameItem(event, '{{.Path | jsquote}}', '{{.Name | jsquote}}')" title="Rename"><i data-lucide="edit-3"></i> Rename</button>{{end}}
              {{if .CanDelete}}<button class="dropdown-item delete-action" onclick="deleteItem(event, '{{.Path | jsquote}}','{{.Name | jsquote}}')" title="Delete"><i data-lucide="trash-2"></i> Delete</button>{{end}}
</div>
</div>
</td>
</tr>
{{else}}
<tr><td colspan="5" class="empty-state"><div class="empty-icon"><i data-lucide="folder-open"></i></div><div class="empty-text">This folder is empty</div></td></tr>
{{end}}
</tbody>`

const INDEX_HTML = `<!DOCTYPE html>
<html lang="en" data-theme="dark">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0, user-scalable=no">
<title>netdisk</title>
<script src="/static/lucide.min.js"></script>
<script src="/static/qrcode.min.js"></script>
<style>
:root {
  --main-bg: rgba(239, 241, 245, 0.72);
  --secondary-bg: #eff1f5;
  --text-color: #4c4f69;
  --border-color: #ccd0da;
  --link-color: #1e66f5;
  --hover-bg: rgba(76, 79, 105, 0.06);
  --player-dock-height: 120px;
  --shadow-color: rgba(0, 0, 0, 0.08);
  --button-bg: #ccd0da;
  --track-color: #1e66f5;
  --track-buffer-color: rgba(30, 102, 245, 0.3);
  --playing-color: #1e66f5;
  --text-muted: #6c6f85;
  --glass-bg: rgba(239, 241, 245, 0.55);
  --glass-border: rgba(76, 79, 105, 0.1);
  --glass-shadow: 0 8px 32px rgba(0, 0, 0, 0.08);
  --accent: #1e66f5;
  --accent-rgb: 30, 102, 245;
  --danger: #d20f39;
  --danger-rgb: 210, 15, 57;
  --green: #40a02b;
}
[data-theme="dark"] {
  --main-bg: rgba(30, 30, 46, 0.72);
  --secondary-bg: #1e1e2e;
  --text-color: #cdd6f4;
  --border-color: #45475a;
  --link-color: #89b4fa;
  --hover-bg: rgba(205, 214, 244, 0.08);
  --header-bg: rgba(30, 30, 46, 0.72);
  --shadow-color: rgba(0, 0, 0, 0.4);
  --button-bg: #313244;
  --track-color: #89b4fa;
  --track-buffer-color: rgba(137, 180, 250, 0.3);
  --playing-color: #89b4fa;
  --text-muted: #a6adc8;
  --glass-bg: rgba(30, 30, 46, 0.55);
  --glass-border: rgba(205, 214, 244, 0.08);
  --glass-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
  --accent: #89b4fa;
  --accent-rgb: 137, 180, 250;
  --danger: #f38ba8;
  --danger-rgb: 243, 139, 168;
  --green: #a6e3a1;
}
* { box-sizing: border-box; }
body {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
  margin: 0;
  padding-bottom: var(--player-dock-height);
  background: var(--secondary-bg);
  color: var(--text-color);
}
.container {
  width: 96%;
  max-width: 1800px;
  margin: 20px auto;
  padding: 20px 24px 20px 24px;
  background: var(--glass-bg);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  box-shadow: var(--glass-shadow);
  border-radius: 10px;
  border: 1px solid var(--glass-border);
}

/* ===== TOPBAR ===== */
.topbar {
  display: flex;
  align-items: center;
  gap: 8px 6px;
  margin-bottom: 0;
  padding-bottom: 10px;
  flex-wrap: wrap;
  position: sticky;
  top: 0;
  z-index: 1000;
  background: var(--glass-bg);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
}
.breadcrumbs {
  display: flex;
  align-items: center;
  gap: 2px;
  font-size: 0.95rem;
  min-width: 0;
  flex-shrink: 1;
  overflow: hidden;
  white-space: nowrap;
  padding: 0 10px;
  height: 34px;
  background: rgba(var(--accent-rgb),0.2);
  border-radius: 6px;
}
.breadcrumbs a {
  text-decoration: none;
  color: var(--text-color);
  padding: 4px 6px;
  border-radius: 6px;
  transition: background .15s;
}
.breadcrumbs a:hover { background: var(--hover-bg); }
.breadcrumbs .sep { color: var(--text-muted); margin: 0 1px; }
.topbar-icon-btn {
  width: 34px;
  height: 34px;
  flex: 0 0 34px;
  padding: 0;
  border: 1px solid transparent;
  background: rgba(var(--accent-rgb),0.2);
  color: var(--text-color);
  border-radius: 6px;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: background .15s;
}
.topbar-icon-btn:hover { background: rgba(var(--accent-rgb),0.35); }
.topbar-icon-btn svg { width: 18px; height: 18px; }

/* ===== SEARCH ===== */
.search-box {
  min-width: 100px;
  max-width: 120px;
}
.search-input-wrap {
  position: relative;
  display: flex;
  align-items: center;
}
.search-input-wrap input {
  width: 100%;
  height: 34px;
  padding: 0 36px 0 12px;
  border: 1px solid var(--border-color);
  background: var(--hover-bg);
  color: var(--text-color);
  border-radius: 8px;
  font-size: 13px;
  outline: none;
  transition: border-color .15s;
}
.search-input-wrap input:focus { border-color: var(--link-color); }
.search-badge {
  position: absolute;
  right: 6px;
  display: none;
  align-items: center;
  gap: 3px;
  cursor: pointer;
  padding: 2px 6px;
  border-radius: 4px;
  transition: background .15s;
}
.search-badge.active { display: flex; }
.search-badge:hover { background: rgba(255,255,255,0.1); }
.search-badge .search-count { color: var(--link-color); font-weight: 600; font-size: 12px; }
.search-badge .search-close { color: var(--danger); font-weight: 700; font-size: 14px; }

/* ===== HEADER INFO ===== */
.header-info {
  font-size: 0.75rem;
  color: var(--text-muted);
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
  white-space: nowrap;
  padding: 0 10px;
  height: 34px;
  background: rgba(var(--accent-rgb),0.2);
  border-radius: 6px;
  margin-left: auto;
}
.header-info a { color: var(--text-muted); text-decoration: none; }
.header-info a:hover { color: var(--link-color); }

/* ===== + NEW BUTTON ===== */
.new-btn-wrapper { position: relative; display: inline-flex; align-items: center; gap: 4px; }
.btn-new {
  width: 34px;
  height: 34px;
  padding: 0;
  border: 1px solid var(--track-color);
  background: var(--track-color);
  color: #fff;
  border-radius: 8px;
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all .15s;
}
.btn-new:hover { opacity: 0.85; }
.btn-new i { width: 16px; height: 16px; }

/* ===== TOOLBAR ===== */
.toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 14px;
  flex-wrap: wrap;
  min-height: 36px;
}
.toolbar-group { display: flex; align-items: center; gap: 8px; }
.toolbar-btn {
  padding: 6px 10px;
  border: 1px solid var(--border-color);
  background: var(--button-bg);
  color: var(--text-color);
  border-radius: 8px;
  cursor: pointer;
  font-size: 13px;
  display: flex;
  align-items: center;
  gap: 4px;
  transition: all .15s;
}
.toolbar-btn:hover { background: var(--hover-bg); }
.toolbar-btn.danger:hover { background: var(--danger); color: #fff; border-color: var(--danger); }
.toolbar-btn i { width: 14px; height: 14px; }

/* ===== BATCH BAR ===== */
.select-btn {
  display: flex;
  width: auto;
  height: 34px;
  padding: 0 12px;
  gap: 4px;
  border: 1px solid var(--track-color);
  background: var(--track-color);
  color: #fff;
  border-radius: 8px;
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  align-items: center;
  justify-content: center;
  transition: all .15s;
}
.select-btn:hover { opacity: 0.85; }
#select-count { font-size: 13px; }
@keyframes slideDown {
  from { opacity: 0; transform: translateY(-8px); }
  to { opacity: 1; transform: translateY(0); }
}

/* ===== FILE TABLE ===== */
.file-table {
  width: 100%;
  border-collapse: collapse;
  table-layout: fixed;
}
.file-table th {
  padding: 4px 12px;
  text-align: left;
  border-bottom: 1px solid var(--border-color);
  font-size: 12px;
  font-weight: 600;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.03em;
  user-select: none;
  position: sticky;
  top: calc(var(--topbar-h, 0px) + var(--batch-h, 0px));
  z-index: 10;
  background: var(--secondary-bg);
}
.th-size, .th-modified { text-align: right !important; }
.file-table th a {
  text-decoration: none;
  color: inherit;
}
.file-table th a:hover { color: var(--link-color); }
.col-check { width: 36px; text-align: center; padding-left: 4px !important; padding-right: 4px !important; }
.col-name { width: auto; min-width: 0; overflow: hidden; }
.col-size { width: 120px; text-align: right; white-space: nowrap; }
.col-modified { width: 180px; text-align: right; white-space: nowrap; padding-right: 12px !important; }
.col-actions { width: 80px; white-space: nowrap; text-align: right; padding-right: 12px !important; }
.view-btn-slot { width: 32px; display: inline-flex; justify-content: center; flex-shrink: 0; }
.row-dropdown { position: relative; display: inline-flex; justify-content: flex-end; }
.col-name-content { display: flex; align-items: center; gap: 8px; }

.file-table tbody tr {
  transition: background-color .15s, padding-left .15s;
  border-bottom: 1px solid rgba(128,128,128,0.08);
}
.file-table tbody tr:hover { background-color: var(--hover-bg); }
.file-table tbody tr:hover td:first-child { padding-left: 16px; }
.file-table td { padding: 3px 12px; vertical-align: middle; font-size: 14px; }

.file-table input[type='checkbox'] { width: 16px; height: 16px; flex-shrink: 0; cursor: pointer; accent-color: var(--track-color); }

.file-icon { width: 16px; height: 16px; flex-shrink: 0; opacity: 0.6; }
.folder-icon { color: var(--link-color); opacity: 0.8; }

.file-name {
  color: var(--text-color);
  text-decoration: none;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  transition: color .15s;
}
.file-name:hover { color: var(--link-color); }
.file-name.playing { color: var(--playing-color); font-weight: 600; }

.m-meta-item { font-size: 12px; color: var(--text-muted); }
.m-date-text { font-family: monospace; font-size: 12px; }

/* ===== ACTION BUTTONS ===== */
.action-btn {
  background: none;
  border: none;
  cursor: pointer;
  padding: 4px;
  color: var(--text-muted);
  border-radius: 6px;
  transition: all .15s;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}
.action-btn:hover { background: rgba(var(--accent-rgb), 0.2) !important; color: #fff; }
  .action-btn i { width: 12px; height: 12px; }
.view-btn { transition: background .15s, color .15s; }

/* ===== DROPDOWN ===== */
.dropdown { position: relative; display: inline-block; }
.dropdown-menu {
  display: none;
  position: absolute;
  top: 100%;
  right: 0;
  margin-top: 4px;
  min-width: 160px;
  background: var(--glass-bg);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  border: 1px solid rgba(255,255,255,0.15) !important;
  border-radius: 10px;
  box-shadow: 0 10px 25px -5px rgba(0,0,0,0.4), 0 8px 10px -6px rgba(0,0,0,0.4) !important;
  z-index: 9999 !important;
  padding: 2px 0;
  animation: dropdownFade .12s ease;
}
.dropdown-menu::before {
  content: "";
  position: absolute;
  top: -10px;
  left: 0;
  right: 0;
  height: 10px;
  background: transparent;
}
@keyframes dropdownFade {
  from { opacity: 0; transform: translateY(-4px); }
  to { opacity: 1; transform: translateY(0); }
}
.dropdown.open .dropdown-menu { display: block; }
.dropdown:hover .dropdown-menu { display: block; }

/* ===== ROW DROPDOWN: horizontal icon toolbar ===== */
.row-dropdown .dropdown-menu {
  display: none !important;
  flex-direction: row !important;
  min-width: 0 !important;
  width: max-content !important;
  padding: 0 4px !important;
  gap: 2px !important;
  position: absolute !important;
  right: 42px !important;
  top: 0 !important;
  bottom: 0 !important;
  margin: auto 0 !important;
  background: var(--glass-bg) !important;
  border: 1px solid var(--glass-border) !important;
  border-radius: 8px !important;
  box-shadow: var(--glass-shadow) !important;
  z-index: 100 !important;
  backdrop-filter: blur(12px) !important;
  align-items: center !important;
}
.row-dropdown:hover .dropdown-menu,
.row-dropdown.open .dropdown-menu,
.row-dropdown.hover-active .dropdown-menu {
  display: flex !important;
}
/* Invisible bridge: covers the gap between buttons and menu */
.row-dropdown .dropdown-menu::before {
  content: "";
  position: absolute !important;
  right: -30px !important;
  top: -8px !important;
  bottom: -8px !important;
  width: 36px !important;
  background: transparent !important;
  display: block !important;
}
.row-dropdown .dropdown-item {
  width: 28px !important;
  height: 28px !important;
  padding: 0 !important;
  justify-content: center !important;
  font-size: 0 !important;
  flex: 0 0 auto !important;
  border-radius: 6px !important;
  display: flex !important;
  align-items: center !important;
}
.row-dropdown .dropdown-item i,
.row-dropdown .dropdown-item svg {
  width: 16px !important;
  height: 16px !important;
  margin: 0 !important;
}
.row-dropdown .dropdown-item.delete-action { color: var(--danger) !important; }
.more-btn {
  z-index: 101 !important;
  position: relative !important;
}
.dropdown-item {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  padding: 5px 12px;
  background: transparent;
  border: none;
  cursor: pointer;
  color: var(--text-color);
  font-size: 13px;
  text-decoration: none;
  transition: background .1s;
  text-align: left;
  white-space: nowrap;
  position: relative;
  z-index: 1;
}
.dropdown-item:hover {
  background: rgba(var(--accent-rgb), 0.3) !important;
  color: #fff !important;
  cursor: pointer;
}
.dropdown-item i { width: 14px; height: 14px; opacity: 0.7; }
.dropdown-item.delete-action { color: var(--danger); }
.dropdown-item.delete-action:hover {
  background: rgba(var(--danger-rgb), 0.4) !important;
  color: #fff !important;
}

/* ===== UPLOAD FLOATING PANEL ===== */
.upload-float {
  display: none;
  position: fixed;
  top: calc(var(--topbar-h, 64px) + 16px);
  left: 50%;
  transform: translateX(-50%);
  z-index: 9000;
  width: min(90vw, 460px);
  padding: 16px 20px;
  background: var(--glass-bg);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border: 1px solid var(--glass-border);
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0,0,0,0.28);
  animation: uploadFade .2s ease;
}
.upload-float.visible { display: block; }
@keyframes uploadFade {
  from { opacity: 0; }
  to { opacity: 1; }
}
.upload-float-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
  font-size: 13px;
  font-weight: 500;
}
.upload-float-header > i { width: 16px; height: 16px; color: var(--track-color); flex-shrink: 0; }
.upload-float-header .upload-count { color: var(--text-muted); font-variant-numeric: tabular-nums; }
.upload-float-header .upload-size {
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.upload-toggle {
  display: none;
  align-items: center;
  justify-content: center;
  background: none;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 2px;
  border-radius: 4px;
  line-height: 0;
  flex-shrink: 0;
  margin-left: auto;
}
.upload-toggle.show { display: inline-flex; }
.upload-toggle i { width: 16px; height: 16px; transition: transform .2s; }
.upload-toggle.open i { transform: rotate(180deg); }
.upload-row {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 12px;
  color: var(--text-muted);
}
.upload-row .upload-speed { flex-shrink: 0; white-space: nowrap; }
.upload-row .upload-percent { flex-shrink: 0; min-width: 34px; text-align: right; font-variant-numeric: tabular-nums; }
.upload-progress-track {
  height: 6px;
  background: rgba(128,128,128,0.2);
  border-radius: 3px;
  overflow: hidden;
  flex: 1;
}
.upload-progress-bar {
  height: 100%;
  background: var(--track-color);
  border-radius: 3px;
  transition: width .15s;
  width: 0%;
}
.upload-filelist {
  max-height: 0;
  overflow-y: auto;
  transition: max-height .2s ease;
  border-top: 1px solid transparent;
}
.upload-filelist.open {
  max-height: 240px;
  margin-top: 10px;
  padding-top: 8px;
  border-top-color: var(--glass-border);
}
.upload-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  padding: 5px 2px;
  font-size: 12px;
}
.upload-item-name {
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.upload-item-meta {
  display: flex;
  gap: 12px;
  color: var(--text-muted);
  font-variant-numeric: tabular-nums;
  flex-shrink: 0;
}
.upload-item.done .upload-item-name,
.upload-item.done .upload-item-meta { color: var(--text-muted); opacity: .6; }
.upload-item.error .upload-item-name { color: var(--danger); }
.upload-item.skipped .upload-item-name,
.upload-item.skipped .upload-item-meta { color: var(--text-muted); opacity: .5; text-decoration: line-through; }

/* ===== DROP OVERLAY ===== */
.drop-overlay {
  display: none;
  position: fixed;
  top: 0; left: 0; right: 0; bottom: 0;
  z-index: 9998;
  background: rgba(0,0,0,0.6);
  backdrop-filter: blur(4px);
  -webkit-backdrop-filter: blur(4px);
  justify-content: center;
  align-items: center;
  pointer-events: none;
}
.drop-overlay.active { display: flex; }
.drop-overlay-inner {
  text-align: center;
  color: #fff;
  font-size: 1.2em;
  font-weight: 500;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}
.drop-overlay-inner i { opacity: 0.8; }

/* ===== DOCK / PLAYER ===== */
.dock {
  position: fixed; left: 0; right: 0; bottom: 0; width: 100%;
  background: var(--glass-bg);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  box-shadow: 0 -4px 20px var(--shadow-color);
  z-index: 1000;
  display: none;
  flex-direction: column;
  padding: 10px 20px;
  box-sizing: border-box;
  height: var(--player-dock-height);
  border-top: 1px solid var(--glass-border);
  transform: translateZ(0);
  backface-visibility: hidden;
  will-change: transform;
}
#player-track-info { display: flex; justify-content: space-between; align-items: center; width: 100%; font-size: 0.9rem; margin-bottom: 8px; }
#current-track-name { font-weight: bold; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; padding-right: 15px; }
#time-display { flex-shrink: 0; font-family: monospace; }
.player-controls { display: flex; align-items: center; justify-content: center; gap: 10px; width: 100%; }
.player-controls button { font-size: 1.2em; background: none; border: none; cursor: pointer; width: auto; min-width: 40px; height: 44px; display: grid; place-items: center; border-radius: 6px; color: var(--text-color); padding: 0 8px; transition: none; }
.player-controls button:hover { background: rgba(var(--accent-rgb), 0.2) !important; color: #fff; }
.speed-btn { font-size: 0.9em; font-weight: bold; }
.seek-text { font-size: 0.9em; font-weight: bold; font-family: monospace; }
#audio-player { display: none; }
#progress-bar {
  -webkit-appearance: none; appearance: none;
  width: 100%; height: 6px; background: transparent;
  cursor: pointer; margin-bottom: 10px;
  background-color: rgba(128,128,128,0.3); border-radius: 3px;
}
#progress-bar::-webkit-slider-runnable-track { height: 6px; background: linear-gradient(to right, var(--track-color) 0%, var(--track-color) 0%, var(--track-buffer-color) 0%, var(--track-buffer-color) 100%); border-radius: 3px; }
#progress-bar::-moz-range-track { height: 6px; background: linear-gradient(to right, var(--track-color) 0%, var(--track-color) 0%, var(--track-buffer-color) 0%, var(--track-buffer-color) 100%); border-radius: 3px; }
#progress-bar::-webkit-slider-thumb { -webkit-appearance: none; appearance: none; margin-top: -5px; width: 16px; height: 16px; background: var(--text-color); border-radius: 50%; border: 2px solid var(--main-bg); box-shadow: 0 0 5px rgba(0,0,0,0.5); }

.extra-links { margin-top: 20px; text-align: center; }
.extra-links a { margin: 0 10px; color: var(--link-color); text-decoration: none; }

/* ===== MODAL ===== */
.modal-overlay { display:none;position:fixed;top:0;left:0;width:100%;height:100%;background:rgba(0,0,0,0.6);z-index:10002;justify-content:center;align-items:center }
.modal-overlay.active { display:flex }
.modal-box {
  background: var(--glass-bg);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border: 1px solid var(--glass-border);
  border-radius: 10px;
  padding: 24px;
  min-width: 320px;
  max-width: 90vw;
  box-shadow: 0 8px 32px rgba(0,0,0,0.4);
}
.modal-box h3 { margin:0 0 16px;font-size:1.1em }
.modal-box input[type=text] { width:100%;padding:8px 12px;border:1px solid var(--border-color);background:var(--secondary-bg);color:var(--text-color);border-radius:8px;box-sizing:border-box;margin-bottom:16px;font-size:14px }
.modal-btns { display:flex;gap:8px;justify-content:flex-end }
.modal-btns button { padding:6px 16px;border-radius:8px;border:none;cursor:pointer;font-size:14px;transition:all .15s }
.modal-btns .btn-ok { background:var(--track-color);color:#fff }
.modal-btns .btn-cancel { background:var(--button-bg);color:var(--text-color) }
.recycle-body { flex:1;overflow-y:auto;background:rgba(0,0,0,0.25);border-radius:6px }
.recycle-btn { padding:5px 12px;border:1px solid var(--border-color);background:var(--button-bg);color:var(--text-color);border-radius:6px;cursor:pointer;font-size:12px }
.recycle-ok { border-color:var(--link-color);color:var(--link-color) }
.recycle-danger { border-color:var(--danger);color:var(--danger);background:none }
.recycle-primary { border:none;background:var(--track-color);color:#fff }
.recycle-row { display:flex;align-items:center;gap:10px;padding:8px 12px;border-bottom:1px solid var(--border-color);cursor:pointer }
.recycle-row:hover { background:rgba(255,255,255,0.04) }
.recycle-check { width:16px;height:16px;flex-shrink:0;cursor:pointer;accent-color:var(--link-color) }
/* ===== SHARE / QR MODAL ===== */
.share-modal-box { width: 460px; max-width: 92vw; max-height: 90vh; overflow-y: auto; }
.share-path { margin: -8px 0 14px; font-size: 12px; color: var(--text-muted); word-break: break-all; }
.share-actions { display: flex; gap: 8px; flex-wrap: wrap; margin-bottom: 16px; }
.share-btn {
  padding: 7px 14px;
  border-radius: 8px;
  border: 1px solid var(--border-color);
  background: var(--button-bg);
  color: var(--text-color);
  cursor: pointer;
  font-size: 13px;
  transition: all .15s;
}
.share-btn:hover:not(:disabled) { opacity: .85; }
.share-btn:disabled { opacity: .4; cursor: not-allowed; }
.share-btn-primary { border-color: var(--track-color); background: var(--track-color); color: #fff; }
.share-btn-danger { border-color: var(--danger); background: none; color: var(--danger); }
.share-field { display: flex; align-items: center; gap: 8px; margin-bottom: 10px; }
.share-field label { flex: 0 0 auto; font-size: 12px; color: var(--text-muted); white-space: nowrap; }
.share-field input[type=text],
.share-field input[type=password],
.share-field input[type=number] {
  min-width: 0;
  height: 34px;
  margin: 0;
  padding: 6px 10px;
  border: 1px solid var(--border-color);
  background: var(--secondary-bg);
  color: var(--text-color);
  border-radius: 8px;
  font-size: 13px;
}
.share-field input[type=text],
.share-field input[type=password] { flex: 1; }
.share-field input[type=number] { width: 90px; }
.share-copy-btn {
  flex: 0 0 34px;
  width: 34px;
  height: 34px;
  padding: 0;
  border: 1px solid var(--border-color);
  background: var(--button-bg);
  color: var(--text-color);
  border-radius: 8px;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}
.share-copy-btn:hover { opacity: .85; }
.share-copy-btn i { width: 15px; height: 15px; }
.qr-wrap { display: none; justify-content: center; margin-top: 14px; padding: 12px; background: #fff; border-radius: 8px; }
.qr-wrap.active { display: flex; }
.qr-wrap canvas { display: block; width: 200px; height: 200px; image-rendering: pixelated; }
.share-status { margin-top: 10px; font-size: 12px; color: var(--text-muted); line-height: 1.5; word-break: break-word; }
.share-status.error { color: var(--danger); }

/* Dialogs must render above the recycle/log modals */
#modal-overlay, #choice-modal { z-index:10050 }
.choice-msg { margin:0 0 20px;font-size:14px;color:var(--text-muted);line-height:1.5 }
.choice-btns { display:flex;gap:8px;justify-content:flex-end;flex-wrap:wrap }
.choice-btns button { padding:6px 14px;border-radius:8px;cursor:pointer;font-size:13px }
.choice-primary { border:none;background:var(--track-color);color:#fff }
.choice-danger { border:1px solid var(--danger);background:none;color:var(--danger) }
.choice-plain { border:1px solid var(--border-color);background:var(--button-bg);color:var(--text-color) }

.key-row { display:flex;align-items:center;justify-content:space-between;padding:8px 0;border-bottom:1px solid var(--border-color) }
.key-row:last-child { border-bottom:none }
.key-label { font-size:13px;color:var(--text-color);flex:1 }
.key-bind { padding:4px 10px;border:1px solid var(--border-color);background:var(--hover-bg);color:var(--text-color);border-radius:6px;font-size:12px;font-family:monospace;cursor:pointer;min-width:60px;text-align:center;transition:all .15s }
.key-bind:hover { border-color:var(--link-color) }
.key-bind.recording { border-color:var(--track-color);background:rgba(var(--accent-rgb),0.15);animation:pulse 1s infinite }
@keyframes pulse { 0%,100%{opacity:1} 50%{opacity:.6} }

/* ===== LIGHTBOX ===== */
.lightbox { display:none;position:fixed;top:0;left:0;width:100%;height:100%;background:rgba(0,0,0,0.95);z-index:10001;flex-direction:column }
.lightbox.active { display:flex }
.lightbox-top { flex:1;display:flex;flex-direction:column;min-height:0;position:relative }
.lightbox-info { padding:8px 16px;color:#fff;font-size:0.85em;background:rgba(0,0,0,0.6);text-align:center;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;flex-shrink:0 }
.lightbox-main { flex:1;display:flex;align-items:center;justify-content:center;min-height:0;position:relative }
.lightbox-main img { max-height:100%;max-width:80vw;object-fit:contain;border-radius:4px }
.lightbox-nav { position:absolute;top:50%;transform:translateY(-50%);width:48px;height:80px;font-size:2em;color:#fff;background:rgba(255,255,255,0.15);border:none;border-radius:8px;cursor:pointer;z-index:5;transition:background .15s;display:flex;align-items:center;justify-content:center }
.lightbox-nav:hover { background:rgba(255,255,255,0.35) }
.lightbox-nav.prev { left:12px }
.lightbox-nav.next { right:12px }
.lightbox-close { position:absolute;top:12px;right:16px;font-size:1.8em;color:#fff;cursor:pointer;background:rgba(0,0,0,0.5);border:none;border-radius:50%;width:40px;height:40px;z-index:10;display:flex;align-items:center;justify-content:center;transition:background .15s }
.lightbox-close:hover { background:rgba(0,0,0,0.9) }
.lightbox-thumbs { height:80px;flex-shrink:0;display:flex;gap:4px;padding:8px 12px;overflow-x:auto;background:rgba(0,0,0,0.8);border-top:1px solid rgba(255,255,255,0.1);scrollbar-width:thin;scrollbar-color:rgba(255,255,255,0.2) transparent }
.lightbox-thumbs::-webkit-scrollbar { height:4px }
.lightbox-thumbs::-webkit-scrollbar-thumb { background:rgba(255,255,255,0.2);border-radius:2px }
.lightbox-thumb { width:64px;height:64px;object-fit:cover;border-radius:4px;border:2px solid transparent;cursor:pointer;flex-shrink:0;opacity:0.6;transition:all .15s }
.lightbox-thumb:hover { opacity:0.9;border-color:rgba(255,255,255,0.4) }
.lightbox-thumb.active { border-color:#89b4fa;opacity:1 }
.lightbox:fullscreen .lightbox-thumbs { display:none }
.lightbox:fullscreen .lightbox-nav { display:none }
.lightbox:fullscreen .lightbox-close { display:none }
.lightbox:fullscreen .lightbox-info { position:absolute;top:0;left:0;right:0;z-index:5;background:rgba(0,0,0,0.4);padding:12px 16px }
.lightbox:fullscreen .lightbox-main img { max-width:100vw;max-height:100vh;border-radius:0 }

/* ===== VIDEO OVERLAY ===== */
.video-overlay { display:none;position:fixed;top:0;left:0;width:100%;height:100%;background:#000;z-index:10001;justify-content:center;align-items:center }
.video-overlay.active { display:flex }
.video-overlay video { max-width:95vw;max-height:92vh;border-radius:4px }
/* Subtitles: white text with a black outline, no background box */
#video-player::cue {
  color:#fff;
  background:none;
  text-shadow:-1px -1px 0 #000,1px -1px 0 #000,-1px 1px 0 #000,1px 1px 0 #000,0 0 4px rgba(0,0,0,0.9);
  font-size:var(--subtitle-size, 1.15rem);
  font-weight:600;
}
.video-overlay-close { position:absolute;top:20px;right:24px;font-size:2.5em;color:#fff;cursor:pointer;background:rgba(0,0,0,0.5);border:none;border-radius:50%;width:48px;height:48px;z-index:10;display:flex;align-items:center;justify-content:center;transition:background .15s }
.video-overlay-close:hover { background:rgba(0,0,0,0.9) }
.video-nav{position:absolute;top:50%;transform:translateY(-50%);width:52px;height:52px;border:none;background:rgba(255,255,255,0.15);color:#fff;border-radius:50%;cursor:pointer;font-size:24px;display:flex;align-items:center;justify-content:center;transition:background .15s;z-index:10}
.video-nav:hover{background:rgba(255,255,255,0.35)}
.video-nav-prev{left:20px}
.video-nav-next{right:20px}

/* ===== SEARCH RESULTS ===== */
.search-item { padding: 8px 12px; cursor: pointer; border-bottom: 1px solid var(--border-color); transition: background .1s; }
.search-item:hover { background: var(--hover-bg); }
.search-item .si-name { font-weight: 500; }
.search-item .si-path { font-size: 0.8em; color: var(--text-muted); }

/* ===== TOAST ===== */
.toast-msg { position:fixed;top:20px;left:50%;transform:translateX(-50%);background:var(--green);color:white;padding:8px 20px;border-radius:10px;font-size:14px;z-index:10003;opacity:0;transition:opacity .3s;pointer-events:none }
.toast-msg.show { opacity:1 }
.toast-msg.error { background:var(--danger); }

/* ===== EMPTY STATE ===== */
.empty-state { text-align:center;padding:48px 16px;color:var(--text-muted);border:none!important }
.empty-icon { margin-bottom:12px;opacity:.4 }
.empty-icon i { width:48px;height:48px }
.empty-text { font-size:1.1em }

/* ===== GRID VIEW ===== */
.grid-view { display:grid;grid-template-columns:repeat(auto-fill,minmax(150px,1fr));gap:12px }
.grid-item { display:flex;flex-direction:column;align-items:center;padding:12px;border:1px solid var(--border-color);border-radius:10px;background:var(--secondary-bg);text-decoration:none;color:var(--text-color);transition:background .2s;position:relative;cursor:pointer }
.grid-item:hover { background:var(--hover-bg) }
.grid-item-focused { background: rgba(137, 180, 250, 0.15) !important; border-color: #89b4fa !important; }
.grid-item-focused .grid-name { color: #a6e3a1 !important; }
.grid-item-focused .grid-icon { color: #a6e3a1 !important; }
.grid-item .grid-check { position:absolute;top:6px;left:6px;z-index:2;display:none }
.grid-item .grid-check input { width:16px;height:16px;cursor:pointer;accent-color:var(--track-color) }
.grid-item:hover .grid-check, .grid-item.grid-checked .grid-check { display:block }
.grid-item.grid-checked { border-color: var(--accent) !important; background: rgba(var(--accent-rgb), 0.1) !important; }
.grid-item .grid-icon { font-size:2.5em;margin-bottom:8px }
.grid-item .grid-name { font-size:0.85em;text-align:center;word-break:break-all;max-height:2.4em;overflow:hidden }
.grid-item .grid-meta { font-size:0.75em;color:var(--text-muted);margin-top:4px }
.grid-item img { object-fit:cover;max-height:120px;width:100%;border-radius:6px;margin-bottom:8px }

/* ===== RESPONSIVE ===== */
@media(min-width:769px) {
  .desktop-row { display:table-row }
  .file-table thead { display:table-header-group }
}
@media(max-width:768px) {
  .container { width: 100% !important; margin: 0 !important; padding: 10px !important; border-radius: 0 !important; }

  .file-table thead { display: none !important; }

  .file-table { display: block; width: 100%; }
  .file-table tbody { display: block; width: 100%; }

  .desktop-row { display: flex !important; align-items: center !important; width: 100% !important; padding: 4px 4px !important; border-bottom: 1px solid rgba(255,255,255,0.06); background: none !important; gap: 6px !important; overflow: visible !important; }
  .desktop-row td { display: block !important; padding: 0 !important; border: none !important; background: none !important; }
  .desktop-row .col-size, .desktop-row .col-modified { display: none !important; }

  .col-check { flex: 0 0 24px !important; }

  .col-name { flex: 1 1 0 !important; min-width: 0 !important; overflow: hidden !important; }
  .col-name .file-name { display: block !important; white-space: nowrap !important; overflow: hidden !important; text-overflow: ellipsis !important; font-size: 0.9rem !important; font-weight: normal !important; color: var(--text-color); }
  .col-name .file-name.playing { color: var(--playing-color) !important; font-weight: 600 !important; }

  .col-actions { flex: 0 0 80px !important; display: flex !important; justify-content: flex-end !important; align-items: center !important; gap: 8px !important; margin-right: 0 !important; }

.action-btn i { width: 12px; height: 12px; }

  .topbar { gap: 6px; margin-bottom: 10px; flex-wrap: wrap; }
  .breadcrumbs { width: calc(100% - 40px); }
  .topbar-icon-btn { width: 34px; height: 34px; }
  .header-info { margin-left: 0; flex-shrink: 0; }
  .search-box { flex: 1; min-width: 0; }
  .search-box input { height: 34px; font-size: 14px; }
  .btn-new { width: 34px; height: 34px; }
  .row-dropdown .dropdown-menu { z-index: 15 !important; }
  .more-btn { z-index: 16 !important; }
}

/* ===== FOCUSED ROW (keyboard nav) ===== */
.file-row-focused { background: rgba(137, 180, 250, 0.15) !important; }
.file-row-focused td { color: #a6e3a1 !important; }
.file-row-focused td:first-child { padding-left: 16px; }
.file-row-focused a.file-name { color: #a6e3a1 !important; }
.file-row-focused .m-meta-item { color: #a6e3a1 !important; }
.file-table th { background: var(--secondary-bg) !important; }

/* ===== SORT DROPDOWN ===== */
.sort-item { display:flex;align-items:center;gap:8px;width:100%;padding:5px 12px;background:transparent;border:none;cursor:pointer;color:var(--text-color);font-size:13px;text-align:left;white-space:nowrap;transition:background .1s }
.sort-item:hover { background:rgba(var(--accent-rgb),0.3) !important;color:#fff !important }
.sort-item i { width:14px;height:14px;opacity:0.7 }

/* ===== SEARCH RESULTS TABLE ===== */
#search-results-area { margin-top:16px }
#search-results-area h3 { margin:0 0 10px;font-size:1em;display:flex;align-items:center;gap:8px }
#search-results-area button { font-size:0.85em }
</style>
</head>
<body>
<div id="drop-overlay" class="drop-overlay"><div class="drop-overlay-inner"><i data-lucide="upload-cloud" style="width:48px;height:48px"></i><div>Drop files or folders here</div></div></div>
<div class="container">
  <div class="topbar">
    <button type="button" class="topbar-icon-btn" onclick="showPageQR()" title="Show page URL / QR code" aria-label="Show page URL and QR code">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
        <rect x="3" y="3" width="7" height="7" rx="1"></rect>
        <rect x="14" y="3" width="7" height="7" rx="1"></rect>
        <rect x="3" y="14" width="7" height="7" rx="1"></rect>
        <path d="M14 14h3v3h-3z"></path>
        <path d="M21 14v.01M14 21v.01M17 21v.01M21 17v4h-4"></path>
      </svg>
    </button>
    <div class="breadcrumbs">
      <a href="/"><i data-lucide="home" style="width:16px;height:16px"></i></a>
      {{range .Breadcrumbs}}<span class="sep">/</span><a href="/{{.Path | urlpath}}">{{.Name}}</a>{{end}}
    </div>
    {{if .CurrentUser}}
    <div class="header-info">
      <span><i data-lucide="user" style="width:12px;height:12px;vertical-align:middle"></i> {{.CurrentUser}}</span>
      {{if .UserPerms}}<span style="{{if contains .UserPerms "w"}}color:red;{{end}}">[{{.UserPerms}}]</span>{{end}}
      <a href="/logout"><i data-lucide="log-out" style="width:14px;height:14px;vertical-align:middle"></i></a>
    </div>
    {{end}}
    {{if .CurrentPath}}
    <div class="search-box">
      <div class="search-input-wrap">
        <input type="text" id="search-input" placeholder="Search...">
        <span id="search-badge" class="search-badge" onclick="clearSearch()"><span class="search-count"></span><span class="search-close">&times;</span></span>
      </div>
    </div>
    {{end}}
    {{if .CurrentPath}}
    <div class="new-btn-wrapper dropdown">
      <button class="select-btn" id="select-btn" onclick="event.stopPropagation();this.parentElement.classList.toggle('open')">
        <span id="select-count"></span><i data-lucide="square-check"></i>
      </button>
      <div class="dropdown-menu">
        <button class="dropdown-item" id="select-toggle-btn" onclick="toggleSelectAll()"><i data-lucide="check-square"></i> <span id="select-toggle-text">Select All</span></button>
        <button class="dropdown-item" onclick="showProperties()"><i data-lucide="info"></i> Properties</button>
        <button class="dropdown-item" onclick="setClipboard('copy')"><i data-lucide="copy"></i> Copy</button>
        <button class="dropdown-item" onclick="setClipboard('move')"><i data-lucide="scissors"></i> Move</button>
        {{if .CanPaste}}<button class="dropdown-item" onclick="pasteItems()"><i data-lucide="clipboard"></i> Paste</button>{{end}}
        <button class="dropdown-item" onclick="downloadSelected()"><i data-lucide="download"></i> Download</button>
        <button class="dropdown-item delete-action" onclick="deleteSelected()"><i data-lucide="trash-2"></i> Delete</button>
      </div>
    </div>
    <div class="new-btn-wrapper dropdown">
      <button class="btn-new" id="sort-btn" onclick="event.stopPropagation();this.parentElement.classList.toggle('open')" title="Sort">
        <i data-lucide="arrow-up-down"></i>
      </button>
      <div class="dropdown-menu">
        <button class="dropdown-item sort-item" data-sort="name" onclick="sortBy('name')"><i data-lucide="type"></i> Name</button>
        <button class="dropdown-item sort-item" data-sort="size" onclick="sortBy('size')"><i data-lucide="hard-drive"></i> Size</button>
        <button class="dropdown-item sort-item" data-sort="mtime" onclick="sortBy('mtime')"><i data-lucide="clock"></i> Modified</button>
      </div>
    </div>
    <div class="new-btn-wrapper dropdown">
      <button class="btn-new">
        <i data-lucide="plus"></i>
      </button>
      <div class="dropdown-menu">
        <form id="upload-form" method="post" enctype="multipart/form-data" action="/_actions/upload/{{.CurrentPath | urlpath}}" style="margin:0">
          <input type="file" id="upload-files" name="file" multiple required style="display:none">
          <input type="file" id="upload-folder" name="file" multiple webkitdirectory required style="display:none">
          <button type="button" class="dropdown-item" onclick="document.getElementById('upload-files').click()"><i data-lucide="upload"></i> Upload Files</button>
          <button type="button" class="dropdown-item" onclick="document.getElementById('upload-folder').click()"><i data-lucide="folder-up"></i> Upload Folder</button>
        </form>
        <button class="dropdown-item" onclick="createNewFile()"><i data-lucide="file-plus"></i> New File</button>
        <button class="dropdown-item" onclick="createNewDirectory()"><i data-lucide="folder-plus"></i> New Folder</button>
        <button class="dropdown-item" onclick="toggleHidden()"><i data-lucide="{{if .ShowHidden}}eye-off{{else}}eye{{end}}"></i> {{if .ShowHidden}}Hide Dotfiles{{else}}Show Dotfiles{{end}}</button>
        <button class="dropdown-item" onclick="openKeySettings()"><i data-lucide="keyboard"></i> Keyboard</button>
        <button class="dropdown-item" onclick="toggleView()"><i id="grid-icon" data-lucide="grid-3x3"></i> <span id="grid-label">Grid View</span></button>
        <button class="dropdown-item" onclick="openRecycleBin()"><i data-lucide="history"></i> Recycle Bin</button>
        <button class="dropdown-item" onclick="openLogViewer()"><i data-lucide="scroll-text"></i> Log</button>
      </div>
    </div>
    {{end}}
  </div>

  {{if .CurrentPath}}
  {{end}}

  <div id="search-results-area" style="display:none">
    <table class="file-table" id="search-results-table">
    <thead><tr class="desktop-row">
    <th class="col-check"><input type="checkbox" id="search-select-all"></th>
    <th class="col-name"><span>Name</span></th>
    <th class="col-size"><span>Size</span></th>
    <th class="col-modified"><span>Modified</span></th>
    <th class="col-actions"></th>
    </tr></thead>
    <tbody id="search-results-body"></tbody>
    </table>
  </div>

  <table class="file-table" id="file-list-table">
  {{template "partial_inner" .}}
  </table>

  {{if not .CurrentPath}}
  <div class="extra-links"><a href="/wc/"><i data-lucide="scissors" style="width:14px;height:14px"></i> webclip</a></div>
  {{end}}
</div>

<!-- Upload floating panel -->
<div class="upload-float" id="upload-float">
  <div class="upload-float-header">
    <span class="upload-count" id="upload-count">(0/0)</span>
    <span class="upload-size" id="upload-size">0 B</span>
    <button type="button" class="upload-toggle" id="upload-toggle" onclick="toggleUploadList()" title="Show files"><i data-lucide="chevron-down"></i></button>
  </div>
  <div class="upload-row">
    <span class="upload-speed" id="upload-speed">--</span>
    <div class="upload-progress-track">
      <div class="upload-progress-bar" id="upload-progress-bar"></div>
    </div>
    <span class="upload-percent" id="upload-total-percent">0%</span>
  </div>
  <div class="upload-filelist" id="upload-filelist"></div>
</div>

<div id="player-dock" class="dock">
  <div id="player-track-info">
    <span id="current-track-name">No track selected</span>
    <span id="loading-hint">(Downloading...)</span>
    <span id="time-display">0:00 / 0:00</span>
  </div>
  <input type="range" id="progress-bar" value="0" step="0.1" class="progress-bar">
  <audio id="audio-player" preload="auto"></audio>
  <div class="player-controls">
    <button id="speed-btn" class="speed-btn">1.0x</button>
    <button id="prev-btn" title="Previous"><i data-lucide="skip-back"></i></button>
    <button id="seek-back-btn" class="seek-text" title="-10s">-10</button>
    <button id="main-play-pause-btn" title="Play/Pause"><i data-lucide="play"></i></button>
    <button id="seek-fwd-btn" class="seek-text" title="+10s">+10</button>
    <button id="next-btn" title="Next"><i data-lucide="skip-forward"></i></button>
    <button id="shuffle-btn" title="Shuffle"><i data-lucide="shuffle"></i></button>
  </div>
</div>

<div id="modal-overlay" class="modal-overlay">
  <div class="modal-box">
    <h3>Title</h3>
    <input type="text" class="modal-input">
    <div class="modal-btns">
      <button class="modal-cancel btn-cancel">Cancel</button>
      <button class="modal-ok btn-ok">OK</button>
    </div>
  </div>
</div>

<div id="key-modal" class="modal-overlay">
  <div class="modal-box" style="max-width:400px;max-height:80vh;overflow-y:auto">
    <h3 style="margin:0 0 12px;display:flex;align-items:center;gap:8px"><i data-lucide="keyboard" style="width:20px;height:20px"></i> Keyboard Shortcuts</h3>
    <div id="key-list"></div>
    <div style="margin-top:12px;display:flex;gap:8px;justify-content:flex-end">
      <button class="btn-cancel" onclick="resetKeys()" style="padding:6px 12px;border:1px solid var(--border-color);background:var(--button-bg);color:var(--text-color);border-radius:8px;cursor:pointer;font-size:13px">Reset</button>
      <button class="btn-ok" onclick="saveKeys()" style="padding:6px 12px;border:none;background:var(--track-color);color:#fff;border-radius:8px;cursor:pointer;font-size:13px">Save</button>
    </div>
  </div>
</div>

<div id="log-modal" class="modal-overlay">
  <div class="modal-box" style="width:80vw;max-width:900px;height:80vh;display:flex;flex-direction:column;overflow:hidden">
    <h3 style="margin:0 0 10px;display:flex;align-items:center;gap:8px"><i data-lucide="scroll-text" style="width:20px;height:20px"></i> Activity Log</h3>
    <div style="display:flex;gap:4px;margin-bottom:10px;flex-wrap:wrap">
      <button class="log-tab active" onclick="loadLog('all',this)" style="padding:4px 10px;border:1px solid var(--border-color);background:var(--track-color);color:#fff;border-radius:6px;cursor:pointer;font-size:12px">All</button>
      <button class="log-tab" onclick="loadLog('web',this)" style="padding:4px 10px;border:1px solid var(--border-color);background:var(--button-bg);color:var(--text-color);border-radius:6px;cursor:pointer;font-size:12px">Web</button>
      <button class="log-tab" onclick="loadLog('webdav',this)" style="padding:4px 10px;border:1px solid var(--border-color);background:var(--button-bg);color:var(--text-color);border-radius:6px;cursor:pointer;font-size:12px">WebDAV</button>
      <button class="log-tab" onclick="loadLog('media',this)" style="padding:4px 10px;border:1px solid var(--border-color);background:var(--button-bg);color:var(--text-color);border-radius:6px;cursor:pointer;font-size:12px">Media</button>
    </div>
    <div id="log-content" style="flex:1;1;overflow-y:auto;background:rgba(0,0,0,0.3);border-radius:6px;padding:10px;font-family:monospace;font-size:12px;line-height:1.6;white-space:pre-wrap;word-break:break-all"></div>
    <div style="margin-top:10px;display:flex;gap:8px;justify-content:flex-end">
      <button onclick="clearLog()" style="padding:5px 12px;border:1px solid var(--border-color);background:var(--button-bg);color:var(--text-color);border-radius:6px;cursor:pointer;font-size:12px">Clear Screen</button>
      <button onclick="deleteLogFile()" style="padding:5px 12px;border:1px solid var(--danger);background:none;color:var(--danger);border-radius:6px;cursor:pointer;font-size:12px">Delete Log</button>
      <button onclick="exportLog()" style="padding:5px 12px;border:1px solid var(--border-color);background:var(--button-bg);color:var(--text-color);border-radius:6px;cursor:pointer;font-size:12px">Export Log</button>
      <button onclick="document.getElementById('log-modal').classList.remove('active')" style="padding:5px 12px;border:none;background:var(--track-color);color:#fff;border-radius:6px;cursor:pointer;font-size:12px">Close</button>
    </div>
  </div>
</div>

<div id="recycle-modal" class="modal-overlay">
  <div class="modal-box" style="width:82vw;max-width:960px;height:80vh;display:flex;flex-direction:column;overflow:hidden">
    <h3 style="margin:0 0 10px;display:flex;align-items:center;gap:8px"><i data-lucide="history" style="width:20px;height:20px"></i> Recycle Bin</h3>
    <div id="recycle-content" class="recycle-body"></div>
    <div style="margin-top:10px;display:flex;gap:8px;align-items:center;flex-wrap:wrap">
      <button onclick="toggleRecycleAll()" class="recycle-btn">Select All</button>
      <button onclick="invertRecycleSelection()" class="recycle-btn">Invert</button>
      <span id="recycle-selcount" style="font-size:12px;color:var(--text-muted);margin-left:4px">0 selected</span>
      <span style="flex:1"></span>
      <button onclick="restoreSelectedRecycle()" class="recycle-btn recycle-ok">Restore</button>
      <button onclick="deleteSelectedRecycle()" class="recycle-btn recycle-danger">Delete</button>
      <button onclick="clearRecycleBin()" class="recycle-btn recycle-danger">Empty Bin</button>
      <button onclick="closeRecycleBin()" class="recycle-btn recycle-primary">Close</button>
    </div>
  </div>
</div>

<div id="choice-modal" class="modal-overlay">
  <div class="modal-box" style="min-width:360px;max-width:520px">
    <h3 id="choice-title">Title</h3>
    <p id="choice-msg" class="choice-msg"></p>
    <div id="choice-btns" class="choice-btns"></div>
  </div>
</div>

<div id="share-modal" class="modal-overlay">
  <div class="modal-box share-modal-box">
    <h3 id="share-title">Share</h3>
    <div class="share-path" id="share-path"></div>
    <div class="share-actions">
      <button type="button" class="share-btn share-btn-primary" id="share-generate-btn" onclick="generateShare()">Generate URL</button>
      <button type="button" class="share-btn share-btn-danger" id="share-cancel-btn" onclick="cancelShare()" disabled>Cancel Share</button>
    </div>
    <div class="share-field">
      <label for="share-url">URL (copy):</label>
      <input type="text" id="share-url" readonly placeholder="No active share link">
      <button type="button" class="share-copy-btn" onclick="copyShareURL()" title="Copy URL"><i data-lucide="copy"></i></button>
    </div>
    <div class="share-field">
      <label for="share-hours">Expire hours (0=never):</label>
      <input type="number" id="share-hours" min="0" step="1" value="24">
    </div>
    <div class="share-field">
      <label for="share-password">Password (optional):</label>
      <input type="password" id="share-password" placeholder="Leave empty for no password" autocomplete="new-password">
    </div>
    <div class="qr-wrap" id="share-qr-wrap"><canvas id="share-qr"></canvas></div>
    <div class="share-status" id="share-status"></div>
    <div class="modal-btns" style="margin-top:16px">
      <button type="button" class="btn-cancel" onclick="closeShareModal()">Close</button>
    </div>
  </div>
</div>

<div id="qr-modal" class="modal-overlay">
  <div class="modal-box share-modal-box">
    <h3>Current page URL</h3>
    <div class="share-field">
      <input type="text" id="page-url" readonly>
      <button type="button" class="share-copy-btn" onclick="copyPageURL()" title="Copy URL"><i data-lucide="copy"></i></button>
    </div>
    <div class="qr-wrap active"><canvas id="page-qr"></canvas></div>
    <div class="modal-btns" style="margin-top:16px">
      <button type="button" class="btn-cancel" onclick="closePageQR()">Close</button>
    </div>
  </div>
</div>

<div id="lightbox" class="lightbox">
  <div class="lightbox-top">
    <button class="lightbox-close" onclick="document.getElementById('lightbox').classList.remove('active')">&times;</button>
    <div class="lightbox-info" id="lightbox-info"></div>
    <div class="lightbox-main">
      <button class="lightbox-nav prev" onclick="lightboxNav(-1)">&#8249;</button>
      <img id="lightbox-img" src="" alt="">
      <button class="lightbox-nav next" onclick="lightboxNav(1)">&#8250;</button>
    </div>
  </div>
  <div class="lightbox-thumbs" id="lightbox-thumbs"></div>
</div>

<audio id="audio-player" preload="auto"></audio>
<div id="video-overlay" class="video-overlay">
  <button class="video-overlay-close" onclick="closeVideoOverlay()">&times;</button>
  <button class="video-nav video-nav-prev" onclick="playPrev()">&#8249;</button>
  <button class="video-nav video-nav-next" onclick="playNext()">&#8250;</button>
  <video id="video-player" controls preload="auto"></video>
  <div id="video-filename" style="position:absolute;top:20px;left:50%;transform:translateX(-50%);background:rgba(0,0,0,0.6);color:#fff;padding:6px 16px;border-radius:8px;font-size:14px;pointer-events:none;display:none;white-space:nowrap;z-index:20"></div>
</div>

<script>
// Init Lucide icons
lucide.createIcons();
window.__USER_PERMS__ = '{{.UserPerms}}';
window.__CAN_DELETE__ = window.__USER_PERMS__.indexOf('d') !== -1;

// Sticky header: measure topbar height for thead offset
(function(){
  var tb=document.querySelector('.topbar');
  if(!tb)return;
  function setH(){
    var h=tb.offsetHeight;
    document.documentElement.style.setProperty('--topbar-h',h+'px');
  }
  setH();
  window.addEventListener('resize',setH);
  new ResizeObserver(setH).observe(tb);
})();

const audioPlayer = document.getElementById('audio-player'),
  videoPlayer = document.getElementById('video-player'),
  mainPlayPauseBtn = document.getElementById('main-play-pause-btn'),
  prevBtn = document.getElementById('prev-btn'),
  nextBtn = document.getElementById('next-btn'),
  shuffleBtn = document.getElementById('shuffle-btn'),
  speedBtn = document.getElementById('speed-btn'),
  seekBackBtn = document.getElementById('seek-back-btn'),
  seekFwdBtn = document.getElementById('seek-fwd-btn'),
  playerDock = document.getElementById('player-dock'),
  trackNameSpan = document.getElementById('current-track-name'),
  timeDisplaySpan = document.getElementById('time-display'),
  progressBar = document.getElementById('progress-bar'),
  loadingHint = document.getElementById('loading-hint');
let activePlayer = audioPlayer;
let currentVolume = 1.0;

// Scale subtitle text with the video's rendered size, so windowed vs fullscreen stay proportional
function updateSubtitleSize() {
  var h = videoPlayer.clientHeight || 0;
  if (!h) return;
  var size = Math.max(14, Math.min(64, h * 0.045));
  videoPlayer.style.setProperty('--subtitle-size', size + 'px');
}
new ResizeObserver(updateSubtitleSize).observe(videoPlayer);
window.addEventListener('resize', updateSubtitleSize);
document.addEventListener('fullscreenchange', updateSubtitleSize);
function showVolumeOverlay() {
  var ov = document.getElementById('volume-overlay');
  if (!ov) {
    ov = document.createElement('div');
    ov.id = 'volume-overlay';
    ov.style.cssText = 'position:fixed;top:20px;left:50%;transform:translateX(-50%);z-index:10003;background:rgba(0,0,0,0.7);color:#fff;padding:8px 18px;border-radius:8px;font-size:15px;font-weight:600;pointer-events:none;transition:opacity .3s';
    document.body.appendChild(ov);
  }
  var icon = currentVolume <= 0 ? '🔇' : currentVolume < 0.3 ? '🔈' : currentVolume < 0.7 ? '🔉' : '🔊';
  ov.textContent = icon + ' ' + Math.round(currentVolume * 100) + '%';
  ov.style.opacity = '1';
  clearTimeout(ov._timer);
  ov._timer = setTimeout(function() { ov.style.opacity = '0'; }, 1500);
}
function adjustVolume(delta) {
  currentVolume = Math.max(0, Math.min(1, currentVolume + delta));
  activePlayer.volume = currentVolume;
  showVolumeOverlay();
}
function showSpeedOverlay() {
  var ov = document.getElementById('volume-overlay');
  if (!ov) {
    ov = document.createElement('div');
    ov.id = 'volume-overlay';
    ov.style.cssText = 'position:fixed;top:20px;left:50%;transform:translateX(-50%);z-index:10003;background:rgba(0,0,0,0.7);color:#fff;padding:8px 18px;border-radius:8px;font-size:15px;font-weight:600;pointer-events:none;transition:opacity .3s';
    document.body.appendChild(ov);
  }
  ov.textContent = '⚡ ' + currentPlaybackRate.toFixed(1) + 'x';
  ov.style.opacity = '1';
  clearTimeout(ov._timer);
  ov._timer = setTimeout(function() { ov.style.opacity = '0'; }, 1500);
}

// Select-all checkbox
document.getElementById('select-all-cb').addEventListener('change', function(e) {
  document.querySelectorAll('.file-cb').forEach(cb => cb.checked = e.target.checked);
  updateBatchBar();
});

// Batch bar logic
function getVisibleCheckboxes() {
  if (isGridView) {
    var grid = document.getElementById('grid-container');
    return grid ? grid.querySelectorAll('.file-cb') : [];
  }
  const searchArea = document.getElementById('search-results-area');
  if (searchArea && searchArea.style.display !== 'none') {
    return document.querySelectorAll('#search-results-body .file-cb');
  }
  return document.querySelectorAll('#file-list-table .file-cb');
}
function getSelectedPaths() {
  var paths = Array.from(document.querySelectorAll('.file-cb:checked')).map(cb => cb.value);
  if (isGridView) {
    paths = paths.concat(Array.from(document.querySelectorAll('.grid-item .file-cb:checked')).map(cb => cb.value));
  }
  return [...new Set(paths)];
}
function toggleSelectAll() {
  const boxes = getVisibleCheckboxes();
  const allChecked = Array.from(boxes).every(cb => cb.checked);
  boxes.forEach(cb => cb.checked = !allChecked);
  updateBatchBar();
}
function updateBatchBar() {
  const n = getSelectedPaths().length;
  const count = document.getElementById('select-count');
  const boxes = getVisibleCheckboxes();
  const total = boxes.length;
  const text = document.getElementById('select-toggle-text');
  if (count) { count.textContent = n > 0 ? n : ''; }
  if (text) { text.textContent = (n === total && total > 0) ? 'Deselect All' : 'Select All'; }
}
document.addEventListener('change', function(e) {
  if (e.target.classList.contains('file-cb')) updateBatchBar();
});

let playlist = {{tojson .Playlist}}, originalPlaylist = [...playlist], currentIndex = -1, isShuffled = false, currentPlaybackRate = 1.0;
function formatTime(seconds) { const min = Math.floor(seconds / 60); const sec = Math.floor(seconds % 60); return isNaN(sec) ? '0:00' : min+':'+(sec < 10 ? '0' : '')+sec; }
function encodePath(path) { return path.split('/').map(encodeURIComponent).join('/'); }
function highlightCurrentTrack() { document.querySelectorAll('.playing').forEach(el => el.classList.remove('playing')); if (currentIndex < 0 || currentIndex >= playlist.length) return; const path = playlist[currentIndex].path; const rows = document.querySelectorAll('tr.desktop-row'); for(let i=0;i<rows.length;i++){if(rows[i].dataset.path===path){const link=rows[i].querySelector('.file-name');if(link)link.classList.add('playing');break}} }
function playFrom(event, path) { event.stopPropagation(); const idx = playlist.findIndex(item => item.path === path); if (idx !== -1) { currentIndex = idx; playCurrent(); } else { console.error('playFrom: path not in playlist', path); } }
function handleSmartView(event, path, name) {
  event.stopPropagation();
  const ext = name.split('.').pop().toLowerCase();
  const imgExts = ['jpg','jpeg','png','gif','webp','svg','bmp','ico'];
  const mediaExts = ['mp3','m4a','ogg','flac','wav','opus','mp4','mkv','webm','avi','mov','flv','wmv','m4v'];
  if (imgExts.includes(ext)) { openLightbox('/' + encodePath(path), name); }
  else if (mediaExts.includes(ext)) { playFrom(event, path); }
  else { window.open('/preview/' + encodePath(path), '_blank'); }
}
function vtype(path){const e=path.split('.').pop().toLowerCase();if(e==='mp4')return'video/mp4';if(e==='webm')return'video/webm';if(e==='ogg')return'video/ogg';if(e==='mov')return'video/quicktime';if(e==='avi')return'video/x-msvideo';if(e==='mkv')return'video/x-matroska';return''}
function srtToVtt(srt){return 'WEBVTT\n\n'+srt.replace(/(\d{2}:\d{2}:\d{2}),(\d{3})/g,'$1.$2')}
async function loadSubtitle(s){const r=await fetch('/'+encodePath(s.path));let t=await r.text();if(!t.startsWith('WEBVTT'))t=srtToVtt(t);const b=new Blob([t],{type:'text/vtt'});return URL.createObjectURL(b)}
// Load subtitle tracks (sidecar files + embedded container tracks); embedded tracks are converted to WebVTT by the server via ffmpeg
async function addSubtitles(v, track){
  let subs = (track && track.subtitles) ? track.subtitles.slice() : [];
  try {
    const r = await fetch('/_actions/subtitles/' + encodePath(track.path));
    if(r.ok){ const j = await r.json(); if(Array.isArray(j)) subs = j; }
  } catch(e){}
  for(let i=0;i<subs.length;i++){
    const s = subs[i];
    const t = document.createElement('track');
    t.kind='subtitles';
    t.label = s.lang || ('Track '+(i+1));
    t.srclang = s.lang || '';
    if(s.embedded){ t.src = '/_actions/subtitle/' + encodePath(track.path) + '?stream=' + s.index; }
    else { t.src = await loadSubtitle(s); }
    if(i===0) t.default = true;
    v.appendChild(t);
  }
}
async function playCurrent() { if (currentIndex < 0 || currentIndex >= playlist.length) return; const track = playlist[currentIndex]; highlightCurrentTrack(); playerDock.style.display = 'flex'; trackNameSpan.textContent = track.name; loadingHint.style.display = 'inline'; setPlayPauseIcon(false); const streamUrl = '/'+encodePath(track.path); try { audioPlayer.style.display='none'; videoPlayer.style.display='none'; if(track.is_video) { document.getElementById('video-overlay').classList.add('active'); videoPlayer.style.display='block'; activePlayer=videoPlayer; videoPlayer.innerHTML=''; await addSubtitles(videoPlayer, track); const src=document.createElement('source'); src.src=streamUrl; src.type=vtype(track.path); videoPlayer.appendChild(src); videoPlayer.load(); videoPlayer.playbackRate=currentPlaybackRate; try { await videoPlayer.play(); } catch(playErr) { console.warn('play() failed, retrying muted:', playErr); videoPlayer.muted=true; await videoPlayer.play(); } loadingHint.style.display='none'; setPlayPauseIcon(true); checkVideoAudio(videoPlayer); } else { document.getElementById('video-overlay').classList.remove('active'); activePlayer=audioPlayer; const response=await fetch(streamUrl); if(!response.ok) { loadingHint.textContent='(Fetch Error '+response.status+')'; return; } const blob=await response.blob(); const blobUrl=URL.createObjectURL(blob); if(audioPlayer.src)URL.revokeObjectURL(audioPlayer.src); audioPlayer.src=blobUrl; audioPlayer.playbackRate=currentPlaybackRate; loadingHint.style.display='none'; setPlayPauseIcon(true); await audioPlayer.play(); } document.title=track.name; } catch(e){console.error("Play Error:",e);loadingHint.textContent="(Error)";} }
function setPlayPauseIcon(playing) {
  const icon = playing ? 'pause' : 'play';
  mainPlayPauseBtn.innerHTML = '<i data-lucide="' + icon + '"></i>';
  lucide.createIcons({ root: mainPlayPauseBtn });
}
function closeVideoOverlay(){const ov=document.getElementById('video-overlay');ov.classList.remove('active');videoPlayer.pause();videoPlayer.removeAttribute('src');videoPlayer.innerHTML='';videoPlayer.load();document.title='netdisk';hidePlayerDock()}
function hidePlayerDock(){playerDock.style.display='none';setPlayPauseIcon(false)}
function playNext() { if (!playlist.length) return; currentIndex = (currentIndex + 1) % playlist.length; playCurrent(); }
function playPrev() { if (!playlist.length) return; currentIndex = (currentIndex - 1 + playlist.length) % playlist.length; playCurrent(); }
mainPlayPauseBtn.onclick = () => { if (activePlayer.paused) activePlayer.play(); else activePlayer.pause(); };
seekBackBtn.onclick = () => { activePlayer.currentTime -= 10; };
seekFwdBtn.onclick = () => { activePlayer.currentTime += 10; };
audioPlayer.onplay = () => { setPlayPauseIcon(true); audioPlayer.playbackRate = currentPlaybackRate; };
audioPlayer.onpause = () => setPlayPauseIcon(false);
videoPlayer.onplay = () => { setPlayPauseIcon(true); videoPlayer.playbackRate = currentPlaybackRate; var fn = document.getElementById('video-filename'); if (fn) fn.style.display = 'none'; };
videoPlayer.onpause = () => { setPlayPauseIcon(false); var fn = document.getElementById('video-filename'); if (fn && currentIndex >= 0 && currentIndex < playlist.length) { fn.textContent = playlist[currentIndex].name; fn.style.display = 'block'; } };
audioPlayer.onended = playNext;
videoPlayer.onended = playNext;
audioPlayer.onloadedmetadata = () => { timeDisplaySpan.textContent = '0:00 / '+formatTime(audioPlayer.duration); progressBar.value = 0; };
audioPlayer.ontimeupdate = () => { const currentTime = audioPlayer.currentTime, duration = audioPlayer.duration; if (duration) { timeDisplaySpan.textContent = formatTime(currentTime)+' / '+formatTime(duration); progressBar.value = (currentTime / duration) * 100; } };
videoPlayer.onloadedmetadata = () => { timeDisplaySpan.textContent = '0:00 / '+formatTime(videoPlayer.duration); progressBar.value = 0; };
videoPlayer.ontimeupdate = () => { const currentTime = videoPlayer.currentTime, duration = videoPlayer.duration; if (duration) { timeDisplaySpan.textContent = formatTime(currentTime)+' / '+formatTime(duration); progressBar.value = (currentTime / duration) * 100; } };
progressBar.oninput = () => { if (activePlayer.duration) { activePlayer.currentTime = (progressBar.value / 100) * activePlayer.duration; } };
prevBtn.onclick = playPrev; nextBtn.onclick = playNext;
shuffleBtn.onclick = () => { isShuffled = !isShuffled; shuffleBtn.style.opacity = isShuffled ? '1' : '0.5'; if (isShuffled) { for (let i = playlist.length - 1; i > 0; i--) { const j = Math.floor(Math.random() * (i + 1));[playlist[i], playlist[j]] = [playlist[j], playlist[i]]; } } else { playlist = [...originalPlaylist]; } };
speedBtn.onclick = () => { const rates = [1.0, 1.5, 2.0, 2.5, 3.0, 0.5]; let next = rates[(rates.indexOf(currentPlaybackRate)+1)%rates.length]; if(isNaN(next)) next=1.0; currentPlaybackRate = next; activePlayer.playbackRate = next; speedBtn.textContent = next.toFixed(1)+'x'; };

function setClipboard(action) { fetch('/_actions/set_clipboard', { method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify({ paths: getSelectedPaths(), action }) }).then(()=>window.location.reload()); }
function pasteItems() { fetch('/_actions/paste', { method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify({ destination: '{{.CurrentPath}}' }) }).then(()=>window.location.reload()); }
function downloadSelected() { fetch('/_actions/zip_download', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ paths: getSelectedPaths() }) }).then(r => r.blob()).then(b => { const url = window.URL.createObjectURL(b); const a = document.createElement('a'); a.href = url; a.download = "download.zip"; a.click(); }); }
function showProperties() {
  const paths = getSelectedPaths();
  if (!paths.length) return;
  fetch('/_actions/properties', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({paths: paths})
  }).then(r => r.json()).then(d => {
    const fmt = d.totalSize >= 1073741824 ? (d.totalSize/1073741824).toFixed(2)+' GB' : d.totalSize >= 1048576 ? (d.totalSize/1048576).toFixed(2)+' MB' : d.totalSize >= 1024 ? (d.totalSize/1024).toFixed(1)+' KB' : d.totalSize+' B';
    const ov = document.getElementById('modal-overlay');
    ov.querySelector('h3').textContent = 'Properties';
    ov.querySelector('input[type=text]').style.display = 'none';
    const msg = document.createElement('div');
    msg.className = 'modal-extra';
    msg.style.cssText = 'font-size:14px;line-height:1.8;color:var(--text-color)';
    msg.innerHTML = 'Files: <b>'+d.files+'</b><br>Folders: <b>'+d.dirs+'</b><br>Total Size: <b>'+fmt+'</b>';
    ov.querySelector('.modal-box').insertBefore(msg, ov.querySelector('.modal-btns'));
    ov.querySelector('.modal-cancel').onclick = () => { ov.classList.remove('active'); msg.remove(); };
    ov.querySelector('.modal-ok').onclick = () => { ov.classList.remove('active'); msg.remove(); };
    ov.classList.add('active');
  });
}
function renameItem(event, path, oldName) { event.stopPropagation(); showPrompt('Rename to:', oldName, function(n){ if(n) fetch('/_actions/rename', { method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify({ path: '{{.CurrentPath}}', old_name: oldName, new_name: n }) }).then(()=>window.location.reload()); }); }

function showPrompt(title, defaultVal, onOk) {
  const ov=document.getElementById('modal-overlay');
  ov.querySelector('h3').textContent=title;
  const inp=ov.querySelector('input[type=text]');
  inp.value=defaultVal||''; inp.style.display='';
  ov.querySelector('.modal-cancel').onclick=()=>{ov.classList.remove('active')};
  ov.querySelector('.modal-ok').onclick=()=>{ov.classList.remove('active');onOk(inp.value)};
  ov.classList.add('active'); inp.focus(); inp.select();
}
function showConfirm(title, onOk) {
  const ov=document.getElementById('modal-overlay');
  ov.querySelector('h3').textContent=title;
  ov.querySelector('input[type=text]').style.display='none';
  ov.querySelector('.modal-cancel').onclick=()=>{ov.classList.remove('active')};
  ov.querySelector('.modal-ok').onclick=()=>{ov.classList.remove('active');onOk()};
  ov.classList.add('active');
}

// Close dropdowns on outside click
document.addEventListener('click', function(e) {
  document.querySelectorAll('.dropdown.open').forEach(d => { if (!d.contains(e.target)) d.classList.remove('open'); });
});

function toggleDropdown(btn) {
  const dd = btn.parentElement;
  const wasOpen = dd.classList.contains('open');
  document.querySelectorAll('.row-dropdown.open').forEach(function(d){d.classList.remove('open')});
  if (!wasOpen) {
    dd.classList.add('open');
    lucide.createIcons();
  }
}

// Hover delay: keep menu visible briefly after mouse leaves
(function(){
  var hoverTimer = null;
  document.addEventListener('mouseenter', function(e) {
    var dd = e.target.closest('.row-dropdown');
    if (dd) {
      clearTimeout(hoverTimer);
      document.querySelectorAll('.row-dropdown.hover-active').forEach(function(d) {
        if (d !== dd) d.classList.remove('hover-active');
      });
      dd.classList.add('hover-active');
    }
  }, true);
  document.addEventListener('mouseleave', function(e) {
    var dd = e.target.closest('.row-dropdown');
    if (dd && !dd.contains(e.relatedTarget)) {
      clearTimeout(hoverTimer);
      hoverTimer = setTimeout(function() {
        dd.classList.remove('hover-active');
      }, 500);
    }
  }, true);
})();

// Close modals on overlay click
document.querySelectorAll('.modal-overlay').forEach(ov => {
  ov.addEventListener('click', function(e) {
    if (e.target === this) {
      this.classList.remove('active');
      this.querySelectorAll('.modal-box .modal-extra').forEach(d => d.remove());
    }
  });
});

// --- Search ---
const searchInput = document.getElementById('search-input');
const searchBadge = document.getElementById('search-badge');
let searchActive = false;
let searchTimer = null;

function searchIcon(parent, name) {
  const i = document.createElement('i');
  i.setAttribute('data-lucide', name);
  parent.appendChild(i);
}

function searchMenuButton(label, icon) {
  const b = document.createElement('button');
  b.type = 'button';
  b.className = 'dropdown-item';
  searchIcon(b, icon);
  b.appendChild(document.createTextNode(' ' + label));
  return b;
}

function buildSearchRow(item) {
  const tr = document.createElement('tr');
  tr.className = 'desktop-row';
  tr.dataset.path = item.path;
  tr.dataset.size = item.size || 0;

  const tdCheck = document.createElement('td');
  tdCheck.className = 'col-check';
  const cb = document.createElement('input');
  cb.type = 'checkbox';
  cb.className = 'file-cb search-cb';
  cb.value = item.path;
  tdCheck.appendChild(cb);

  const tdName = document.createElement('td');
  tdName.className = 'col-name';
  const link = document.createElement('a');
  link.className = 'file-name';
  link.href = '/' + encodePath(item.path);
  const icon = document.createElement('i');
  icon.className = 'file-icon' + (item.is_dir ? ' folder-icon' : '');
  icon.setAttribute('data-lucide', item.is_dir ? 'folder' : 'file');
  link.appendChild(icon);
  link.appendChild(document.createTextNode(item.is_dir ? '[' + item.name + ']' : item.name));
  tdName.appendChild(link);

  const tdSize = document.createElement('td');
  tdSize.className = 'col-size';
  const sizeSpan = document.createElement('span');
  sizeSpan.className = 'm-meta-item';
  sizeSpan.textContent = item.is_dir ? '--' : formatFileSizeJS(item.size);
  tdSize.appendChild(sizeSpan);

  const tdMtime = document.createElement('td');
  tdMtime.className = 'col-modified';
  const mtimeSpan = document.createElement('span');
  mtimeSpan.className = 'm-meta-item m-date-text';
  mtimeSpan.textContent = item.formatted_mtime;
  tdMtime.appendChild(mtimeSpan);

  const tdActions = document.createElement('td');
  tdActions.className = 'col-actions';
  const viewSlot = document.createElement('div');
  viewSlot.className = 'view-btn-slot';
  const ext = (item.name.split('.').pop() || '').toLowerCase();
  const imgExts = ['jpg','jpeg','png','gif','webp','svg','bmp','ico'];
  const mediaExts = ['mp3','m4a','ogg','flac','wav','opus','mp4','mkv','webm','avi','mov','flv','wmv','m4v'];
  const textExts = ['txt','md','json','xml','html','css','js','py','sh','conf','ini','log','cfg','go','ts','tsx','jsx','java','c','cpp','h','rs','rb','php'];
  if (!item.is_dir) {
    let viewIcon = '';
    if (imgExts.includes(ext)) viewIcon = 'image';
    else if (mediaExts.includes(ext)) viewIcon = 'play';
    else if (textExts.includes(ext)) viewIcon = 'eye';
    if (viewIcon) {
      const viewBtn = document.createElement('button');
      viewBtn.type = 'button';
      viewBtn.className = 'action-btn view-btn';
      viewBtn.title = (viewIcon === 'play') ? 'Play' : 'Preview';
      searchIcon(viewBtn, viewIcon);
      viewBtn.onclick = function(e) { handleSmartView(e, item.path, item.name); };
      viewSlot.appendChild(viewBtn);
    }
  }
  tdActions.appendChild(viewSlot);

  const dropdown = document.createElement('div');
  dropdown.className = 'dropdown row-dropdown';
  const moreBtn = document.createElement('button');
  moreBtn.type = 'button';
  moreBtn.className = 'action-btn more-btn';
  moreBtn.title = 'More';
  searchIcon(moreBtn, 'more-vertical');
  moreBtn.onclick = function(e) { e.stopPropagation(); this.parentElement.classList.toggle('open'); };
  dropdown.appendChild(moreBtn);

  const menu = document.createElement('div');
  menu.className = 'dropdown-menu';
  if (item.can_edit) {
    const edit = document.createElement('a');
    edit.className = 'dropdown-item';
    edit.title = 'Edit';
    edit.href = '/edit/' + encodePath(item.path);
    searchIcon(edit, 'pencil');
    edit.appendChild(document.createTextNode(' Edit'));
    menu.appendChild(edit);
  }
  const share = searchMenuButton('Share', 'share-2');
  share.onclick = function(e) { shareFile(e, item.path); };
  menu.appendChild(share);
  if (item.can_rename) {
    const rename = searchMenuButton('Rename', 'edit-3');
    rename.onclick = function(e) { renameItem(e, item.path, item.name); };
    menu.appendChild(rename);
  }
  if (item.can_delete) {
    const del = searchMenuButton('Delete', 'trash-2');
    del.className += ' delete-action';
    del.onclick = function(e) { deleteItem(e, item.path, item.name); };
    menu.appendChild(del);
  }
  dropdown.appendChild(menu);
  tdActions.appendChild(dropdown);

  tr.appendChild(tdCheck);
  tr.appendChild(tdName);
  tr.appendChild(tdSize);
  tr.appendChild(tdMtime);
  tr.appendChild(tdActions);
  return tr;
}

function doSearch() {
  const q = (searchInput.value || '').trim();
  const dir = '{{.CurrentPath}}';
  const fileList = document.getElementById('file-list-table');
  if (!q) { clearSearch(); return; }
  fetch('/api/search?q=' + encodeURIComponent(q) + '&dir=' + encodeURIComponent(dir)).then(r => r.json()).then(items => {
    const area = document.getElementById('search-results-area');
    const tbody = document.getElementById('search-results-body');
    tbody.innerHTML = '';
    fileList.style.display = 'none';
    var gridContainer = document.getElementById('grid-container');
    if (gridContainer) gridContainer.style.display = 'none';
    searchActive = true;
    searchBadge.classList.add('active');
    searchBadge.querySelector('.search-count').textContent = items.length;
    items.forEach(item => {
      tbody.appendChild(buildSearchRow(item));
    });
    lucide.createIcons();
    area.style.display = 'block';
    document.getElementById('search-select-all').checked = false;
    document.getElementById('search-select-all').onchange = function(e) { document.querySelectorAll('.search-cb').forEach(cb => cb.checked = e.target.checked); updateBatchBar(); };
  });
}

function clearSearch() {
  searchInput.value = '';
  searchActive = false;
  searchBadge.classList.remove('active');
  document.getElementById('search-results-area').style.display = 'none';
  if (isGridView) {
    document.getElementById('file-list-table').style.display = 'none';
    var gridContainer = document.getElementById('grid-container');
    if (gridContainer) gridContainer.style.display = '';
  } else {
    document.getElementById('file-list-table').style.display = '';
  }
}

function formatFileSizeJS(size) {
  if (size >= 1073741824) return (size / 1073741824).toFixed(1) + ' GB';
  if (size >= 1048576) return (size / 1048576).toFixed(1) + ' MB';
  if (size >= 1024) return (size / 1024).toFixed(1) + ' KB';
  return size + ' B';
}

if (searchInput) {
  searchInput.addEventListener('input', function() {
    clearTimeout(searchTimer);
    const q = searchInput.value.trim();
    if (!q) { clearSearch(); return; }
    searchTimer = setTimeout(doSearch, 300);
  });
  searchInput.addEventListener('keydown', function(e) {
    if (e.key === 'Enter') { e.preventDefault(); clearTimeout(searchTimer); doSearch(); }
    if (e.key === 'Escape') { clearSearch(); searchInput.blur(); }
  });
}

// --- Grid View ---
let isGridView = localStorage.getItem('cd_view') === 'grid';
function toggleView() {
  isGridView = !isGridView;
  localStorage.setItem('cd_view', isGridView ? 'grid' : 'list');
  updateGridIcon();
  applyView();
}
function updateGridIcon() {
  const el = document.getElementById('grid-icon');
  const label = document.getElementById('grid-label');
  if (!el) return;
  const gridSvg = '<svg id="grid-icon" xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="7" height="7" x="3" y="3" rx="1"/><rect width="7" height="7" x="14" y="3" rx="1"/><rect width="7" height="7" x="14" y="14" rx="1"/><rect width="7" height="7" x="3" y="14" rx="1"/></svg>';
  const listSvg = '<svg id="grid-icon" xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="8" x2="21" y1="6" y2="6"/><line x1="8" x2="21" y1="12" y2="12"/><line x1="8" x2="21" y1="18" y2="18"/><line x1="3" x2="3.01" y1="6" y2="6"/><line x1="3" x2="3.01" y1="12" y2="12"/><line x1="3" x2="3.01" y1="18" y2="18"/></svg>';
  el.outerHTML = isGridView ? listSvg : gridSvg;
  if (label) label.textContent = isGridView ? 'Detail View' : 'Grid View';
}
if (isGridView) updateGridIcon();
function applyView() {
  const table = document.getElementById('file-list-table');
  const items = Array.from(table.querySelectorAll('tbody tr'));
  if(isGridView) {
    table.style.display='none';
    let grid = document.getElementById('grid-container');
    if(!grid) { grid=document.createElement('div'); grid.id='grid-container'; grid.className='grid-view'; table.parentNode.insertBefore(grid, table.nextSibling); }
    grid.innerHTML='';
    items.forEach(tr=>{
      const path=tr.dataset.path, name=tr.dataset.name||tr.querySelector('.file-name').textContent.trim();
      const isDir=tr.dataset.isdir==='true';
      const displayName=isDir&&name.startsWith('[')?name.slice(1,-1):name;
      const ext=displayName.split('.').pop().toLowerCase();
      const imgExts=['jpg','jpeg','png','gif','webp','svg'];
      const a=document.createElement('div'); a.className='grid-item';
      a.dataset.path=path; a.dataset.name=name; a.dataset.isdir=isDir;
      a.onclick=function(e){
        if(e.target.closest('.grid-check')) return;
        if(isDir) window.location.href='/'+encodePath(path);
        else { var fe={stopPropagation:function(){}}; handleSmartView(fe,path,name); }
      };

      const checkLabel=document.createElement('label');
      checkLabel.className='grid-check';
      checkLabel.onclick=function(e){e.stopPropagation();};
      const cb=document.createElement('input');
      cb.type='checkbox'; cb.className='file-cb'; cb.value=path;
      cb.onchange=function(){gridCheckChange(this);};
      checkLabel.appendChild(cb);
      a.appendChild(checkLabel);

      if(!isDir && imgExts.includes(ext)) {
        const img=document.createElement('img');
        img.src='/'+encodePath(path);
        img.alt=displayName;
        img.loading='lazy';
        a.appendChild(img);
      } else {
        let icon='&#128196;';
        if(isDir) icon='&#128193;';
        else if(['mp3','m4a','ogg','flac','wav','opus'].includes(ext)) icon='&#127925;';
        else if(['mp4','mkv','avi','mov','webm'].includes(ext)) icon='&#127909;';
        else if(['zip','rar','7z','tar','gz'].includes(ext)) icon='&#128230;';
        const iconDiv=document.createElement('div');
        iconDiv.className='grid-icon';
        iconDiv.innerHTML=icon;
        a.appendChild(iconDiv);
      }
      const nameDiv=document.createElement('div');
      nameDiv.className='grid-name';
      nameDiv.textContent=displayName;
      a.appendChild(nameDiv);
      grid.appendChild(a);
    });
  } else {
    table.style.display='';
    const grid=document.getElementById('grid-container');
    if(grid) grid.remove();
  }
}
function gridCheckChange(cb) {
  const item = cb.closest('.grid-item');
  if(cb.checked) item.classList.add('grid-checked');
  else item.classList.remove('grid-checked');
  updateBatchBar();
}
if(isGridView) applyView();

// --- Upload with directory support + drag-and-drop ---
const uploadForm = document.getElementById('upload-form');
const uploadFiles = document.getElementById('upload-files');
const uploadFolder = document.getElementById('upload-folder');
const dropOverlay = document.getElementById('drop-overlay');

if(uploadForm) {
  uploadFiles.addEventListener('change', function() {
    if(!uploadFiles.files.length) return;
    uploadDirFiles(Array.from(uploadFiles.files));
    uploadFiles.value = '';
  });
  uploadFolder.addEventListener('change', function() {
    if(!uploadFolder.files.length) return;
    uploadDirFiles(Array.from(uploadFolder.files));
    uploadFolder.value = '';
  });
}

let uploadListOpen = false;
function toggleUploadList() {
  uploadListOpen = !uploadListOpen;
  const list = document.getElementById('upload-filelist');
  const btn = document.getElementById('upload-toggle');
  if (list) list.classList.toggle('open', uploadListOpen);
  if (btn) btn.classList.toggle('open', uploadListOpen);
}

function uploadDirFiles(fileList) {
  if (!fileList.length) return;
  const currentPath = '{{.CurrentPath}}';
  const panel = document.getElementById('upload-float');
  const bar = document.getElementById('upload-progress-bar');
  const totalPctEl = document.getElementById('upload-total-percent');
  const speedEl = document.getElementById('upload-speed');
  const countEl = document.getElementById('upload-count');
  const sizeEl = document.getElementById('upload-size');
  const listEl = document.getElementById('upload-filelist');
  const toggleBtn = document.getElementById('upload-toggle');

  // One upload request per file, so per-file progress and speed can be shown.
  // The whole relative path (including subdirectories) is sent in "paths"; the
  // request URL is just the base directory.
  const jobs = fileList.map(f => {
    const rel = f.webkitRelativePath || f.name;
    return {
      file: f, rel: rel,
      target: currentPath,
      size: f.size, loaded: 0, done: false,
      el: null, pctEl: null, spdEl: null,
      lastLoaded: 0, lastTime: Date.now(),
    };
  });

  const totalSize = jobs.reduce((s, j) => s + j.size, 0);
  let totalBytes = totalSize || 1;
  let jobsLen = jobs.length;
  let completed = 0, failures = 0, skipped = 0, lastError = '', active = 0;
  let uploadConflict = 'overwrite';

  // Initialize UI
  panel.classList.add('visible');
  bar.style.width = '0%';
  totalPctEl.textContent = '0%';
  speedEl.textContent = '--';
  countEl.textContent = '(0/' + jobsLen + ')';
  sizeEl.textContent = formatFileSizeJS(totalSize);
  listEl.innerHTML = '';
  listEl.classList.toggle('open', uploadListOpen);
  toggleBtn.classList.toggle('open', uploadListOpen);
  toggleBtn.classList.toggle('show', jobsLen > 1);

  jobs.forEach(job => {
    const row = document.createElement('div');
    row.className = 'upload-item';
    const nm = document.createElement('div');
    nm.className = 'upload-item-name';
    nm.textContent = job.rel;
    nm.title = job.rel;
    const meta = document.createElement('div');
    meta.className = 'upload-item-meta';
    const pct = document.createElement('span');
    pct.textContent = '0%';
    const spd = document.createElement('span');
    spd.textContent = '--';
    meta.appendChild(pct); meta.appendChild(spd);
    row.appendChild(nm); row.appendChild(meta);
    listEl.appendChild(row);
    job.el = row; job.pctEl = pct; job.spdEl = spd; job.nameEl = nm;
  });

  function totalLoaded() {
    let sum = 0;
    jobs.forEach(j => { sum += (j.done ? j.size : j.loaded); });
    return sum;
  }
  function updateTotal() {
    const pct = Math.min(100, Math.round(totalLoaded() / totalBytes * 100));
    bar.style.width = pct + '%';
    totalPctEl.textContent = pct + '%';
  }

  // Overall speed: throttled to 500ms
  let totalLastLoaded = 0, totalLastTime = Date.now(), speedTimer = null;
  function updateSpeed() {
    if (speedTimer) return;
    speedTimer = setTimeout(() => {
      speedTimer = null;
      const sum = totalLoaded();
      const now = Date.now();
      const dt = (now - totalLastTime) / 1000;
      if (dt > 0) speedEl.textContent = formatSpeed((sum - totalLastLoaded) / dt);
      totalLastLoaded = sum;
      totalLastTime = now;
    }, 500);
  }

  function allDone() {
    if (failures > 0) {
      panel.classList.remove('visible');
      alert('Upload failed (' + failures + '/' + jobsLen + '): ' + lastError);
    } else if (skipped > 0) {
      showToast('Skipped ' + skipped + ' file(s)');
      setTimeout(() => window.location.reload(), 1200);
    } else {
      window.location.reload();
    }
  }
  function finishOne() {
    completed++;
    countEl.textContent = '(' + completed + '/' + jobsLen + ')';
    updateTotal();
    if (completed >= jobsLen) allDone();
  }

  const MAX_CONCURRENT = 3;
  let next = 0;
  function pump() {
    while (active < MAX_CONCURRENT && next < jobsLen) {
      const j = jobs[next++];
      if (j.skipped) continue;
      startJob(j);
    }
  }

  function startJob(job) {
    active++;
    const fd = new FormData();
    fd.append('paths', JSON.stringify([job.rel]));
    fd.append('conflict', uploadConflict);
    fd.append('file', job.file);
    const xhr = new XMLHttpRequest();
    xhr.open('POST', '/_actions/upload/' + encodePath(job.target) + '?conflict=' + encodeURIComponent(uploadConflict));
    xhr.setRequestHeader('X-Requested-With', 'XMLHttpRequest');
    xhr.upload.onprogress = (ev) => {
      if (!ev.lengthComputable) return;
      job.loaded = ev.loaded;
      const pct = job.size ? Math.min(100, Math.round(ev.loaded / job.size * 100)) : 100;
      job.pctEl.textContent = pct + '%';
      const now = Date.now();
      const dt = (now - job.lastTime) / 1000;
      if (dt > 0.3) {
        job.spdEl.textContent = formatSpeed((ev.loaded - job.lastLoaded) / dt);
        job.lastLoaded = ev.loaded;
        job.lastTime = now;
      }
      updateTotal();
      updateSpeed();
    };
    const onEnd = (ok, errMsg, info) => {
      active--;
      if (ok) {
        job.done = true;
        job.loaded = job.size;
        job.el.classList.add('done');
        job.pctEl.textContent = '100%';
        job.spdEl.textContent = '';
        if (info && info.skipped) skipped += info.skipped;
      } else {
        failures++;
        lastError = errMsg;
        job.el.classList.add('error');
        job.pctEl.textContent = 'Error';
      }
      finishOne();
      pump();
    };
    xhr.onload = () => {
      const ok = xhr.status >= 200 && xhr.status < 300;
      let info = null;
      if (ok) { try { info = JSON.parse(xhr.responseText); } catch (e) {} }
      onEnd(ok, (xhr.responseText || '').trim() || ('HTTP ' + xhr.status), info);
    };
    xhr.onerror = () => onEnd(false, 'Network error');
    xhr.send(fd);
  }

  // --- Conflict pre-flight -------------------------------------------------
  // 1) Ask the server which target files and top-level folders already exist.
  // 2) Resolve folder conflicts (merge / rename / skip / cancel).
  // 3) Resolve file conflicts (overwrite / rename / skip / cancel).
  function jobWebPath(j) { return currentPath ? currentPath + '/' + j.rel : j.rel; }
  function jobTopWeb(j) {
    const i = j.rel.indexOf('/');
    if (i <= 0) return '';
    const top = j.rel.slice(0, i);
    return currentPath ? currentPath + '/' + top : top;
  }
  function checkUpload() {
    const paths = jobs.filter(j => !j.skipped).map(jobWebPath);
    const dirs = [], seen = {};
    jobs.forEach(j => {
      if (j.skipped) return;
      const t = jobTopWeb(j);
      if (t && !seen[t]) { seen[t] = 1; dirs.push(t); }
    });
    return fetch('/_actions/upload_check', {
      method: 'POST', headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({paths: paths, dirs: dirs, base: currentPath})
    }).then(r => r.json());
  }
  function markSkipped() {
    jobs.forEach(j => {
      if (j.skipped && !j.counted) {
        j.counted = true; completed++; skipped++;
        j.el.classList.add('skipped');
        j.pctEl.textContent = 'Skipped';
        j.spdEl.textContent = '';
      }
    });
    countEl.textContent = '(' + completed + '/' + jobsLen + ')';
  }

  async function preflight() {
    try {
      let d = await checkUpload();

      const dc = (d && d.dir_conflicts) || [];
      if (dc.length) {
        const choice = await showChoice('Folder already exists',
          dc.length + ' folder(s) already exist. How should they be handled?', [
          {label:'Merge', value:'merge', style:'primary'},
          {label:'Rename folder', value:'rename'},
          {label:'Skip folder', value:'skip'},
          {label:'Cancel', value:null},
        ]);
        if (choice === null) { panel.classList.remove('visible'); return; }
        if (choice === 'skip') {
          const set = {};
          dc.forEach(x => { set[x.path] = 1; });
          jobs.forEach(j => { if (set[jobTopWeb(j)]) j.skipped = true; });
        } else if (choice === 'rename') {
          const map = {};
          dc.forEach(x => { map[x.path] = x.name; });
          jobs.forEach(j => {
            const t = jobTopWeb(j);
            if (t && map[t]) {
              const rest = j.rel.slice(j.rel.indexOf('/') + 1);
              j.rel = map[t] + '/' + rest;
              if (j.nameEl) { j.nameEl.textContent = j.rel; j.nameEl.title = j.rel; }
            }
          });
        }
        d = await checkUpload(); // re-check files after folder handling
      }

      const fc = (d && d.conflicts) || [];
      if (fc.length) {
        const choice = await askConflict(fc.length + ' file(s) already exist. How should they be handled?');
        if (choice === null) { panel.classList.remove('visible'); return; }
        uploadConflict = choice;
        if (choice === 'skip') {
          // Skip client-side: don't even send the conflicting files, so the
          // behavior never depends on request/field ordering on the server.
          const set = {};
          fc.forEach(p => { set[p] = 1; });
          jobs.forEach(j => { if (set[jobWebPath(j)]) j.skipped = true; });
        }
      }
    } catch (e) { /* network error: fall through and just try */ }

    markSkipped();
    const realSize = jobs.reduce((s, j) => s + (j.skipped ? 0 : j.size), 0);
    totalBytes = realSize || 1;
    sizeEl.textContent = formatFileSizeJS(realSize);
    if (completed >= jobsLen) { allDone(); return; }
    pump();
  }

  preflight();
}

function formatSpeed(bytesPerSec) {
  if (bytesPerSec >= 1048576) return (bytesPerSec / 1048576).toFixed(1) + ' MB/s';
  if (bytesPerSec >= 1024) return (bytesPerSec / 1024).toFixed(1) + ' KB/s';
  return Math.round(bytesPerSec) + ' B/s';
}

// --- Drag and Drop ---
let dragCounter = 0;
document.addEventListener('dragenter', (e) => { e.preventDefault(); dragCounter++; dropOverlay.classList.add('active'); });
document.addEventListener('dragleave', (e) => { e.preventDefault(); dragCounter--; if (dragCounter <= 0) { dragCounter = 0; dropOverlay.classList.remove('active'); } });
document.addEventListener('dragover', (e) => { e.preventDefault(); });
document.addEventListener('drop', (e) => {
  e.preventDefault();
  dragCounter = 0;
  dropOverlay.classList.remove('active');
  if (!e.dataTransfer.items.length) return;
  const allFiles = [];
  const entries = [];
  for (const item of e.dataTransfer.items) {
    const entry = item.webkitGetAsEntry ? item.webkitGetAsEntry() : null;
    if (entry) entries.push(entry);
    else if (item.kind === 'file') allFiles.push(item.getAsFile());
  }
  if (!entries.length && !allFiles.length) return;
  let pendingEntries = entries.length;
  if (pendingEntries === 0) { uploadDirFiles(allFiles); return; }
  entries.forEach(entry => traverseEntry(entry, '', allFiles, () => { pendingEntries--; if (pendingEntries <= 0) uploadDirFiles(allFiles); }));
});

function traverseEntry(entry, basePath, outFiles, cb) {
  if (entry.isFile) {
    entry.file(f => { try { Object.defineProperty(f, 'webkitRelativePath', { value: basePath + f.name, configurable: true }); } catch (e) {} outFiles.push(f); cb(); });
  } else if (entry.isDirectory) {
    const reader = entry.createReader();
    const dirPath = basePath + entry.name + '/';
    let allEntries = [];
    function readBatch() {
      reader.readEntries(batch => {
        if (batch.length === 0) {
          let pending = allEntries.length;
          if (pending === 0) { cb(); return; }
          allEntries.forEach(e => traverseEntry(e, dirPath, outFiles, () => { pending--; if (pending <= 0) cb(); }));
        } else {
          allEntries = allEntries.concat(Array.from(batch));
          readBatch();
        }
      });
    }
    readBatch();
  } else { cb(); }
}

// --- Toast ---
function showToast(msg, isError){ const t=document.createElement('div'); t.className='toast-msg' + (isError ? ' error' : ''); t.textContent=msg; document.body.appendChild(t); requestAnimationFrame(()=>t.classList.add('show')); setTimeout(()=>{t.classList.remove('show');setTimeout(()=>t.remove(),300)},1500); }
function showToastMsg(msg){showToast(msg)}

// Long-lived toast (for messages that need time to read, e.g. unsupported audio codec)
function showAudioHint(msg){
  const t=document.createElement('div');
  t.className='toast-msg';
  t.style.maxWidth='420px';
  t.textContent=msg;
  document.body.appendChild(t);
  requestAnimationFrame(()=>t.classList.add('show'));
  setTimeout(()=>{t.classList.remove('show');setTimeout(()=>t.remove(),300)},6000);
}

// After video playback starts, detect whether a decodable audio track exists; if the browser cannot decode it (e.g. AC-3/DTS/TrueHD), show a hint
let _audioCheckToken = 0;
function checkVideoAudio(v){
  const token = ++_audioCheckToken;
  setTimeout(()=>{
    if (token !== _audioCheckToken) return;
    if (!v || v !== videoPlayer || v.style.display === 'none') return;
    let hasAudio = null;
    if (typeof v.mozHasAudio === 'boolean') hasAudio = v.mozHasAudio;          // Firefox
    else if (typeof v.webkitAudioDecodedByteCount === 'number') hasAudio = v.webkitAudioDecodedByteCount > 0; // Chrome/Safari
    if (hasAudio === false) showAudioHint('Note: this file uses an audio codec the browser cannot decode (e.g. AC-3 / DTS / TrueHD), so there is no sound');
  }, 2500);
}

// --- Share / QR ---
let shareState = { path: '', token: '', url: '' };

function absoluteURL(u) {
  try { return new URL(u, window.location.origin).href; } catch (e) { return window.location.origin + u; }
}

function renderQR(canvas, text) {
  if (!canvas || !text || typeof qrcode !== 'function') return false;
  try {
    const qr = qrcode(0, 'M');
    qr.addData(text, 'Byte');
    qr.make();
    const count = qr.getModuleCount();
    const quiet = 4;
    const cell = Math.max(2, Math.floor(200 / (count + quiet * 2)));
    const dim = (count + quiet * 2) * cell;
    canvas.width = dim;
    canvas.height = dim;
    const ctx = canvas.getContext('2d');
    ctx.fillStyle = '#ffffff';
    ctx.fillRect(0, 0, dim, dim);
    ctx.fillStyle = '#000000';
    for (let r = 0; r < count; r++) {
      for (let c = 0; c < count; c++) {
        if (qr.isDark(r, c)) ctx.fillRect((c + quiet) * cell, (r + quiet) * cell, cell, cell);
      }
    }
    return true;
  } catch (e) {
    console.error('QR generation failed', e);
    return false;
  }
}

function formatExpiry(iso) {
  if (!iso) return 'This link never expires.';
  const d = new Date(iso);
  if (isNaN(d.getTime()) || d.getFullYear() < 2000) return 'This link never expires.';
  return 'Expires: ' + d.toLocaleString();
}

function applyShare(share) {
  const urlInput = document.getElementById('share-url');
  const status = document.getElementById('share-status');
  const qrWrap = document.getElementById('share-qr-wrap');
  const cancelBtn = document.getElementById('share-cancel-btn');
  const passwordInput = document.getElementById('share-password');
  status.className = 'share-status';
  if (share && share.url) {
    shareState.token = share.token || '';
    shareState.url = absoluteURL(share.url);
    urlInput.value = shareState.url;
    cancelBtn.disabled = false;
    status.textContent = formatExpiry(share.expires_at) + (share.has_password ? ' · Password protected' : '');
    passwordInput.placeholder = share.has_password ? 'Leave empty to remove password' : 'Leave empty for no password';
    if (renderQR(document.getElementById('share-qr'), shareState.url)) qrWrap.classList.add('active');
    else qrWrap.classList.remove('active');
  } else {
    shareState.token = '';
    shareState.url = '';
    urlInput.value = '';
    cancelBtn.disabled = true;
    status.textContent = 'No active share link.';
    passwordInput.placeholder = 'Leave empty for no password';
    qrWrap.classList.remove('active');
  }
}

function openShare(path) {
  shareState = { path: path || '', token: '', url: '' };
  const modal = document.getElementById('share-modal');
  const cleanPath = (path || '').replace(/\/+$/, '');
  const name = cleanPath.split('/').pop() || 'Root';
  document.getElementById('share-title').textContent = 'Share: ' + name;
  document.getElementById('share-path').textContent = '/' + (path || '');
  document.getElementById('share-hours').value = '24';
  document.getElementById('share-password').value = '';
  applyShare(null);
  document.getElementById('share-status').textContent = 'Loading...';
  modal.classList.add('active');
  fetch('/api/share?path=' + encodeURIComponent(path || '')).then(r => r.json()).then(d => {
    if (d && d.success && d.share) applyShare(d.share);
    else applyShare(null);
  }).catch(() => {
    const status = document.getElementById('share-status');
    status.className = 'share-status error';
    status.textContent = 'Could not load share status.';
  });
}

function closeShareModal() {
  document.getElementById('share-modal').classList.remove('active');
}

function generateShare() {
  const path = shareState.path || '';
  const hours = Math.max(0, parseInt(document.getElementById('share-hours').value, 10) || 0);
  const password = document.getElementById('share-password').value || '';
  const btn = document.getElementById('share-generate-btn');
  const status = document.getElementById('share-status');
  btn.disabled = true;
  status.className = 'share-status';
  status.textContent = 'Generating link...';
  fetch('/api/share', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({path: path, hours: hours, password: password})
  }).then(r => r.json()).then(d => {
    if (d && d.success) {
      applyShare({token: d.token, url: d.url, expires_at: d.expires_at, has_password: d.has_password});
      copyToClipboard(shareState.url);
      showToast('Share link copied!');
    } else {
      status.className = 'share-status error';
      status.textContent = (d && d.message) || 'Failed to generate link.';
    }
  }).catch(() => {
    status.className = 'share-status error';
    status.textContent = 'Failed to generate link.';
  }).finally(() => { btn.disabled = false; });
}

function cancelShare() {
  if (!shareState.token) return;
  const token = shareState.token;
  showConfirm('Cancel this share link? The URL will stop working immediately.', function() {
    fetch('/api/share/' + encodeURIComponent(token), {method: 'DELETE'}).then(r => r.json()).then(d => {
      if (d && d.success) {
        applyShare(null);
        showToast('Share link cancelled');
      } else {
        showToast('Failed to cancel share', true);
      }
    }).catch(() => showToast('Failed to cancel share', true));
  });
}

function copyShareURL() {
  if (!shareState.url) return;
  copyToClipboard(shareState.url);
  showToast('Share link copied!');
}

function shareCurrent() {
  openShare('{{.CurrentPath}}');
}

function shareFile(event, path) {
  event.stopPropagation();
  openShare(path);
}

function showPageQR() {
  const url = window.location.href;
  document.getElementById('page-url').value = url;
  renderQR(document.getElementById('page-qr'), url);
  document.getElementById('qr-modal').classList.add('active');
}

function closePageQR() {
  document.getElementById('qr-modal').classList.remove('active');
}

function copyPageURL() {
  copyToClipboard(window.location.href);
  showToast('URL copied!');
}

function copyToClipboard(text){if(navigator.clipboard&&window.isSecureContext){navigator.clipboard.writeText(text).catch(()=>{})}else{const ta=document.createElement('textarea');ta.value=text;ta.style.position='fixed';ta.style.left='-9999px';document.body.appendChild(ta);ta.select();document.execCommand('copy');document.body.removeChild(ta)}}

// --- Lightbox ---
let lightboxImages = [], lightboxIdx = 0;
let slideshowTimer = null;
function collectImages() {
  const imgExts = ['jpg', 'jpeg', 'png', 'gif', 'webp', 'svg'];
  const uniquePaths = new Set();
  lightboxImages = [];
  document.querySelectorAll('[data-path]').forEach(el => {
    const p = el.dataset.path || el.getAttribute('href')?.replace(/^\//, '');
    if (!p) return;
    const ext = (p.split('.').pop() || '').toLowerCase();
    if (imgExts.includes(ext) && !uniquePaths.has(p)) {
      uniquePaths.add(p);
      lightboxImages.push({ src: '/' + encodePath(p), path: p, info: p.split('/').pop() });
    }
  });
}
function openLightbox(src, info) {
  collectImages();
  lightboxIdx = lightboxImages.findIndex(img => img.src === src || window.location.origin + img.src === src);
  if (lightboxIdx === -1) lightboxIdx = 0;
  showLightbox();
}
function showLightbox() {
  const lb = document.getElementById('lightbox');
  if (!lb || !lightboxImages.length) return;
  const img = document.getElementById('lightbox-img');
  const infoEl = document.getElementById('lightbox-info');
  const thumbsEl = document.getElementById('lightbox-thumbs');
  img.src = lightboxImages[lightboxIdx].src;
  infoEl.textContent = lightboxImages[lightboxIdx].info + ' (' + (lightboxIdx + 1) + '/' + lightboxImages.length + ')' + (slideshowTimer ? ' [slideshow]' : '');
  if (thumbsEl.children.length !== lightboxImages.length) {
    thumbsEl.innerHTML = lightboxImages.map(function(item, i) {
      return '<img class="lightbox-thumb' + (i === lightboxIdx ? ' active' : '') + '" src="' + item.src + '" onclick="lightboxGoTo(' + i + ')" loading="lazy">';
    }).join('');
  } else {
    Array.from(thumbsEl.children).forEach(function(t, i) {
      t.classList.toggle('active', i === lightboxIdx);
    });
  }
  var activeThumb = thumbsEl.children[lightboxIdx];
  if (activeThumb) activeThumb.scrollIntoView({ block: 'nearest', inline: 'center' });
  lb.classList.add('active');
}
function lightboxGoTo(idx) { lightboxIdx = idx; showLightbox(); }
function lightboxNav(dir) {
  var next = lightboxIdx + dir;
  if (next >= 0 && next < lightboxImages.length) { lightboxIdx = next; showLightbox(); }
  else if (slideshowTimer) { stopSlideshow(); }
}
function toggleSlideshow() {
  if (slideshowTimer) { stopSlideshow(); }
  else {
    var lb = document.getElementById('lightbox');
    if (lb.requestFullscreen) lb.requestFullscreen();
    slideshowTimer = setInterval(function() { lightboxNav(1); }, 3000);
    showLightbox();
  }
}
function stopSlideshow() { clearInterval(slideshowTimer); slideshowTimer = null; showLightbox(); }
document.addEventListener('fullscreenchange', function() {
  if (!document.fullscreenElement && slideshowTimer) stopSlideshow();
});
function deleteLightboxCurrent() {
  if (!window.__CAN_DELETE__) { showToast('No delete permission', true); return; }
  if (!lightboxImages.length || lightboxIdx < 0) return;
  var img = lightboxImages[lightboxIdx];
  var name = img.info;
  showConfirm('Delete "' + name + '"?', function() {
    fetch('/_actions/delete', { method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify({ path: img.path }) }).then(function(res) {
      if (res.ok) {
        lightboxImages.splice(lightboxIdx, 1);
        if (lightboxImages.length === 0) { stopSlideshow(); document.getElementById('lightbox').classList.remove('active'); }
        else { if (lightboxIdx >= lightboxImages.length) lightboxIdx = lightboxImages.length - 1; showLightbox(); }
        collectImages();
      }
    });
  });
}

// --- Keyboard Shortcuts ---
const defaultKeys = {
  'nav_down': 'j',
  'nav_up': 'k',
  'nav_enter': 'l',
  'nav_parent': 'h',
  'select_all': 'a',
  'play_pause': 'P',
  'next': 'J',
  'prev': 'K',
  'seek_fwd': 'ArrowRight',
  'seek_back': 'ArrowLeft',
  'speed_up': '>',
  'speed_down': '<',
  'toggle_speed': 'S',
  'slideshow': 'P',
  'delete': 'd',
  'help': '?'
};
const keyLabels = {
  'nav_down': 'Next Row',
  'nav_up': 'Prev Row',
  'nav_enter': 'Open / Enter',
  'nav_parent': 'Parent Directory',
  'select_all': 'Select All',
  'search': 'Focus Search',
  'play_pause': 'Play / Pause',
  'next': 'Next Track / Image',
  'prev': 'Prev Track / Image',
  'seek_fwd': 'Seek Forward (+5s)',
  'seek_back': 'Seek Back (-5s)',
  'speed_up': 'Speed Up (>)',
  'speed_down': 'Speed Down (<)',
  'toggle_speed': 'Cycle Speed (S)',
  'slideshow': 'Slideshow (in lightbox)',
  'delete': 'Delete Selected',
  'help': 'Shortcuts Help'
};
let customKeys = JSON.parse(localStorage.getItem('cd_keys') || 'null') || {...defaultKeys};
let recordingKey = null;

// --- Log Viewer ---
function openLogViewer() {
  document.getElementById('log-modal').classList.add('active');
  loadLog('all', document.querySelector('.log-tab.active'));
}
function loadLog(category, btn) {
  document.querySelectorAll('.log-tab').forEach(t => { t.style.background='var(--button-bg)'; t.style.color='var(--text-color)'; t.classList.remove('active'); });
  if (btn) { btn.style.background='var(--track-color)'; btn.style.color='#fff'; btn.classList.add('active'); }
  const el = document.getElementById('log-content');
  el.textContent = 'Loading...';
  fetch('/api/logs?category='+encodeURIComponent(category)+'&_t='+Date.now(), {cache:'no-store'}).then(r=>r.json()).then(lines=>{
    el.textContent = lines.length ? lines.join('\n') : '';
    el.scrollTop = el.scrollHeight;
  }).catch(err=>{
    el.textContent = 'Error loading logs: ' + err;
  });
}
function clearLog() {
  fetch('/api/logs', {method:'DELETE'}).then(()=>{ document.getElementById('log-content').textContent='(no logs)'; });
}
function deleteLogFile() {
  showConfirm('Delete log file permanently?', function() {
    fetch('/api/logs/delete', {method:'POST'}).then(()=>{ document.getElementById('log-content').textContent='(no logs)'; showToast('Log deleted'); });
  });
}
function exportLog() {
  const text = document.getElementById('log-content').textContent;
  if (!text || text === 'Loading...' || text === '(no logs)') { showToast('No log content to export'); return; }
  const blob = new Blob([text], {type: 'text/plain'});
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = 'netdisk-log-' + new Date().toISOString().replace(/[:.]/g,'-').slice(0,19) + '.txt';
  a.click();
  URL.revokeObjectURL(url);
  showToast('Log exported');
}

// --- Recycle Bin ---
let recycleCache = [];
function escapeHtml(s){ return String(s).replace(/[&<>"']/g, function(c){ return {'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]; }); }
function openRecycleBin(){ document.getElementById('recycle-modal').classList.add('active'); loadRecycle(); }
function closeRecycleBin(){ document.getElementById('recycle-modal').classList.remove('active'); }
function loadRecycle(){
  const el = document.getElementById('recycle-content');
  el.innerHTML = '<div style="padding:20px;text-align:center;color:var(--text-muted)">Loading...</div>';
  fetch('/api/recycle', {cache:'no-store'}).then(r=>r.json()).then(items=>{
    recycleCache = Array.isArray(items) ? items : [];
    if (!recycleCache.length) {
      el.innerHTML = '<div style="padding:30px;text-align:center;color:var(--text-muted)">Recycle bin is empty</div>';
      updateRecycleSelCount();
      return;
    }
    let html = '';
    recycleCache.forEach(function(it, i){
      html += '<label class="recycle-row">'
        + '<input type="checkbox" class="recycle-check" data-idx="' + i + '" onchange="updateRecycleSelCount()">'
        + '<i data-lucide="' + (it.is_dir ? 'folder' : 'file') + '" style="width:16px;height:16px;flex-shrink:0;color:var(--text-muted)"></i>'
        + '<div style="flex:1;min-width:0"><div style="font-size:13px;white-space:nowrap;overflow:hidden;text-overflow:ellipsis">' + escapeHtml(it.name) + '</div>'
        + '<div style="font-size:11px;color:var(--text-muted);word-break:break-all">' + escapeHtml(it.path) + '</div></div>'
        + '<div style="font-size:11px;color:var(--text-muted);white-space:nowrap">' + (it.is_dir ? '--' : formatFileSizeJS(it.size)) + ' &middot; ' + escapeHtml(it.mtime) + '</div>'
        + '</label>';
    });
    el.innerHTML = html;
    if (window.lucide) lucide.createIcons({ root: el });
    updateRecycleSelCount();
  }).catch(e=>{ el.innerHTML = '<div style="padding:20px;color:var(--danger)">Error: ' + escapeHtml(e) + '</div>'; });
}
function recycleChecks(){ return Array.from(document.querySelectorAll('.recycle-check')); }
function selectedRecyclePaths(){
  return recycleChecks().filter(c=>c.checked).map(c=>recycleCache[+c.dataset.idx].path);
}
function updateRecycleSelCount(){
  const n = recycleChecks().filter(c=>c.checked).length;
  const el = document.getElementById('recycle-selcount');
  if (el) el.textContent = n + ' selected';
}
function toggleRecycleAll(){
  const checks = recycleChecks();
  const allChecked = checks.length > 0 && checks.every(c=>c.checked);
  checks.forEach(c=>{ c.checked = !allChecked; });
  updateRecycleSelCount();
}
function invertRecycleSelection(){
  recycleChecks().forEach(c=>{ c.checked = !c.checked; });
  updateRecycleSelCount();
}
// Generic multi-option dialog. options: [{label, value, style}] where style is
// 'primary' | 'danger' | 'plain'. Resolves to the chosen value, or null if cancelled.
function showChoice(title, message, options){
  return new Promise(function(resolve){
    const ov = document.getElementById('choice-modal');
    document.getElementById('choice-title').textContent = title;
    document.getElementById('choice-msg').textContent = message;
    const box = document.getElementById('choice-btns');
    box.innerHTML = '';
    function close(v){ ov.classList.remove('active'); resolve(v); }
    options.forEach(function(opt){
      const b = document.createElement('button');
      b.textContent = opt.label;
      b.className = 'choice-' + (opt.style || 'plain');
      b.onclick = function(){ close(opt.value); };
      box.appendChild(b);
    });
    ov.onclick = function(e){ if (e.target === ov) close(null); };
    ov.classList.add('active');
  });
}

// Ask the user how to resolve name conflicts (overwrite / rename / skip / cancel).
function askConflict(message){
  return showChoice('File already exists', message, [
    {label:'Overwrite', value:'overwrite', style:'danger'},
    {label:'Auto rename', value:'rename', style:'primary'},
    {label:'Skip', value:'skip'},
    {label:'Cancel', value:null},
  ]);
}

// Restore items, prompting for a conflict policy when destinations already exist.
function restorePaths(paths, onDone){
  fetch('/api/recycle/restore', {method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({paths:paths})})
    .then(function(r){ return r.json().then(function(d){ return {status:r.status, data:d}; }); })
    .then(function(res){
      if (res.status === 409 && res.data.conflicts) {
        askConflict(res.data.conflicts.length + ' destination(s) already exist. How should they be handled?')
          .then(function(choice){
            if (!choice) return;
            fetch('/api/recycle/restore', {method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({paths:paths, conflict:choice})})
              .then(r=>r.json()).then(onDone).catch(function(){ showToast('Restore failed', true); });
          });
        return;
      }
      onDone(res.data);
    })
    .catch(function(){ showToast('Restore failed', true); });
}
function restoreSelectedRecycle(){
  const paths = selectedRecyclePaths();
  if (!paths.length) { showToast('No items selected', true); return; }
  restorePaths(paths, function(d){
    showToast('Restored ' + ((d && d.restored) || 0));
    if (d && d.restored > 0) window.location.reload(); else loadRecycle();
  });
}
function deleteSelectedRecycle(){
  const paths = selectedRecyclePaths();
  if (!paths.length) { showToast('No items selected', true); return; }
  showConfirm('Delete ' + paths.length + ' item(s) permanently?', function(){
    fetch('/api/recycle/delete', {method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({paths:paths})})
      .then(()=>loadRecycle()).catch(()=>showToast('Delete failed', true));
  });
}
function clearRecycleBin(){
  showConfirm('Empty the recycle bin? This cannot be undone.', function(){
    fetch('/api/recycle/clear', {method:'POST'}).then(()=>loadRecycle()).catch(()=>showToast('Failed', true));
  });
}

// --- Keyboard Shortcuts ---
function openKeySettings() {
  const list = document.getElementById('key-list');
  list.innerHTML = '';
  var sections = [
    { title: 'File Navigation', keys: [
      ['j / k', 'Next / Prev row'],
      ['h', 'Parent directory'],
      ['l / Enter', 'Open / Smart preview'],
      ['Space', 'Toggle select'],
      ['a', 'Select all'],
      ['d', 'Delete selected'],
      ['s', 'Focus search'],
      ['?', 'This help']
    ]},
    { title: 'Media Controls', keys: [
      ['P / Space', 'Play / Pause'],
      ['J / K', 'Next / Prev track'],
      ['< / >', 'Speed down / up'],
      ['S', 'Cycle speed'],
      ['- / +', 'Volume down / up'],
      ['f', 'Fullscreen']
    ]},
    { title: 'Video Overlay', keys: [
      ['h / l', 'Rewind 5s / Forward 5s'],
      ['H / L', 'Rewind 60s / Forward 60s'],
      ['← / →', 'Seek -5s / +5s'],
      ['f', 'Fullscreen'],
      ['Escape', 'Close']
    ]},
    { title: 'Lightbox', keys: [
      ['j / k', 'Prev / Next image'],
      ['← / →', 'Prev / Next image'],
      ['P', 'Slideshow on/off'],
      ['d', 'Delete current image'],
      ['L', 'Close lightbox'],
      ['Escape', 'Close']
    ]}
  ];
  sections.forEach(function(sec) {
    var h = document.createElement('div');
    h.style.cssText = 'font-size:11px;font-weight:600;color:var(--link-color);text-transform:uppercase;letter-spacing:0.05em;padding:10px 0 4px;border-bottom:1px solid var(--border-color)';
    h.textContent = sec.title;
    list.appendChild(h);
    sec.keys.forEach(function(pair) {
      var row = document.createElement('div');
      row.className = 'key-row';
      row.innerHTML = '<span class="key-label">' + pair[1] + '</span><span class="key-bind" style="cursor:default">' + pair[0] + '</span>';
      list.appendChild(row);
    });
  });
  document.getElementById('key-modal').classList.add('active');
}
function formatKey(k) {
  const map = {' ':'Space','ArrowRight':'→','ArrowLeft':'←','ArrowDown':'↓','ArrowUp':'↑','Backspace':'⌫'};
  return map[k] || k;
}
function startRecord(el) {
  if (recordingKey) recordingKey.classList.remove('recording');
  recordingKey = el;
  el.classList.add('recording');
  el.textContent = 'Press a key...';
  const handler = function(e) {
    e.preventDefault();
    e.stopPropagation();
    el.classList.remove('recording');
    el.textContent = formatKey(e.key);
    customKeys[el.dataset.action] = e.key;
    document.removeEventListener('keydown', handler, true);
    recordingKey = null;
  };
  document.addEventListener('keydown', handler, true);
}
function saveKeys() {
  localStorage.setItem('cd_keys', JSON.stringify(customKeys));
  document.getElementById('key-modal').classList.remove('active');
  showToast('Shortcuts saved');
}
function resetKeys() {
  customKeys = {...defaultKeys};
  localStorage.removeItem('cd_keys');
  openKeySettings();
  showToast('Shortcuts reset');
}

// --- Keyboard file navigation ---
let focusedRowIndex = -1;
function getFileRows() {
  if (isGridView) {
    var grid = document.getElementById('grid-container');
    return grid ? Array.from(grid.querySelectorAll('.grid-item')) : [];
  }
  var searchArea = document.getElementById('search-results-area');
  if (searchArea && searchArea.style.display !== 'none') {
    return Array.from(document.querySelectorAll('#search-results-body tr.desktop-row'));
  }
  return Array.from(document.querySelectorAll('#file-list-table tbody tr.desktop-row'));
}
function getFirstVisibleRow() {
  var rows = getFileRows();
  for (var i = 0; i < rows.length; i++) {
    var rect = rows[i].getBoundingClientRect();
    if (rect.bottom > 0) return i;
  }
  return 0;
}
function updateFocusedRow() {
  var rows = getFileRows();
  rows.forEach(function(r) { r.classList.remove('file-row-focused'); });
  if (focusedRowIndex >= 0 && focusedRowIndex < rows.length) {
    rows[focusedRowIndex].classList.add('file-row-focused');
    rows[focusedRowIndex].scrollIntoView({ block: 'center' });
  }
}
function fileNav(dir) {
  var rows = getFileRows();
  if (!rows.length) return;
  if (focusedRowIndex < 0) focusedRowIndex = dir > 0 ? getFirstVisibleRow() : rows.length - 1;
  else focusedRowIndex = Math.max(0, Math.min(rows.length - 1, focusedRowIndex + dir));
  updateFocusedRow();
}
function enterFocusedItem() {
  var rows = getFileRows();
  if (focusedRowIndex < 0 || focusedRowIndex >= rows.length) return;
  var row = rows[focusedRowIndex];
  var isDir = row.dataset.isdir === 'true';
  var path = row.dataset.path;
  var name = row.dataset.name;
  if (isDir) { window.location.href = '/' + encodePath(path); }
  else {
    var fakeEvent = { stopPropagation: function(){} };
    handleSmartView(fakeEvent, path, name);
  }
}
function toggleFocusedCheck() {
  var rows = getFileRows();
  if (focusedRowIndex < 0 || focusedRowIndex >= rows.length) return;
  var row = rows[focusedRowIndex];
  var cb = row.querySelector('.file-cb');
  if (!cb) return;
  cb.checked = !cb.checked;
  if (isGridView) {
    if (cb.checked) row.classList.add('grid-checked');
    else row.classList.remove('grid-checked');
  }
  updateBatchBar();
}
function goParentDir() {
  var path = '{{.CurrentPath}}';
  if (!path) return;
  var parts = path.replace(/\/+$/, '').split('/');
  parts.pop();
  window.location.href = '/' + encodePath(parts.join('/'));
}

// --- Global keyboard handler ---
document.addEventListener('keydown', function(e) {
  if (e.key === ' ') {
    var ov = document.getElementById('modal-overlay');
    if (ov && ov.classList.contains('active')) {
      var inp = ov.querySelector('input[type=text]');
      if (inp && inp.style.display !== 'none') return;
      var okBtn = ov.querySelector('.modal-ok');
      if (okBtn) { e.preventDefault(); okBtn.click(); return; }
    }
  }
  if (document.activeElement && (document.activeElement.tagName === 'INPUT' || document.activeElement.tagName === 'TEXTAREA')) return;

  // Lightbox navigation
  const lb = document.getElementById('lightbox');
  const isLightboxActive = lb && lb.classList.contains('active');
  if (isLightboxActive) {
    if (e.key === 'Escape') { stopSlideshow(); lb.classList.remove('active'); e.preventDefault(); }
    else if (e.key === 'ArrowRight' || e.key === 'J' || e.key === 'j') { lightboxNav(1); e.preventDefault(); }
    else if (e.key === 'ArrowLeft' || e.key === 'K' || e.key === 'k') { lightboxNav(-1); e.preventDefault(); }
    else if (e.key === 'P' || e.key === 'p') { toggleSlideshow(); e.preventDefault(); }
    else if (e.key === 'L') { stopSlideshow(); lb.classList.remove('active'); e.preventDefault(); }
    else if (e.key === 'd') { deleteLightboxCurrent(); e.preventDefault(); }
    return;
  }

  // Video overlay navigation
  const vo = document.getElementById('video-overlay');
  if (vo && vo.classList.contains('active')) {
    if (e.key === 'Escape') { closeVideoOverlay(); e.preventDefault(); }
    else if (e.key === 'J') { playNext(); e.preventDefault(); }
    else if (e.key === 'K') { playPrev(); e.preventDefault(); }
    else if (e.key === 'P' || e.key === ' ') { if (activePlayer.paused) activePlayer.play(); else activePlayer.pause(); e.preventDefault(); }
    else if (e.key === 'f') { var vc = document.getElementById('video-player'); if (document.fullscreenElement) document.exitFullscreen(); else vc.requestFullscreen(); e.preventDefault(); }
    else if (e.key === 'h') { activePlayer.currentTime -= 5; e.preventDefault(); }
    else if (e.key === 'l') { activePlayer.currentTime += 5; e.preventDefault(); }
    else if (e.key === 'H') { activePlayer.currentTime -= 60; e.preventDefault(); }
    else if (e.key === 'L') { activePlayer.currentTime += 60; e.preventDefault(); }
    else if (e.key === 'ArrowLeft') { activePlayer.currentTime -= 5; e.preventDefault(); }
    else if (e.key === 'ArrowRight') { activePlayer.currentTime += 5; e.preventDefault(); }
    else if (e.key === '<') { currentPlaybackRate = Math.max(0.5, currentPlaybackRate - 0.5); activePlayer.playbackRate = currentPlaybackRate; speedBtn.textContent = currentPlaybackRate.toFixed(1) + 'x'; showSpeedOverlay(); e.preventDefault(); }
    else if (e.key === '>') { currentPlaybackRate = Math.min(3.0, currentPlaybackRate + 0.5); activePlayer.playbackRate = currentPlaybackRate; speedBtn.textContent = currentPlaybackRate.toFixed(1) + 'x'; showSpeedOverlay(); e.preventDefault(); }
    else if (e.key === '-') { adjustVolume(-0.1); e.preventDefault(); }
    else if (e.key === '=' || e.key === '+') { adjustVolume(0.1); e.preventDefault(); }
    return;
  }

  // Escape: close modals
  if (e.key === 'Escape') {
    var km = document.getElementById('key-modal');
    if (km && km.classList.contains('active')) { km.classList.remove('active'); e.preventDefault(); return; }
    var lm = document.getElementById('log-modal');
    if (lm && lm.classList.contains('active')) { lm.classList.remove('active'); e.preventDefault(); return; }
    var mo = document.getElementById('modal-overlay');
    if (mo && mo.classList.contains('active')) { mo.classList.remove('active'); e.preventDefault(); return; }
    return;
  }

  // --- Uppercase: media controls ---
  if (e.key === 'J') { playNext(); e.preventDefault(); return; }
  if (e.key === 'K') { playPrev(); e.preventDefault(); return; }
  if (e.key === 'P') { if (activePlayer.src) { if (activePlayer.paused) activePlayer.play(); else activePlayer.pause(); } e.preventDefault(); return; }
  if (e.key === 'S') { speedBtn.click(); e.preventDefault(); return; }

  // --- Lowercase: file navigation ---
  if (e.key === 'j') { fileNav(1); e.preventDefault(); return; }
  if (e.key === 'k') { fileNav(-1); e.preventDefault(); return; }
  if (e.key === 'h') { goParentDir(); e.preventDefault(); return; }
  if (e.key === 'l' || e.key === 'Enter') { enterFocusedItem(); e.preventDefault(); return; }
  if (e.key === ' ') { toggleFocusedCheck(); e.preventDefault(); return; }
  if (e.key === 'a') { toggleSelectAll(); e.preventDefault(); return; }
  if (e.key === 'd') { if (getSelectedPaths().length > 0) deleteSelected(); e.preventDefault(); return; }
  if (e.key === 's') { var si = document.getElementById('search-input'); if (si) { si.focus(); si.select(); } e.preventDefault(); return; }
  if (e.key === '<') { currentPlaybackRate = Math.max(0.5, currentPlaybackRate - 0.5); activePlayer.playbackRate = currentPlaybackRate; speedBtn.textContent = currentPlaybackRate.toFixed(1) + 'x'; showSpeedOverlay(); e.preventDefault(); return; }
  if (e.key === '>') { currentPlaybackRate = Math.min(3.0, currentPlaybackRate + 0.5); activePlayer.playbackRate = currentPlaybackRate; speedBtn.textContent = currentPlaybackRate.toFixed(1) + 'x'; showSpeedOverlay(); e.preventDefault(); return; }
  if (e.key === '-') { adjustVolume(-0.1); e.preventDefault(); return; }
  if (e.key === '=' || e.key === '+') { adjustVolume(0.1); e.preventDefault(); return; }
  if (e.key === '?') { openKeySettings(); e.preventDefault(); return; }
  if (e.key === 'Backspace') { goParentDir(); e.preventDefault(); return; }
});

// --- Delete ---
function deleteItem(event, path, name) {
  event.stopPropagation();
  showConfirm('Delete "'+name+'"?', function(){
    fetch('/_actions/delete', { method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify({ path: path }) }).then(res => {
      if(res.ok) { removeRowByPath(path); }
      else { alert("Delete failed."); }
    });
  });
}
function removeRowByPath(path){document.querySelectorAll('tr.desktop-row').forEach(r=>{if(r.dataset.path===path)r.remove()});updateBatchBar();}
function deleteSelected() {
  if (!window.__CAN_DELETE__) { showToast('No delete permission', true); return; }
  const paths = getSelectedPaths();
  if(!paths.length) return;
  showConfirm('Delete '+paths.length+' items to recycle bin?', function(){
    fetch('/_actions/batch_delete', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ paths }) }).then(res => {
      if(res.ok) { paths.forEach(p => removeRowByPath(p)); }
      else { alert("Batch delete failed."); }
    });
  });
}
function createNewFile() { showPrompt('Filename:', '', function(n){ if(n) fetch('/_actions/new_file', { method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify({ path: '{{.CurrentPath}}', filename: n }) }).then(()=>window.location.reload()); }); }
function createNewDirectory() { showPrompt('Folder:', '', function(n){ if(n) fetch('/_actions/new_directory', { method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify({ path: '{{.CurrentPath}}', dirname: n }) }).then(()=>window.location.reload()); }); }
function toggleHidden() {
  var url = new URL(window.location);
  if (url.searchParams.get('show_hidden') === '1') { url.searchParams.delete('show_hidden'); }
  else { url.searchParams.set('show_hidden', '1'); }
  window.location.href = url.toString();
}

// --- Sort ---
let currentSortBy = '{{.SortBy}}';
let currentOrder = '{{.Order}}';
function sortBy(field) {
  if (field === currentSortBy) {
    currentOrder = currentOrder === 'asc' ? 'desc' : 'asc';
  } else {
    currentSortBy = field;
    currentOrder = 'asc';
  }
  clientSort();
  updateSortHighlight();
  document.querySelectorAll('.new-btn-wrapper.dropdown.open').forEach(d => d.classList.remove('open'));
}
function clientSort() {
  const tbody = document.querySelector('#file-list-table tbody');
  if (!tbody) return;
  const rows = Array.from(tbody.querySelectorAll('tr.desktop-row'));
  if (!rows.length) return;
  const dirs = rows.filter(r => r.dataset.isdir === 'true');
  const files = rows.filter(r => r.dataset.isdir !== 'true');
  const desc = currentOrder === 'desc';
  function comparator(a, b) {
    let va, vb;
    switch (currentSortBy) {
      case 'size':
        va = parseInt(a.dataset.size) || 0;
        vb = parseInt(b.dataset.size) || 0;
        return desc ? vb - va : va - vb;
      case 'mtime':
        va = parseInt(a.dataset.mtime) || 0;
        vb = parseInt(b.dataset.mtime) || 0;
        return desc ? vb - va : va - vb;
      default:
        va = (a.dataset.name || '').toLowerCase();
        vb = (b.dataset.name || '').toLowerCase();
        return desc ? vb.localeCompare(va) : va.localeCompare(vb);
    }
  }
  dirs.sort(comparator);
  files.sort(comparator);
  tbody.innerHTML = '';
  dirs.concat(files).forEach(r => tbody.appendChild(r));
  lucide.createIcons();
  if (isGridView) applyView();
}
function updateSortHighlight() {
  document.querySelectorAll('.sort-item').forEach(function(el) {
    var existing = el.querySelector('.sort-arrow');
    if (existing) existing.remove();
    el.style.color = '';
    if (el.dataset.sort === currentSortBy) {
      var s = document.createElement('span');
      s.className = 'sort-arrow';
      s.style.cssText = 'margin-left:auto;font-size:11px;opacity:0.7';
      s.textContent = currentOrder === 'asc' ? '\u25B2' : '\u25BC';
      el.appendChild(s);
      el.style.color = 'var(--link-color)';
    }
  });
  var btn = document.getElementById('sort-btn');
  if (btn) {
    var svg = btn.querySelector('svg');
    if (svg) {
      svg.outerHTML = '<i data-lucide="' + (currentOrder === 'asc' ? 'arrow-up' : 'arrow-down') + '" style="width:16px;height:16px"></i>';
      lucide.createIcons();
    }
  }
}
updateSortHighlight();
</script>
</body>
</html>`

const WEBCLIP_HTML = `<!DOCTYPE html>
<html lang="en" data-theme="dark">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0, user-scalable=no">
<title>webclip</title>
<script src="/static/lucide.min.js"></script>
<style>
:root {
  --main-bg: rgba(30,30,46,0.72);
  --secondary-bg: #1e1e2e;
  --text-color: #cdd6f4;
  --text-muted: #a6adc8;
  --border-color: #45475a;
  --link-color: #89b4fa;
  --hover-bg: rgba(205,214,244,0.06);
  --glass-bg: rgba(30,30,46,0.55);
  --glass-border: rgba(205,214,244,0.08);
  --glass-shadow: 0 8px 32px rgba(0,0,0,0.3);
  --accent: #89b4fa;
  --danger: #f38ba8;
  --success: #a6e3a1;
}
* { box-sizing: border-box; }
body {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
  margin: 0;
  padding: 20px;
  background: var(--secondary-bg);
  color: var(--text-color);
  min-height: 100vh;
}
.container {
  width: 96%;
  max-width: 1800px;
  margin: 0 auto;
}

/* ===== BREADCRUMBS ===== */
.topbar {
  display: flex;
  align-items: center;
  gap: 8px 24px;
  margin-bottom: 12px;
  flex-wrap: wrap;
}
.breadcrumbs {
  display: flex;
  align-items: center;
  gap: 2px;
  font-size: 0.95rem;
}
.breadcrumbs a {
  text-decoration: none;
  color: var(--text-color);
  padding: 4px 6px;
  border-radius: 6px;
  transition: background .15s;
}
.breadcrumbs a:hover { background: var(--hover-bg); }
.breadcrumbs .sep { color: var(--text-muted); margin: 0 1px; }

/* ===== INPUT CARD ===== */
.input-card {
  background: var(--glass-bg);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border: 1px solid var(--glass-border);
  border-radius: 10px;
  padding: 14px;
  margin-bottom: 12px;
}
.input-card textarea {
  width: 100%;
  min-height: 100px;
  padding: 10px 12px;
  border: 1px solid rgba(255,255,255,0.08);
  background: rgba(0,0,0,0.2);
  color: var(--text-color);
  border-radius: 8px;
  font-family: inherit;
  font-size: 14px;
  line-height: 1.5;
  resize: vertical;
  outline: none;
  transition: border-color .15s, box-shadow .15s;
}
.input-card textarea:focus {
  border-color: var(--accent);
  box-shadow: 0 0 0 3px rgba(137,180,250,0.2);
}
.input-card textarea::placeholder { color: var(--text-muted); }

/* ===== BUTTONS ===== */
.btn-row {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}
.btn-spacer { flex: 1 1 0; }
.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  height: 36px;
  padding: 0 16px;
  border: 1px solid var(--border-color);
  background: var(--glass-bg);
  color: var(--text-color);
  border-radius: 10px;
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  transition: all .15s;
  white-space: nowrap;
}
.btn i { width: 15px; height: 15px; }
.btn:hover { background: var(--hover-bg); border-color: rgba(255,255,255,0.15); }
.btn-primary {
  background: var(--accent);
  border-color: var(--accent);
  color: #fff;
}
.btn-primary:hover { opacity: 0.85; background: var(--accent); }
.btn-danger { border-color: var(--danger); color: var(--danger); }
.btn-danger:hover { background: rgba(239,68,68,0.15); }
.btn-icon { width: 36px; padding: 0; }
.btn-success { border-color: var(--success); color: var(--success); }
.btn-success:hover { background: rgba(34,197,94,0.15); }
.btn-spacer { margin-left: auto; }

/* ===== CLIP CARDS ===== */
#card-list { display: flex; flex-direction: column; gap: 6px; }
.clip-card {
  background: var(--glass-bg);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  border: 1px solid var(--glass-border);
  border-radius: 10px;
  padding: 10px 14px;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: background .15s, border-color .15s, transform .15s;
  overflow: hidden;
}
.clip-card:hover {
  background: rgba(255,255,255,0.08);
  border-color: rgba(255,255,255,0.12);
}
.clip-card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
  font-size: 12px;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}
.clip-card-header i { width: 14px; height: 14px; opacity: 0.6; }
.clip-card-body {
  font-size: 13px;
  line-height: 1.5;
  color: var(--text-color);
  word-break: break-word;
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
}
.clip-card-body a { color: var(--link-color); text-decoration: none; }
.clip-card-body a:hover { text-decoration: underline; }
.clip-card-file {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  color: var(--link-color);
  text-decoration: none;
  padding: 4px 10px;
  background: rgba(137,180,250,0.1);
  border-radius: 8px;
  transition: background .15s;
}
.clip-card-file:hover { background: rgba(137,180,250,0.2); text-decoration: none; }
.clip-card-file i { width: 16px; height: 16px; }
.clip-card-actions {
  display: flex;
  gap: 4px;
  flex-shrink: 0;
  opacity: 0;
  transition: opacity .15s;
}
.clip-card:hover .clip-card-actions { opacity: 1; }
.ghost-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border: none;
  background: transparent;
  color: var(--text-muted);
  border-radius: 8px;
  cursor: pointer;
  transition: all .15s;
  flex-shrink: 0;
}
.ghost-btn i { width: 18px; height: 18px; }
.ghost-btn:hover { background: var(--hover-bg); color: var(--text-color); }
.ghost-btn.danger:hover { background: rgba(239,68,68,0.15); color: var(--danger); }

/* ===== DROP OVERLAY ===== */
.drop-overlay {
  display: none;
  position: fixed;
  top: 0; left: 0; right: 0; bottom: 0;
  z-index: 9998;
  background: rgba(0,0,0,0.6);
  backdrop-filter: blur(4px);
  -webkit-backdrop-filter: blur(4px);
  justify-content: center;
  align-items: center;
  pointer-events: none;
}
.drop-overlay.active { display: flex; }
.drop-overlay-inner {
  text-align: center;
  color: #fff;
  font-size: 1.2em;
  font-weight: 500;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}
.drop-overlay-inner i { opacity: 0.8; }

/* ===== TOAST ===== */
.toast-msg {
  position: fixed;
  top: 20px;
  left: 50%;
  transform: translateX(-50%);
  background: var(--success);
  color: #fff;
  padding: 8px 20px;
  border-radius: 10px;
  font-size: 14px;
  z-index: 10001;
  opacity: 0;
  transition: opacity .3s;
  pointer-events: none;
}
.toast-msg.show { opacity: 1; }

/* ===== EMPTY STATE ===== */
.empty-state {
  text-align: center;
  padding: 36px 16px;
  color: var(--text-muted);
}
.empty-state i { width: 48px; height: 48px; opacity: 0.3; margin-bottom: 12px; }

/* ===== CONFIRM MODAL ===== */
.modal-overlay{display:none;position:fixed;top:0;left:0;width:100%;height:100%;background:rgba(0,0,0,0.6);z-index:10000;justify-content:center;align-items:center}
.modal-overlay.active{display:flex}
.modal-box{background:var(--glass-bg);backdrop-filter:blur(20px);-webkit-backdrop-filter:blur(20px);border:1px solid var(--glass-border);border-radius:10px;padding:24px;min-width:320px;max-width:90vw;box-shadow:0 8px 32px rgba(0,0,0,0.4)}
.modal-box h3{margin:0 0 16px;font-size:1.1em}
.modal-box p{margin:0 0 20px;font-size:14px;color:var(--text-muted);line-height:1.5}
.modal-btns{display:flex;gap:8px;justify-content:flex-end}
.modal-btns button{padding:8px 18px;border-radius:8px;border:none;cursor:pointer;font-size:14px;font-weight:500;transition:all .15s}
.modal-btns .btn-ok{background:var(--danger);color:#fff}
.modal-btns .btn-ok:hover{opacity:.85}
.modal-btns .btn-cancel{background:var(--button-bg);color:var(--text-color);border:1px solid var(--border-color)}
.modal-btns .btn-cancel:hover{background:var(--hover-bg)}

/* ===== LIGHTBOX ===== */
.wc-lightbox{display:none;position:fixed;top:0;left:0;width:100%;height:100%;background:rgba(0,0,0,0.92);z-index:10000;justify-content:center;align-items:center}
.wc-lightbox.active{display:flex}
.wc-lightbox-img{max-width:85vw;max-height:85vh;object-fit:contain;border-radius:6px;user-select:none;-webkit-user-drag:none}
.wc-lightbox-close{position:absolute;top:16px;right:20px;width:40px;height:40px;border:none;background:rgba(255,255,255,0.15);color:#fff;border-radius:50%;cursor:pointer;font-size:20px;display:flex;align-items:center;justify-content:center;transition:background .15s;z-index:2}
.wc-lightbox-close:hover{background:rgba(255,255,255,0.3)}
.wc-nav{position:absolute;top:50%;transform:translateY(-50%);width:48px;height:48px;border:none;background:rgba(255,255,255,0.12);color:#fff;border-radius:50%;cursor:pointer;font-size:22px;display:flex;align-items:center;justify-content:center;transition:background .15s;z-index:2}
.wc-nav:hover{background:rgba(255,255,255,0.3)}
.wc-nav-prev{left:16px}
.wc-nav-next{right:16px}
.wc-nav.hidden{display:none}
.wc-lightbox-count{position:absolute;bottom:20px;left:50%;transform:translateX(-50%);color:rgba(255,255,255,0.6);font-size:14px;font-family:monospace;z-index:2}

/* ===== MOBILE ===== */
@media(max-width: 640px) {
  body { padding: 10px; }
  .input-card { padding: 10px; }
  .btn { height: 38px; padding: 0 12px; font-size: 13px; }
  .clip-card { padding: 8px 10px; flex-wrap: wrap; }
  .clip-card-body { flex-basis: 100%; }
  .clip-card-actions { opacity: 1; }
}
</style>
</head>
<body>
<div class="container">
  <div class="topbar">
    <div class="breadcrumbs">
      <a href="/"><i data-lucide="home" style="width:16px;height:16px"></i></a>
      <span class="sep">/</span>
      <span style="color:var(--text-muted)">webclip</span>
    </div>
  </div>

  <div class="input-card">
    <form id="text-form">
      <textarea id="text-input" placeholder="Paste text, notes, links..." rows="5"></textarea>
      <div class="btn-row" style="margin-top:14px">
        <input type="file" id="file-input" multiple onchange="handleFiles(this.files)" style="display:none">
        <button type="button" class="btn btn-icon" onclick="document.getElementById('file-input').click()"><i data-lucide="upload"></i></button>
        <button type="submit" class="btn btn-primary btn-icon"><i data-lucide="send"></i></button>
        <span class="btn-spacer"></span>
        <button type="button" class="btn btn-success btn-icon" onclick="copyAll()"><i data-lucide="copy"></i></button>
        <button type="button" class="btn btn-danger btn-icon" onclick="clearAll()"><i data-lucide="trash-2"></i></button>
      </div>
    </form>
  </div>

  <div id="card-list"></div>
  <div id="empty-state" class="empty-state" style="display:none">
    <i data-lucide="clipboard-list"></i>
    <div>No clips yet</div>
  </div>
</div>

<div id="wc-lightbox" class="wc-lightbox" onclick="if(event.target===this)closeWcLightbox()">
  <button class="wc-lightbox-close" onclick="closeWcLightbox()">&times;</button>
  <button class="wc-nav wc-nav-prev" id="wc-nav-prev" onclick="wcNav(-1)">&#8249;</button>
  <button class="wc-nav wc-nav-next" id="wc-nav-next" onclick="wcNav(1)">&#8250;</button>
  <img id="wc-lightbox-img" class="wc-lightbox-img" src="" alt="">
  <div id="wc-lightbox-count" class="wc-lightbox-count"></div>
</div>

<div id="confirm-modal" class="modal-overlay">
  <div class="modal-box">
    <h3 id="confirm-title">Confirm</h3>
    <p id="confirm-msg"></p>
    <div class="modal-btns">
      <button class="btn-cancel" onclick="closeConfirm(false)">Cancel</button>
      <button class="btn-ok" onclick="closeConfirm(true)">OK</button>
    </div>
  </div>
</div>

<div id="toast" class="toast-msg">Copied!</div>

<script>
lucide.createIcons();
const API='/wc/api/cards';
let currentCards=[];
function wcFileUrl(name){return '/wc/file/'+encodeURIComponent(name)}

function showToast(msg){const t=document.getElementById('toast');t.textContent=msg;t.classList.add('show');setTimeout(()=>t.classList.remove('show'),1500)}
function linkify(text){return text.replace(/(https?:\/\/[^\s<]+)/g,'<a href="$1" target="_blank" rel="noopener">$1</a>')}
function copyToClipboard(text){if(navigator.clipboard&&window.isSecureContext){navigator.clipboard.writeText(text).then(()=>showToast('Copied!')).catch(()=>{fallbackCopy(text)})}else{fallbackCopy(text);showToast('Copied!')}}
function fallbackCopy(text){const ta=document.createElement("textarea");ta.value=text;ta.style.position="fixed";ta.style.left="-9999px";document.body.appendChild(ta);ta.focus();ta.select();try{document.execCommand('copy')}catch(e){}document.body.removeChild(ta);showToast('Copied!')}

let confirmResolve=null;
function showConfirm(title,msg){return new Promise(resolve=>{confirmResolve=resolve;document.getElementById('confirm-title').textContent=title;document.getElementById('confirm-msg').textContent=msg;document.getElementById('confirm-modal').classList.add('active')})}
function closeConfirm(val){document.getElementById('confirm-modal').classList.remove('active');if(confirmResolve){confirmResolve(val);confirmResolve=null}}

let wcImages=[],wcIdx=0;
function openWcLightbox(idx){wcImages=currentCards.filter(c=>c.type==='image');wcIdx=idx;const c=wcImages[wcIdx];document.getElementById('wc-lightbox-img').src=wcFileUrl(c.filename);document.getElementById('wc-lightbox').classList.add('active');wcUpdateNav()}
function closeWcLightbox(){document.getElementById('wc-lightbox').classList.remove('active')}
function wcNav(dir){wcIdx+=dir;const c=wcImages[wcIdx];document.getElementById('wc-lightbox-img').src=wcFileUrl(c.filename);wcUpdateNav()}
function wcUpdateNav(){document.getElementById('wc-nav-prev').classList.toggle('hidden',wcIdx<=0);document.getElementById('wc-nav-next').classList.toggle('hidden',wcIdx>=wcImages.length-1);document.getElementById('wc-lightbox-count').textContent=(wcIdx+1)+'/'+wcImages.length}
document.addEventListener('keydown',e=>{if(!document.getElementById('wc-lightbox').classList.contains('active'))return;if(e.key==='Escape')closeWcLightbox();if(e.key==='ArrowLeft'&&wcIdx>0)wcNav(-1);if(e.key==='ArrowRight'&&wcIdx<wcImages.length-1)wcNav(1)});

function isImageFile(name){return /\.(jpg|jpeg|png|gif|webp|svg|bmp|ico)$/i.test(name)}

async function load(){
  const r=await fetch(API);
  currentCards=await r.json();
  const list=document.getElementById('card-list');
  const empty=document.getElementById('empty-state');
  list.innerHTML='';
  if(!currentCards.length){empty.style.display='';lucide.createIcons();return}
  empty.style.display='none';
  currentCards.forEach(c=>{
    const d=document.createElement('div');
    d.className='clip-card';
    const body=document.createElement('div');
    body.className='clip-card-body';
    if(c.type==='text'){
      body.innerHTML=linkify(c.content);
    }else if(c.type==='image'){
      const imgIdx=currentCards.filter(x=>x.type==='image').indexOf(c);
      const wrap=document.createElement('div');
      wrap.style.cssText='cursor:pointer;display:inline-flex;align-items:center;gap:8px;color:var(--link-color);text-decoration:none;padding:4px 10px;background:rgba(137,180,250,0.1);border-radius:8px;transition:background .15s';
      wrap.innerHTML='<i data-lucide="image"></i>'+c.original_name;
      wrap.onmouseenter=()=>wrap.style.background='rgba(137,180,250,0.2)';
      wrap.onmouseleave=()=>wrap.style.background='rgba(137,180,250,0.1)';
      wrap.onclick=()=>openWcLightbox(imgIdx);
      body.appendChild(wrap);
    }else{
      const a=document.createElement('a');
      a.href=wcFileUrl(c.filename);
      a.target='_blank';
      a.rel='noopener';
      a.className='clip-card-file';
      a.innerHTML='<i data-lucide="download"></i>'+c.original_name;
      body.appendChild(a);
    }
    d.appendChild(body);
    const actions=document.createElement('div');
    actions.className='clip-card-actions';
    const cp=document.createElement('button');
    cp.className='ghost-btn';
    cp.title='Copy';
    cp.innerHTML='<i data-lucide="copy"></i>';
    cp.onclick=()=>{const val=c.type==='text'?c.content:(window.location.origin+wcFileUrl(c.filename));copyToClipboard(val)};
    const del=document.createElement('button');
    del.className='ghost-btn danger';
    del.title='Delete';
    del.innerHTML='<i data-lucide="trash-2"></i>';
    del.onclick=()=>fetch(API+'/'+c.id,{method:'DELETE'}).then(load);
    actions.appendChild(cp);
    actions.appendChild(del);
    d.appendChild(actions);
    list.appendChild(d);
  });
  lucide.createIcons();
}

document.getElementById('text-form').onsubmit=async(e)=>{
  e.preventDefault();
  const t=document.getElementById('text-input').value.trim();
  if(!t)return;
  const f=new FormData();
  f.append('text',t);
  await fetch(API,{method:'POST',body:f});
  document.getElementById('text-input').value='';
  load();
};
async function handleFiles(files){
  const f=new FormData();
  Array.from(files).forEach(x=>f.append('file',x));
  await fetch(API,{method:'POST',body:f});
  load();
}
async function clearAll(){
  if(!currentCards.length)return;
  const ok=await showConfirm('Clear All','Are you sure you want to delete all clips? This cannot be undone.');
  if(ok)fetch(API+'/clear',{method:'POST'}).then(load);
}
function copyAll(){
  if(!currentCards.length)return;
  const allText=currentCards.map(c=>c.type==='text'?c.content:(window.location.origin+wcFileUrl(c.filename))).join('\n\n');
  copyToClipboard(allText);
}
load();
</script>
</body>
</html>`

const PREVIEW_HTML = `<!DOCTYPE html>
<html lang="en" data-theme="dark">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>{{.Filename}} - Preview</title>
<link rel="stylesheet" href="/static/hljs-dark.min.css">
<script src="/static/lucide.min.js"></script>
<style>
:root{--bg:#1e1e2e;--text:#cdd6f4;--border:#45475a;--code-bg:#181825}
body{font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,sans-serif;margin:0;padding:20px;background:var(--bg);color:var(--text)}
.container{width:96%;max-width:none;margin:0 auto;transition:max-width .2s}
.container.narrow{max-width:900px}
.header{display:flex;justify-content:space-between;align-items:center;padding-bottom:10px;border-bottom:1px solid var(--border);position:sticky;top:0;z-index:10;background:var(--bg);padding-top:10px}
.header-actions{display:flex;gap:8px;align-items:center}
.header-actions button,.header-actions a{height:28px;padding:0 12px;border:1px solid #89b4fa;color:#89b4fa;background:none;border-radius:6px;cursor:pointer;font-size:0.85em;text-decoration:none;transition:all .15s;display:inline-flex;align-items:center;justify-content:center;gap:5px;line-height:1}
.header-actions button:hover,.header-actions a:hover{background:rgba(137,180,250,0.15)}
.filename{font-size:1.1em;font-weight:600;display:flex;align-items:center;gap:6px}
.filename i{opacity:.5}
@media(max-width:768px){
.header{flex-direction:column;align-items:flex-start;gap:8px}
.header-actions{width:100%;justify-content:flex-end}
}
.md-content{line-height:1.7;font-size:16px}
.md-content h1,.md-content h2,.md-content h3{margin-top:1.5em;margin-bottom:0.5em}
.md-content code{background:#313244;padding:2px 6px;border-radius:4px;font-size:0.9em}
.md-content pre{background:var(--code-bg);padding:16px;border-radius:8px;overflow-x:auto}
.md-content pre code{background:none;padding:0}
.md-content a{color:#89b4fa}
.md-content blockquote{border-left:4px solid #89b4fa;padding-left:16px;margin:16px 0;color:#a6adc8}
.md-content table{border-collapse:collapse;width:100%;margin:16px 0}
.md-content th,.md-content td{border:1px solid var(--border);padding:8px 12px;text-align:left}
.md-content th{background:#313244}
.code-content{background:var(--code-bg);padding:16px;border-radius:8px;overflow-x:auto;font-family:monospace;font-size:14px;line-height:1.5}
.code-content pre{margin:0;background:none;padding:0}
</style>
</head>
<body>
<div class="container" id="preview-container">
<div class="header">
<h1 style="margin:0;font-size:1.3em"><i data-lucide="file-text" style="width:18px;height:18px;opacity:.5"></i> {{.Path}}</h1>
<div class="header-actions">
<button class="toggle-btn" onclick="toggleWidth()" title="Toggle width"><i data-lucide="maximize-2" style="width:14px;height:14px"></i></button>
<a href="/edit/{{.Path | urlpath}}"><i data-lucide="pencil" style="width:14px;height:14px"></i> Edit</a>
<button onclick="window.close()"><i data-lucide="x" style="width:14px;height:14px"></i> Close</button>
</div>
</div>
{{if .IsMarkdown}}
<div class="md-content" id="md-content"></div>
{{else if .IsCode}}
<div class="code-content"><pre><code class="language-{{.Language}}">{{.Content}}</code></pre></div>
{{else}}
<pre>{{.Content}}</pre>
{{end}}
</div>
<script src="/static/marked.min.js"></script>
<script src="/static/highlight.min.js"></script>
<script>
lucide.createIcons();
{{if .IsMarkdown}}
document.getElementById('md-content').innerHTML = marked.parse({{.Content}});
{{end}}
document.querySelectorAll('pre code').forEach(el => hljs.highlightElement(el));
function toggleWidth(){
  const c=document.getElementById('preview-container');
  c.classList.toggle('narrow');
  const isNarrow=c.classList.contains('narrow');
  localStorage.setItem('cd_preview_width',isNarrow?'narrow':'wide');
  const btn=document.querySelector('.toggle-btn i');
  if(btn){btn.setAttribute('data-lucide',isNarrow?'minimize-2':'maximize-2');lucide.createIcons();}
}
if(localStorage.getItem('cd_preview_width')==='narrow'){document.getElementById('preview-container').classList.add('narrow');const btn=document.querySelector('.toggle-btn i');if(btn){btn.setAttribute('data-lucide','minimize-2');lucide.createIcons();}}
</script>
</body>
</html>`
