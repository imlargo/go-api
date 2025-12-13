package main

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/nicolailuther/butter/internal/cache"
	"github.com/nicolailuther/butter/internal/cache/redis"
	"github.com/nicolailuther/butter/internal/config"
	"github.com/nicolailuther/butter/internal/database"
	"github.com/nicolailuther/butter/internal/enums"
	"github.com/nicolailuther/butter/internal/models"
	"github.com/nicolailuther/butter/internal/repositories"
	"github.com/nicolailuther/butter/internal/services"
	"github.com/nicolailuther/butter/internal/store"
	"github.com/nicolailuther/butter/pkg/kv"
	"github.com/nicolailuther/butter/pkg/storage"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

	serviceContainer := services.NewService(store, logger, &cfg, cacheKeys, cacheService)
	fileService := services.NewFileService(serviceContainer, storage, cfg.Storage.BucketName)
	println(fileService)

	// Get all accounts
	var accounts []*models.Account
	result := db.FindInBatches(&accounts, 100, func(tx *gorm.DB, batch int) error {

		var toUpdate []*models.Account

		for _, account := range accounts {
			username := extractUsername(account)
			if username != "" {
				println("Updating account ID ", account.ID, " with username ", username)
				toUpdate = append(toUpdate, &models.Account{
					ID:       account.ID,
					Username: username,
				})
			}
		}

		if len(toUpdate) > 0 {
			if err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{
					"username",
				}),
			}).CreateInBatches(toUpdate, 100).Error; err != nil {
				fmt.Printf("Error saving batch %d: %v\n", batch, err)
			}
		}

		return nil
	})

	if result.Error != nil {
		logger.Fatal("Could not fetch clients: ", result.Error)
		return
	}
}

func extractUsername(account *models.Account) string {
	switch account.Platform {
	case enums.PlatformInstagram:
		return extractInstagramUsername(account.AccountUrl)
	default:
		return ""
	}
}

func extractInstagramUsername(accountUrl string) string {
	// Handle empty or whitespace-only input
	accountUrl = strings.TrimSpace(accountUrl)
	if accountUrl == "" {
		return ""
	}

	// Add protocol if missing
	if !strings.HasPrefix(accountUrl, "http://") && !strings.HasPrefix(accountUrl, "https://") {
		accountUrl = "https://" + accountUrl
	}

	// Parse the URL
	u, err := url.Parse(accountUrl)
	if err != nil {
		return ""
	}

	// Check if it's an Instagram URL
	if !strings.Contains(u.Host, "instagram.com") {
		return ""
	}

	// Split the path and find the username
	pathParts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(pathParts) == 0 {
		return ""
	}

	username := pathParts[0]

	// Validate username (basic validation)
	if username == "" || username == "www" {
		return ""
	}

	// Remove common Instagram path prefixes that aren't usernames
	invalidPrefixes := []string{"p", "reel", "reels", "stories", "tv", "explore", "accounts", "direct"}
	for _, prefix := range invalidPrefixes {
		if username == prefix {
			return ""
		}
	}

	return username
}
