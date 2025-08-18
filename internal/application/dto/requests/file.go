package requestsdto

import (
	"mime/multipart"
)

type UploadFileRequest struct {
	File *multipart.FileHeader `form:"file" binding:"required"`
}

type GetPresignedURLRequest struct {
	ExpiryMins int `json:"expiry_minutes,omitempty"`
}
