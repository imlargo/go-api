package jobs

import (
	"fmt"
	"time"

	"github.com/nicolailuther/butter/internal/models"
	"github.com/nicolailuther/butter/internal/store"
	"github.com/nicolailuther/butter/pkg/email"
	"github.com/nicolailuther/butter/pkg/email/templates"
	"go.uber.org/zap"
)

type SendMarketplaceMessageDigestTask struct {
	store       *store.Store
	emailClient email.EmailClient
	logger      *zap.SugaredLogger
}

func NewSendMarketplaceMessageDigestTask(
	store *store.Store,
	emailClient email.EmailClient,
	logger *zap.SugaredLogger,
) Job {
	return &SendMarketplaceMessageDigestTask{
		store:       store,
		emailClient: emailClient,
		logger:      logger,
	}
}

func (t *SendMarketplaceMessageDigestTask) Execute() error {
	t.logger.Info("Starting marketplace message digest email task")

	// Find conversations with unread messages
	conversations, err := t.store.ChatConversations.GetConversationsWithUnreadMessages()
	if err != nil {
		return fmt.Errorf("failed to get conversations with unread messages: %w", err)
	}

	now := time.Now()

	// Group conversations by user (buyers and sellers separately)
	buyerConversations := make(map[uint][]*models.ChatConversation)
	sellerConversations := make(map[uint][]*models.ChatConversation)

	for _, conversation := range conversations {
		// Group by buyer if they have unread messages and should receive email
		if conversation.BuyerUnreadCount > 0 && t.shouldSendDigest(conversation.BuyerLastEmailCheckedAt, now) {
			buyerConversations[conversation.BuyerID] = append(buyerConversations[conversation.BuyerID], conversation)
		}

		// Group by seller if they have unread messages and should receive email
		if conversation.SellerUnreadCount > 0 && t.shouldSendDigest(conversation.SellerLastEmailCheckedAt, now) {
			sellerConversations[conversation.SellerID] = append(sellerConversations[conversation.SellerID], conversation)
		}
	}

	emailsSent := 0

	// Send consolidated emails to buyers
	for buyerID, convs := range buyerConversations {
		if err := t.sendConsolidatedUnreadMessagesEmail(buyerID, convs, true); err != nil {
			t.logger.Errorw("Failed to send consolidated buyer unread messages email",
				"error", err,
				"buyerID", buyerID,
				"conversationCount", len(convs),
			)
		} else {
			emailsSent++
			// Update last checked timestamp for all conversations
			updateTime := time.Now()
			for _, conv := range convs {
				if err := t.store.ChatConversations.UpdateBuyerLastEmailCheckedAt(conv.ID, &updateTime); err != nil {
					t.logger.Errorw("Failed to update buyer last email checked timestamp",
						"error", err,
						"conversationID", conv.ID,
					)
				}
			}
		}
	}

	// Send consolidated emails to sellers
	for sellerID, convs := range sellerConversations {
		if err := t.sendConsolidatedUnreadMessagesEmail(sellerID, convs, false); err != nil {
			t.logger.Errorw("Failed to send consolidated seller unread messages email",
				"error", err,
				"sellerID", sellerID,
				"conversationCount", len(convs),
			)
		} else {
			emailsSent++
			// Update last checked timestamp for all conversations
			updateTime := time.Now()
			for _, conv := range convs {
				if err := t.store.ChatConversations.UpdateSellerLastEmailCheckedAt(conv.ID, &updateTime); err != nil {
					t.logger.Errorw("Failed to update seller last email checked timestamp",
						"error", err,
						"conversationID", conv.ID,
					)
				}
			}
		}
	}

	t.logger.Infow("Marketplace message digest email task completed",
		"emailsSent", emailsSent,
	)

	return nil
}

func (t *SendMarketplaceMessageDigestTask) GetName() TaskLabel {
	return TaskSendMarketplaceMessageDigest
}

// shouldSendDigest determines if a digest should be sent based on last checked time
func (t *SendMarketplaceMessageDigestTask) shouldSendDigest(lastCheckedAt *time.Time, now time.Time) bool {
	if lastCheckedAt == nil {
		// Never checked before, send email
		return true
	}

	timeSinceLastCheck := now.Sub(*lastCheckedAt)
	return timeSinceLastCheck >= models.EmailBatchingThreshold
}

// sendUnreadMessagesEmail sends an email notification about unread messages
func (t *SendMarketplaceMessageDigestTask) sendUnreadMessagesEmail(conversation *models.ChatConversation, receiverID uint, isForBuyer bool, unreadCount int) error {
	// Get receiver user information using preloaded data if available
	var receiver *models.User
	if isForBuyer {
		receiver = conversation.Buyer
	} else {
		receiver = conversation.Seller
	}
	if receiver == nil {
		var err error
		receiver, err = t.store.Users.GetByID(receiverID)
		if err != nil {
			return fmt.Errorf("failed to get receiver user: %w", err)
		}
	}

	// Get sender information - in marketplace chat, sender is the other party
	var senderID uint
	var senderName string
	if isForBuyer {
		senderID = conversation.SellerID
		if conversation.Seller != nil {
			senderName = conversation.Seller.Name
		}
	} else {
		senderID = conversation.BuyerID
		if conversation.Buyer != nil {
			senderName = conversation.Buyer.Name
		}
	}

	// Fallback to fetching sender if not preloaded
	if senderName == "" {
		sender, err := t.store.Users.GetByID(senderID)
		if err != nil {
			return fmt.Errorf("failed to get sender user: %w", err)
		}
		senderName = sender.Name
	}

	// Get service information if available
	serviceTitle := "Marketplace Chat"
	if conversation.Service != nil {
		serviceTitle = conversation.Service.Title
	} else if conversation.ServiceID != 0 {
		t.logger.Warnw("Conversation has non-zero ServiceID but Service is nil. Possible data integrity issue.",
			"conversationID", conversation.ID,
			"serviceID", conversation.ServiceID,
		)
	}

	// Prepare email data
	emailData := templates.UnreadMessagesEmailData{
		SenderName:     senderName,
		ReceiverName:   receiver.Name,
		ConversationID: conversation.ID,
		UnreadCount:    unreadCount,
		ServiceTitle:   serviceTitle,
		IsForBuyer:     isForBuyer,
	}

	// Generate email content
	subject, htmlBody, textBody := templates.NewUnreadMessagesNotification(emailData)

	// Send email
	_, err := t.emailClient.SendEmail(&email.SendEmailParams{
		From:    "noreply@notifications.hellobutter.io",
		To:      []string{receiver.Email},
		Subject: subject,
		Html:    htmlBody,
		Text:    textBody,
	})
	if err != nil {
		return fmt.Errorf("failed to send unread messages email: %w", err)
	}

	t.logger.Infow("Sent unread messages email",
		"conversationID", conversation.ID,
		"receiverID", receiverID,
		"unreadCount", unreadCount,
	)

	return nil
}

// sendConsolidatedUnreadMessagesEmail sends a consolidated email notification about unread messages across multiple conversations
func (t *SendMarketplaceMessageDigestTask) sendConsolidatedUnreadMessagesEmail(receiverID uint, conversations []*models.ChatConversation, isForBuyer bool) error {
	// Get receiver user information
	receiver, err := t.store.Users.GetByID(receiverID)
	if err != nil {
		return fmt.Errorf("failed to get receiver user: %w", err)
	}

	// Build list of conversation digest items
	var conversationItems []templates.ConversationDigestItem
	totalUnread := 0

	for _, conversation := range conversations {
		var senderID uint
		var senderName string
		var unreadCount int

		if isForBuyer {
			senderID = conversation.SellerID
			if conversation.Seller != nil {
				senderName = conversation.Seller.Name
			}
			unreadCount = conversation.BuyerUnreadCount
		} else {
			senderID = conversation.BuyerID
			if conversation.Buyer != nil {
				senderName = conversation.Buyer.Name
			}
			unreadCount = conversation.SellerUnreadCount
		}

		// Fallback to fetching sender if not preloaded
		if senderName == "" {
			sender, err := t.store.Users.GetByID(senderID)
			if err != nil {
				t.logger.Errorw("Failed to get sender user",
					"error", err,
					"senderID", senderID,
					"conversationID", conversation.ID,
				)
				senderName = "Unknown User"
			} else {
				senderName = sender.Name
			}
		}

		// Get service information if available
		serviceTitle := "Marketplace Chat"
		if conversation.Service != nil {
			serviceTitle = conversation.Service.Title
		} else if conversation.ServiceID != 0 {
			t.logger.Warnw("Conversation has non-zero ServiceID but Service is nil. Possible data integrity issue.",
				"conversationID", conversation.ID,
				"serviceID", conversation.ServiceID,
			)
		}

		conversationItems = append(conversationItems, templates.ConversationDigestItem{
			ConversationID: conversation.ID,
			SenderName:     senderName,
			ServiceTitle:   serviceTitle,
			UnreadCount:    unreadCount,
		})

		totalUnread += unreadCount
	}

	// Prepare email data
	emailData := templates.UnreadMessagesDigestData{
		ReceiverName:  receiver.Name,
		Conversations: conversationItems,
		TotalUnread:   totalUnread,
	}

	// Generate email content
	subject, htmlBody, textBody := templates.NewUnreadMessagesDigest(emailData)

	// Send email
	_, err = t.emailClient.SendEmail(&email.SendEmailParams{
		From:    "noreply@notifications.hellobutter.io",
		To:      []string{receiver.Email},
		Subject: subject,
		Html:    htmlBody,
		Text:    textBody,
	})
	if err != nil {
		return fmt.Errorf("failed to send consolidated unread messages email: %w", err)
	}

	t.logger.Infow("Sent consolidated unread messages email",
		"receiverID", receiverID,
		"conversationCount", len(conversations),
		"totalUnread", totalUnread,
	)

	return nil
}
