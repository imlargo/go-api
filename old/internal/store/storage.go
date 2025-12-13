package store

import (
	"github.com/nicolailuther/butter/internal/repositories"
	"gorm.io/gorm"
)

type Store struct {
	DB                          *gorm.DB // Expose DB for transactions. Use sparingly and prefer repository methods when possible.
	AccountAnalytics            repositories.AccountAnalyticRepository
	Clients                     repositories.ClientRepository
	MarketplaceCategories       repositories.MarketplaceCategoryRepository
	MarketplaceSellers          repositories.MarketplaceSellerRepository
	OnlyfansAccounts            repositories.OnlyfansAccountRepository
	Posts                       repositories.PostRepository
	TextOverlays                repositories.TextOverlayRepository
	Accounts                    repositories.AccountRepository
	Files                       repositories.FileRepository
	MarketplaceOrders           repositories.MarketplaceOrderRepository
	MarketplaceServices         repositories.MarketplaceServiceRepository
	OnlyfansLinks               repositories.OnlyfansTrackingLinkRepository
	PostingGoals                repositories.PostingGoalRepository
	ChatConversations           repositories.ChatConversationRepository
	MarketplaceServicePackages  repositories.MarketplaceServicePackageRepository
	OnlyfansTransactions        repositories.OnlyfansTransactionRepository
	PushSubscriptions           repositories.PushNotificationSubscriptionRepository
	ChatMessages                repositories.ChatMessageRepository
	MarketplaceResults          repositories.MarketplaceServiceResultRepository
	Notifications               repositories.NotificationRepository
	PostAnalytics               repositories.PostAnalyticRepository
	ReferralCode                repositories.ReferralCodeRepository
	ReferralDiscounts           repositories.ReferralDiscountRepository
	Users                       repositories.UserRepository
	ContentFolders              repositories.ContentFolderRepository
	Contents                    repositories.ContentRepository
	ContentFiles                repositories.ContentFileRepository
	GeneratedContents           repositories.GeneratedContentRepository
	GeneratedContentFiles       repositories.GeneratedContentFileRepository
	ContentAccounts             repositories.ContentAccountRepository
	MarketplaceDeliverables     repositories.MarketplaceDeliverableRepository
	MarketplaceRevisionRequests repositories.MarketplaceRevisionRequestRepository
	MarketplaceDisputes         repositories.MarketplaceDisputeRepository
	MarketplaceOrderTimelines   repositories.MarketplaceOrderTimelineRepository
	Subscriptions               repositories.SubscriptionRepository
	Payments                    repositories.PaymentRepository
	RepurposerTasks             repositories.RepurposerTaskRepository
	GenerationLocks             repositories.GenerationLockRepository
	GenerationStatus            repositories.GenerationStatusRepository
	SyncStatus                  repositories.SyncStatusRepository
}

func NewStorage(container *repositories.Repository, db *gorm.DB) *Store {
	store := &Store{
		DB:                          db,
		AccountAnalytics:            repositories.NewAccountAnalyticRepository(container),
		Clients:                     repositories.NewClientRepository(container),
		MarketplaceCategories:       repositories.NewMarketplaceCategoryRepository(container),
		MarketplaceSellers:          repositories.NewMarketplaceSellerRepository(container),
		OnlyfansAccounts:            repositories.NewOnlyfansAccountRepository(container),
		Posts:                       repositories.NewPostRepository(container),
		TextOverlays:                repositories.NewTextOverlayRepository(container),
		Accounts:                    repositories.NewAccountRepository(container),
		Files:                       repositories.NewFileRepository(container),
		MarketplaceOrders:           repositories.NewMarketplaceOrderRepository(container),
		MarketplaceServices:         repositories.NewMarketplaceServiceRepository(container),
		OnlyfansLinks:               repositories.NewOnlyfansTrackingLinkRepository(container),
		PostingGoals:                repositories.NewPostingGoalRepository(container),
		ChatConversations:           repositories.NewChatConversationRepository(container),
		MarketplaceServicePackages:  repositories.NewMarketplaceServicePackageRepository(container),
		OnlyfansTransactions:        repositories.NewOnlyfansTransactionRepository(container),
		PushSubscriptions:           repositories.NewPushSubscriptionRepository(container),
		ChatMessages:                repositories.NewChatMessageRepository(container),
		MarketplaceResults:          repositories.NewMarketplaceServiceResultRepository(container),
		Notifications:               repositories.NewNotificationRepository(container),
		PostAnalytics:               repositories.NewPostAnalyticRepository(container),
		ReferralCode:                repositories.NewReferralCodeRepository(container),
		ReferralDiscounts:           repositories.NewReferralDiscountRepository(container),
		Users:                       repositories.NewUserRepository(container),
		ContentFolders:              repositories.NewContentFolderRepository(container),
		Contents:                    repositories.NewContentRepository(container),
		ContentFiles:                repositories.NewContentFileRepository(container),
		GeneratedContents:           repositories.NewGeneratedContentRepository(container),
		GeneratedContentFiles:       repositories.NewGeneratedContentFileRepository(container),
		ContentAccounts:             repositories.NewContentAccountRepository(container),
		MarketplaceDeliverables:     repositories.NewMarketplaceDeliverableRepository(container),
		MarketplaceRevisionRequests: repositories.NewMarketplaceRevisionRequestRepository(container),
		MarketplaceDisputes:         repositories.NewMarketplaceDisputeRepository(container),
		MarketplaceOrderTimelines:   repositories.NewMarketplaceOrderTimelineRepository(container),
		Subscriptions:               repositories.NewSubscriptionRepository(container),
		Payments:                    repositories.NewPaymentRepository(container),
		RepurposerTasks:             repositories.NewRepurposerTaskRepository(container),
		GenerationLocks:             repositories.NewGenerationLockRepository(container),
		GenerationStatus:            repositories.NewGenerationStatusRepository(container),
		SyncStatus:                  repositories.NewSyncStatusRepository(container),
	}

	return store
}
