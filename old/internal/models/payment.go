package models

import (
	"time"

	"gorm.io/gorm"
)

// PaymentProvider represents the payment service provider
// Currently only Stripe is supported (which handles both fiat and crypto payments)
type PaymentProvider string

const (
	PaymentProviderStripe PaymentProvider = "stripe"
	// All payments go through Stripe, which supports both fiat (card) and crypto
)

// PaymentMethodType represents the payment method used
type PaymentMethodType string

const (
	PaymentMethodCard    PaymentMethodType = "card"
	PaymentMethodCrypto  PaymentMethodType = "crypto"
	PaymentMethodUnknown PaymentMethodType = "unknown"
	// Stripe supports various methods
)

// PaymentStatus represents the current status of a payment
type PaymentStatus string

const (
	PaymentStatusPending    PaymentStatus = "pending"
	PaymentStatusProcessing PaymentStatus = "processing"
	PaymentStatusCompleted  PaymentStatus = "completed"
	PaymentStatusFailed     PaymentStatus = "failed"
	PaymentStatusRefunded   PaymentStatus = "refunded"
	PaymentStatusCanceled   PaymentStatus = "canceled"
)

// PaymentType represents the type/purpose of the payment
type PaymentType string

const (
	PaymentTypeMarketplaceOrder PaymentType = "marketplace_order"
	PaymentTypeSubscription     PaymentType = "subscription"
	// Future payment types can be added here
)

// Payment represents a universal payment record
// All payments are processed through Stripe, which supports both fiat (card) and crypto payments
type Payment struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Core payment information
	Provider    PaymentProvider `json:"provider" gorm:"not null;index;default:'stripe'"` // Always 'stripe'
	Status      PaymentStatus   `json:"status" gorm:"not null;index"`
	PaymentType PaymentType     `json:"payment_type" gorm:"not null;index"`
	Amount      int64           `json:"amount" gorm:"not null"` // Amount in smallest currency unit (cents)
	Currency    string          `json:"currency" gorm:"not null;default:'usd'"`

	// User/Entity relationships
	UserID            uint   `json:"user_id" gorm:"index;not null"`
	RelatedEntityType string `json:"related_entity_type"` // e.g., "marketplace_order", "subscription"
	RelatedEntityID   uint   `json:"related_entity_id" gorm:"index"`

	// Stripe fields (all payments use Stripe)
	StripeCheckoutSessionID string `json:"stripe_checkout_session_id" gorm:"index"`
	StripePaymentIntentID   string `json:"stripe_payment_intent_id" gorm:"index"`
	StripeChargeID          string `json:"stripe_charge_id" gorm:"index"`
	StripeInvoiceID         string `json:"stripe_invoice_id" gorm:"index"`
	StripeCustomerID        string `json:"stripe_customer_id" gorm:"index"`

	// Generic payment URL (for checkout pages)
	PaymentURL string `json:"payment_url"`

	// Payment method information (card, crypto, etc. - all processed via Stripe)
	PaymentMethodType  PaymentMethodType `json:"payment_method_type"`  // card, crypto, etc.
	PaymentMethodLast4 string            `json:"payment_method_last4"` // Last 4 digits of card, or crypto address suffix

	// Transaction details
	ProcessedAt  *time.Time `json:"processed_at"`
	CompletedAt  *time.Time `json:"completed_at"`
	RefundedAt   *time.Time `json:"refunded_at"`
	RefundAmount int64      `json:"refund_amount" gorm:"default:0"`

	// Error handling
	FailureReason string `json:"failure_reason"`
	FailureCode   string `json:"failure_code"`

	// Additional metadata (stored as JSON for flexibility)
	Metadata string `json:"metadata" gorm:"type:jsonb;default:'{}'"`

	// Notes for admin/internal use
	InternalNotes string `json:"internal_notes" gorm:"type:text"`

	// Foreign key relationship
	User *User `json:"-" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
