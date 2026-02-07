package server

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"copyrem/internal/config"
	"copyrem/internal/converter"
)

func maxUploadBytes() int64 {
	return int64(MaxUploadMB) * 1024 * 1024
}

func ConvertHandler(cfg config.Params) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		limit := maxUploadBytes()
		r.Body = http.MaxBytesReader(w, r.Body, limit)
		if err := r.ParseMultipartForm(limit); err != nil {
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
		prefix := "copyrem-"
		inPath := filepath.Join(dir, prefix+randomName()+ext)
		outPath := filepath.Join(dir, prefix+randomName()+".mp3")

		defer func() {
			_ = os.Remove(inPath)
			_ = os.Remove(outPath)
		}()

		outFile, err := os.Create(inPath)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create temp file")
			return
		}
		_, err = outFile.ReadFrom(inF)
		outFile.Close()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to save upload")
			return
		}

		if err := converter.Convert(cfg, inPath, outPath); err != nil {
			writeError(w, http.StatusInternalServerError, "conversion failed")
			return
		}

		outName := safeDownloadFilename(trimExt(fh.Filename)) + DownloadSuffix
		w.Header().Set("Content-Type", "audio/mpeg")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", outName))
		http.ServeFile(w, r, outPath)
	}
}

func randomName() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "out"
	}
	return fmt.Sprintf("%x", b)
}

func trimExt(name string) string {
	return strings.TrimSuffix(name, filepath.Ext(name))
}
