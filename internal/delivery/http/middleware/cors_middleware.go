package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSConfig holds CORS middleware configuration
type CORSConfig struct {
	AllowedOrigins []string
	Environment    string
}

// NewCORSMiddleware creates a CORS middleware with whitelist support
func NewCORSMiddleware(config CORSConfig) gin.HandlerFunc {
	// Build origin map for O(1) lookup
	originMap := make(map[string]bool)
	for _, origin := range config.AllowedOrigins {
		originMap[strings.TrimSpace(origin)] = true
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// In development, allow all. In production, check whitelist.
		if config.Environment != "production" || originMap[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Request-ID")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
