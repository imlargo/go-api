package database

import (
	"github.com/nicolailuther/butter/internal/models"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {

	err := db.AutoMigrate(
		&models.User{},
		&models.Notification{},
		&models.PushNotificationSubscription{},
		&models.Client{},
		&models.Account{},
		&models.File{},
		&models.TextOverlay{},
		&models.Post{},
		&models.ReferralCode{},
		&models.ReferralDiscount{},
		&models.AccountAnalytic{},
		&models.PostAnalytic{},
		&models.PostingGoal{},
		&models.MarketplaceCategory{},
		&models.MarketplaceSeller{},
		&models.MarketplaceService{},
		&models.MarketplaceServiceResult{},
		&models.MarketplaceServicePackage{},
		&models.Payment{}, // Universal payment model
		&models.MarketplaceOrder{},
		&models.MarketplaceDeliverable{},
		&models.MarketplaceRevisionRequest{},
		&models.MarketplaceDispute{},
		&models.MarketplaceOrderTimeline{},
		&models.OnlyfansAccount{},
		&models.OnlyfansTransaction{},
		&models.OnlyfansTrackingLink{},
		&models.ChatConversation{},
		&models.ChatMessage{},
		&models.ContentFolder{},
		&models.Content{},
		&models.ContentFile{},
		&models.ContentAccount{},
		&models.GeneratedContent{},
		&models.GeneratedContentFile{},
		// Payment system models
		&models.Subscription{},
		// Task queue system models
		&models.RepurposerTask{},
		&models.AccountGenerationLock{},
		&models.AccountGenerationStatus{},
		// Sync system models
		&models.AccountSyncStatus{},
	)

	return err
}
