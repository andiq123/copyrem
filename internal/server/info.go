package server

import (
	"encoding/json"
	"net/http"
)

type UploadInfo struct {
	MaxUploadMB       int      `json:"max_upload_mb"`
	AllowedExtensions []string `json:"allowed_extensions"`
	DownloadSuffix    string   `json:"download_suffix"`
}

func InfoHandler() http.HandlerFunc {
	info := UploadInfo{
		MaxUploadMB:       MaxUploadMB,
		AllowedExtensions: AllowedExtensions,
		DownloadSuffix:    DownloadSuffix,
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(info)
	}
}
