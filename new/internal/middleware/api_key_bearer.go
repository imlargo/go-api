package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/imlargo/go-api/pkg/medusa/responses"
)

func BearerApiKeyMiddleware(apiKey string) gin.HandlerFunc {

	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")

		if authHeader == "" {
			ctx.Abort()
			responses.ErrorUnauthorized(ctx, "authorization header is missing")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			ctx.Abort()
			responses.ErrorUnauthorized(ctx, "authorization header must be in format 'Bearer token'")
			return
		}

		apiKeyHeader := parts[1]
		if apiKeyHeader == "" {
			ctx.Abort()
			responses.ErrorUnauthorized(ctx, "authorization header is missing")
			return
		}

		if apiKeyHeader != apiKey {
			ctx.Abort()
			responses.ErrorUnauthorized(ctx, "invalid API key")
			return
		}

		ctx.Next()
	}
}
