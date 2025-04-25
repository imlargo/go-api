package main

import (
	"log"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	api "github.com/imlargo/go-api/cmd/app"
	"github.com/imlargo/go-api/internal/db"
	"github.com/imlargo/go-api/internal/env"
	"github.com/imlargo/go-api/internal/ratelimiter"
	"github.com/imlargo/go-api/internal/store"
)

// @title Default API
// @version 1.0
// @description Default backend service for a web application.

// @contact.name Default
// @contact.url https://default.dev
// @license.name MIT
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {

	errEnv := env.Initialize()
	if errEnv != nil {
		log.Println(errEnv.Error())
		return
	}

	config := api.SetupConfig()

	if config.Ratelimiter.Enabled {
		log.Printf("Initializing rate limiter with config: %s request every %.2f seconds", strconv.Itoa(config.Ratelimiter.RequestsPerTimeFrame), config.Ratelimiter.TimeFrame.Seconds())
	}

	rl := ratelimiter.NewTokenBucketLimiter(config.Ratelimiter)

	db, dbErr := db.ConnectDB()
	if dbErr != nil {
		log.Fatalf("Failed to connect to database: %v", dbErr)
	}

	storage := store.NewStorage(db)

	app := &api.Application{
		Config:      config,
		RateLimiter: rl,
		Storage:     storage,
	}

	gin.SetMode(os.Getenv("GIN_MODE"))
	router := app.Mount()
	app.SetupDocs(router)

	log.Printf("Server running on %s:%s", os.Getenv("API_URL"), os.Getenv("PORT"))

	if err := router.Run(":" + app.Config.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
