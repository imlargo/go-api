package services

import (
	"github.com/imlargo/go-api-template/internal/cache"
	"github.com/imlargo/go-api-template/internal/config"
	"github.com/imlargo/go-api-template/internal/store"
	"github.com/imlargo/go-api-template/pkg/kv"
	"go.uber.org/zap"
)

type Service struct {
	store     *store.Store
	logger    *zap.SugaredLogger
	config    *config.AppConfig
	cacheKeys *cache.CacheKeys
	cache     kv.KeyValueStore
}

func NewService(
	store *store.Store,
	logger *zap.SugaredLogger,
	config *config.AppConfig,
	cacheKeys *cache.CacheKeys,
	cache kv.KeyValueStore,
) *Service {
	return &Service{
		store,
		logger,
		config,
		cacheKeys,
		cache,
	}
}
