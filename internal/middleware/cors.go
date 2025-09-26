package middleware

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewCorsMiddleware(host string, origins []string) gin.HandlerFunc {
	allowedOrigins := append(origins, host)
	
	// Add common development origins
	devOrigins := []string{
		"http://localhost:3000",
		"http://localhost:3001", 
		"http://localhost:5173", // Vite default
		"http://localhost:8080",
		"http://127.0.0.1:3000",
		"http://127.0.0.1:5173",
	}
	allowedOrigins = append(allowedOrigins, devOrigins...)

	config := cors.Config{
		AllowOrigins:  allowedOrigins,
		AllowWildcard: false, // More secure than wildcard
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodHead,
			http.MethodOptions,
		},
		AllowCredentials: true,
		AllowHeaders: []string{
			"Origin",
			"Accept",
			"Authorization",
			"Content-Type",
			"X-API-Key",
			"X-Requested-With",
			"Cache-Control",
		},
		ExposeHeaders: []string{
			"Content-Length",
			"X-Total-Count",
			"X-RateLimit-Limit",
			"X-RateLimit-Remaining",
		},
		MaxAge: 12 * time.Hour, // Cache preflight for 12 hours
	}

	return cors.New(config)
}
