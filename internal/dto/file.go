package dto

import (
	"io"
	"mime/multipart"
)

type PresignedURLResponse struct {
	Url       string `json:"url"`
	ExpiresAt string `json:"expires_at"`
}

type FileDownloadDto struct {
	Content     io.ReadCloser
	ContentType string
	Size        int64
}

type UploadFileRequest struct {
	File *multipart.FileHeader `form:"file" binding:"required"`
}

type GetPresignedURLRequest struct {
	ExpiryMins int `json:"expiry_minutes,omitempty"`
}
