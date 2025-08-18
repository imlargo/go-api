package services

import (
	"context"
	"errors"
	"time"

	"github.com/imlargo/go-api-template/internal/enums"
	"github.com/imlargo/go-api-template/internal/models"
	"github.com/imlargo/go-api-template/pkg/push"
	"github.com/imlargo/go-api-template/pkg/sse"
)

type NotificationService interface {
	DispatchNotification(userID uint, title, message, notifType string) error

	DispatchSSE(notification *models.Notification) error
	SubscribeSSE(ctx context.Context, userID uint, deviceID string) (sse.Connection, error)
	UnsubscribeSSE(userID uint, deviceID string) error

	DispatchPush(userID uint, notification *models.Notification) error
	SubscribePush(userID uint, endpoint string, p256dh string, auth string) (*models.PushNotificationSubscription, error)
	UnsubscribePush(subscriptionID uint) error

	GetSSESubscriptions() map[string]interface{}
	GetPushSubscription(subscriptionID uint) (*models.PushNotificationSubscription, error)

	GetUserNotifications(userID uint) ([]*models.Notification, error)
	MarkNotificationsAsRead(userID uint) error
}

type notificationService struct {
	*Service
	SSE  sse.SSEManager
	Push push.PushNotifier
}

func NewNotificationService(service *Service, sse sse.SSEManager, push push.PushNotifier) NotificationService {
	return &notificationService{
		Service: service,
		SSE:     sse,
		Push:    push,
	}
}

func (s *notificationService) DispatchNotification(userID uint, title, message string, notifType string) error {
	notification := &models.Notification{
		UserID:      userID,
		Title:       title,
		Description: message,
		Category:    enums.NotificationType(notifType),
		Read:        false,
	}

	err := s.DispatchSSE(notification)
	if err != nil {
		s.logger.Errorln("Error dispatching SSE notification:", err.Error())
	}

	err = s.DispatchPush(userID, notification)
	if err != nil {
		s.logger.Errorln("Error dispatching push notification:", err.Error())
	}

	return nil
}

func (s *notificationService) DispatchSSE(notification *models.Notification) error {
	// 1. Guardar en base de datos
	err := s.store.Notifications.Create(notification)
	if err != nil {
		s.logger.Errorln("Error saving notification to database:", err.Error())
		// return err
	}

	return s.SSE.Send(notification.UserID, &sse.Message{
		Event: "notification",
		Data:  notification,
	})
}

func (s *notificationService) DispatchPush(userID uint, notification *models.Notification) error {

	subs, err := s.store.PushSubscriptions.GetSubscriptionsByUser(userID)

	if err == nil {
		for _, subscription := range subs {
			notification := map[string]interface{}{
				"title":    notification.Title,
				"message":  notification.Description,
				"category": notification.Category,
			}

			err = s.Push.Send(&push.Subscription{
				Endpoint: subscription.Endpoint,
				P256dh:   subscription.P256dh,
				Auth:     subscription.Auth,
			}, notification)
			if err != nil {
				s.store.PushSubscriptions.Delete(subscription.ID)
				continue
			}
		}
	}

	return nil
}

func (s *notificationService) SubscribeSSE(ctx context.Context, userID uint, deviceID string) (sse.Connection, error) {
	return s.SSE.Subscribe(ctx, userID, deviceID)
}

func (s *notificationService) UnsubscribeSSE(userID uint, deviceID string) error {
	return s.SSE.Unsubscribe(userID, deviceID)
}

func (s *notificationService) SubscribePush(userID uint, endpoint string, p256dh string, auth string) (*models.PushNotificationSubscription, error) {
	sub := &models.PushNotificationSubscription{
		UserID:   userID,
		Endpoint: endpoint,
		P256dh:   p256dh,
		Auth:     auth,
	}

	err := s.store.PushSubscriptions.Create(sub)

	if err != nil {
		return nil, err
	}

	return sub, nil
}

func (s *notificationService) UnsubscribePush(subscriptionID uint) error {
	return s.store.PushSubscriptions.Delete(subscriptionID)
}

func (s *notificationService) GetSSESubscriptions() map[string]interface{} {
	return s.SSE.GetSSESubscriptions()
}

func (s *notificationService) GetUserNotifications(userID uint) ([]*models.Notification, error) {
	if userID == 0 {
		return nil, errors.New("user ID is required")
	}

	notifications, err := s.store.Notifications.GetByUser(userID)
	if err != nil {
		return nil, err
	}

	return notifications, nil
}

func (s *notificationService) MarkNotificationsAsRead(userID uint) error {
	if userID == 0 {
		return errors.New("user ID is required")
	}

	now := time.Now()
	err := s.store.Notifications.MarkAsRead(userID, now)
	if err != nil {
		return err
	}

	return nil
}

func (s *notificationService) GetPushSubscription(subscriptionID uint) (*models.PushNotificationSubscription, error) {
	if subscriptionID == 0 {
		return nil, errors.New("subscription ID is required")
	}

	subscription, err := s.store.PushSubscriptions.GetByID(subscriptionID)
	if err != nil {
		return nil, err
	}

	return subscription, nil
}
