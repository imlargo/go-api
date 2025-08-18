package dto

import (
	"mime/multipart"
)

type UploadFileRequest struct {
	File *multipart.FileHeader `form:"file" binding:"required"`
}

type GetPresignedURLRequest struct {
	ExpiryMins int `json:"expiry_minutes,omitempty"`
}

type PresignedURLResponse struct {
	Url       string `json:"url"`
	ExpiresAt string `json:"expires_at"`
}
