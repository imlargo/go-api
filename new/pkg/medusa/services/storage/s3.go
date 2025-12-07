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

type s3Storage struct {
	client   *s3.Client
	config   StorageConfig
	provider StorageProvider
}

func NewR2Storage(provider StorageProvider, config StorageConfig) (FileStorage, error) {
	client, err := GetStorageClient(provider, config)
	if err != nil {
		return nil, err
	}

	return &s3Storage{
		client: client,
		config: config,
	}, nil
}

func (s *s3Storage) Upload(key string, reader io.Reader, contentType string, size int64) (*FileResult, error) {
	ctx := context.Background()

	// Prepare the put object input
	input := &s3.PutObjectInput{
		Bucket:        aws.String(s.config.BucketName),
		Key:           aws.String(key),
		Body:          reader,
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(size),
	}

	// Upload the file
	result, err := s.client.PutObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to R2: %w", err)
	}

	// Create the file model
	file := &FileResult{
		Key:         key,
		Size:        size,
		ContentType: contentType,
		Etag:        clearStringQuotes(aws.ToString(result.ETag)),
	}

	// Set the URL based on configuration
	if s.config.UsePublicURL {
		file.Url = s.GetPublicURL(key)
	}

	return file, nil
}

func (s *s3Storage) Download(key string) (io.ReadCloser, error) {
	ctx := context.Background()

	input := &s3.GetObjectInput{
		Bucket: aws.String(s.config.BucketName),
		Key:    aws.String(key),
	}

	result, err := s.client.GetObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to download file from R2: %w", err)
	}

	return result.Body, nil
}

func (s *s3Storage) Delete(key string) error {
	ctx := context.Background()

	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s.config.BucketName),
		Key:    aws.String(key),
	}

	_, err := s.client.DeleteObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete file from R2: %w", err)
	}

	return nil
}

func (s *s3Storage) GetPresignedURL(key string, expiry time.Duration) (string, error) {
	ctx := context.Background()

	// Create a presign client
	presignClient := s3.NewPresignClient(s.client)

	request := &s3.GetObjectInput{
		Bucket: aws.String(s.config.BucketName),
		Key:    aws.String(key),
	}

	// Generate presigned URL
	result, err := presignClient.PresignGetObject(ctx, request, func(opts *s3.PresignOptions) {
		opts.Expires = expiry
	})

	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return result.URL, nil
}

func (s *s3Storage) GetPublicURL(key string) string {
	if s.config.PublicDomain != "" {
		return fmt.Sprintf("https://%s/%s", s.config.PublicDomain, key)
	}

	// Default R2 public URL format
	return fmt.Sprintf("https://pub-%s.r2.dev/%s", s.config.AccountID, key)
}

func (s *s3Storage) BulkDelete(keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	ctx := context.Background()

	var allErrors []string

	// Process files in batches
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

	// Return all collected errors
	if len(allErrors) > 0 {
		return fmt.Errorf("bulk delete completed with errors: %v, %v", len(allErrors), allErrors[0])
	}

	return nil
}

// deleteBatch deletes a single batch of files (up to 1000 objects)
func (s *s3Storage) deleteBatch(ctx context.Context, keys []string) error {
	// Convert keys to ObjectIdentifier slice
	objects := make([]types.ObjectIdentifier, len(keys))
	for i, key := range keys {
		objects[i] = types.ObjectIdentifier{
			Key: aws.String(key),
		}
	}

	// Prepare delete input
	input := &s3.DeleteObjectsInput{
		Bucket: aws.String(s.config.BucketName),
		Delete: &types.Delete{
			Objects: objects,
			Quiet:   aws.Bool(false), // Set to true to reduce response size if you don't need error details
		},
	}

	// Execute bulk delete
	result, err := s.client.DeleteObjects(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to execute bulk delete: %w", err)
	}

	// Check for individual deletion errors
	if len(result.Errors) > 0 {
		var errorMessages []string
		for _, deleteError := range result.Errors {
			errorMessages = append(errorMessages, fmt.Sprintf("key %s: %s",
				aws.ToString(deleteError.Key), aws.ToString(deleteError.Message)))
		}
		return fmt.Errorf("some files failed to delete: %v", errorMessages)
	}

	return nil
}

func clearStringQuotes(s string) string {
	var quotesRegex = regexp.MustCompile(`"`)
	return quotesRegex.ReplaceAllString(s, "")
}

func (s *s3Storage) GetFileForDownload(key string) (*FileDownload, error) {
	ctx := context.Background()

	input := &s3.GetObjectInput{
		Bucket: aws.String(s.config.BucketName),
		Key:    aws.String(key),
	}

	result, err := s.client.GetObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get file from R2: %w", err)
	}

	// Extraer Content-Type
	contentType := "application/octet-stream"
	if result.ContentType != nil {
		contentType = *result.ContentType
	}

	// Extraer tamaño
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
