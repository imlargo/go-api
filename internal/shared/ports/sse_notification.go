package ports

import (
	"context"

	"github.com/imlargo/go-api-template/internal/domain/models"
)

type SSENotificationDispatcher interface {
	Send(notification *models.Notification) error
	Subscribe(ctx context.Context, userID uint, deviceID string) (SSENotificationConnection, error)
	Unsubscribe(userID uint, deviceID string) error
	GetSSESubscriptions() map[string]interface{}
}

type SSENotificationConnection interface {
	GetChannel() <-chan *models.Notification
	GetContext() context.Context
	UpdateLastSeen()
}
