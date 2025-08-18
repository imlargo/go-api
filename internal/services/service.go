package services

import (
	"github.com/imlargo/go-api-template/internal/config"
	"github.com/imlargo/go-api-template/internal/store"
	"go.uber.org/zap"
)

type Service struct {
	store  *store.Store
	logger *zap.SugaredLogger
	config *config.AppConfig
}

func NewService(
	store *store.Store,
	logger *zap.SugaredLogger,
	config *config.AppConfig,
) *Service {
	return &Service{
		store,
		logger,
		config,
	}
}
