package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SecurityHeadersMiddleware adds security headers to responses
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")
		
		// Enable XSS filtering
		c.Header("X-XSS-Protection", "1; mode=block")
		
		// Prevent page from being displayed in frame/iframe
		c.Header("X-Frame-Options", "DENY")
		
		// Force HTTPS (customize based on your needs)
		// c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		
		// Disable referrer information
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		
		// Content Security Policy (customize based on your needs)
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self';")
		
		// Permissions Policy (formerly Feature Policy)
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=()")

		c.Next()
	}
}

// RequestSizeLimitMiddleware limits the size of request bodies
func RequestSizeLimitMiddleware(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxSize {
			c.JSON(413, gin.H{"error": "Request body too large"})
			c.Abort()
			return
		}
		
		// Set a hard limit on request body size
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)
		
		c.Next()
	}
}