package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
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

func allowedExtension(ext string) bool {
	for _, e := range AllowedExtensions {
		if ext == e {
			return true
		}
	}
	return false
}

type uploadError struct {
	status int
	err    error
}

func (e uploadError) Error() string { return e.err.Error() }

func (e uploadError) Status() int { return e.status }

func ParseUpload(w http.ResponseWriter, r *http.Request) (inPath, baseName string, err error) {
	limit := int64(MaxUploadMB) * 1024 * 1024
	r.Body = http.MaxBytesReader(w, r.Body, limit)
	if err := r.ParseMultipartForm(2 << 20); err != nil {
		if err.Error() == "http: request body too large" {
			return "", "", uploadError{http.StatusRequestEntityTooLarge, fmt.Errorf("file too large (max %d MB)", MaxUploadMB)}
		}
		return "", "", uploadError{http.StatusBadRequest, fmt.Errorf("invalid form")}
	}
	fhs, ok := r.MultipartForm.File["file"]
	if !ok || len(fhs) == 0 {
		return "", "", uploadError{http.StatusBadRequest, fmt.Errorf("missing file")}
	}
	fh := fhs[0]
	ext := strings.ToLower(filepath.Ext(fh.Filename))
	if !allowedExtension(ext) {
		return "", "", uploadError{http.StatusBadRequest, fmt.Errorf("unsupported format. Allowed: %s", allowedExtensionsStr)}
	}
	inF, err := fh.Open()
	if err != nil {
		return "", "", uploadError{http.StatusInternalServerError, fmt.Errorf("failed to read upload")}
	}
	defer inF.Close()
	dir := os.TempDir()
	inPath = filepath.Join(dir, "copyrem-"+randHex(8)+ext)
	dst, err := os.Create(inPath)
	if err != nil {
		return "", "", uploadError{http.StatusInternalServerError, fmt.Errorf("failed to create temp file")}
	}
	_, err = dst.ReadFrom(inF)
	dst.Close()
	if err != nil {
		_ = os.Remove(inPath)
		return "", "", uploadError{http.StatusInternalServerError, fmt.Errorf("failed to save upload")}
	}
	baseName = safeDownloadFilename(strings.TrimSuffix(fh.Filename, filepath.Ext(fh.Filename)))
	return inPath, baseName, nil
}

func uploadStatus(err error) int {
	if err == nil {
		return 0
	}
	if ue, ok := err.(uploadError); ok {
		return ue.Status()
	}
	return http.StatusBadRequest
}
