package models

import (
	"time"

	"github.com/nicolailuther/butter/internal/enums"
)

type ReferralCode struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Code               string                   `json:"code" gorm:"unique;not null"`
	Description        string                   `json:"description"`
	Status             enums.ReferralCodeStatus `json:"status" gorm:"default:'active'"`
	ExpiresAt          *time.Time               `json:"expires_at"`
	UsageLimit         *int                     `json:"usage_limit"`
	Clicks             int                      `json:"clicks" gorm:"default:0"`
	Registrations      int                      `json:"registrations" gorm:"default:0"`
	DiscountPercentage *float64                 `json:"discount_percentage" gorm:"default:null"`
	StripeCouponID     string                   `json:"stripe_coupon_id" gorm:"default:null"`
	UserID             uint                     `json:"user_id" gorm:"not null;index"`

	User *User `json:"-" gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
