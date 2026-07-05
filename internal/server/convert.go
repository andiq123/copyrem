package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"copyrem/internal/config"
	"copyrem/internal/converter"
)

// convertSem bounds concurrent ffmpeg jobs to CPU count; extra jobs queue.
var convertSem = make(chan struct{}, max(1, runtime.NumCPU()))

func ConvertHandler(cfg config.Params, store *JobStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		inPath, baseName, status, err := ParseUpload(w, r)
		if err != nil {
			writeError(w, status, err.Error())
			return
		}
		dir := filepath.Dir(inPath)
		outPath := filepath.Join(dir, randHex(8)+".mp3")
		job := store.Create(inPath, outPath, baseName+DownloadSuffix)

		intensity := 1.0
		if val := r.FormValue("intensity"); val != "" {
			if f, err := strconv.ParseFloat(val, 64); err == nil {
				intensity = f
			}
		}

		go func() {
			// Cap concurrent ffmpeg jobs to CPU count so each runs at full
			// speed instead of N jobs thrashing. Cancelled-while-queued exits fast.
			select {
			case convertSem <- struct{}{}:
				defer func() { <-convertSem }()
			case <-job.Ctx.Done():
				_ = os.Remove(inPath)
				return
			}

			err := converter.ConvertWithProgress(job.Ctx, cfg, inPath, outPath, intensity, func(pct int) {
				store.SetPercent(job.ID, pct)
			})
			if err != nil {
				_ = os.Remove(inPath)
				_ = os.Remove(outPath)
				if job.Ctx.Err() == context.Canceled {
					return
				}
				log.Printf("job %s failed: %v", job.ID, err)
				store.SetFailed(job.ID, err.Error())
				return
			}
			store.SetDone(job.ID)
		}()

		writeJSON(w, http.StatusOK, struct {
			JobID string `json:"job_id"`
		}{job.ID})
	}
}

func ProgressHandler(store *JobStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		id := strings.TrimPrefix(r.URL.Path, "/convert/progress/")
		if id == "" || store.Get(id) == nil {
			writeError(w, http.StatusNotFound, "job not found")
			return
		}

		flusher, ok := w.(http.Flusher)
		if !ok {
			writeError(w, http.StatusInternalServerError, "streaming not supported")
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-r.Context().Done():
				if j := store.Get(id); j != nil && j.Status != JobDone {
					store.Cancel(id)
				}
				return
			case <-ticker.C:
				j := store.Get(id)
				if j == nil {
					return
				}

				switch j.Status {
				case JobDone:
					fmt.Fprintf(w, "data: {\"percent\":100,\"done\":true,\"filename\":%q}\n\n", j.OriginalName)
					flusher.Flush()
					return
				case JobFailed:
					fmt.Fprintf(w, "data: {\"percent\":%d,\"error\":%q}\n\n", j.Percent, j.Error)
					flusher.Flush()
					return
				default:
					fmt.Fprintf(w, "data: {\"percent\":%d}\n\n", j.Percent)
					flusher.Flush()
				}
			}
		}
	}
}

func CancelHandler(store *JobStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		id := strings.TrimPrefix(r.URL.Path, "/convert/cancel/")
		if id == "" || store.Get(id) == nil {
			writeError(w, http.StatusNotFound, "job not found")
			return
		}

		store.Cancel(id)
		writeJSON(w, http.StatusOK, struct{ Cancelled bool `json:"cancelled"` }{true})
	}
}

func DownloadHandler(store *JobStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		id := strings.TrimPrefix(r.URL.Path, "/convert/download/")
		job := store.Get(id)
		if job == nil {
			writeError(w, http.StatusNotFound, "job not found")
			return
		}
		if job.Status != JobDone {
			writeError(w, http.StatusConflict, "job not ready")
			return
		}

		w.Header().Set("Content-Type", "audio/mpeg")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", job.OriginalName))
		http.ServeFile(w, r, job.OutPath)
	}
}

func PreviewHandler(store *JobStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		rest := strings.TrimPrefix(r.URL.Path, "/convert/preview/")
		id, kind, ok := strings.Cut(rest, "/")
		if !ok || id == "" {
			writeError(w, http.StatusNotFound, "job not found")
			return
		}

		job := store.Get(id)
		if job == nil {
			writeError(w, http.StatusNotFound, "job not found")
			return
		}

		var path string
		switch kind {
		case "original":
			path = job.InPath
		case "remixed":
			if job.Status != JobDone {
				writeError(w, http.StatusConflict, "job not ready")
				return
			}
			path = job.OutPath
		default:
			writeError(w, http.StatusNotFound, "not found")
			return
		}

		if path == "" {
			writeError(w, http.StatusNotFound, "file not found")
			return
		}
		if _, err := os.Stat(path); err != nil {
			writeError(w, http.StatusNotFound, "file not found")
			return
		}

		w.Header().Set("Content-Type", audioMIME(path))
		w.Header().Set("Content-Disposition", "inline")
		http.ServeFile(w, r, path)
	}
}

func audioMIME(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".mp3":
		return "audio/mpeg"
	case ".m4a":
		return "audio/mp4"
	case ".wav":
		return "audio/wav"
	case ".flac":
		return "audio/flac"
	case ".aac":
		return "audio/aac"
	case ".ogg":
		return "audio/ogg"
	default:
		return "audio/mpeg"
	}
}
