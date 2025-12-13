package handlers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/imlargo/go-api/pkg/medusa/core/handler"
	"github.com/imlargo/go-api/pkg/medusa/core/responses"
	"github.com/imlargo/go-api/pkg/medusa/services/sse"
)

type Handler struct {
	*handler.Handler
	sseService sse.SSEManager
}

func NewSSEHandler(handler *handler.Handler) *Handler {
	sse := sse.NewSSEManager()
	return &Handler{
		Handler:    handler,
		sseService: sse,
	}
}

func (h *Handler) Listen(c *gin.Context) {
	userIDStr := c.Query("user_id")
	deviceID := c.Query("device_id")

	if userIDStr == "" || deviceID == "" {
		responses.ErrorBadRequest(c, "user_id and device_id are required")
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		responses.ErrorBadRequest(c, "invalid user_id")
		return
	}

	// SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Cache-Control")

	client, err := h.sseService.Subscribe(c.Request.Context(), uint(userID), deviceID)
	if err != nil {
		responses.ErrorBadRequest(c, fmt.Sprintf("error subscribing: %v", err))
		return
	}

	c.SSEvent("connected", gin.H{
		"user_id":   userID,
		"device_id": deviceID,
		"timestamp": time.Now().Unix(),
	})
	c.Writer.Flush()

	pingTicker := time.NewTicker(30 * time.Second)
	defer pingTicker.Stop()

	for {
		select {
		case notification, ok := <-client.GetChannel():
			if !ok {
				return // Closed channel
			}

			client.UpdateLastSeen()

			c.SSEvent("notification", notification)
			c.Writer.Flush()

		case <-pingTicker.C:
			c.SSEvent("ping", gin.H{"timestamp": time.Now().Unix()})
			c.Writer.Flush()

		case <-c.Request.Context().Done():
			return

		case <-client.GetContext().Done():
			return
		}
	}
}

func (h *Handler) Publish(c *gin.Context) {
	var payload dto.SendNotificationRequestPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		responses.ErrorBindJson(c, err)
		return
	}

	if payload.UserID == 0 {
		responses.ErrorBadRequest(c, "user_id is required")
		return
	}

	if payload.Notification.ID == 0 {
		responses.ErrorBadRequest(c, "notification_id is required")
		return
	}

	err := h.notificationService.DispatchSSE(&payload.Notification)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, fmt.Sprintf("error sending notification: %v", err))
		return
	}

	if !c.IsAborted() {
		responses.Ok(c, "Notification sent successfully")
	}
}
