package dto

import (
	"time"

	"github.com/nicolailuther/butter/internal/enums"
)

// CreateReferralCodeRequest represents the request to create a referral code
type CreateReferralCodeRequest struct {
	Code               string     `json:"code" binding:"omitempty,max=50"`
	Description        string     `json:"description" binding:"max=500"`
	ExpiresAt          *time.Time `json:"expires_at"`
	UsageLimit         *int       `json:"usage_limit" binding:"omitempty,min=1"`
	DiscountPercentage *float64   `json:"discount_percentage" binding:"omitempty,min=0,max=100"`
}

// UpdateReferralCodeRequest represents the request to update a referral code
type UpdateReferralCodeRequest struct {
	Description        string                   `json:"description" binding:"max=500"`
	Status             enums.ReferralCodeStatus `json:"status" binding:"omitempty,oneof=active inactive"`
	ExpiresAt          *time.Time               `json:"expires_at"`
	UsageLimit         *int                     `json:"usage_limit" binding:"omitempty,min=1"`
	DiscountPercentage *float64                 `json:"discount_percentage" binding:"omitempty,min=0,max=100"`
}

// ReferralCodeResponse represents a referral code response
type ReferralCodeResponse struct {
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
	DiscountPercentage *float64                 `json:"discount_percentage"`
	UserID             uint                     `json:"user_id" gorm:"not null;index"`

	ReferralLink   string  `json:"referral_link"`
	ConversionRate float64 `json:"conversion_rate"`
	RemainingUses  *int    `json:"remaining_uses"`

	// User information (only populated for admins)
	UserName  string `json:"user_name,omitempty"`
	UserEmail string `json:"user_email,omitempty"`
}

// ReferralMetricsResponse represents metrics for the referral dashboard
type ReferralMetricsResponse struct {
	TotalCodes         int     `json:"total_codes"`
	ActiveCodes        int     `json:"active_codes"`
	TotalClicks        int     `json:"total_clicks"`
	TotalRegistrations int     `json:"total_registrations"`
	ConversionRate     float64 `json:"conversion_rate"`
}

// TrackClickRequest represents a click tracking request
type TrackClickRequest struct {
	Code string `json:"code" binding:"required"`
}

// CheckCodeAvailabilityResponse represents availability check response
type CheckCodeAvailabilityResponse struct {
	Available bool   `json:"available"`
	Message   string `json:"message"`
}

// ReferredUserSubscriptionStatus represents the subscription status of a referred user
type ReferredUserSubscriptionStatus string

const (
	ReferredUserSubscriptionStatusActive   ReferredUserSubscriptionStatus = "active"
	ReferredUserSubscriptionStatusCanceled ReferredUserSubscriptionStatus = "canceled"
	ReferredUserSubscriptionStatusInactive ReferredUserSubscriptionStatus = "inactive"
)

// ReferredUserResponse represents a user who registered using a referral code
type ReferredUserResponse struct {
	ID                 uint                           `json:"id"`
	Name               string                         `json:"name"`
	Email              string                         `json:"email"`
	RegisteredAt       time.Time                      `json:"registered_at"`
	PlanType           string                         `json:"plan_type"`
	SubscriptionStatus ReferredUserSubscriptionStatus `json:"subscription_status"`
	MonthsSubscribed   int                            `json:"months_subscribed"`
	CurrentPeriodEnd   *time.Time                     `json:"current_period_end,omitempty"`
	ReferralCode       string                         `json:"referral_code"`
}

// ReferredUsersSummary represents summary statistics for referred users
type ReferredUsersSummary struct {
	TotalReferred        int `json:"total_referred"`
	ActiveSubscribers    int `json:"active_subscribers"`
	CanceledSubscribers  int `json:"canceled_subscribers"`
	InactiveSubscribers  int `json:"inactive_subscribers"`
	TotalMonthsGenerated int `json:"total_months_generated"`
}

// ReferredUsersListResponse represents the paginated response for referred users
type ReferredUsersListResponse struct {
	Users    []*ReferredUserResponse `json:"users"`
	Summary  ReferredUsersSummary    `json:"summary"`
	Total    int64                   `json:"total"`
	Page     int                     `json:"page"`
	PageSize int                     `json:"page_size"`
}

// ReferredUsersFilter represents filters for referred users list
type ReferredUsersFilter struct {
	Status    string `form:"status"`
	DateFrom  string `form:"date_from"`
	DateTo    string `form:"date_to"`
	SortBy    string `form:"sort_by"`
	SortOrder string `form:"sort_order"`
	Page      int    `form:"page"`
	PageSize  int    `form:"page_size"`
}
