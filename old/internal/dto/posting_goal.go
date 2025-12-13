package dto

import (
	"github.com/nicolailuther/butter/internal/enums"
	"github.com/nicolailuther/butter/internal/models"
)

type ClientPostingProgressSummary struct {
	ID           uint         `json:"id"`
	Name         string       `json:"name"`
	ProfileImage *models.File `json:"profile_image,omitempty" `

	Accounts []*AccountPostingProgressSummary `json:"accounts,omitempty"`

	TotalMarketingCost float64 `json:"total_marketing_cost"`
}

type AccountPostingProgressSummary struct {
	ID           uint              `json:"id"`
	Username     string            `json:"username"`
	Name         string            `json:"name"`
	Platform     enums.Platform    `json:"platform"`
	AccountRole  enums.AccountRole `json:"account_role"`
	ProfileImage *models.File      `json:"profile_image"`

	TotalPosts      int `json:"total_posts"`
	PostGoal        int `json:"post_goal"`
	TotalSlideshows int `json:"total_slideshows"`
	SlideshowGoal   int `json:"slideshow_goal"`
	TotalStories    int `json:"total_stories"`
	StoryGoal       int `json:"story_goal"`
	DaysCompleted   int `json:"days_completed"`
	TotalDays       int `json:"total_days"`

	MarketingCost float64 `json:"marketing_cost"`
}
