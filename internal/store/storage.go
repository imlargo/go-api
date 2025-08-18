package store

import (
	"github.com/imlargo/go-api-template/internal/infrastructure/cache"
	"github.com/imlargo/go-api-template/internal/repositories"
	"github.com/imlargo/go-api-template/pkg/kv"
	"gorm.io/gorm"
)

type Store struct {
	Files             repositories.FileRepository
	PushSubscriptions repositories.PushNotificationSubscriptionRepository
	Notifications     repositories.NotificationRepository
	Users             repositories.UserRepository
}

func NewStorage(db *gorm.DB, cacheService kv.KeyValueStore, cacheKeys cache.CacheKeys) *Store {
	return &Store{
		Files:             repositories.NewFileRepository(db),
		Notifications:     repositories.NewNotificationRepository(db),
		PushSubscriptions: repositories.NewPushSubscriptionRepository(db),
		Users:             repositories.NewUserRepository(db, cacheService, cacheKeys),
	}
}
