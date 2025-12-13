package dto

import (
	"mime/multipart"

	"github.com/nicolailuther/butter/internal/enums"
)

type CreateFolder struct {
	Name     string `json:"name"`
	ClientID uint   `json:"client_id"`
	ParentID uint   `json:"parent_id"`
}

type UpdateFolder struct {
	Name     string `json:"name"`
	ParentID *uint  `json:"parent_id"` // Pointer to handle null values
}

type GetFolderFilters struct {
	ClientID uint `json:"client_id"`
	FolderID uint `json:"folder_id"`
}

type CreateContent struct {
	ClientID     uint                    `json:"client_id" form:"client_id"`
	Type         enums.ContentType       `json:"type" form:"type"`
	FolderID     uint                    `json:"folder_id" form:"folder_id"`
	ContentFiles []*multipart.FileHeader `json:"content_files" form:"content_files"`
}

type UpdateContent struct {
	Name        string            `json:"name"`
	Type        enums.ContentType `json:"type"`
	Enabled     bool              `json:"enabled"`
	FolderID    *uint             `json:"folder_id"` // Pointer to allow null for root folder
	UseMirror   bool              `json:"use_mirror"`
	UseOverlays bool              `json:"use_overlays"`
}

type GenerateContent struct {
	AccountID uint              `json:"account_id"`
	Type      enums.ContentType `json:"type"`
	Quantity  int               `json:"quantity"`
}

type UpdateContentAccount struct {
	Enabled             *bool `json:"enabled"`
	TimesPosted         *int  `json:"times_posted"`
	AccountTotalViews   *int  `json:"total_views"`
	AccountAverageViews *int  `json:"avg_views"`
	TimesGenerated      *int  `json:"times_generated"`
}

type UpdateGeneratedContent struct {
	IsPosted    *bool `json:"is_posted"`
	MaybePosted *bool `json:"maybe_posted"`
}
