package server

import (
	"net/http"
	"os"
	"strings"

	"copyrem/internal/config"
)

func NewMux(cfg config.Params, staticDir string) *http.ServeMux {
	mux := http.NewServeMux()
	store := NewJobStore()

	mux.HandleFunc("/convert", RateLimitConvert(ConvertHandler(cfg, store)))
	mux.HandleFunc("/convert/progress/", ProgressHandler(store))
	mux.HandleFunc("/convert/cancel/", CancelHandler(store))
	mux.HandleFunc("/convert/download/", DownloadHandler(store))
	mux.HandleFunc("/convert/preview/", PreviewHandler(store))
	mux.Handle("/", http.FileServer(http.Dir(staticDir)))

	return mux
}

func Chain(next http.Handler) http.Handler {
	return SecurityHeaders(CORS(next))
}

func allowedOrigins() map[string]bool {
	origins := map[string]bool{
		"http://localhost:5173": true,
		"http://127.0.0.1:5173": true,
	}
	if s := os.Getenv("CORS_ORIGINS"); s != "" {
		for o := range strings.SplitSeq(s, ",") {
			if o = strings.TrimSpace(o); o != "" {
				origins[o] = true
			}
		}
	}
	return origins
}

func CORS(next http.Handler) http.Handler {
	allowed := allowedOrigins()
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
