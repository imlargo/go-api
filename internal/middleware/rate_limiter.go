package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/nicolailuther/butter/internal/responses"
	"github.com/nicolailuther/butter/pkg/ratelimiter"
)

func NewRateLimiterMiddleware(rl ratelimiter.RateLimiter) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ip := ctx.ClientIP()
		allow, retryAfter := rl.Allow(ip)
		if !allow {
			message := "Rate limit exceeded. Try again in " + fmt.Sprintf("%.2f", retryAfter)
			responses.ErrorTooManyRequests(ctx, message)
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
