package server

import "net/http"

func InfoHandler() http.HandlerFunc {
	type info struct {
		MaxUploadMB       int      `json:"max_upload_mb"`
		AllowedExtensions []string `json:"allowed_extensions"`
		DownloadSuffix    string   `json:"download_suffix"`
	}
	data := info{MaxUploadMB, AllowedExtensions, DownloadSuffix}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		writeJSON(w, http.StatusOK, data)
	}
}
