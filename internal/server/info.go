package server

import (
	"encoding/json"
	"net/http"
)

func InfoHandler() http.HandlerFunc {
	type info struct {
		MaxUploadMB       int      `json:"max_upload_mb"`
		AllowedExtensions []string `json:"allowed_extensions"`
		DownloadSuffix    string   `json:"download_suffix"`
	}
	body, _ := json.Marshal(info{MaxUploadMB, AllowedExtensions, DownloadSuffix})
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write(body)
	}
}
