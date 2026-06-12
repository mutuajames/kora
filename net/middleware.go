package net

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/secure"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestIDMiddleware ensures every request has a traceable ID.
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-Id")
		if id == "" {
			id = uuid.New().String()
		}
		c.Header("X-Request-Id", id)
		c.Set("request_id", id)
		c.Next()
	}
}

// SecurityHeadersMiddleware sets production security headers via gin-contrib/secure.
func SecurityHeadersMiddleware(tlsEnabled bool) gin.HandlerFunc {
	cfg := secure.Config{
		SSLRedirect:          tlsEnabled,
		IsDevelopment:        false,
		STSSeconds:            31536000,
		STSIncludeSubdomains: true,
		FrameDeny:            true,
		ContentTypeNosniff:   true,
		BrowserXssFilter:     true,
		ContentSecurityPolicy: "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval' https://unpkg.com https://cdn.jsdelivr.net; style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; img-src 'self' data:;",
		ReferrerPolicy:       "same-origin",
	}
	if !tlsEnabled {
		cfg.STSSeconds = 0
		cfg.SSLRedirect = false
	}
	return secure.New(cfg)
}

// CORSMiddleware configures cross-origin access.
func CORSMiddleware(allowedOrigins []string) gin.HandlerFunc {
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{"*"}
	}
	return cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Kora-CSRF-Token", "X-Request-Id"},
		ExposeHeaders:    []string{"X-Request-Id", "Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * 3600, // 12 hours
	})
}
