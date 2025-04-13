package middlewares

import (
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/imlargo/go-api/internal/env"
)

func NewCorsMiddleware() gin.HandlerFunc {

	host := os.Getenv(env.API_URL)
	allowedOrigins := []string{
		host,
	}

	config := cors.Config{
		AllowOrigins: allowedOrigins,
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
	}

	return cors.New(config)
}
