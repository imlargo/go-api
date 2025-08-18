package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/imlargo/go-api-template/internal/presentation/http/responses"
	"github.com/imlargo/go-api-template/internal/shared/ports"
)

func AuthTokenMiddleware(jwtAuthenticator ports.JWTAuthenticator) gin.HandlerFunc {
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

		tokenData, err := jwtAuthenticator.ValidateToken(token, false)
		if err != nil {
			ctx.Abort()
			responses.ErrorUnauthorized(ctx, err.Error())
			return
		}

		ctx.Set("userID", tokenData.UserID)

		ctx.Next()
	}
}
