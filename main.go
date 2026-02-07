package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"copyrem/internal/config"
	"copyrem/internal/server"
)

const (
	readTimeout  = 30 * time.Second
	writeTimeout = 60 * time.Second
	idleTimeout  = 120 * time.Second
)

func main() {
	addr := flag.String("addr", defaultAddr(), "Listen address")
	staticDir := flag.String("static", "frontend/dist", "Directory to serve React build from")
	flag.Parse()

	cfg := config.Default()
	mux := server.NewMux(cfg, *staticDir)
	handler := server.Chain(mux)

	srv := &http.Server{
		Addr:         *addr,
		Handler:      handler,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	log.Printf("CopyRem server listening on %s", *addr)
	if err := srv.ListenAndServe(); err != nil {
		fmt.Fprintf(os.Stderr, "server: %v\n", err)
		os.Exit(1)
	}
}

func defaultAddr() string {
	if p := os.Getenv("PORT"); p != "" {
		return ":" + p
	}
	return ":8080"
}
