package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type WebClipCard struct {
	ID           string  `json:"id"`
	Timestamp    float64 `json:"timestamp"`
	Type         string  `json:"type"`
	Content      string  `json:"content"`
	Filename     string  `json:"filename,omitempty"`
	OriginalName string  `json:"original_name,omitempty"`
}

func webclipEnsureDirs() {
	os.MkdirAll(webclipDir, 0755)
	os.MkdirAll(webclipFiles, 0755)
}

func webclipHomeHandler(w http.ResponseWriter, r *http.Request) {
	executeTemplate(w, webclipTmpl, nil)
}

func webclipGetFileHandler(w http.ResponseWriter, r *http.Request) {
	webclipEnsureDirs()
	filename := strings.TrimPrefix(r.URL.Path, "/wc/file/")
	if !validNameComponent(filename) {
		http.NotFound(w, r)
		return
	}
	filePath := filepath.Join(webclipFiles, filename)
	if !pathInsideVolume(webclipFiles, filePath) {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, filePath)
}

func webclipCardsAPIHandler(w http.ResponseWriter, r *http.Request) {
	webclipEnsureDirs()

	if r.Method == http.MethodGet {
		entries, err := os.ReadDir(webclipDir)
		if err != nil {
			json.NewEncoder(w).Encode([]WebClipCard{})
			return
		}

		var cardFiles []string
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
				cardFiles = append(cardFiles, e.Name())
			}
		}
		sort.Sort(sort.Reverse(sort.StringSlice(cardFiles)))

		var cards []WebClipCard
		for _, fn := range cardFiles {
			data, err := os.ReadFile(filepath.Join(webclipDir, fn))
			if err != nil {
				continue
			}
			var card WebClipCard
			if json.Unmarshal(data, &card) == nil {
				cards = append(cards, card)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cards)
		return
	}

	if r.Method == http.MethodPost {
		// Stream the request body to avoid ParseMultipartForm caching large files to /tmp
		mr, err := r.MultipartReader()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{"status": "error", "message": err.Error()})
			return
		}

		var results []WebClipCard
		var textContent string

		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				writeLog("webclip", fmt.Sprintf("(%s) STREAM-ERROR ctxErr=%v err=%v", clientIP(r), r.Context().Err(), err))
				break
			}

			switch part.FormName() {
			case "text":
				b, _ := io.ReadAll(io.LimitReader(part, 1<<20))
				textContent = string(b)
				part.Close()

			case "file":
				filename := part.FileName()
				if filename == "" {
					part.Close()
					continue
				}
				cardID := fmt.Sprintf("%.5f_%s", float64(time.Now().UnixNano())/1e9, randomHex(2))
				originalName := getSafeName(filename)
				ext := filepath.Ext(originalName)
				base := strings.TrimSuffix(originalName, ext)
				uniqueFilename := fmt.Sprintf("%s_%d%s", base, time.Now().Unix(), ext)
				savePath := filepath.Join(webclipFiles, uniqueFilename)

				dst, err := os.Create(savePath)
				if err != nil {
					writeLog("webclip", fmt.Sprintf("(%s) CREATE-FAILED %s: %v", clientIP(r), savePath, err))
					part.Close()
					continue
				}
				_, copyErr := io.Copy(dst, part)
				closeErr := dst.Close()
				ct := part.Header.Get("Content-Type")
				part.Close()
				if copyErr != nil || closeErr != nil {
					os.Remove(savePath)
					writeLog("webclip", fmt.Sprintf("(%s) COPY-ERROR %s: copyErr=%v closeErr=%v ctxErr=%v",
						clientIP(r), savePath, copyErr, closeErr, r.Context().Err()))
					continue
				}

				cardType := "file"
				if strings.HasPrefix(ct, "image/") {
					cardType = "image"
				}

				card := WebClipCard{
					ID:           cardID,
					Timestamp:    float64(time.Now().UnixNano()) / 1e9,
					Type:         cardType,
					Content:      originalName,
					Filename:     uniqueFilename,
					OriginalName: originalName,
				}

				cardData, _ := json.MarshalIndent(card, "", "  ")
				os.WriteFile(filepath.Join(webclipDir, cardID+".json"), cardData, 0644)
				results = append(results, card)
				logActivity("WebClip", "UPLOAD", uniqueFilename, r)

			default:
				part.Close()
			}
		}

		if strings.TrimSpace(textContent) != "" {
			cardID := fmt.Sprintf("%.5f", float64(time.Now().UnixNano())/1e9)
			card := WebClipCard{
				ID:        cardID,
				Timestamp: float64(time.Now().UnixNano()) / 1e9,
				Type:      "text",
				Content:   textContent,
			}
			cardData, _ := json.MarshalIndent(card, "", "  ")
			os.WriteFile(filepath.Join(webclipDir, cardID+".json"), cardData, 0644)
			results = append(results, card)
			logActivity("WebClip", "ADD_TEXT", "Card ID: "+cardID, r)
		}

		if len(results) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{"status": "error", "message": "No content provided"})
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "success", "cards": results})
	}
}

func webclipDeleteCardHandler(w http.ResponseWriter, r *http.Request) {
	cardID := strings.TrimPrefix(r.URL.Path, "/wc/api/cards/")
	if !validNameComponent(cardID) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "error", "message": "Card not found"})
		return
	}
	cardPath := filepath.Join(webclipDir, cardID+".json")
	if !pathInsideVolume(webclipDir, cardPath) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "error", "message": "Card not found"})
		return
	}

	if _, err := os.Stat(cardPath); os.IsNotExist(err) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "error", "message": "Card not found"})
		return
	}

	data, err := os.ReadFile(cardPath)
	if err == nil {
		var card WebClipCard
		if json.Unmarshal(data, &card) == nil && validNameComponent(card.Filename) {
			filePath := filepath.Join(webclipFiles, card.Filename)
			if pathInsideVolume(webclipFiles, filePath) {
				os.Remove(filePath)
			}
		}
	}
	os.Remove(cardPath)
	logActivity("WebClip", "DELETE", "Card ID: "+cardID, r)
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success"})
}

func webclipClearAllHandler(w http.ResponseWriter, r *http.Request) {
	os.RemoveAll(webclipDir)
	webclipEnsureDirs()
	logActivity("WebClip", "CLEAR_ALL", "All cards cleared", r)
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success"})
}

func randomHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func copyIO(dst interface{ Write([]byte) (int, error) }, src interface{ Read([]byte) (int, error) }) {
	buf := make([]byte, 32*1024)
	for {
		n, err := src.Read(buf)
		if n > 0 {
			dst.Write(buf[:n])
		}
		if err != nil {
			break
		}
	}
}
