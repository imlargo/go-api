package models

import "time"

type AccountAnalytic struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	Followers int    `json:"followers" gorm:"default:0"`
	Date      string `json:"date" gorm:"type:date;index:idx_date;index:idx_account_date"`
	AccountID uint   `json:"account_id" gorm:"index:idx_account;index:idx_account_date;default:null"`

	Account *Account `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

type PostAnalytic struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	TotalViews int    `json:"total_views" gorm:"default:0"`
	Date       string `json:"date" gorm:"type:date;index:idx_date;index:idx_post_date;not null"`
	PostID     uint   `json:"post_id" gorm:"index:idx_post;index:idx_post_date;index:idx_account_post;default:null"`
	AccountID  uint   `json:"account_id" gorm:"index:idx_account;index:idx_account_post;default:null"`

	Post    *Post    `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Account *Account `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
