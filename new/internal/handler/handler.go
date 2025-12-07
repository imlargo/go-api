package handlers

import (
	"github.com/imlargo/go-api/pkg/medusa/logger"
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
