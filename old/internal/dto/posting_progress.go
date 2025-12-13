package dto

import "github.com/nicolailuther/butter/internal/enums"

type PostingProgressRawResult struct {
	ClientID              uint              `json:"client_id"`
	ClientName            string            `json:"client_name"`
	ClientProfileImageID  *uint             `json:"client_profile_image_id"`
	AccountID             uint              `json:"account_id"`
	AccountUsername       string            `json:"account_username"`
	AccountName           string            `json:"account_name"`
	AccountPlatform       enums.Platform    `json:"account_platform"`
	AccountRole           enums.AccountRole `json:"account_role"`
	AccountProfileImageID *uint             `json:"account_profile_image_id"`
	TotalPosts            int               `json:"total_posts"`
	PostGoal              int               `json:"post_goal"`
	TotalSlideshows       int               `json:"total_slideshows"`
	SlideshowGoal         int               `json:"slideshow_goal"`
	TotalStories          int               `json:"total_stories"`
	StoryGoal             int               `json:"story_goal"`
	DaysCompleted         int               `json:"days_completed"`
	TotalDays             int               `json:"total_days"`
	MarketingCost         float64           `json:"marketing_cost"`
}
