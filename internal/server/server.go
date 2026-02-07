package server

import (
	_ "embed"
	"net/http"
	"os"
	"strings"

	"copyrem/internal/config"
)

//go:embed static/build.html
var buildHTML []byte

func NewMux(cfg config.Params, staticDir string) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/info", InfoHandler())
	mux.HandleFunc("/convert", RateLimitConvert(ConvertHandler(cfg)))

	var staticHandler http.Handler
	if staticDir != "" {
		if info, err := os.Stat(staticDir); err == nil && info.IsDir() {
			staticHandler = http.FileServer(http.Dir(staticDir))
		}
	}
	if staticHandler == nil {
		staticHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/" || r.URL.Path == "/index.html" {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.Write(buildHTML)
				return
			}
			http.NotFound(w, r)
		})
	}
	mux.Handle("/", staticHandler)

	return mux
}

func Chain(next http.Handler) http.Handler {
	return SecurityHeaders(CORS(next))
}

func AllowedOriginsForCORS() map[string]bool {
	origins := map[string]bool{
		"http://localhost:5173":  true,
		"http://127.0.0.1:5173": true,
	}
	if s := os.Getenv("CORS_ORIGINS"); s != "" {
		for _, o := range strings.Split(s, ",") {
			if o = strings.TrimSpace(o); o != "" {
				origins[o] = true
			}
		}
	}
	return origins
}

func CORS(next http.Handler) http.Handler {
	allowed := AllowedOriginsForCORS()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && allowed[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Max-Age", "86400")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
