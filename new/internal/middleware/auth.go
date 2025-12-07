package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/imlargo/go-api/pkg/medusa/jwt"
	"github.com/imlargo/go-api/pkg/medusa/responses"
)

func AuthTokenMiddleware(jwtAuthenticator *jwt.JWT) gin.HandlerFunc {
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

		tokenData, err := jwtAuthenticator.ParseToken(token)
		if err != nil {
			ctx.Abort()
			responses.ErrorUnauthorized(ctx, err.Error())
			return
		}

		ctx.Set("userID", tokenData.UserID)

		ctx.Next()
	}
}
