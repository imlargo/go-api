package dto

import "github.com/nicolailuther/butter/internal/models"

type SendNotificationRequestPayload struct {
	UserID       uint                `json:"user_id" binding:"required"`
	Notification models.Notification `json:"notification" binding:"required"`
}

type NotificationSubscriptionPayload struct {
	UserID   uint   `json:"user_id"`
	DeviceID string `json:"device_id"`
}

type PushNotificationRequestPayload struct {
	UserID   uint   `json:"userId"`
	Title    string `json:"title"`
	Message  string `json:"message"`
	Category string `json:"category"`
}
