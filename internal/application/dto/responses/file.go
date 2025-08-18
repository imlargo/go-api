package responsesdto

import "io"

type PresignedURLResponse struct {
	Url       string `json:"url"`
	ExpiresAt string `json:"expires_at"`
}

type FileDownloadDto struct {
	Content     io.ReadCloser
	ContentType string
	Size        int64
}
