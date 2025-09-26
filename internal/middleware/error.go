package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ErrorHandlingMiddleware provides centralized error handling and panic recovery
func ErrorHandlingMiddleware(logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				requestID := c.GetString(RequestIDKey)
				stack := string(debug.Stack())
				
				logger.With(
					"request_id", requestID,
					"error", err,
					"stack_trace", stack,
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
					"client_ip", c.ClientIP(),
				).Error("Panic recovered")

				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Internal server error",
					"request_id": requestID,
				})
				c.Abort()
			}
		}()

		c.Next()

		// Handle any errors that were added during processing
		if len(c.Errors) > 0 {
			requestID := c.GetString(RequestIDKey)
			
			for _, ginErr := range c.Errors {
				logger.With(
					"request_id", requestID,
					"error", ginErr.Error(),
					"error_type", ginErr.Type,
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
					"client_ip", c.ClientIP(),
				).Error("Request error")
			}

			// If response hasn't been written, send error response
			if !c.Writer.Written() {
				switch c.Errors.Last().Type {
				case gin.ErrorTypeBind:
					c.JSON(http.StatusBadRequest, gin.H{
						"error": "Invalid request data",
						"request_id": requestID,
					})
				case gin.ErrorTypePublic:
					c.JSON(http.StatusBadRequest, gin.H{
						"error": c.Errors.Last().Error(),
						"request_id": requestID,
					})
				default:
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": "Internal server error",
						"request_id": requestID,
					})
				}
			}
		}
	}
}