package pubsub

import "time"

// Metrics interface for observability
type Metrics interface {
	IncrementPublished(topic string)
	IncrementConsumed(topic string)
	IncrementFailed(topic string)
	IncrementRetry(topic string)
	RecordPublishLatency(topic string, duration time.Duration)
	RecordProcessingLatency(topic string, duration time.Duration)
}
