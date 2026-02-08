package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"copyrem/internal/separator"
)

const (
	SeparateDownloadSuffixVocals  = "_vocals.mp3"
	SeparateDownloadSuffixInst    = "_instrumental.mp3"
)

func SeparateHandler(store *JobStore) http.HandlerFunc {
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
		base := strings.TrimSuffix(filepath.Base(inPath), filepath.Ext(inPath))
		outVocals := filepath.Join(dir, base+SeparateDownloadSuffixVocals)
		outInstrumental := filepath.Join(dir, base+SeparateDownloadSuffixInst)
		job := store.CreateWithTwoOutputs(inPath, outVocals, outInstrumental, baseName)

		go func() {
			store.SetRunning(job.ID)
			err := separator.SeparateWithProgress(job.Ctx, inPath, outVocals, outInstrumental, func(pct int) {
				store.SetPercent(job.ID, pct)
			})
			_ = os.Remove(inPath)
			if err != nil {
				if job.Ctx.Err() == context.Canceled {
					return
				}
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

func SeparateProgressHandler(store *JobStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		id := strings.TrimPrefix(r.URL.Path, "/separate/progress/")
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

func SeparateCancelHandler(store *JobStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		id := strings.TrimPrefix(r.URL.Path, "/separate/cancel/")
		if id == "" || store.Get(id) == nil {
			writeError(w, http.StatusNotFound, "job not found")
			return
		}

		store.Cancel(id)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte(`{"cancelled":true}`))
	}
}

func SeparateDownloadHandler(store *JobStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		path := strings.TrimPrefix(r.URL.Path, "/separate/download/")
		parts := strings.SplitN(path, "/", 2)
		if len(parts) != 2 {
			writeError(w, http.StatusNotFound, "job not found")
			return
		}
		id, stem := parts[0], parts[1]
		job := store.Get(id)
		if job == nil || job.OutPath2 == "" {
			writeError(w, http.StatusNotFound, "job not found")
			return
		}
		if job.Status != JobDone {
			writeError(w, http.StatusConflict, "job not ready")
			return
		}

		var filePath, name string
		switch stem {
		case "vocals":
			filePath = job.OutPath
			name = job.OriginalName + SeparateDownloadSuffixVocals
		case "instrumental":
			filePath = job.OutPath2
			name = job.OriginalName + SeparateDownloadSuffixInst
		default:
			writeError(w, http.StatusNotFound, "job not found")
			return
		}

		w.Header().Set("Content-Type", "audio/mpeg")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", name))
		http.ServeFile(w, r, filePath)
	}
}
