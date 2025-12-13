package handlers

import (
	"github.com/imlargo/go-api/pkg/medusa/core/handler"
)

type Handler struct {
	handler.Handler
}

func NewHandler(handler handler.Handler) *Handler {
	return &Handler{Handler: handler}
}
