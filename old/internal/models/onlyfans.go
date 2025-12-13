package models

import (
	"time"

	"github.com/nicolailuther/butter/internal/enums"
	"gorm.io/gorm"
)

type OnlyfansAccount struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	ExternalID  string                   `json:"external_id" gorm:"uniqueIndex"`
	RealID      uint                     `json:"real_id" gorm:"uniqueIndex;not null"`
	Email       string                   `json:"email" gorm:"index"`
	Username    string                   `json:"username"`
	Name        string                   `json:"name"`
	Subscribers int                      `json:"subscribers"`
	AuthStatus  enums.OnlyfansAuthStatus `json:"auth_status" gorm:"default:pending"`

	ClientID uint `json:"client_id" gorm:"index;default:null"`

	Client *Client `json:"client"`
}

type OnlyfansTransaction struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	ExternalID  string                    `json:"external_id" gorm:"uniqueIndex"`
	RevenueType enums.OnlyfansRevenueType `json:"revenue_type"`
	Amount      float64                   `json:"amount"`

	ClientID          uint `json:"client_id" gorm:"index"`
	OnlyfansAccountID uint `json:"onlyfans_account_id" gorm:"index;default:null"`

	Client          *Client          `json:"-"`
	OnlyfansAccount *OnlyfansAccount `json:"-"`
}

type OnlyfansTrackingLink struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	ExternalID  int     `json:"external_id" gorm:"uniqueIndex"`
	Name        string  `json:"name"`
	Url         string  `json:"url"`
	Clicks      int     `json:"clicks" gorm:"default:0"`
	Subscribers int     `json:"subscribers" gorm:"default:0"`
	Revenue     float64 `json:"revenue" gorm:"default:0"`

	ClientID          uint `json:"client_id" gorm:"index"`
	OnlyfansAccountID uint `json:"onlyfans_account_id" gorm:"index"`

	Client          *Client          `swaggerignore:"true" json:"-"`
	OnlyfansAccount *OnlyfansAccount `json:"account"`
}
