package main

import (
	"time"

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
	"gorm.io/gorm"
)

type UserReport struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	TierLevel    int    `json:"tier_level"`
	ClientCount  int    `json:"client_count"`
	AccountCount int    `json:"account_count"`
}

type TopContentReportDetailed struct {
	ContentID        uint       `json:"content_id"`
	ContentName      string     `json:"content_name"`
	ContentType      string     `json:"content_type"`
	OwnerUserName    string     `json:"owner_user_name"`
	OwnerUserEmail   string     `json:"owner_user_email"`
	OwnerTierLevel   int        `json:"owner_tier_level"`
	ClientName       string     `json:"client_name"`
	AssignedAccounts int        `json:"assigned_accounts"`
	TimesPosted      int        `json:"times_posted"`
	TotalViews       int        `json:"total_views"`
	AverageViews     int        `json:"average_views"`
	TimesGenerated   int        `json:"times_generated"`
	LastGeneratedAt  *time.Time `json:"last_generated_at"` // Use *time.Time to handle NULL
	ContentEnabled   bool       `json:"content_enabled"`
	UseMirror        bool       `json:"use_mirror"`
	UseOverlays      bool       `json:"use_overlays"`
}

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

	_, err = GetUserReport(db)
	if err != nil {
		logger.Fatal("Error generating user report: ", err)
	}

	_, err = GetTop20ContentsByViews(db)
	if err != nil {
		logger.Fatal("Error generating top content report: ", err)
	}

}

// Optimized version using a single query with JOINs and aggregations
func GetUserReport(db *gorm.DB) ([]UserReport, error) {
	var reports []UserReport

	err := db.Raw(`
		SELECT 
			u.id,
			u.name,
			u.email,
			u.tier_level,
			COUNT(DISTINCT c.id) as client_count,
			COUNT(DISTINCT a.id) as account_count
		FROM users u
		LEFT JOIN clients c ON u.id = c.user_id
		LEFT JOIN accounts a ON c.id = a.client_id AND a.deleted_at IS NULL
		WHERE u.deleted_at IS NULL
		GROUP BY u.id, u.name, u.email, u.tier_level
		ORDER BY client_count DESC, account_count DESC, u.name
	`).Scan(&reports).Error

	return reports, err
}

func GetTop20ContentsByViews(db *gorm.DB) ([]TopContentReportDetailed, error) {
	var reports []TopContentReportDetailed

	/*
		select
		  p.account_id,
		  a.username,
		  COUNT(*) as total_posts,
		  SUM(p.is_deleted::int) as deleted_posts,
		  COUNT(*) - SUM(p.is_deleted::int) as active_posts,
		  ROUND(AVG(p.is_deleted::int) * 100, 2) as deleted_percentage
		from
		  posts p
		  left join accounts a on p.account_id = a.id
		where
		  p.created_at >= '2025-09-09'
		  and p.created_at <= '2025-09-20 23:59:59'
		group by
		  p.account_id,
		  a.username
		order by
		  deleted_percentage desc,
		  deleted_posts desc;
	*/

	err := db.Raw(`
		SELECT 
			c.id as content_id,
			c.name as content_name,
			c.type as content_type,
			u.name as owner_user_name,
			u.email as owner_user_email,
			u.tier_level as owner_tier_level,
			cl.name as client_name,
			COUNT(DISTINCT ca.account_id) as assigned_accounts,
			c.times_posted,
			c.total_views,
			c.average_views as average_views,
			c.times_generated,
			c.last_generated_at,
			c.enabled as content_enabled,
			c.use_mirror,
			c.use_overlays
		FROM contents c
		INNER JOIN clients cl ON c.client_id = cl.id
		INNER JOIN users u ON cl.user_id = u.id AND u.deleted_at IS NULL
		LEFT JOIN content_accounts ca ON c.id = ca.content_id
		GROUP BY 
			c.id, 
			c.name, 
			c.type, 
			u.name, 
			u.email,
			u.tier_level,
			cl.name,
			c.times_posted, 
			c.total_views, 
			c.average_views,
			c.times_generated,
			c.last_generated_at,
			c.enabled,
			c.use_mirror,
			c.use_overlays
		ORDER BY c.total_views DESC, c.average_views DESC
		LIMIT 20
	`).Scan(&reports).Error

	return reports, err
}
