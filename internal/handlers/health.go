package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HealthHandler struct {
	*Handler
	db    *gorm.DB
	redis interface{ Ping() error }
}

type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Checks    map[string]string `json:"checks"`
	Version   string            `json:"version,omitempty"`
}

func NewHealthHandler(h *Handler, db *gorm.DB, redis interface{ Ping() error }) *HealthHandler {
	return &HealthHandler{
		Handler: h,
		db:      db,
		redis:   redis,
	}
}

// Health godoc
// @Summary Health check endpoint
// @Description Returns the health status of the API and its dependencies
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse "Healthy"
// @Failure 503 {object} responses.ErrorResponse "Service Unavailable"
// @Router /health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	checks := make(map[string]string)
	overallStatus := "healthy"

	// Check database
	if h.db != nil {
		sqlDB, err := h.db.DB()
		if err != nil {
			checks["database"] = "error: " + err.Error()
			overallStatus = "unhealthy"
		} else {
			err = sqlDB.PingContext(ctx)
			if err != nil {
				checks["database"] = "unreachable: " + err.Error()
				overallStatus = "unhealthy"
			} else {
				checks["database"] = "healthy"
			}
		}
	} else {
		checks["database"] = "not configured"
	}

	// Check Redis
	if h.redis != nil {
		err := h.redis.Ping()
		if err != nil {
			checks["redis"] = "unreachable: " + err.Error()
			overallStatus = "unhealthy"
		} else {
			checks["redis"] = "healthy"
		}
	} else {
		checks["redis"] = "not configured"
	}

	response := HealthResponse{
		Status:    overallStatus,
		Timestamp: time.Now().Format(time.RFC3339),
		Checks:    checks,
		Version:   "1.0.0", // TODO: get from build info
	}

	if overallStatus == "unhealthy" {
		c.JSON(http.StatusServiceUnavailable, response)
		return
	}

	c.JSON(http.StatusOK, response)
}

// Readiness godoc
// @Summary Readiness check endpoint
// @Description Returns whether the API is ready to serve requests
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string "Ready"
// @Router /ready [get]
func (h *HealthHandler) Readiness(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]string{
		"status":    "ready",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// Liveness godoc
// @Summary Liveness check endpoint
// @Description Returns whether the API is alive (basic ping)
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string "Alive"
// @Router /live [get]
func (h *HealthHandler) Liveness(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]string{
		"status":    "alive",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}