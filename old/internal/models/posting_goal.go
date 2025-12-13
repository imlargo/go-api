package models

import "time"

type PostingGoal struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	PostGoal      int `json:"post_goal" gorm:"default:0"`
	SlideshowGoal int `json:"slideshow_goal" gorm:"default:0"`
	StoryGoal     int `json:"story_goal" gorm:"default:0"`

	TotalPosts      int `json:"total_posts" gorm:"default:0"`
	TotalSlideshows int `json:"total_slideshows" gorm:"default:0"`
	TotalStories    int `json:"total_stories" gorm:"default:0"`

	AccountID     uint    `json:"account_id" gorm:"index;default:null"`
	MarketingCost float64 `json:"marketing_cost" gorm:"default:0"`

	Account *Account `json:"-" gorm:"foreignKey:AccountID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
