package main

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/nicolailuther/butter/internal/cache"
	"github.com/nicolailuther/butter/internal/cache/redis"
	"github.com/nicolailuther/butter/internal/config"
	"github.com/nicolailuther/butter/internal/database"
	"github.com/nicolailuther/butter/internal/metrics"
	"github.com/nicolailuther/butter/internal/repositories"
	"github.com/nicolailuther/butter/internal/store"
	"github.com/nicolailuther/butter/pkg/kv"
	"github.com/nicolailuther/butter/pkg/ratelimiter"
	"github.com/nicolailuther/butter/pkg/storage"
	"go.uber.org/zap"
)

// @title Butter API
// @version 1.0
// @description Backend service for Butter

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
	db, err := database.NewPostgres(cfg.Database.URL)
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
	store := store.NewStorage(repositoryContainer, db)

	// Metrics
	metricsService := metrics.NewPrometheusMetrics()

	router := gin.Default()

	// Configure trusted proxies to properly handle X-Forwarded-For headers
	// This is critical for users behind proxies (e.g., UK users)
	// Trust Cloudflare and localhost only by default
	// If deployed behind other reverse proxies, configure via environment
	trustedProxies := []string{
		"127.0.0.1", // localhost
		"::1",       // localhost IPv6
		// Cloudflare IP ranges (updated as of 2024)
		"173.245.48.0/20",
		"103.21.244.0/22",
		"103.22.200.0/22",
		"103.31.4.0/22",
		"141.101.64.0/18",
		"108.162.192.0/18",
		"190.93.240.0/20",
		"188.114.96.0/20",
		"197.234.240.0/22",
		"198.41.128.0/17",
		"162.158.0.0/15",
		"104.16.0.0/13",
		"104.24.0.0/14",
		"172.64.0.0/13",
		"131.0.72.0/22",
	}

	// Only trust private networks if explicitly running in development mode
	if gin.Mode() == gin.DebugMode {
		trustedProxies = append(trustedProxies, "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16")
	}

	router.SetTrustedProxies(trustedProxies)

	app := &Application{
		Config:      cfg,
		Store:       store,
		Storage:     storage,
		Metrics:     metricsService,
		Cache:       cacheService,
		CacheKeys:   cacheKeys,
		RateLimiter: rateLimiter,
		Router:      router,
		Logger:      logger,
		RedisClient: redisClient,
	}

	app.Mount()

	logger.Info("Starting API...")

	app.Run()
}
