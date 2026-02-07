package server

import (
	"path/filepath"
	"regexp"
	"strings"
)

func safeDownloadFilename(name string) string {
	base := filepath.Base(name)
	base = strings.TrimSpace(base)
	re := regexp.MustCompile(`[^a-zA-Z0-9._\- ]`)
	base = re.ReplaceAllString(base, "")
	re = regexp.MustCompile(`\s+`)
	base = re.ReplaceAllString(base, " ")
	if base == "" {
		base = "audio"
	}
	if len(base) > 200 {
		base = base[:200]
	}
	return base
}
