package server

import (
	_ "embed"
	"net/http"
	"os"
	"strings"

	"copyrem/internal/config"

	"github.com/gin-gonic/gin"
)

//go:embed static/build.html
var buildHTML []byte

func RegisterRoutes(r *gin.Engine, cfg config.Params) {
	r.Use(securityMiddleware())
	r.Use(corsMiddleware())

	r.GET("/api/info", gin.WrapF(InfoHandler()))
	r.POST("/convert", gin.WrapF(RateLimitConvert(ConvertHandler(cfg))))

	if info, err := os.Stat("frontend/dist"); err == nil && info.IsDir() {
		r.Static("/assets", "frontend/dist/assets")
		r.StaticFile("/robots.txt", "frontend/dist/robots.txt")
		r.StaticFile("/sitemap.xml", "frontend/dist/sitemap.xml")
		r.NoRoute(func(c *gin.Context) {
			c.File("frontend/dist/index.html")
		})
	} else {
		r.NoRoute(func(c *gin.Context) {
			c.Data(http.StatusOK, "text/html; charset=utf-8", buildHTML)
		})
	}
}

func securityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		c.Header("Content-Security-Policy",
			"default-src 'self'; script-src 'self'; style-src 'self' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; connect-src 'self'; img-src 'self' data:; frame-ancestors 'none'; base-uri 'self'")
		c.Next()
	}
}

func corsMiddleware() gin.HandlerFunc {
	allowed := AllowedOriginsForCORS()
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && allowed[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type")
			c.Header("Access-Control-Max-Age", "86400")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
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
