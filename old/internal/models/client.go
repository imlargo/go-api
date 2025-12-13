package models

import (
	"time"

	"github.com/nicolailuther/butter/internal/enums"
)

type Client struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Name              string               `json:"name" gorm:"not null"`
	CompanyPercentage int                  `json:"company_percentage" gorm:"default:0"`
	Industry          enums.ClientIndustry `json:"industry"`

	ProfileImageID uint `json:"profile_image_id" gorm:"default:null;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	UserID         uint `json:"user_id" gorm:"index;not null"` // Should be not null

	ProfileImage *File      `json:"profile_image" gorm:"foreignKey:ProfileImageID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	User         *User      `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Accounts     []*Account `json:"-"`
	Users        []*User    `json:"-" gorm:"many2many:user_clients;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" swaggerignore:"true"`
}
