package services

import (
	"github.com/imlargo/go-api-template/internal/store"
	"go.uber.org/zap"
)

type Service struct {
	store  *store.Store
	logger *zap.SugaredLogger
}

func NewService(
	store *store.Store,
	logger *zap.SugaredLogger,
) *Service {
	return &Service{
		store,
		logger,
	}
}
