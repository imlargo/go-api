package handlers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/gin-gonic/gin"
	requestsdto "github.com/imlargo/go-api-template/internal/application/dto/requests"
	"github.com/imlargo/go-api-template/internal/application/services"
	"github.com/imlargo/go-api-template/internal/enums"
	"github.com/imlargo/go-api-template/internal/models"

	"github.com/imlargo/go-api-template/internal/presentation/http/responses"
)

type NotificationController interface {
	SubscribeSSE(c *gin.Context)
	UnsubscribeSSE(c *gin.Context)
	DispatchSSE(c *gin.Context)

	SubscribePush(c *gin.Context)
	DispatchPush(c *gin.Context)

	GetPushSubscription(c *gin.Context)

	GetSSESubscriptions(c *gin.Context)
	GetUserNotifications(c *gin.Context)
	MarkNotificationsAsRead(c *gin.Context)

	DispatchNotification(c *gin.Context) // Deprecated
}

type NotificationControllerImpl struct {
	notificationService services.NotificationService
}

func NewNotificationController(notificationService services.NotificationService) NotificationController {
	return &NotificationControllerImpl{notificationService: notificationService}
}

// @Summary		Subscribe to notifications
// @Router			/api/v1/notifications/subscribe [get]
// @Description	Subscribe to real time notifications using Server-Sent Events (SSE)
// @Tags			notifications
// @Accept			json
// @Produce		text/event-stream
// @Param			user_id query	int	true	"User ID"
// @Param			device_id query	string	true	"Device ID"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (u *NotificationControllerImpl) SubscribeSSE(c *gin.Context) {
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

	client, err := u.notificationService.SubscribeSSE(c.Request.Context(), uint(userID), deviceID)
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

// @Summary		Send notification
// @Router			/api/v1/notifications/send [post]
// @Description	Send realtime notifications to a user
// @Tags			notifications
// @Accept			json
// @Produce		json
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Param		payload body	requestsdto.SendNotificationRequestPayload	true	"Notification Payload"
// @Security     PushApiKey
func (h *NotificationControllerImpl) DispatchSSE(c *gin.Context) {
	var payload requestsdto.SendNotificationRequestPayload
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

// @Summary		Unsuscribe SSE
// @Router			/api/v1/notifications/unsubscribe [post]
// @Description	Unsubscribe from real time notifications using SSE
// @Tags			notifications
// @Accept			json
// @Produce		json
// @Param			payload body	requestsdto.NotificationSubscriptionPayload	true	"Unsubscribe Payload"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (h *NotificationControllerImpl) UnsubscribeSSE(c *gin.Context) {
	var payload requestsdto.NotificationSubscriptionPayload

	if err := c.ShouldBindJSON(&payload); err != nil {
		responses.ErrorBindJson(c, err)
		return
	}

	err := h.notificationService.UnsubscribeSSE(payload.UserID, payload.DeviceID)
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, fmt.Sprintf("error unsubscribing: %v", err))
		return
	}

	if !c.IsAborted() {
		responses.Ok(c, gin.H{
			"message":   "Unsubscription successful",
			"user_id":   payload.UserID,
			"device_id": payload.DeviceID,
		})
		return
	}
}

// @Summary		Subscribe to Push Notifications
// @Router			/api/v1/notifications/push/subscribe/{userID} [post]
// @Description	Subscribe to Push Notifications
// @Tags			notifications
// @Param			userID	path	string				true	"User ID"
// @Accept			json
// @Produce		json
// @Param			sub	body	webpush.Subscription	true	"Subscription Payload"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (u *NotificationControllerImpl) SubscribePush(c *gin.Context) {
	var sub webpush.Subscription

	userID := c.Param("userID")
	if userID == "" {
		responses.ErrorBadRequest(c, "userID is required")
	}

	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		responses.ErrorBadRequest(c, "invalid userID")
		return
	}

	if err := c.ShouldBindJSON(&sub); err != nil {
		responses.ErrorBindJson(c, err)
		return
	}

	subscription, err2 := u.notificationService.SubscribePush(uint(userIDInt), sub.Endpoint, sub.Keys.P256dh, sub.Keys.Auth)
	if err2 != nil {
		responses.ErrorInternalServer(c)
		return
	}

	if !c.IsAborted() {
		responses.Ok(c, subscription)
	}

}

// @Summary		Send Push Notification
// @Router			/api/v1/notifications/push/send [post]
// @Description	 Send Push Notification
// @Tags			notifications
// @Accept			json
// @Produce		json
// @Param		payload body	requestsdto.PushNotificationRequestPayload	true	"Push Notification Payload"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     PushApiKey
func (u *NotificationControllerImpl) DispatchPush(c *gin.Context) {

	var payload requestsdto.PushNotificationRequestPayload

	if err := c.ShouldBindJSON(&payload); err != nil {
		responses.ErrorBindJson(c, err)
		return
	}

	err := u.notificationService.DispatchPush(payload.UserID, &models.Notification{
		Title:       payload.Title,
		Description: payload.Message,
		Category:    enums.NotificationType(payload.Category),
	})

	if err != nil {
		responses.ErrorInternalServerWithMessage(c, fmt.Sprintf("error sending push notification: %v", err))
		return
	}

	if !c.IsAborted() {
		responses.Ok(c, "Notification sent successfully")
	}
}

// @Summary		Get SSE Subscriptions
// @Router			/api/v1/notifications/subscriptions [get]
// @Description	Get all SSE subscriptions
// @Tags			notifications
// @Accept			json
// @Produce		json
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     PushApiKey
func (u *NotificationControllerImpl) GetSSESubscriptions(c *gin.Context) {
	subscriptions := u.notificationService.GetSSESubscriptions()
	if subscriptions == nil {
		responses.ErrorInternalServer(c)
		return
	}

	if !c.IsAborted() {
		responses.Ok(c, subscriptions)
	}
}

// @Summary		Get User Notifications
// @Router			/api/v1/notifications [get]
// @Description	Get all notifications for a user
// @Tags			notifications
// @Accept			json
// @Produce		json
// @Param			user_id query	int	true	"User ID"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error
// @Security     BearerAuth
func (u *NotificationControllerImpl) GetUserNotifications(c *gin.Context) {
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		responses.ErrorBadRequest(c, "user_id is required")
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		responses.ErrorBadRequest(c, "invalid user_id")
		return
	}

	notifications, err := u.notificationService.GetUserNotifications(uint(userID))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, fmt.Sprintf("error fetching notifications: %v", err))
		return
	}

	responses.Ok(c, notifications)
}

// @Summary		Mark Notifications as Read
// @Router			/api/v1/notifications/read [post]
// @Description	Mark notifications as read for a user since a specific time
// @Tags			notifications
// @Accept			json
// @Produce		json
// @Param			user_id query	int	true	"User ID"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (u *NotificationControllerImpl) MarkNotificationsAsRead(c *gin.Context) {
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		responses.ErrorBadRequest(c, "user_id is required")
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		responses.ErrorBadRequest(c, "invalid user_id")
		return
	}

	err = u.notificationService.MarkNotificationsAsRead(uint(userID))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, fmt.Sprintf("error marking notifications as read: %v", err))
		return
	}

	responses.Ok(c, "Notifications marked as read successfully")
}

// @Summary		Get Push Subscription
// @Router			/api/v1/notifications/push/subscriptions/{id} [get]
// @Description	Get a specific push notification subscription by ID
// @Tags			notifications
// @Accept			json
// @Produce		json
// @Param			id  path	int	true	"Subscription ID"
// @Success		200	{object}	models.PushNotificationSubscription	"Push Notification Subscription"
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		404	{object}	responses.ErrorResponse	"Not Found"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     BearerAuth
func (u *NotificationControllerImpl) GetPushSubscription(c *gin.Context) {

	subscriptionIDStr := c.Param("id")
	if subscriptionIDStr == "" {
		responses.ErrorBadRequest(c, "subscriptionID is required")
		return
	}

	subscriptionID, err := strconv.Atoi(subscriptionIDStr)
	if err != nil {
		responses.ErrorBadRequest(c, "invalid subscriptionID")
		return
	}

	subscription, err := u.notificationService.GetPushSubscription(uint(subscriptionID))
	if err != nil {
		responses.ErrorInternalServerWithMessage(c, fmt.Sprintf("error fetching subscription: %v", err))
		return
	}

	if subscription == nil {
		responses.ErrorNotFound(c, "Subscription not found")
		return
	}

	responses.Ok(c, subscription)
}

// @Summary		Dispatch Notification
// @Router			/api/v1/notifications/dispatch [post]
// @Description	Dispatch a notification (deprecated)
// @Tags			notifications
// @Accept			json
// @Produce		json
// @Failure		400	{object}	responses.ErrorResponse	"Bad Request"
// @Failure		500	{object}	responses.ErrorResponse	"Internal Server Error"
// @Security     PushApiKey
func (u *NotificationControllerImpl) DispatchNotification(c *gin.Context) {
	u.notificationService.DispatchNotification(36, "Test Notification", "This is a test notification", string(enums.NotificationTypeBase))
}
