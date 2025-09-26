package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/imlargo/go-api/internal/responses"
)

type HealthHandler struct {
	*Handler
}

func NewHealthHandler(handler *Handler) *HealthHandler {
	return &HealthHandler{
		Handler: handler,
	}
}

// HealthCheck returns the health status of the API
// @Summary Check API health status
// @Description Returns the current status of the API service
// @Tags Health
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	responses.Ok(c, "ok")
}
