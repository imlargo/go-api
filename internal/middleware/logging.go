package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	RequestIDKey = "X-Request-ID"
)

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(RequestIDKey)
		if requestID == "" {
			requestID = uuid.New().String()
		}
		
		c.Header(RequestIDKey, requestID)
		c.Set(RequestIDKey, requestID)
		c.Next()
	}
}

// StructuredLoggingMiddleware provides structured logging with request context
func StructuredLoggingMiddleware(logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		method := c.Request.Method
		requestID := c.GetString(RequestIDKey)

		// Log request start
		logger.With(
			"request_id", requestID,
			"method", method,
			"path", path,
			"query", raw,
			"user_agent", c.GetHeader("User-Agent"),
			"client_ip", c.ClientIP(),
		).Info("Request started")

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)
		statusCode := c.Writer.Status()

		// Log request completion
		logLevel := "info"
		if statusCode >= 400 && statusCode < 500 {
			logLevel = "warn"
		} else if statusCode >= 500 {
			logLevel = "error"
		}

		logEntry := logger.With(
			"request_id", requestID,
			"method", method,
			"path", path,
			"query", raw,
			"status_code", statusCode,
			"latency_ms", latency.Milliseconds(),
			"client_ip", c.ClientIP(),
			"user_agent", c.GetHeader("User-Agent"),
		)

		switch logLevel {
		case "warn":
			logEntry.Warn("Request completed with client error")
		case "error":
			logEntry.Error("Request completed with server error")
		default:
			logEntry.Info("Request completed successfully")
		}
	}
}