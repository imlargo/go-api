package main

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/imlargo/go-api-template/internal"
	"github.com/imlargo/go-api-template/internal/cache"
	"github.com/imlargo/go-api-template/internal/cache/redis"
	"github.com/imlargo/go-api-template/internal/config"
	postgres "github.com/imlargo/go-api-template/internal/database"
	"github.com/imlargo/go-api-template/internal/metrics"
	"github.com/imlargo/go-api-template/internal/repositories"
	"github.com/imlargo/go-api-template/internal/store"
	"github.com/imlargo/go-api-template/pkg/kv"
	"github.com/imlargo/go-api-template/pkg/ratelimiter"
	"github.com/imlargo/go-api-template/pkg/storage"
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
	storage, err := storage.NewR2Storage(storage.StorageConfig{
		BucketName:      cfg.Storage.BucketName,
		AccountID:       cfg.Storage.AccountID,
		AccessKeyID:     cfg.Storage.AccessKeyID,
		SecretAccessKey: cfg.Storage.SecretAccessKey,
		PublicDomain:    cfg.Storage.PublicDomain,
		UsePublicURL:    cfg.Storage.UsePublicURL,
	})
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
		Storage:     storage,
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
