package main

import (
	"log"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	api "github.com/imlargo/go-api/cmd/app"
	"github.com/imlargo/go-api/internal/env"
	"github.com/imlargo/go-api/internal/ratelimiter"
)

// @contact.name imlargo
// @contact.url https://imlargo.dev
// @license.name MIT
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
func main() {

	errEnv := env.Initialize()
	if errEnv != nil {
		log.Println(errEnv.Error())
		return
	}

	gin.SetMode(os.Getenv("GIN_MODE"))

	config := api.SetupConfig()

	if config.Ratelimiter.Enabled {
		log.Printf("Initializing rate limiter with config: %s request every %.2f seconds", strconv.Itoa(config.Ratelimiter.RequestsPerTimeFrame), config.Ratelimiter.TimeFrame.Seconds())
	}

	rl := ratelimiter.NewTokenBucketLimiter(config.Ratelimiter)

	app := &api.Application{
		Config:      config,
		RateLimiter: rl,
	}

	router := app.Mount()

	app.SetupDocs(router)

	if err := router.Run(":" + app.Config.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
