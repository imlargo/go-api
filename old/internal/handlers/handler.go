package handlers

import (
	"go.uber.org/zap"
)

type Handler struct {
	logger *zap.SugaredLogger
}

func NewHandler(
	logger *zap.SugaredLogger,
) *Handler {
	return &Handler{
		logger: logger,
	}
}
