package middleware

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewCorsMiddleware(host string, origins []string) gin.HandlerFunc {

	allowedOrigins := append(origins, host)

	config := cors.Config{
		AllowOrigins:  allowedOrigins,
		AllowWildcard: true,
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
		},
	}

	return cors.New(config)
}
