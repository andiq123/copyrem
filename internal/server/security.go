package server

import (
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	reUnsafe = regexp.MustCompile(`[^a-zA-Z0-9._\- ]`)
	reSpaces = regexp.MustCompile(`\s+`)
)

func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; script-src 'self'; style-src 'self' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; connect-src 'self'; img-src 'self' data:; frame-ancestors 'none'; base-uri 'self'")
		next.ServeHTTP(w, r)
	})
}

func safeDownloadFilename(name string) string {
	base := reUnsafe.ReplaceAllString(strings.TrimSpace(filepath.Base(name)), "")
	base = reSpaces.ReplaceAllString(base, " ")
	if base == "" {
		return "audio"
	}
	if len(base) > 200 {
		return base[:200]
	}
	return base
}
