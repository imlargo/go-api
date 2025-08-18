package storage

import (
	"io"
	"time"

	responsesdto "github.com/imlargo/go-api-template/internal/application/dto/responses"
	"github.com/imlargo/go-api-template/internal/domain/models"
)

type StorageAdapter interface {
	Upload(key string, reader io.Reader, contentType string, size int64) (*models.File, error)
	Download(key string) (io.ReadCloser, error)
	Delete(key string) error
	GetPresignedURL(key string, expiry time.Duration) (string, error)
	GetPublicURL(key string) string
	BulkDelete(keys []string) error
	GetFileForDownload(key string) (*responsesdto.FileDownloadDto, error)
}
