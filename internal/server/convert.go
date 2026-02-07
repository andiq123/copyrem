package server

import (
	"encoding/json"
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

		limit := int64(MaxUploadMB) * 1024 * 1024
		r.Body = http.MaxBytesReader(w, r.Body, limit)
		if err := r.ParseMultipartForm(2 << 20); err != nil {
			if err.Error() == "http: request body too large" {
				writeError(w, http.StatusRequestEntityTooLarge, fmt.Sprintf("file too large (max %d MB)", MaxUploadMB))
				return
			}
			writeError(w, http.StatusBadRequest, "invalid form")
			return
		}

		fhs, ok := r.MultipartForm.File["file"]
		if !ok || len(fhs) == 0 {
			writeError(w, http.StatusBadRequest, "missing file")
			return
		}
		fh := fhs[0]
		ext := strings.ToLower(filepath.Ext(fh.Filename))
		if !allowedExtension(ext) {
			writeError(w, http.StatusBadRequest, "unsupported format. Allowed: "+AllowedExtensionsComma())
			return
		}

		inF, err := fh.Open()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to read upload")
			return
		}
		defer inF.Close()

		dir := os.TempDir()
		inPath := filepath.Join(dir, "copyrem-"+randHex(8)+ext)
		outPath := filepath.Join(dir, "copyrem-"+randHex(8)+".mp3")

		dst, err := os.Create(inPath)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create temp file")
			return
		}
		_, err = dst.ReadFrom(inF)
		dst.Close()
		if err != nil {
			_ = os.Remove(inPath)
			writeError(w, http.StatusInternalServerError, "failed to save upload")
			return
		}

		name := safeDownloadFilename(strings.TrimSuffix(fh.Filename, filepath.Ext(fh.Filename))) + DownloadSuffix
		job := store.Create(inPath, outPath, name)

		go func() {
			store.SetRunning(job.ID)
			if err := converter.ConvertWithProgress(cfg, inPath, outPath, func(pct int) {
				store.SetPercent(job.ID, pct)
			}); err != nil {
				store.SetFailed(job.ID, "conversion failed")
				return
			}
			store.SetDone(job.ID)
		}()

		writeJSON(w, http.StatusOK, map[string]string{"job_id": job.ID})
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
				return
			case <-ticker.C:
				j := store.Get(id)
				if j == nil {
					return
				}

				evt := map[string]any{"percent": j.Percent}
				if j.Status == JobDone {
					evt["done"] = true
				}
				if j.Status == JobFailed {
					evt["error"] = j.Error
				}

				data, _ := json.Marshal(evt)
				fmt.Fprintf(w, "data: %s\n\n", data)
				flusher.Flush()

				if j.Status == JobDone || j.Status == JobFailed {
					return
				}
			}
		}
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
