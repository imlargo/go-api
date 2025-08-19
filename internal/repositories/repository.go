package repositories

import (
	"github.com/imlargo/go-api-template/internal/cache"
	"github.com/imlargo/go-api-template/pkg/kv"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Repository struct {
	db        *gorm.DB
	cacheKeys *cache.CacheKeys
	cache     kv.KeyValueStore
	logger    *zap.SugaredLogger
}

func NewRepository(
	db *gorm.DB,
	cacheKeys *cache.CacheKeys,
	cache kv.KeyValueStore,
	logger *zap.SugaredLogger,
) *Repository {
	return &Repository{
		db,
		cacheKeys,
		cache,
		logger,
	}
}
