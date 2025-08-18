package repositories

import (
	"github.com/imlargo/go-api-template/internal/domain/models"
	"gorm.io/gorm"
)

type PushNotificationSubscriptionRepository interface {
	Create(subscription *models.PushNotificationSubscription) error
	GetSubscriptionsByUser(id uint) ([]*models.PushNotificationSubscription, error)
	Delete(id uint) error
	GetByID(id uint) (*models.PushNotificationSubscription, error)
}

type pushSubscriptionRepositoryImpl struct {
	db *gorm.DB
}

func NewPushSubscriptionRepository(db *gorm.DB) PushNotificationSubscriptionRepository {
	return &pushSubscriptionRepositoryImpl{
		db: db,
	}
}

func (r *pushSubscriptionRepositoryImpl) Create(subscription *models.PushNotificationSubscription) error {
	return r.db.Create(subscription).Error
}

func (r *pushSubscriptionRepositoryImpl) GetSubscriptionsByUser(id uint) ([]*models.PushNotificationSubscription, error) {
	var subscriptions []*models.PushNotificationSubscription
	if err := r.db.Where("user_id = ?", id).Find(&subscriptions).Error; err != nil {
		return nil, err
	}
	return subscriptions, nil
}

func (r *pushSubscriptionRepositoryImpl) Delete(id uint) error {
	if err := r.db.Where("id = ?", id).Delete(&models.PushNotificationSubscription{}).Error; err != nil {
		return err
	}
	return nil
}

func (r *pushSubscriptionRepositoryImpl) GetByID(id uint) (*models.PushNotificationSubscription, error) {
	var subscription models.PushNotificationSubscription
	if err := r.db.First(&subscription, id).Error; err != nil {
		return nil, err
	}
	return &subscription, nil
}
