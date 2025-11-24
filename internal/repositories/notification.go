package repositories

import (
	"time"

	"github.com/nicolailuther/butter/internal/models"
)

type NotificationRepository interface {
	Create(notification *models.Notification) error
	GetByUser(id uint) ([]*models.Notification, error)
	MarkAsRead(userID uint, since time.Time) error
}

type notificationRepository struct {
	*Repository
}

func NewNotificationRepository(
	r *Repository,
) NotificationRepository {
	return &notificationRepository{
		Repository: r,
	}
}

func (r *notificationRepository) Create(notification *models.Notification) error {
	return r.db.Create(notification).Error
}

func (r *notificationRepository) GetByUser(userID uint) ([]*models.Notification, error) {
	var notifications []*models.Notification
	if err := r.db.Order("created_at desc").Where(&models.Notification{UserID: userID}).Limit(100).Find(&notifications).Error; err != nil {
		return nil, err
	}
	return notifications, nil
}

func (r *notificationRepository) MarkAsRead(userID uint, since time.Time) error {
	result := r.db.Model(&models.Notification{}).
		Where("user_id = ? AND created_at < ? AND read = ?", userID, since, false).
		Update("read", true)

	if result.Error != nil {
		return result.Error
	}

	return nil
}
