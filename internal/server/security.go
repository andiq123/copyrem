package server

import (
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
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
