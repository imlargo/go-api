package pubsub

import (
	"context"
	"time"
)

// Message represents a message in the pub/sub system
type Message struct {
	ID          string            `json:"id"`
	Topic       string            `json:"topic"`
	Payload     []byte            `json:"payload"`
	Headers     map[string]string `json:"headers"`
	PublishedAt time.Time         `json:"published_at"`
	RetryCount  int               `json:"retry_count"`
}

// HandlerFunc is the signature for message handlers
type HandlerFunc func(ctx context.Context, msg *Message) error
