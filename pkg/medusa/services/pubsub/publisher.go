package pubsub

import (
	"context"
	"time"
)

// Publisher defines the interface for publishing messages
type Publisher interface {
	// Publish sends a message to a topic
	Publish(ctx context.Context, topic string, msg *Message) error

	// PublishBatch sends multiple messages atomically
	PublishBatch(ctx context.Context, messages []*Message) error

	// Request sends a message and waits for a reply (RPC pattern)
	Request(ctx context.Context, topic string, msg *Message, timeout time.Duration) (*Message, error)

	// Close closes the publisher and releases resources
	Close() error
}

// PublishOptions configures publish behavior
type PublishOptions struct {
	Mandatory   bool
	Immediate   bool
	ContentType string
	Priority    int
	TTL         time.Duration
	Persistent  bool
	Headers     map[string]interface{}
}

// PublishOption is a functional option for Publish
type PublishOption func(*PublishOptions)

// WithMandatory sets mandatory flag
func WithMandatory(mandatory bool) PublishOption {
	return func(o *PublishOptions) {
		o.Mandatory = mandatory
	}
}

// WithPriority sets message priority
func WithPriority(priority int) PublishOption {
	return func(o *PublishOptions) {
		o.Priority = priority
	}
}

// WithTTL sets message time-to-live
func WithTTL(ttl time.Duration) PublishOption {
	return func(o *PublishOptions) {
		o.TTL = ttl
	}
}

// WithPersistent sets persistence mode
func WithPersistent(persistent bool) PublishOption {
	return func(o *PublishOptions) {
		o.Persistent = persistent
	}
}

// WithHeaders sets custom headers
func WithHeaders(headers map[string]interface{}) PublishOption {
	return func(o *PublishOptions) {
		o.Headers = headers
	}
}
