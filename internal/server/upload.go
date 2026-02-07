package server

import (
	"strings"
)

const (
	MaxUploadMB   = 80
	DownloadSuffix = "_modified.mp3"
)

var (
	AllowedExtensions = []string{".mp3", ".m4a", ".wav", ".flac", ".aac", ".ogg"}
)

func AllowedExtensionsComma() string {
	return strings.Join(AllowedExtensions, ", ")
}

func allowedExtension(ext string) bool {
	for _, e := range AllowedExtensions {
		if ext == e {
			return true
		}
	}
	return false
}
