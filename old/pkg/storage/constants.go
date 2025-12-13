package storage

// Storage operation configuration constants
const (
	// MaxBatchSize is the maximum number of items that can be deleted in a single batch operation
	// This limit is imposed by R2/S3 for bulk delete operations
	MaxBatchSize = 1000
)
