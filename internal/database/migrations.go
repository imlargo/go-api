package postgres

import (
	"github.com/imlargo/go-api/internal/models"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {

	err := db.AutoMigrate(
		&models.User{},
		&models.Notification{},
		&models.PushNotificationSubscription{},
		&models.File{},
	)

	return err
}
