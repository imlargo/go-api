package models

import (
	"time"

	"github.com/nicolailuther/butter/internal/enums"
	"gorm.io/gorm"
)

type SubscriptionType string

const (
	SubscriptionTypeTier  SubscriptionType = "tier"
	SubscriptionTypeAddon SubscriptionType = "addon"
)

type SubscriptionStatus string

const (
	SubscriptionStatusActive            SubscriptionStatus = "active"
	SubscriptionStatusPastDue           SubscriptionStatus = "past_due"
	SubscriptionStatusCanceled          SubscriptionStatus = "canceled"
	SubscriptionStatusTrialing          SubscriptionStatus = "trialing"
	SubscriptionStatusUnpaid            SubscriptionStatus = "unpaid"
	SubscriptionStatusIncomplete        SubscriptionStatus = "incomplete"
	SubscriptionStatusIncompleteExpired SubscriptionStatus = "incomplete_expired"
)

// Subscription represents a user's subscription (tier or add-on)
type Subscription struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	UserID                 uint               `json:"user_id" gorm:"index;not null"`
	StripeSubscriptionID   string             `json:"stripe_subscription_id" gorm:"unique;not null"`
	StripePriceID          string             `json:"stripe_price_id" gorm:"not null"`
	StripeProductID        string             `json:"stripe_product_id"`
	SubscriptionType       SubscriptionType   `json:"subscription_type" gorm:"not null"`
	TierLevel              enums.TierLevel    `json:"tier_level" gorm:"not null;default:0"` // Subscription tier level
	Status                 SubscriptionStatus `json:"status" gorm:"not null"`
	CurrentPeriodStart     time.Time          `json:"current_period_start"`
	CurrentPeriodEnd       time.Time          `json:"current_period_end"`
	CancelAtPeriodEnd      bool               `json:"cancel_at_period_end" gorm:"default:false"`
	CanceledAt             *time.Time         `json:"canceled_at"`
	TrialStart             *time.Time         `json:"trial_start"`
	TrialEnd               *time.Time         `json:"trial_end"`
	LatestInvoiceID        string             `json:"latest_invoice_id"`
	DefaultPaymentMethodID string             `json:"default_payment_method_id"`

	User *User `json:"-" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
