package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/nicolailuther/butter/internal/responses"
)

type HealthHandler struct {
	*Handler
}

func NewHealthHandler(handler *Handler) *HealthHandler {
	return &HealthHandler{
		Handler: handler,
	}
}

// @Summary 		Health
// @Router			/health [get]
// @Description	Get health status
// @Tags			health
// @Produce		json
func (h *HealthHandler) GetHealth(c *gin.Context) {
	responses.Ok(c, "ok")
}
