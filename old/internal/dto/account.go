package dto

import (
	"github.com/nicolailuther/butter/internal/enums"
	"github.com/nicolailuther/butter/internal/models"
)

type AccountInsightsResponse struct {
	CPM               float64 `json:"cpm"`
	MediaViews        int     `json:"media_views"`
	DaysSinceLastPost int     `json:"days_since_last_post"`
	Leaders           string  `json:"leaders"`
	Posters           string  `json:"posters"`
}
type AccountResponse struct {
	models.Account
	DailyPosts      int     `json:"_daily_posts"`
	WeeklyPosts     int     `json:"_weekly_posts"`
	ViewsPercentage float64 `json:"_views_percentage"`
}

type CreateAccountRequest struct {
	Username           string            `json:"username" form:"username"`
	Platform           enums.Platform    `json:"platform" form:"platform"`
	CrossPromo         string            `json:"cross_promo" form:"cross_promo"`
	TrackingLinkUrl    string            `json:"tracking_link_url" form:"tracking_link_url"`
	DailyMarketingCost float64           `json:"daily_marketing_cost" form:"daily_marketing_cost"`
	AccountRole        enums.AccountRole `json:"account_role" form:"account_role"`
	ClientID           uint              `json:"client_id" form:"client_id"`
}

type UpdateAccountRequest struct {
	Username               *string           `json:"username"`
	Enabled                bool              `json:"enabled" `
	PostingGoal            *int              `json:"posting_goal"`
	SlideshowPostingGoal   *int              `json:"slideshow_posting_goal"`
	StoryPostingGoal       *int              `json:"story_posting_goal"`
	BioRequest             string            `json:"bio_request"`
	CrossPromo             string            `json:"cross_promo"`
	TrackingLinkUrl        string            `json:"tracking_link_url"`
	DailyMarketingCost     float64           `json:"daily_marketing_cost"`
	AccountRole            enums.AccountRole `json:"account_role"`
	Mirroring              bool              `json:"mirroring"`
	TextOverlays           bool              `json:"text_overlays"`
	OnlyfansTrackingLinkID uint              `json:"onlyfans_tracking_link_id"`
	AutoGenerateEnabled    *bool             `json:"auto_generate_enabled"`
	AutoGenerateHour       *int              `json:"auto_generate_hour"`
}

type AccountLimitStatusResponse struct {
	CurrentCount  int  `json:"current_count"`
	Limit         int  `json:"limit"`
	Remaining     int  `json:"remaining"`
	CanCreateMore bool `json:"can_create_more"`
}
