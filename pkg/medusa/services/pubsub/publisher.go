package pubsub

import (
	"context"
	"fmt"
	"time"
)

// Publisher defines the interface for publishing messages
type Publisher interface {
	Publish(ctx context.Context, topic string, payload interface{}, opts ...PublishOption) error
	Close() error
}

// PublishOption configures publish behavior
type PublishOption func(*publishOptions)

type publishOptions struct {
	headers    map[string]string
	priority   uint8
	expiration string
	persistent bool
}

func WithHeaders(headers map[string]string) PublishOption {
	return func(o *publishOptions) {
		o.headers = headers
	}
}

func WithPriority(priority uint8) PublishOption {
	return func(o *publishOptions) {
		o.priority = priority
	}
}

func WithExpiration(duration time.Duration) PublishOption {
	return func(o *publishOptions) {
		o.expiration = fmt.Sprintf("%d", duration.Milliseconds())
	}
}

func WithPersistent(persistent bool) PublishOption {
	return func(o *publishOptions) {
		o.persistent = persistent
	}
}
