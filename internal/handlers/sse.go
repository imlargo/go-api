package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/imlargo/go-api/pkg/medusa/core/handler"
)

type Handler struct {
	*handler.Handler
}

func NewSSEHandler(handler *handler.Handler) *Handler {
	return &Handler{Handler: handler}
}

func (h *Handler) Listen(c *gin.Context) {

}

func (h *Handler) Publish(c *gin.Context) {

}
