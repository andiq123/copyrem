package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"copyrem/internal/config"
	"copyrem/internal/server"
)

func main() {
	cfg, err := config.Load("settings.json")
	if err != nil {
		log.Printf("settings.json not found, using defaults")
	}
	mux := server.NewMux(cfg, "frontend/dist")
	handler := server.Chain(mux)

	addr := defaultAddr()
	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 5 * time.Minute,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("CopyRem server listening on %s", addr)
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
