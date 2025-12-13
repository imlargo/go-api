package models

import (
	"time"

	"gorm.io/gorm"
)

// ReferralDiscount tracks which users have used referral discounts
// This ensures each customer can only use a discount once
type ReferralDiscount struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	UserID             uint      `json:"user_id" gorm:"not null;uniqueIndex:idx_user_discount"`
	ReferralCodeID     uint      `json:"referral_code_id" gorm:"not null;index"`
	DiscountPercentage float64   `json:"discount_percentage" gorm:"not null"`
	StripeCouponID     string    `json:"stripe_coupon_id" gorm:"not null"`
	SubscriptionID     uint      `json:"subscription_id" gorm:"index"`
	AppliedAt          time.Time `json:"applied_at"`

	User         *User         `json:"-" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ReferralCode *ReferralCode `json:"-" gorm:"foreignKey:ReferralCodeID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Subscription *Subscription `json:"-" gorm:"foreignKey:SubscriptionID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
}
