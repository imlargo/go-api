package repositories

import (
	"time"

	"github.com/nicolailuther/butter/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ChatConversationRepository interface {
	GetAll() ([]*models.ChatConversation, error)
	GetConversationsWithUnreadMessages() ([]*models.ChatConversation, error)
	GetByID(id uint) (*models.ChatConversation, error)
	Create(conversation *models.ChatConversation) error
	Update(conversation *models.ChatConversation) error
	Delete(id uint) error
	GetByBuyerAndService(buyerID, serviceID uint) (*models.ChatConversation, error)
	GetAllWithFilters(filters *models.ChatConversation) ([]*models.ChatConversation, error)

	IncrementBuyerUnreadCount(conversationID uint) error
	IncrementSellerUnreadCount(conversationID uint) error

	MarkBuyerConversationAsRead(conversationID uint) error
	MarkSellerConversationAsRead(conversationID uint) error

	UpdateLastMessage(conversationID uint, messageID uint) error
	UpdateBuyerLastEmailCheckedAt(conversationID uint, timestamp *time.Time) error
	UpdateSellerLastEmailCheckedAt(conversationID uint, timestamp *time.Time) error
}

type chatConversationRepository struct {
	*Repository
}

func NewChatConversationRepository(
	r *Repository,
) ChatConversationRepository {
	return &chatConversationRepository{
		Repository: r,
	}
}

func (r *chatConversationRepository) Create(conversation *models.ChatConversation) error {
	return r.db.Create(conversation).Error
}

func (r *chatConversationRepository) GetByID(id uint) (*models.ChatConversation, error) {
	var conversation models.ChatConversation

	if err := r.db.
		Preload("Buyer").Preload("Seller").Preload("SellerProfile").Preload("Service").Preload("LastMessage").
		Preload("Order").Preload("Order.Client").Preload("Order.ServicePackage").Preload("Order.Payment").
		First(&conversation, id).Error; err != nil {
		return nil, err
	}
	return &conversation, nil
}

func (r *chatConversationRepository) Update(conversation *models.ChatConversation) error {
	return r.db.Model(conversation).Clauses(clause.Returning{}).Updates(conversation).Error
}

func (r *chatConversationRepository) Delete(id uint) error {
	var conversation models.ChatConversation
	conversation.ID = id
	return r.db.Delete(&conversation).Error
}

func (r *chatConversationRepository) GetAll() ([]*models.ChatConversation, error) {
	var conversations []*models.ChatConversation
	if err := r.db.Preload("Service").Find(&conversations).Error; err != nil {
		return nil, err
	}
	return conversations, nil
}

func (r *chatConversationRepository) GetByBuyerAndService(buyerID, serviceID uint) (*models.ChatConversation, error) {
	var conversation models.ChatConversation
	if err := r.db.Where(&models.ChatConversation{
		BuyerID:   buyerID,
		ServiceID: serviceID,
	}).First(&conversation).Error; err != nil {
		return nil, err
	}
	return &conversation, nil
}

func (r *chatConversationRepository) IncrementBuyerUnreadCount(conversationID uint) error {
	return r.db.Model(&models.ChatConversation{}).
		Where(&models.ChatConversation{ID: conversationID}).
		UpdateColumn("buyer_unread_count", gorm.Expr("buyer_unread_count + 1")).Error
}

func (r *chatConversationRepository) IncrementSellerUnreadCount(conversationID uint) error {
	return r.db.Model(&models.ChatConversation{}).
		Where(&models.ChatConversation{ID: conversationID}).
		UpdateColumn("seller_unread_count", gorm.Expr("seller_unread_count + 1")).Error
}

func (r *chatConversationRepository) GetAllWithFilters(filters *models.ChatConversation) ([]*models.ChatConversation, error) {
	var conversations []*models.ChatConversation
	if err := r.db.
		Preload("Buyer").Preload("Seller").Preload("SellerProfile").Preload("Service").Preload("LastMessage").
		Preload("Order").Preload("Order.Client").Preload("Order.ServicePackage").Preload("Order.Payment").
		Model(&models.ChatConversation{}).Where(filters).Order("updated_at DESC").Find(&conversations).Error; err != nil {
		return nil, err
	}

	return conversations, nil
}

func (r *chatConversationRepository) MarkBuyerConversationAsRead(conversationID uint) error {
	return r.db.Model(&models.ChatConversation{}).
		Where(&models.ChatConversation{ID: conversationID}).
		UpdateColumn("buyer_unread_count", 0).Error
}

func (r *chatConversationRepository) MarkSellerConversationAsRead(conversationID uint) error {
	return r.db.Model(&models.ChatConversation{}).
		Where(&models.ChatConversation{ID: conversationID}).
		UpdateColumn("seller_unread_count", 0).Error
}

func (r *chatConversationRepository) UpdateLastMessage(conversationID uint, messageID uint) error {
	return r.db.Model(&models.ChatConversation{}).
		Where(&models.ChatConversation{ID: conversationID}).
		Update("last_message_id", messageID).Error
}

// GetConversationsWithUnreadMessages retrieves conversations that have unread messages
func (r *chatConversationRepository) GetConversationsWithUnreadMessages() ([]*models.ChatConversation, error) {
	var conversations []*models.ChatConversation
	if err := r.db.
		Preload("Service").
		Preload("Buyer").
		Preload("Seller").
		Where("buyer_unread_count > 0 OR seller_unread_count > 0").
		Find(&conversations).Error; err != nil {
		return nil, err
	}
	return conversations, nil
}

// UpdateBuyerLastEmailCheckedAt updates the buyer's last email checked timestamp
func (r *chatConversationRepository) UpdateBuyerLastEmailCheckedAt(conversationID uint, timestamp *time.Time) error {
	return r.db.Model(&models.ChatConversation{}).
		Where("id = ?", conversationID).
		Update("buyer_last_email_checked_at", timestamp).Error
}

// UpdateSellerLastEmailCheckedAt updates the seller's last email checked timestamp
func (r *chatConversationRepository) UpdateSellerLastEmailCheckedAt(conversationID uint, timestamp *time.Time) error {
	return r.db.Model(&models.ChatConversation{}).
		Where("id = ?", conversationID).
		Update("seller_last_email_checked_at", timestamp).Error
}
