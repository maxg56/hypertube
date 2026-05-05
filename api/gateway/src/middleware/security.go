package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityHeadersMiddleware sets HTTP security headers on every response.
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'none'")
		c.Next()
	}
}
