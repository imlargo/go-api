package models

import (
	"time"

	"github.com/nicolailuther/butter/internal/enums"
)

type Post struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at" gorm:"index:idx_posts_created_at_account_id"`
	UpdatedAt time.Time `json:"updated_at"`

	Platform   enums.Platform    `json:"platform"`
	Type       enums.ContentType `json:"type" gorm:"index;default:'video'"`
	Url        string            `json:"url"`
	TotalViews int               `json:"total_views"`
	IsTracked  bool              `json:"is_tracked"`
	IsDeleted  bool              `json:"is_deleted" gorm:"default:false"`
	Track      bool              `json:"track" gorm:"default:true"`

	AccountContentID   uint `json:"account_content_id" gorm:"index;default:null"`
	ContentID          uint `json:"content_id" gorm:"index;default:null"`
	AccountID          uint `json:"account_id" gorm:"index:idx_posts_created_at_account_id;index;default:null"`
	GeneratedContentID uint `json:"generated_content_id" gorm:"default:null"`
	ThumbnailID        uint `json:"thumbnail_id" gorm:"default:null"`
	TextOverlayID      uint `json:"text_overlay_id" gorm:"default:null"`

	Thumbnail        *File             `json:"thumbnail" gorm:"foreignKey:ThumbnailID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	AccountContent   *ContentAccount   `json:"-" gorm:"foreignKey:AccountContentID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Content          *Content          `json:"-" gorm:"foreignKey:ContentID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	GeneratedContent *GeneratedContent `json:"generated_content" gorm:"foreignKey:GeneratedContentID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	TextOverlay      *TextOverlay      `json:"-" gorm:"foreignKey:TextOverlayID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Account          *Account          `json:"-" gorm:"foreignKey:AccountID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}
