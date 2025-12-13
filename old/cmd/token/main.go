package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/nicolailuther/butter/internal/cache"
	"github.com/nicolailuther/butter/internal/cache/redis"
	"github.com/nicolailuther/butter/internal/config"
	"github.com/nicolailuther/butter/internal/database"
	"github.com/nicolailuther/butter/internal/repositories"
	"github.com/nicolailuther/butter/internal/services"
	"github.com/nicolailuther/butter/internal/store"
	"github.com/nicolailuther/butter/pkg/jwt"
	"github.com/nicolailuther/butter/pkg/kv"
	"github.com/nicolailuther/butter/pkg/storage"
	"go.uber.org/zap"
)

func main() {
	cfg := config.LoadConfig()

	// Logger
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

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

	jwtAuthenticator := jwt.NewJwt(jwt.Config{
		Secret:   cfg.Auth.JwtSecret,
		Issuer:   cfg.Auth.JwtIssuer,
		Audience: cfg.Auth.JwtAudience,
	})

	// parse email from CLI: prefer --email flag, fallback to first positional arg
	emailFlag := flag.String("email", "", "Email for which to generate an access token")
	flag.Parse()

	email := strings.TrimSpace(*emailFlag)
	if email == "" && flag.NArg() > 0 {
		email = strings.TrimSpace(flag.Arg(0))
	}

	if email == "" {
		fmt.Fprintln(os.Stderr, "Usage: go run cmd/token/main.go --email user@example.com")
		os.Exit(2)
	}

	user, err := store.Users.GetByEmail(strings.ToLower(email))
	if err != nil {
		logger.Fatal("Failed to find user: ", err)
	}

	accessExpiration := time.Now().Add(cfg.Auth.TokenExpiration)
	accessToken, err := jwtAuthenticator.GenerateToken(user.ID, accessExpiration)
	if err != nil {
		logger.Fatal("Failed to generate access token: ", err)
	}

	println(accessToken)
}
