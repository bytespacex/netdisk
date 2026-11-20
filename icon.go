package main

import "net/http"

// iconSVG is the built-in favicon (also served at /icon.svg for other programs).
const iconSVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64">
<path d="M6 17a5 5 0 0 1 5-5h13l6 7h23a5 5 0 0 1 5 5v26a5 5 0 0 1-5 5H11a5 5 0 0 1-5-5V17z" fill="#58a6ff"/>
<path d="M6 24h52v5H6z" fill="#3b82c4" opacity=".55"/>
</svg>`

func iconHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	_, _ = w.Write([]byte(iconSVG))
}
