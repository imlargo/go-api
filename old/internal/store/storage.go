package store

import (
	"github.com/nicolailuther/butter/internal/repositories"
	"gorm.io/gorm"
)

type Store struct {
	DB                         *gorm.DB // Expose DB for transactions. Use sparingly and prefer repository methods when possible.
	MarketplaceSellers         repositories.MarketplaceSellerRepository
	OnlyfansAccounts           repositories.OnlyfansAccountRepository
	Posts                      repositories.PostRepository
	TextOverlays               repositories.TextOverlayRepository
	Files                      repositories.FileRepository
	MarketplaceServices        repositories.MarketplaceServiceRepository
	OnlyfansLinks              repositories.OnlyfansTrackingLinkRepository
	PostingGoals               repositories.PostingGoalRepository
	MarketplaceServicePackages repositories.MarketplaceServicePackageRepository
	OnlyfansTransactions       repositories.OnlyfansTransactionRepository
	PushSubscriptions          repositories.PushNotificationSubscriptionRepository
	Notifications              repositories.NotificationRepository
	PostAnalytics              repositories.PostAnalyticRepository
	ReferralCode               repositories.ReferralCodeRepository
	ReferralDiscounts          repositories.ReferralDiscountRepository
	Users                      repositories.UserRepository
	Subscriptions              repositories.SubscriptionRepository
	Payments                   repositories.PaymentRepository
	RepurposerTasks            repositories.RepurposerTaskRepository
	SyncStatus                 repositories.SyncStatusRepository
}

func NewStorage(container *repositories.Repository, db *gorm.DB) *Store {
	store := &Store{
		DB:                         db,
		MarketplaceSellers:         repositories.NewMarketplaceSellerRepository(container),
		OnlyfansAccounts:           repositories.NewOnlyfansAccountRepository(container),
		Posts:                      repositories.NewPostRepository(container),
		TextOverlays:               repositories.NewTextOverlayRepository(container),
		Files:                      repositories.NewFileRepository(container),
		MarketplaceServices:        repositories.NewMarketplaceServiceRepository(container),
		OnlyfansLinks:              repositories.NewOnlyfansTrackingLinkRepository(container),
		PostingGoals:               repositories.NewPostingGoalRepository(container),
		MarketplaceServicePackages: repositories.NewMarketplaceServicePackageRepository(container),
		OnlyfansTransactions:       repositories.NewOnlyfansTransactionRepository(container),
		PushSubscriptions:          repositories.NewPushSubscriptionRepository(container),
		Notifications:              repositories.NewNotificationRepository(container),
		PostAnalytics:              repositories.NewPostAnalyticRepository(container),
		ReferralCode:               repositories.NewReferralCodeRepository(container),
		ReferralDiscounts:          repositories.NewReferralDiscountRepository(container),
		Users:                      repositories.NewUserRepository(container),
		Subscriptions:              repositories.NewSubscriptionRepository(container),
		Payments:                   repositories.NewPaymentRepository(container),
		RepurposerTasks:            repositories.NewRepurposerTaskRepository(container),
		SyncStatus:                 repositories.NewSyncStatusRepository(container),
	}

	return store
}
