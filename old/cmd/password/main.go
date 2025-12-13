package main

import (
	"os"

	"github.com/nicolailuther/butter/internal/cache"
	"github.com/nicolailuther/butter/internal/cache/redis"
	"github.com/nicolailuther/butter/internal/config"
	"github.com/nicolailuther/butter/internal/database"
	"github.com/nicolailuther/butter/internal/repositories"
	"github.com/nicolailuther/butter/internal/services"
	"github.com/nicolailuther/butter/internal/store"
	"github.com/nicolailuther/butter/pkg/kv"
	"github.com/nicolailuther/butter/pkg/storage"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Logger
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	// Read email and new password from positional command-line arguments.
	// Usage: go run cmd/password/main.go <email> <newPassword>
	if len(os.Args) < 3 {
		logger.Fatalf("Usage: %s <email> <newPassword>", os.Args[0])
	}

	email := os.Args[1]
	newPassword := os.Args[2]

	cfg := config.LoadConfig()

	// Database
	db, err := database.NewPostgres(cfg.Database.URL)
	if err != nil {
		logger.Fatal("Could not initialize database: ", err)
	}

	// Storage
	_, err = storage.NewR2Storage(storage.StorageConfig{
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

	_ = services.NewService(store, logger, &cfg, cacheKeys, cacheService)

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		logger.Fatal("Failed to hash password: ", err)
	}

	user, err := store.Users.GetByEmail(email)
	if err != nil {
		logger.Fatal("Failed to find user: ", err)
	}

	// Update user password
	user.Password = string(hashedPassword)

	if err := store.Users.Update(user); err != nil {
		logger.Fatal("Failed to update password: ", err)
	}
}
