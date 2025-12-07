package service

import (
	"github.com/imlargo/go-api/internal/config"
	"github.com/imlargo/go-api/internal/store"
	"github.com/imlargo/go-api/pkg/medusa/core/logger"
)

type Service struct {
	store  *store.Store
	logger *logger.Logger
	config *config.Config
}

func NewService(
	store *store.Store,
	logger *logger.Logger,
	config *config.Config,
) *Service {
	return &Service{
		store,
		logger,
		config,
	}
}
