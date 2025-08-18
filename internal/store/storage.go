package store

import (
	"github.com/imlargo/go-api-template/internal/infrastructure/cache"
	"github.com/imlargo/go-api-template/internal/repositories"
	"gorm.io/gorm"
)

type Storage struct {
	Files             repositories.FileRepository
	PushSubscriptions repositories.PushNotificationSubscriptionRepository
	Notifications     repositories.NotificationRepository
	Users             repositories.UserRepository
}

func NewStorage(db *gorm.DB, cacheService cache.CacheService, cacheKeys cache.CacheKeys) *Storage {
	return &Storage{
		Files:             repositories.NewFileRepository(db),
		Notifications:     repositories.NewNotificationRepository(db),
		PushSubscriptions: repositories.NewPushSubscriptionRepository(db),
		Users:             repositories.NewUserRepository(db, cacheService, cacheKeys),
	}
}
