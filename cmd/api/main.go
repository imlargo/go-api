package main

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/imlargo/go-api/internal"
	"github.com/imlargo/go-api/internal/cache"
	"github.com/imlargo/go-api/internal/cache/redis"
	"github.com/imlargo/go-api/internal/config"
	postgres "github.com/imlargo/go-api/internal/database"
	"github.com/imlargo/go-api/internal/metrics"
	"github.com/imlargo/go-api/internal/repositories"
	"github.com/imlargo/go-api/internal/store"
	"github.com/imlargo/go-api/pkg/kv"
	"github.com/imlargo/go-api/pkg/ratelimiter"
	"github.com/imlargo/go-api/pkg/storage"
	"go.uber.org/zap"
)

// @title Go api
// @version 1.0
// @description Backend service template

// @contact.name Default
// @contact.url https://default.dev
// @license.name MIT
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

// @securityDefinitions.apiKey PushApiKey
// @in header
// @name X-API-Key
func main() {
	cfg := config.LoadConfig()

	// Logger
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	// Rate Limiter
	if cfg.RateLimiter.Enabled {
		logger.Infof("Initializing rate limiter with config: %s request every %.2f seconds", strconv.Itoa(cfg.RateLimiter.RequestsPerTimeFrame), cfg.RateLimiter.TimeFrame.Seconds())
	}
	rateLimiter := ratelimiter.NewTokenBucketLimiter(ratelimiter.Config{
		RequestsPerTimeFrame: cfg.RateLimiter.RequestsPerTimeFrame,
		TimeFrame:            cfg.RateLimiter.TimeFrame,
	})

	// Database
	db, err := postgres.NewPostgres(cfg.Database.URL)
	if err != nil {
		logger.Fatal("Could not initialize database: ", err)
	}

	// Storage
	fileStorage := storage.NewEmptyStorage()
	if cfg.Storage.Enabled {
		fileStorage, err = storage.NewR2Storage(storage.StorageConfig{
			BucketName:      cfg.Storage.BucketName,
			AccountID:       cfg.Storage.AccountID,
			AccessKeyID:     cfg.Storage.AccessKeyID,
			SecretAccessKey: cfg.Storage.SecretAccessKey,
			PublicDomain:    cfg.Storage.PublicDomain,
			UsePublicURL:    cfg.Storage.UsePublicURL,
		})
	}

	if err != nil {
		logger.Fatal("Could not initialize storage service: ", err)
		return
	}

	// Redis
	redisClient, err := redis.NewRedisClient(cfg.Redis.RedisURL)
	if err != nil {
		logger.Fatal("Could not initialize Redis client: ", err)
		return
	}

	// Cache
	cacheProvider := redis.NewRedisCache(redisClient)
	cacheService := kv.NewKeyValueStore(cacheProvider)
	cacheKeys := cache.NewCacheKeys(kv.NewBuilder("api", "v1"))

	// Repositories
	repositoryContainer := repositories.NewRepository(db, cacheKeys, cacheService, logger)
	store := store.NewStorage(repositoryContainer)

	// Metrics
	metricsService := metrics.NewPrometheusMetrics()

	router := gin.Default()

	app := &internal.Application{
		Config:      cfg,
		Store:       store,
		Storage:     fileStorage,
		Metrics:     metricsService,
		Cache:       cacheService,
		CacheKeys:   cacheKeys,
		RateLimiter: rateLimiter,
		Router:      router,
		Logger:      logger,
	}

	app.Mount()

	logger.Info("Starting API...")

	app.Run()
}
