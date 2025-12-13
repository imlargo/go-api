package models

import (
	"time"
)

const (
	// EmailBatchingThreshold is the duration after which pending chat messages are sent as a digest email
	EmailBatchingThreshold = 12 * time.Hour
)

type ChatConversation struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	BuyerID         uint `json:"buyer_id" gorm:"index;not null;uniqueIndex:idx_buyer_service"`
	SellerID        uint `json:"seller_id" gorm:"index;not null"`
	SellerProfileID uint `json:"seller_profile_id" gorm:"index;not null"`

	ServiceID uint `json:"service_id" gorm:"index;default:null;uniqueIndex:idx_buyer_service"`
	OrderID   uint `json:"order_id" gorm:"index;default:null"`

	BuyerUnreadCount  int  `json:"buyer_unread_count" gorm:"default:0"`
	SellerUnreadCount int  `json:"seller_unread_count" gorm:"default:0"`
	LastMessageID     uint `json:"last_message_id" gorm:"default:null"`

	// Email notification tracking - when we last checked for unread messages to send digest emails
	BuyerLastEmailCheckedAt  *time.Time `json:"buyer_last_email_checked_at" gorm:"default:null"`
	SellerLastEmailCheckedAt *time.Time `json:"seller_last_email_checked_at" gorm:"default:null"`

	Buyer         *User              `json:"_buyer" gorm:"foreignKey:BuyerID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Seller        *User              `json:"_seller" gorm:"foreignKey:SellerID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	SellerProfile *MarketplaceSeller `json:"seller_profile" gorm:"foreignKey:SellerProfileID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	Service     *MarketplaceService `json:"service,omitempty" gorm:"foreignKey:ServiceID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Order       *MarketplaceOrder   `json:"order,omitempty" gorm:"foreignKey:OrderID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	LastMessage *ChatMessage        `json:"last_message,omitempty" gorm:"foreignKey:LastMessageID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
}

type ChatMessage struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	ConversationID uint   `json:"conversation_id" gorm:"index;not null"`
	SenderID       uint   `json:"sender_id" gorm:"default:null"`
	Content        string `json:"content"`
	IsAutomated    bool   `json:"is_automated" gorm:"default:false"`

	Conversation *ChatConversation `json:"-" gorm:"foreignKey:ConversationID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Sender       *User             `json:"-" gorm:"foreignKey:SenderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
