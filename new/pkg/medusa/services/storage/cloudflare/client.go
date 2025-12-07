package cloudflare

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/imlargo/go-api/pkg/medusa/services/storage"
)

func newR2Client(r2Config storage.StorageConfig) (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(r2Config.AccessKeyID, r2Config.SecretAccessKey, "")),
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
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", r2Config.AccountID))
		o.UsePathStyle = true
	})

	return client, nil
}
