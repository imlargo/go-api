package handler

import (
	"github.com/imlargo/go-api/pkg/medusa/core/logger"
)

type Handler struct {
	logger *logger.Logger
}

func NewHandler(
	logger *logger.Logger,
) *Handler {
	return &Handler{
		logger: logger,
	}
}
