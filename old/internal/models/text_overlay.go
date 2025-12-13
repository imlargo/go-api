package models

import (
	"time"

	"gorm.io/gorm"
)

type TextOverlay struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Content  string `json:"content" gorm:"not null"`
	Enabled  bool   `json:"enabled" gorm:"not null;default:true"`
	ClientID uint   `json:"client_id" gorm:"not null"`

	Client   *Client    `json:"-" swaggerignore:"true"`
	Accounts []*Account `gorm:"many2many:text_overlay_accounts" json:"accounts"`
}
