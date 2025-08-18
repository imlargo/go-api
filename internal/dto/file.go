package dto

import (
	"mime/multipart"
)

type UploadFileRequest struct {
	File *multipart.FileHeader `form:"file" binding:"required"`
}

type CreatePresignedUrl struct {
	ExpiryMins int `json:"expiry_minutes,omitempty"`
}

type PresignedURL struct {
	Url       string `json:"url"`
	ExpiresAt string `json:"expires_at"`
}
