package middleware

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/imlargo/go-api/internal/responses"
	"github.com/imlargo/go-api/pkg/ratelimiter"
)

func NewRateLimiterMiddleware(rl ratelimiter.RateLimiter) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ip := ctx.ClientIP()
		allow, retryAfter := rl.Allow(ip)
		
		// Add basic rate limit headers (without specific limiter details)
		if !allow {
			ctx.Header("X-RateLimit-Remaining", "0")
			ctx.Header("Retry-After", strconv.Itoa(int(retryAfter)))
		}
		
		if !allow {
			message := "Rate limit exceeded. Try again in " + fmt.Sprintf("%.2f", retryAfter) + " seconds"
			responses.ErrorToManyRequests(ctx, message)
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
