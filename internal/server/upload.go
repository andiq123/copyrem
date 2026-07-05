package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const (
	MaxUploadMB    = 80
	DownloadSuffix = "_modified.mp3"
)

var (
	AllowedExtensions    = []string{".mp3", ".m4a", ".wav", ".flac", ".aac", ".ogg"}
	allowedExtensionsStr = strings.Join(AllowedExtensions, ", ")
)

func ParseUpload(w http.ResponseWriter, r *http.Request) (inPath, baseName string, status int, err error) {
	limit := int64(MaxUploadMB) * 1024 * 1024
	r.Body = http.MaxBytesReader(w, r.Body, limit)
	if err := r.ParseMultipartForm(2 << 20); err != nil {
		if err.Error() == "http: request body too large" {
			return "", "", http.StatusRequestEntityTooLarge, fmt.Errorf("file too large (max %d MB)", MaxUploadMB)
		}
		return "", "", http.StatusBadRequest, fmt.Errorf("invalid form")
	}
	fhs, ok := r.MultipartForm.File["file"]
	if !ok || len(fhs) == 0 {
		return "", "", http.StatusBadRequest, fmt.Errorf("missing file")
	}
	fh := fhs[0]
	ext := strings.ToLower(filepath.Ext(fh.Filename))
	if !slices.Contains(AllowedExtensions, ext) {
		return "", "", http.StatusBadRequest, fmt.Errorf("unsupported format. Allowed: %s", allowedExtensionsStr)
	}
	inF, err := fh.Open()
	if err != nil {
		return "", "", http.StatusInternalServerError, fmt.Errorf("failed to read upload")
	}
	defer inF.Close()
	inPath = filepath.Join(os.TempDir(), "copyrem-"+randHex(8)+ext)
	dst, err := os.Create(inPath)
	if err != nil {
		return "", "", http.StatusInternalServerError, fmt.Errorf("failed to create temp file")
	}
	_, err = dst.ReadFrom(inF)
	dst.Close()
	if err != nil {
		_ = os.Remove(inPath)
		return "", "", http.StatusInternalServerError, fmt.Errorf("failed to save upload")
	}
	baseName = safeDownloadFilename(strings.TrimSuffix(fh.Filename, filepath.Ext(fh.Filename)))
	return inPath, baseName, http.StatusOK, nil
}
