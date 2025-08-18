package main

import (
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/imlargo/go-api-template/internal"
	"github.com/imlargo/go-api-template/internal/cache"
	"github.com/imlargo/go-api-template/internal/cache/redis"
	"github.com/imlargo/go-api-template/internal/config"
	postgres "github.com/imlargo/go-api-template/internal/database"
	"github.com/imlargo/go-api-template/internal/metrics"
	"github.com/imlargo/go-api-template/internal/store"
	cachekey "github.com/imlargo/go-api-template/pkg/keybuilder"
	"github.com/imlargo/go-api-template/pkg/kv"
	"github.com/imlargo/go-api-template/pkg/ratelimiter"
	"github.com/imlargo/go-api-template/pkg/storage"
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

	// Database
	db, err := postgres.NewPostgres(cfg.Database.URL)
	if err != nil {
		log.Fatal("Could not initialize database: ", err)
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
		log.Fatal("Could not initialize storage service: ", err)
		return
	}

	// Redis
	redisClient, err := redis.NewRedisClient(cfg.Redis.RedisURL)
	if err != nil {
		log.Fatal("Could not initialize Redis client: ", err)
		return
	}

	// Cache
	cacheProvider := redis.NewRedisCache(redisClient)
	cacheService := kv.NewKeyValueStore(cacheProvider)
	cacheKeys := cache.NewCacheKeys(cachekey.NewBuilder("api", "v1"))

	// Repositories
	store := store.NewStorage(db, cacheService, cacheKeys)

	// Rate Limiter
	if cfg.RateLimiter.Enabled {
		log.Printf("Initializing rate limiter with config: %s request every %.2f seconds", strconv.Itoa(cfg.RateLimiter.RequestsPerTimeFrame), cfg.RateLimiter.TimeFrame.Seconds())
	}
	rateLimiter := ratelimiter.NewTokenBucketLimiter(ratelimiter.Config{
		RequestsPerTimeFrame: cfg.RateLimiter.RequestsPerTimeFrame,
		TimeFrame:            cfg.RateLimiter.TimeFrame,
	})

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
	}

	app.Mount()

	println("Starting API...")

	app.Run()
}
