package storage

import (
	"errors"
	"io"
	"time"
)

type EmptyStorage struct{}

func NewEmptyStorage() FileStorage {
	return &EmptyStorage{}
}

var (
	errorFileStorageNotEnabled = errors.New("file storage not enabled")
)

func (s *EmptyStorage) Upload(key string, reader io.Reader, contentType string, size int64) (*FileResult, error) {
	return nil, errorFileStorageNotEnabled
}

func (s *EmptyStorage) Download(key string) (io.ReadCloser, error) {
	return nil, errorFileStorageNotEnabled
}

func (s *EmptyStorage) Delete(key string) error {
	return errorFileStorageNotEnabled
}

func (s *EmptyStorage) GetPresignedURL(key string, expiry time.Duration) (string, error) {
	return "", errorFileStorageNotEnabled
}

func (s *EmptyStorage) GetPublicURL(key string) string {
	return ""
}

func (s *EmptyStorage) BulkDelete(keys []string) error {
	return errorFileStorageNotEnabled
}

func (s *EmptyStorage) GetFileForDownload(key string) (*FileDownload, error) {
	return nil, errorFileStorageNotEnabled
}
