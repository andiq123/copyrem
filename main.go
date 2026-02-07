package main

import (
	"log"
	"os"

	"copyrem/internal/config"
	"copyrem/internal/server"

	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	cfg := config.Default()
	server.RegisterRoutes(r, cfg)

	addr := defaultAddr()
	log.Printf("CopyRem server listening on %s", addr)
	r.Run(addr)
}

func defaultAddr() string {
	if p := os.Getenv("PORT"); p != "" {
		return ":" + p
	}
	return ":8080"
}
