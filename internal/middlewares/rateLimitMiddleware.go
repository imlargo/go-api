package middlewares

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/imlargo/go-api/internal/ratelimiter"
	"github.com/imlargo/go-api/internal/responses"
)

func RateLimiterMiddleware(rl ratelimiter.Limiter, cfg ratelimiter.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if cfg.Enabled {

			ip := ctx.ClientIP()
			allow, retryAfter := rl.Allow(ip)
			if !allow {
				message := "Rate limit exceeded. Try again in " + fmt.Sprintf("%.2f", retryAfter)
				responses.ErrorToManyRequests(ctx, message)
				ctx.Abort()
				return
			}
		}

		ctx.Next()
	}
}
