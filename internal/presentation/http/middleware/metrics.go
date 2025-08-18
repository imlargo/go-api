package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/imlargo/go-api-template/internal/metrics"
)

func NewMetricsMiddleware(metrics metrics.MetricsService) gin.HandlerFunc {
	return func(c *gin.Context) {

		// Skip if it's an SSE request
		if c.GetHeader("Accept") == "text/event-stream" || c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		start := time.Now()

		// Process the request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Get request info
		method := c.Request.Method
		path := c.FullPath()
		status := strconv.Itoa(c.Writer.Status())

		// If no path was resolved (404), use the real URL
		if path == "" {
			path = c.Request.URL.Path
		}

		// Record metrics
		metrics.RecordHTTPRequest(method, path, status)
		metrics.RecordHTTPDuration(method, path, status, duration)
	}
}
