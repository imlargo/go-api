package models

import (
	"time"

	"github.com/nicolailuther/butter/internal/enums"
	"gorm.io/gorm"
)

type MarketplaceCategory struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Name        string `json:"name" gorm:"not null;unique"`
	Description string `json:"description" gorm:"not null"`
	Icon        string `json:"icon"`
}

type MarketplaceSeller struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	UserID   uint   `json:"user_id" gorm:"default:null"`
	Nickname string `json:"nickname"`
	Bio      string `json:"bio"`

	User *User `json:"-"`
}

type MarketplaceServiceResult struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	FileID    uint `json:"file_id"`
	ServiceID uint `json:"service_id" gorm:"index;default:null"`

	File    *File               `json:"file"`
	Service *MarketplaceService `json:"-"`
}

type MarketplaceService struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Title              string  `json:"title"`
	Description        string  `json:"description"`
	TermsIn            string  `json:"terms_in"`
	Expectations       string  `json:"expectations"`
	StepsToStart       string  `json:"steps_to_start"`
	Disclaimer         string  `json:"disclaimer"`
	AvailableSpots     int     `json:"available_spots"`
	PlatformCommission float64 `json:"platform_commission" gorm:"default:0"`

	CategoryID uint `json:"marketplace_category_id" gorm:"index"`
	SellerID   uint `json:"seller_id" gorm:"index"`
	UserID     uint `json:"user_id" gorm:"index"`
	ImageID    uint `json:"file_id" gorm:"default:null"`

	Image *File `json:"image"`

	Category *MarketplaceCategory `json:"-" gorm:"foreignKey:CategoryID"`
	User     *User                `json:"-" gorm:"foreignKey:UserID"`

	Seller *MarketplaceSeller `json:"_seller" gorm:"foreignKey:SellerID"`
}

type MarketplaceServicePackage struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Name              string  `json:"name"`
	Description       string  `json:"description"`
	Price             float64 `json:"price"`
	DurationDays      int     `json:"duration_days" gorm:"default:0"`
	IncludedRevisions int     `json:"included_revisions" gorm:"default:2"`

	ServiceID uint `json:"service_id" gorm:"index"`

	Service *MarketplaceService `json:"-"`
}

type MarketplaceOrder struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	StartDate          time.Time                    `json:"start_date"`
	DueDate            time.Time                    `json:"due_date"`
	CompletedAt        time.Time                    `json:"completed_at" gorm:"default:null"`
	AutoCompletionDate time.Time                    `json:"auto_completion_date" gorm:"default:null"`
	Status             enums.MarketplaceOrderStatus `json:"status"`
	RequiredInfo       string                       `json:"required_info"`
	TotalRevisions     int                          `json:"total_revisions" gorm:"default:0"`
	RevisionsUsed      int                          `json:"revisions_used" gorm:"default:0"`
	TotalDeliveries    int                          `json:"total_deliveries" gorm:"default:0"`
	DaysExtended       int                          `json:"days_extended" gorm:"default:0"`
	ServiceID          uint                         `json:"service_id" gorm:"index"`
	CategoryID         uint                         `json:"category_id" gorm:"index"`
	BuyerID            uint                         `json:"buyer_id" gorm:"index"`
	ClientID           uint                         `json:"client_id" gorm:"index"`
	ServicePackageID   uint                         `json:"service_package_id"`
	PaymentID          uint                         `json:"payment_id" gorm:"index;default:null"`
	ConversationID     uint                         `json:"conversation_id" gorm:"default:null"`

	Payment *Payment `json:"payment" gorm:"foreignKey:PaymentID"`

	Category       *MarketplaceCategory       `json:"_category" gorm:"foreignKey:CategoryID"`
	Service        *MarketplaceService        `json:"_service" gorm:"foreignKey:ServiceID"`
	Buyer          *User                      `json:"_buyer" gorm:"foreignKey:BuyerID"`
	Client         *Client                    `json:"_client" gorm:"foreignKey:ClientID"`
	ServicePackage *MarketplaceServicePackage `json:"_package" gorm:"foreignKey:ServicePackageID"`

	Conversation *ChatConversation `json:"-" gorm:"foreignKey:ConversationID"`
}

type MarketplaceDeliverable struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Description  string                  `json:"description" gorm:"type:text;not null"`
	DeliveryNote string                  `json:"delivery_note" gorm:"type:text"`
	Status       enums.DeliverableStatus `json:"status" gorm:"default:'pending'"`
	ReviewedAt   time.Time               `json:"reviewed_at" gorm:"default:null"`
	ReviewNotes  string                  `json:"review_notes" gorm:"type:text"`

	OrderID  uint `json:"order_id" gorm:"index;not null"`
	SellerID uint `json:"seller_id" gorm:"index;not null"`

	Order  *MarketplaceOrder `json:"_order" gorm:"foreignKey:OrderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Seller *User             `json:"_seller" gorm:"foreignKey:SellerID"`
}

type MarketplaceRevisionRequest struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	RevisionNumber int                         `json:"revision_number" gorm:"not null"`
	Reason         string                      `json:"reason" gorm:"type:text;not null"`
	SellerResponse string                      `json:"seller_response" gorm:"type:text"`
	Status         enums.RevisionRequestStatus `json:"status" gorm:"default:'pending'"`
	RespondedAt    time.Time                   `json:"responded_at" gorm:"default:null"`

	OrderID       uint `json:"order_id" gorm:"index;not null"`
	DeliverableID uint `json:"deliverable_id" gorm:"index;not null"`

	Order       *MarketplaceOrder       `json:"_order" gorm:"foreignKey:OrderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Deliverable *MarketplaceDeliverable `json:"_deliverable" gorm:"foreignKey:DeliverableID"`
}

type MarketplaceDispute struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	ReasonCategory  enums.DisputeReasonCategory `json:"reason_category" gorm:"not null"`
	Description     string                      `json:"description" gorm:"type:text;not null"`
	Status          enums.DisputeStatus         `json:"status" gorm:"default:'open'"`
	Resolution      enums.DisputeResolution     `json:"resolution"`
	ResolutionNotes string                      `json:"resolution_notes" gorm:"type:text"`
	RefundAmount    float64                     `json:"refund_amount" gorm:"default:0"`
	ResolvedAt      time.Time                   `json:"resolved_at" gorm:"default:null"`

	OrderID    uint `json:"order_id" gorm:"index;not null;unique"`
	OpenedBy   uint `json:"opened_by" gorm:"index;not null"`
	ResolvedBy uint `json:"resolved_by" gorm:"index;default:null"`

	Order    *MarketplaceOrder `json:"_order" gorm:"foreignKey:OrderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Opener   *User             `json:"_opener" gorm:"foreignKey:OpenedBy"`
	Resolver *User             `json:"_resolver" gorm:"foreignKey:ResolvedBy"`
}

type MarketplaceOrderTimeline struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	EventType   enums.OrderTimelineEventType `json:"event_type" gorm:"not null"`
	Description string                       `json:"description" gorm:"type:text;not null"`
	Metadata    string                       `json:"metadata" gorm:"type:jsonb"`

	OrderID uint `json:"order_id" gorm:"index;not null"`
	ActorID uint `json:"actor_id" gorm:"index:default:null"`

	Order *MarketplaceOrder `json:"_order" gorm:"foreignKey:OrderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Actor *User             `json:"_actor" gorm:"foreignKey:ActorID"`
}
