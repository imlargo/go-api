package models

import (
	"time"

	"github.com/nicolailuther/butter/internal/enums"
	"gorm.io/gorm"
)

type Account struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Username             string              `json:"username" gorm:"not null"`
	Name                 string              `json:"name" gorm:"not null"`
	Platform             enums.Platform      `json:"platform" gorm:"not null"`
	Followers            int                 `json:"followers" gorm:"default:0"`
	AccountUrl           string              `json:"account_url" gorm:"not null"`
	Enabled              bool                `json:"enabled"  gorm:"default:true"`
	PostingGoal          int                 `json:"posting_goal" gorm:"default:0"`
	SlideshowPostingGoal int                 `json:"slideshow_posting_goal" gorm:"default:0"`
	StoryPostingGoal     int                 `json:"story_posting_goal" gorm:"default:0"`
	AverageViews         int                 `json:"average_views" gorm:"default:0"`
	Bio                  string              `json:"bio"`
	BioLink              string              `json:"bio_link"`
	BioRequest           string              `json:"bio_request"`
	CrossPromo           string              `json:"cross_promo"`
	TrackingLinkUrl      string              `json:"tracking_link_url"`
	DailyMarketingCost   float64             `json:"daily_marketing_cost" gorm:"default:0"`
	AccountRole          enums.AccountRole   `json:"account_role" gorm:"default:backup;not null" `
	Status               enums.AccountStatus `json:"status" gorm:"default:active;not null"`
	LastTrackedAt        *time.Time          `json:"last_tracked_at"`
	AutoGenerateEnabled  bool                `json:"auto_generate_enabled" gorm:"default:false"`
	AutoGenerateHour     *int                `json:"auto_generate_hour" gorm:"default:null"`

	OnlyfansTrackingLinkID uint `json:"onlyfans_tracking_link_id" gorm:"default:null"`
	ClientID               uint `json:"client_id" gorm:"index;not null"` // Should not be null
	ProfileImageID         uint `json:"profile_image_id" gorm:"default:null"`

	Client               *Client               `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ProfileImage         *File                 `json:"profile_image" gorm:"foreignKey:ProfileImageID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	OnlyfansTrackingLink *OnlyfansTrackingLink `json:"onlyfans_tracking_link" gorm:"constraint:OnDelete:SET NULL;"`
	Users                []*User               `json:"-" gorm:"many2many:user_accounts;"`
	TextOverlays         []*TextOverlay        `gorm:"many2many:text_overlay_accounts" json:"-"`
}
