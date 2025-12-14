package pubsub

import (
	"context"
	"time"
)

// Publisher defines the contract for publishing messages to topics/exchanges
type Publisher interface {
	// Publish sends a message to the specified topic with options
	Publish(ctx context.Context, topic string, message *Message, opts ...PublishOption) error
	// Close gracefully shuts down the publisher
	Close() error
}

// PublishOption configures message publishing
type PublishOption func(*PublishOptions)

// PublishOptions holds options for publishing
type PublishOptions struct {
	Persistent    bool
	Priority      uint8
	Expiration    time.Duration
	Headers       map[string]interface{}
	CorrelationID string
	ReplyTo       string
	Mandatory     bool
	Immediate     bool
}

// WithPersistent makes the message persistent
func WithPersistent() PublishOption {
	return func(o *PublishOptions) {
		o.Persistent = true
	}
}

// WithPriority sets message priority (0-9)
func WithPriority(priority uint8) PublishOption {
	return func(o *PublishOptions) {
		o.Priority = priority
	}
}

// WithExpiration sets message TTL
func WithExpiration(ttl time.Duration) PublishOption {
	return func(o *PublishOptions) {
		o.Expiration = ttl
	}
}

// WithHeaders adds custom headers
func WithHeaders(headers map[string]interface{}) PublishOption {
	return func(o *PublishOptions) {
		o.Headers = headers
	}
}

// WithCorrelationID sets correlation ID for request-reply pattern
func WithCorrelationID(id string) PublishOption {
	return func(o *PublishOptions) {
		o.CorrelationID = id
	}
}

// WithReplyTo sets the reply-to queue for RPC pattern
func WithReplyTo(queue string) PublishOption {
	return func(o *PublishOptions) {
		o.ReplyTo = queue
	}
}
