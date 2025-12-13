package dto

import "github.com/nicolailuther/butter/internal/enums"

type CreatePostRequest struct {
	Url                string            `json:"url"`
	AccountID          uint              `json:"account_id" gorm:"index:idx_posts_created_at_account_id;index"`
	GeneratedContentID uint              `json:"generated_content_id" gorm:"default:null"`
	ContentType        enums.ContentType `json:"content_type"`
}

type TrackPostRequest struct {
	Url       string `json:"url"`
	AccountID uint   `json:"account_id" gorm:"index:idx_posts_created_at_account_id;index"`
}

type SyncPostsRequest struct {
	AccountID uint `json:"account_id" binding:"required"`
}
