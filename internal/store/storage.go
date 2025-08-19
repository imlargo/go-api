package store

import (
	"github.com/imlargo/go-api-template/internal/repositories"
)

type Store struct {
	Files             repositories.FileRepository
	PushSubscriptions repositories.PushNotificationSubscriptionRepository
	Notifications     repositories.NotificationRepository
	Users             repositories.UserRepository
}

func NewStorage(container *repositories.Repository) *Store {
	return &Store{
		Files:             repositories.NewFileRepository(container),
		Notifications:     repositories.NewNotificationRepository(container),
		PushSubscriptions: repositories.NewPushSubscriptionRepository(container),
		Users:             repositories.NewUserRepository(container),
	}
}
