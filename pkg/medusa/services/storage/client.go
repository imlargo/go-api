package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var ErrUnsupportedStorageProvider = errors.New("unsupported storage provider")

func NewStorageClient(provider StorageProvider, config StorageConfig) (*s3.Client, error) {
	switch provider {
	case StorageProviderR2:
		return NewR2Client(config)
	default:
		return nil, ErrUnsupportedStorageProvider
	}
}

func NewR2Client(r2cfg StorageConfig) (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(r2cfg.AccessKeyID, r2cfg.SecretAccessKey, "")),
		config.WithRegion("auto"),
		// Cloudflare R2 does not support AWS S3 checksum calculation or validation features.
		// These settings are disabled to ensure compatibility and prevent errors when interacting with R2.
		config.WithRequestChecksumCalculation(aws.RequestChecksumCalculationUnset),
		config.WithResponseChecksumValidation(aws.ResponseChecksumValidationUnset),
	)

	if err != nil {
		return nil, fmt.Errorf("error loading AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", r2cfg.AccountID))
		o.UsePathStyle = true
	})

	return client, nil
}
