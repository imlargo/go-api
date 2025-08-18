package services

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/imlargo/go-api-template/internal/domain/enums"
	"github.com/imlargo/go-api-template/internal/domain/models"
	"github.com/imlargo/go-api-template/internal/shared/ports"
	"github.com/imlargo/go-api-template/internal/store"
)

type NotificationService interface {
	DispatchNotification(userID uint, title, message, notifType string) error

	DispatchSSE(notification *models.Notification) error
	SubscribeSSE(ctx context.Context, userID uint, deviceID string) (ports.SSENotificationConnection, error)
	UnsubscribeSSE(userID uint, deviceID string) error

	DispatchPush(userID uint, notification *models.Notification) error
	SubscribePush(userID uint, endpoint string, p256dh string, auth string) (*models.PushNotificationSubscription, error)
	UnsubscribePush(subscriptionID uint) error

	GetSSESubscriptions() map[string]interface{}
	GetPushSubscription(subscriptionID uint) (*models.PushNotificationSubscription, error)

	GetUserNotifications(userID uint) ([]*models.Notification, error)
	MarkNotificationsAsRead(userID uint) error
}

type notificationServiceImpl struct {
	store *store.Storage
	SSE   ports.SSENotificationDispatcher
	Push  ports.PushNotifier
}

func NewNotificationService(store *store.Storage, sse ports.SSENotificationDispatcher, push ports.PushNotifier) NotificationService {
	return &notificationServiceImpl{
		store: store,
		SSE:   sse,
		Push:  push,
	}
}

func (d *notificationServiceImpl) DispatchNotification(userID uint, title, message string, notifType string) error {
	notification := &models.Notification{
		UserID:      userID,
		Title:       title,
		Description: message,
		Category:    enums.NotificationType(notifType),
		Read:        false,
	}

	err := d.DispatchSSE(notification)
	if err != nil {
		log.Println("Error dispatching SSE notification:", err.Error())
	}

	err = d.DispatchPush(userID, notification)
	if err != nil {
		log.Println("Error dispatching push notification:", err.Error())
	}

	return nil
}

func (d *notificationServiceImpl) DispatchSSE(notification *models.Notification) error {
	// 1. Guardar en base de datos
	err := d.store.Notifications.Create(notification)
	if err != nil {
		log.Println("Error saving notification to database:", err.Error())
		// return err
	}

	return d.SSE.Send(notification)
}

func (d *notificationServiceImpl) DispatchPush(userID uint, notification *models.Notification) error {

	subs, err := d.store.PushSubscriptions.GetSubscriptionsByUser(userID)

	if err == nil {
		for _, subscription := range subs {
			webpushSub := &webpush.Subscription{
				Endpoint: subscription.Endpoint,
				Keys: webpush.Keys{
					P256dh: subscription.P256dh,
					Auth:   subscription.Auth,
				},
			}

			notification := map[string]interface{}{
				"title":    notification.Title,
				"message":  notification.Description,
				"category": notification.Category,
			}

			err = d.Push.SendNotification(webpushSub, notification)
			if err != nil {
				d.store.PushSubscriptions.Delete(subscription.ID)
				continue
			}
		}
	}

	return nil
}

func (d *notificationServiceImpl) SubscribeSSE(ctx context.Context, userID uint, deviceID string) (ports.SSENotificationConnection, error) {
	return d.SSE.Subscribe(ctx, userID, deviceID)
}

func (d *notificationServiceImpl) UnsubscribeSSE(userID uint, deviceID string) error {
	return d.SSE.Unsubscribe(userID, deviceID)
}

func (d *notificationServiceImpl) SubscribePush(userID uint, endpoint string, p256dh string, auth string) (*models.PushNotificationSubscription, error) {
	sub := &models.PushNotificationSubscription{
		UserID:   userID,
		Endpoint: endpoint,
		P256dh:   p256dh,
		Auth:     auth,
	}

	err := d.store.PushSubscriptions.Create(sub)

	if err != nil {
		return nil, err
	}

	return sub, nil
}

func (d *notificationServiceImpl) UnsubscribePush(subscriptionID uint) error {
	return d.store.PushSubscriptions.Delete(subscriptionID)
}

func (d *notificationServiceImpl) GetSSESubscriptions() map[string]interface{} {
	return d.SSE.GetSSESubscriptions()
}

func (d *notificationServiceImpl) GetUserNotifications(userID uint) ([]*models.Notification, error) {
	if userID == 0 {
		return nil, errors.New("user ID is required")
	}

	notifications, err := d.store.Notifications.GetByUser(userID)
	if err != nil {
		return nil, err
	}

	return notifications, nil
}

func (d *notificationServiceImpl) MarkNotificationsAsRead(userID uint) error {
	if userID == 0 {
		return errors.New("user ID is required")
	}

	now := time.Now()
	err := d.store.Notifications.MarkAsRead(userID, now)
	if err != nil {
		return err
	}

	return nil
}

func (d *notificationServiceImpl) GetPushSubscription(subscriptionID uint) (*models.PushNotificationSubscription, error) {
	if subscriptionID == 0 {
		return nil, errors.New("subscription ID is required")
	}

	subscription, err := d.store.PushSubscriptions.GetByID(subscriptionID)
	if err != nil {
		return nil, err
	}

	return subscription, nil
}
