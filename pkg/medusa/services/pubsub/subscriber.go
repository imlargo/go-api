package pubsub

import "context"

// Subscriber defines the contract for subscribing to topics and consuming messages
type Subscriber interface {
	// Subscribe registers a handler for messages on a topic
	Subscribe(ctx context.Context, topic string, handler MessageHandler, opts ...SubscribeOption) error
	// Unsubscribe removes a subscription from a topic
	Unsubscribe(topic string) error
	// Close gracefully shuts down the subscriber
	Close() error
}

// SubscribeOption configures subscription behavior
type SubscribeOption func(*SubscribeOptions)

// SubscribeOptions holds options for subscribing
type SubscribeOptions struct {
	QueueName     string
	Durable       bool
	AutoDelete    bool
	Exclusive     bool
	NoWait        bool
	ConsumerTag   string
	QoS           *QoSConfig
	RetryPolicy   *RetryPolicy
	DeadLetter    *DeadLetterConfig
	Middleware    []Middleware
	ErrorHandler  ErrorHandler
	AutoAck       bool
	PrefetchCount int
}

// WithQueueName specifies a custom queue name
func WithQueueName(name string) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.QueueName = name
	}
}

// WithDurable makes the queue durable
func WithDurable() SubscribeOption {
	return func(o *SubscribeOptions) {
		o.Durable = true
	}
}

// WithAutoDelete enables auto-deletion of the queue
func WithAutoDelete() SubscribeOption {
	return func(o *SubscribeOptions) {
		o.AutoDelete = true
	}
}

// WithExclusive makes the queue exclusive
func WithExclusive() SubscribeOption {
	return func(o *SubscribeOptions) {
		o.Exclusive = true
	}
}

// WithConsumerTag sets a custom consumer tag
func WithConsumerTag(tag string) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.ConsumerTag = tag
	}
}

// WithQoS sets Quality of Service parameters
func WithQoS(qos QoSConfig) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.QoS = &qos
	}
}

// WithRetryPolicy sets retry behavior
func WithRetryPolicy(policy RetryPolicy) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.RetryPolicy = &policy
	}
}

// WithDeadLetter configures dead letter queue
func WithDeadLetter(config DeadLetterConfig) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.DeadLetter = &config
	}
}

// WithMiddleware adds middleware to the subscription
func WithMiddleware(middleware ...Middleware) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.Middleware = append(o.Middleware, middleware...)
	}
}

// WithErrorHandler sets custom error handling
func WithErrorHandler(handler ErrorHandler) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.ErrorHandler = handler
	}
}

// WithAutoAck enables automatic acknowledgment
func WithAutoAck() SubscribeOption {
	return func(o *SubscribeOptions) {
		o.AutoAck = true
	}
}

// WithPrefetchCount sets the prefetch count
func WithPrefetchCount(count int) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.PrefetchCount = count
	}
}
