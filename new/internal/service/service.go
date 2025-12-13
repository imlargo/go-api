package service

import (
	"github.com/imlargo/go-api/internal/config"
	"github.com/imlargo/go-api/internal/store"
	medusaservice "github.com/imlargo/go-api/pkg/medusa/core/service"
)

type Service struct {
	medusaservice.Service
	store  *store.Store
	config *config.Config
}

func NewService(
	medusa medusaservice.Service,
	store *store.Store,
	config *config.Config,
) *Service {
	return &Service{
		medusa,
		store,
		config,
	}
}
