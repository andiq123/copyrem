package server

import (
	"encoding/json"
	"net/http"
)

var infoJSON []byte

func init() {
	infoJSON, _ = json.Marshal(struct {
		MaxUploadMB       int      `json:"max_upload_mb"`
		AllowedExtensions []string `json:"allowed_extensions"`
		DownloadSuffix    string   `json:"download_suffix"`
	}{MaxUploadMB, AllowedExtensions, DownloadSuffix})
}

func InfoHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write(infoJSON)
	}
}
