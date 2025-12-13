package jobs

// Configuration constants for job execution

const (
	// DefaultBatchSize is the default number of items to process in a single batch
	DefaultBatchSize = 250

	// DefaultNumWorkers is the default number of concurrent workers for parallel processing
	// Reduced from 5 to 3 to prevent overwhelming the Instagram API and avoid timeout errors
	DefaultNumWorkers = 3

	// SaveAnalytics determines whether analytics should be saved during post tracking
	SaveAnalytics = true

	// UpdatePosts determines whether posts should be updated during post tracking
	UpdatePosts = true
)

// Update types for account tracking
type updateType int

const (
	updateTypeMinimal updateType = iota
	updateTypeFull
)
