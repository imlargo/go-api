package models

import (
	"time"

	"github.com/imlargo/go-api-template/internal/domain/enums"
)

type Notification struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at" gorm:"index"`

	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Category    enums.NotificationType `json:"category"`
	Read        bool                   `json:"read" gorm:"default:false;index:idx_user_read"`
	UserID      uint                   `json:"user_id" gorm:"index;index:idx_user_read"`

	User *User `swaggerignore:"true" json:"-"`
}

type PushNotificationSubscription struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	Endpoint string `json:"endpoint"`
	P256dh   string `json:"p256dh"`
	Auth     string `json:"auth"`

	UserID uint `json:"user_id" gorm:"index"`

	User *User `swaggerignore:"true" json:"-"`
}

func (s *PushNotificationSubscription) ToWebPush() map[string]string {
	return map[string]string{
		"endpoint": s.Endpoint,
		"p256dh":   s.P256dh,
		"auth":     s.Auth,
	}
}
