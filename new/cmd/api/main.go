package main

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/imlargo/go-api/internal/config"
	"github.com/imlargo/go-api/internal/database"
	"github.com/imlargo/go-api/internal/store"
	"github.com/imlargo/go-api/pkg/medusa/core/app"
	"github.com/imlargo/go-api/pkg/medusa/core/logger"
	"github.com/imlargo/go-api/pkg/medusa/core/ratelimiter"
	medusarepo "github.com/imlargo/go-api/pkg/medusa/core/repository"
	"github.com/imlargo/go-api/pkg/medusa/core/responses"
	"github.com/imlargo/go-api/pkg/medusa/core/server/http"
	"github.com/imlargo/go-api/pkg/medusa/services/cache"
	"github.com/imlargo/go-api/pkg/medusa/services/storage"
)

func main() {
	cfg := config.LoadConfig()

	logger := logger.NewLogger()
	defer logger.Sync()

	router := gin.Default()
	srv := http.NewServer(
		router,
		logger,
		http.WithServerHost(cfg.Server.Host),
		http.WithServerPort(cfg.Server.Port),
	)

	app := app.NewApp(
		app.WithName("butter"),
		app.WithServer(srv),
	)

	Mount(app, cfg, router, logger)

	app.Run(context.Background())
}

func Mount(app *app.App, cfg config.Config, router *gin.Engine, logger *logger.Logger) {

	// Ping
	router.GET("/ping", func(c *gin.Context) {
		responses.SuccessOK(c, "hello")
	})

	// Rate Limiter
	if cfg.RateLimiter.Enabled {
		logger.Info("Rate Limiter is enabled")
	}
	_ = ratelimiter.NewTokenBucketLimiter(ratelimiter.Config{
		RequestsPerTimeFrame: cfg.RateLimiter.RequestsPerTimeFrame,
		TimeFrame:            cfg.RateLimiter.TimeFrame,
	})

	// Database
	db, err := database.NewPostgresDatabase(cfg.Database.URL)
	if err != nil {
		logger.Fatal("Could not connect to the database: " + err.Error())
		return
	}

	// Storage
	_, err = storage.NewFileStorage(storage.StorageProviderR2, cfg.Storage)
	if err != nil {
		logger.Fatal("Could not initialize storage: " + err.Error())
		return
	}

	// Redis
	redisClient, err := database.NewRedisClient(cfg.Redis.RedisURL)
	if err != nil {
		logger.Fatal("Could not connect to Redis: " + err.Error())
		return
	}

	// Cache
	_ = cache.NewRedisCache(redisClient)

	// Repositories
	medusaStore := medusarepo.NewStore(db, logger)
	_ = store.NewStore(medusaStore)
}
