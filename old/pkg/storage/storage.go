package storage

import (
	"io"
	"time"
)

type FileStorage interface {
	Upload(key string, reader io.Reader, contentType string, size int64) (*FileResult, error)
	Download(key string) (io.ReadCloser, error)
	Delete(key string) error
	GetPresignedURL(key string, expiry time.Duration) (string, error)
	GetPublicURL(key string) string
	BulkDelete(keys []string) error
	GetFileForDownload(key string) (*FileDownload, error)
	FileExists(key string) (bool, error)
}
