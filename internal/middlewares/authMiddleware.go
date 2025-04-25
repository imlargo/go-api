package middlewares

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/imlargo/go-api/internal/auth"
	"github.com/imlargo/go-api/internal/responses"
)

func AuthTokenMiddleware() gin.HandlerFunc {
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

		token := parts[1]
		if token == "" {
			ctx.Abort()
			responses.ErrorUnauthorized(ctx, "token is empty")
			return
		}

		jwtAuthenticator := auth.NewJWTAuthenticator()
		tokenData, err := jwtAuthenticator.ValidateToken(token, false)
		if err != nil {
			ctx.Abort()
			responses.ErrorUnauthorized(ctx, err.Error())
			return
		}

		ctx.Set("user_id", tokenData.UserID)

		ctx.Next()
	}
}
