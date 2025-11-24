package models

import (
	"time"

	"github.com/nicolailuther/butter/internal/enums"
	"gorm.io/gorm"
)

type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Name            string          `json:"name" gorm:"not null"`
	Email           string          `json:"email" gorm:"unique;not null"`
	Password        string          `json:"password" gorm:"not null"`
	Role            enums.UserRole  `json:"role"`
	ChangedPassword bool            `json:"changed_password"`
	TierLevel       enums.TierLevel `json:"tier_level"` // User's subscription tier level

	// Stripe-related fields for subscription management
	StripeCustomerID   string    `json:"stripe_customer_id" gorm:"default:null"`
	SubscriptionStatus string    `json:"subscription_status"`
	CurrentPeriodEnd   time.Time `json:"current_period_end"`

	// Onboarding preferences
	Type     enums.UserType `json:"type" gorm:"type:varchar(50)"`
	Industry enums.Industry `json:"industry" gorm:"type:varchar(50)"`
	Goal     enums.Goal     `json:"goal" gorm:"type:varchar(50)"`
	TeamSize enums.TeamSize `json:"team_size" gorm:"type:varchar(50)"`

	CreatedBy      uint `json:"created_by" gorm:"index;default:null" `
	ReferralCodeID uint `json:"referral_code_id" gorm:"default:null"`

	Creator *User `json:"creator" gorm:"foreignKey:CreatedBy;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" swaggerignore:"true"`

	AssignedClients []*Client  `gorm:"many2many:user_clients" json:"assigned_clients"`
	Accounts        []*Account `gorm:"many2many:user_accounts;" json:"assigned_accounts"`

	ReferralCode *ReferralCode `gorm:"foreignKey:ReferralCodeID" json:"-"`
}
