package service

import (
	"github.com/imlargo/go-api/pkg/medusa/core/logger"
)

type Service struct {
	logger *logger.Logger
}

func NewService(
	logger *logger.Logger,
) *Service {
	return &Service{
		logger,
	}
}
