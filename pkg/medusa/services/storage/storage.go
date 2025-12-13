package storage

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// FileStorage defines the interface for file storage operations
type FileStorage interface {
	Upload(key string, reader io.Reader, contentType string, size int64) (*FileResult, error)
	Download(key string) (io.ReadCloser, error)
	Delete(key string) error
	GetPresignedURL(key string, expiry time.Duration) (string, error)
	GetPublicURL(key string) string
	BulkDelete(keys []string) error
	GetFileForDownload(key string) (*FileDownload, error)
}

type fileStorage struct {
	client   *s3.Client
	config   StorageConfig
	provider StorageProvider
}

var quotesRegex = regexp.MustCompile(`"`)

// NewFileStorage creates a new file storage instance with the specified provider
func NewFileStorage(provider StorageProvider, config StorageConfig) (FileStorage, error) {
	if !provider.IsValid() {
		return nil, fmt.Errorf("unsupported storage provider: %s", provider)
	}

	client, err := NewStorageClient(provider, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage client: %w", err)
	}

	return &fileStorage{
		client:   client,
		config:   config,
		provider: provider,
	}, nil
}

// Upload uploads a file to the storage
func (s *fileStorage) Upload(key string, reader io.Reader, contentType string, size int64) (*FileResult, error) {
	ctx := context.Background()

	input := &s3.PutObjectInput{
		Bucket:        aws.String(s.config.BucketName),
		Key:           aws.String(key),
		Body:          reader,
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(size),
	}

	result, err := s.client.PutObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	file := &FileResult{
		Key:         key,
		Size:        size,
		ContentType: contentType,
		Etag:        clearStringQuotes(aws.ToString(result.ETag)),
	}

	if s.config.UsePublicURL {
		file.Url = s.GetPublicURL(key)
	}

	return file, nil
}

// Download downloads a file from the storage
func (s *fileStorage) Download(key string) (io.ReadCloser, error) {
	ctx := context.Background()

	input := &s3.GetObjectInput{
		Bucket: aws.String(s.config.BucketName),
		Key:    aws.String(key),
	}

	result, err := s.client.GetObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	return result.Body, nil
}

// Delete removes a file from the storage
func (s *fileStorage) Delete(key string) error {
	ctx := context.Background()

	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s.config.BucketName),
		Key:    aws.String(key),
	}

	_, err := s.client.DeleteObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// GetPresignedURL generates a presigned URL for temporary access to a file
func (s *fileStorage) GetPresignedURL(key string, expiry time.Duration) (string, error) {
	ctx := context.Background()

	presignClient := s3.NewPresignClient(s.client)

	request := &s3.GetObjectInput{
		Bucket: aws.String(s.config.BucketName),
		Key:    aws.String(key),
	}

	result, err := presignClient.PresignGetObject(ctx, request, func(opts *s3.PresignOptions) {
		opts.Expires = expiry
	})

	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return result.URL, nil
}

// GetPublicURL returns the public URL for a file
func (s *fileStorage) GetPublicURL(key string) string {
	// Use custom domain if configured
	if s.config.PublicDomain != "" {
		return fmt.Sprintf("https://%s/%s", s.config.PublicDomain, key)
	}

	// Generate provider-specific public URLs
	switch s.provider {
	case StorageProviderR2:
		return fmt.Sprintf("https://pub-%s.r2.dev/%s", s.config.AccountID, key)
	default:
		return ""
	}
}

// BulkDelete deletes multiple files in batches
func (s *fileStorage) BulkDelete(keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	ctx := context.Background()
	var allErrors []string

	// Process files in batches of MaxBatchSize
	for i := 0; i < len(keys); i += MaxBatchSize {
		end := i + MaxBatchSize
		if end > len(keys) {
			end = len(keys)
		}

		batch := keys[i:end]
		if err := s.deleteBatch(ctx, batch); err != nil {
			allErrors = append(allErrors, fmt.Sprintf("batch %d-%d: %v", i, end-1, err))
		}
	}

	if len(allErrors) > 0 {
		return fmt.Errorf("bulk delete completed with %d error(s): %s",
			len(allErrors), allErrors[0])
	}

	return nil
}

// deleteBatch deletes a single batch of files (up to MaxBatchSize objects)
func (s *fileStorage) deleteBatch(ctx context.Context, keys []string) error {
	objects := make([]types.ObjectIdentifier, len(keys))
	for i, key := range keys {
		objects[i] = types.ObjectIdentifier{
			Key: aws.String(key),
		}
	}

	input := &s3.DeleteObjectsInput{
		Bucket: aws.String(s.config.BucketName),
		Delete: &types.Delete{
			Objects: objects,
			Quiet:   aws.Bool(false),
		},
	}

	result, err := s.client.DeleteObjects(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to execute bulk delete: %w", err)
	}

	// Check for individual deletion errors
	if len(result.Errors) > 0 {
		errorMessages := make([]string, 0, len(result.Errors))
		for _, deleteError := range result.Errors {
			errorMessages = append(errorMessages,
				fmt.Sprintf("key %s: %s",
					aws.ToString(deleteError.Key),
					aws.ToString(deleteError.Message)))
		}
		return fmt.Errorf("failed to delete %d file(s): %v",
			len(result.Errors), errorMessages)
	}

	return nil
}

// GetFileForDownload retrieves a file with its metadata for download
func (s *fileStorage) GetFileForDownload(key string) (*FileDownload, error) {
	ctx := context.Background()

	input := &s3.GetObjectInput{
		Bucket: aws.String(s.config.BucketName),
		Key:    aws.String(key),
	}

	result, err := s.client.GetObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	contentType := "application/octet-stream"
	if result.ContentType != nil {
		contentType = *result.ContentType
	}

	size := int64(0)
	if result.ContentLength != nil {
		size = *result.ContentLength
	}

	return &FileDownload{
		Content:     result.Body,
		ContentType: contentType,
		Size:        size,
	}, nil
}

// clearStringQuotes removes quotes from strings (commonly found in ETags)
func clearStringQuotes(s string) string {
	return quotesRegex.ReplaceAllString(s, "")
}
