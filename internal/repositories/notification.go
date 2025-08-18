package repositories

import (
	"time"

	"github.com/imlargo/go-api-template/internal/domain/models"
	"gorm.io/gorm"
)

type NotificationRepository interface {
	Create(notification *models.Notification) error
	GetByUser(id uint) ([]*models.Notification, error)
	MarkAsRead(userID uint, since time.Time) error
}

type notificationRepositoryImpl struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepositoryImpl{
		db: db,
	}
}

func (r *notificationRepositoryImpl) Create(notification *models.Notification) error {
	return r.db.Create(notification).Error
}

func (r *notificationRepositoryImpl) GetByUser(userID uint) ([]*models.Notification, error) {
	var notifications []*models.Notification
	if err := r.db.Order("created_at desc").Where(&models.Notification{UserID: userID}).Limit(100).Find(&notifications).Error; err != nil {
		return nil, err
	}
	return notifications, nil
}

func (r *notificationRepositoryImpl) MarkAsRead(userID uint, since time.Time) error {
	result := r.db.Model(&models.Notification{}).
		Where("user_id = ? AND created_at < ? AND read = ?", userID, since, false).
		Update("read", true)

	if result.Error != nil {
		return result.Error
	}

	return nil
}
