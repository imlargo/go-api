package handlers

import (
	"github.com/imlargo/go-api/pkg/logger"
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
