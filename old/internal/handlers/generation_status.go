package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nicolailuther/butter/internal/enums"
	_ "github.com/nicolailuther/butter/internal/models"
	"github.com/nicolailuther/butter/internal/responses"
	"github.com/nicolailuther/butter/internal/services"
	"github.com/redis/go-redis/v9"
)

type GenerationStatusHandler struct {
	*Handler
	statusService services.GenerationStatusService
	redisClient   *redis.Client
}

func NewGenerationStatusHandler(handler *Handler, statusService services.GenerationStatusService, redisClient *redis.Client) *GenerationStatusHandler {
	return &GenerationStatusHandler{
		Handler:       handler,
		statusService: statusService,
		redisClient:   redisClient,
	}
}

// @Summary		Get generation status for account
// @Router			/api/v2/content/generation/status [get]
// @Description	Get current generation status for an account
// @Tags			generation
// @Produce		json
// @Param			account_id query uint true "Account ID to check status"
// @Param			content_type query string false "Content type filter (video, story, slideshow)"
// @Success		200	{object}	models.AccountGenerationStatus "Generation status retrieved successfully"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"No active generation found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *GenerationStatusHandler) GetGenerationStatus(c *gin.Context) {
	accountIDStr := c.Query("account_id")
	if accountIDStr == "" {
		responses.ErrorBadRequest(c, "account_id is required")
		return
	}

	accountID, err := strconv.ParseUint(accountIDStr, 10, 32)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid account_id: "+err.Error())
		return
	}

	contentTypeStr := c.DefaultQuery("content_type", "video")
	contentType := enums.ContentType(contentTypeStr)

	status, err := h.statusService.GetActiveStatus(uint(accountID), contentType)
	if err != nil {
		responses.ErrorNotFound(c, "Active generation")
		return
	}

	responses.Ok(c, status)
}

// @Summary		Get latest generation status for account
// @Router			/api/v2/content/generation/latest-status [get]
// @Description	Get the most recent generation status for an account (active or completed)
// @Tags			generation
// @Produce		json
// @Param			account_id query uint true "Account ID to check status"
// @Param			content_type query string false "Content type filter (video, story, slideshow)"
// @Success		200	{object}	models.AccountGenerationStatus "Latest generation status retrieved successfully"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"No generation found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *GenerationStatusHandler) GetLatestGenerationStatus(c *gin.Context) {
	accountIDStr := c.Query("account_id")
	if accountIDStr == "" {
		responses.ErrorBadRequest(c, "account_id is required")
		return
	}

	accountID, err := strconv.ParseUint(accountIDStr, 10, 32)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid account_id: "+err.Error())
		return
	}

	contentTypeStr := c.DefaultQuery("content_type", "video")
	contentType := enums.ContentType(contentTypeStr)

	status, err := h.statusService.GetLatestStatus(uint(accountID), contentType)
	if err != nil {
		responses.ErrorNotFound(c, "Generation status")
		return
	}

	responses.Ok(c, status)
}

// @Summary		Get all generation status records
// @Router			/api/v2/content/generation/history [get]
// @Description	Get generation history for an account
// @Tags			generation
// @Produce		json
// @Param			account_id query uint true "Account ID to get history"
// @Param			limit query int false "Maximum number of results (default: 10)"
// @Success		200	{array}		models.AccountGenerationStatus "Generation history retrieved successfully"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *GenerationStatusHandler) GetGenerationHistory(c *gin.Context) {
	accountIDStr := c.Query("account_id")
	if accountIDStr == "" {
		responses.ErrorBadRequest(c, "account_id is required")
		return
	}

	accountID, err := strconv.ParseUint(accountIDStr, 10, 32)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid account_id: "+err.Error())
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}

	history, err := h.statusService.GetByAccountID(uint(accountID), limit)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, "Failed to get generation history: "+err.Error())
		return
	}

	responses.Ok(c, history)
}

// @Summary		Stream generation events via SSE
// @Router			/api/v2/content/generation/events [get]
// @Description	Subscribe to real-time generation status updates via Server-Sent Events
// @Tags			generation
// @Produce		text/event-stream
// @Param			account_id query uint true "Account ID to subscribe to updates"
// @Success		200	"Event stream"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Security     BearerAuth
func (h *GenerationStatusHandler) StreamGenerationEvents(c *gin.Context) {
	accountIDStr := c.Query("account_id")
	if accountIDStr == "" {
		responses.ErrorBadRequest(c, "account_id is required")
		return
	}

	accountID, err := strconv.ParseUint(accountIDStr, 10, 32)
	if err != nil {
		responses.ErrorBadRequest(c, "Invalid account_id: "+err.Error())
		return
	}

	// Set headers for SSE
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	// Create context with cancellation
	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	// Subscribe to Redis pub/sub for task events
	pubsub := h.redisClient.Subscribe(ctx, "repurposer:events")
	defer pubsub.Close()

	// Send initial connection message
	fmt.Fprintf(c.Writer, "event: connected\ndata: {\"message\": \"Connected to generation events\"}\n\n")
	c.Writer.Flush()

	// Poll for events
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-pubsub.Channel():
			if msg == nil {
				continue
			}

			// Parse event
			var event map[string]interface{}
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				continue
			}

			// Check if event is for this account
			if eventAccountID, ok := event["account_id"].(float64); ok && uint(eventAccountID) == uint(accountID) {
				// Send event to client
				eventType, _ := event["event_type"].(string)
				data, _ := json.Marshal(event)
				fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", eventType, string(data))
				c.Writer.Flush()
			}
		case <-ticker.C:
			// Send heartbeat
			fmt.Fprintf(c.Writer, ": heartbeat\n\n")
			c.Writer.Flush()

			// Check if connection is still alive
			if _, err := c.Writer.Write([]byte{}); err != nil {
				if err != io.EOF {
					h.logger.Warnw("SSE connection error", "error", err, "account_id", accountID)
				}
				return
			}
		}
	}
}
