package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/imlargo/go-api-template/internal/presentation/http/responses"
	"github.com/imlargo/go-api-template/internal/shared/ports"
)

func NewRateLimiterMiddleware(rl ports.RateLimiter) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ip := ctx.ClientIP()
		allow, retryAfter := rl.Allow(ip)
		if !allow {
			message := "Rate limit exceeded. Try again in " + fmt.Sprintf("%.2f", retryAfter)
			responses.ErrorToManyRequests(ctx, message)
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
