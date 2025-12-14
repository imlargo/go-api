package pubsub

import (
	"context"
	"time"
)

// Message represents a message in the pub/sub system
type Message struct {
	ID            string
	Topic         string
	Payload       []byte
	Headers       map[string]string
	CorrelationID string
	CausationID   string
	Timestamp     time.Time
	ContentType   string
	Priority      int
	TTL           time.Duration
	Metadata      map[string]interface{}
}

// MessageHandler processes incoming messages
type MessageHandler func(ctx context.Context, msg *Message) error
