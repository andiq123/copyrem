package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"copyrem/internal/config"
	"copyrem/internal/converter"
)

func ConvertHandler(cfg config.Params, store *JobStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		inPath, baseName, err := ParseUpload(w, r)
		if err != nil {
			writeError(w, uploadStatus(err), err.Error())
			return
		}
		dir := filepath.Dir(inPath)
		outPath := filepath.Join(dir, randHex(8)+".mp3")
		job := store.Create(inPath, outPath, baseName+DownloadSuffix)

		go func() {
			store.SetRunning(job.ID)
			err := converter.ConvertWithProgress(job.Ctx, cfg, inPath, outPath, func(pct int) {
				store.SetPercent(job.ID, pct)
			})
			_ = os.Remove(inPath)
			if err != nil {
				if job.Ctx.Err() == context.Canceled {
					return
				}
				store.SetFailed(job.ID, "conversion failed")
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
					fmt.Fprint(w, "data: {\"percent\":100,\"done\":true}\n\n")
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

		store.Cancel(id)
	}
}
