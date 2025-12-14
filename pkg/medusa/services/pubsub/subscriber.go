package pubsub

import "context"

// Subscriber defines the interface for consuming messages
type Subscriber interface {
	// Subscribe registers a handler for a topic
	Subscribe(ctx context.Context, topic string, handler MessageHandler, opts ...SubscribeOption) error

	// Unsubscribe removes a subscription
	Unsubscribe(topic string) error

	// Close closes the subscriber and releases resources
	Close() error
}

// SubscribeOptions configures subscription behavior
type SubscribeOptions struct {
	QueueName         string
	ConsumerTag       string
	AutoAck           bool
	Exclusive         bool
	PrefetchCount     int
	PrefetchSize      int
	RetryPolicy       *RetryPolicy
	DeadLetterQueue   *DeadLetterQueue
	CircuitBreaker    *CircuitBreakerConfig
	Middleware        MiddlewareChain
	ConcurrentWorkers int
}

// SubscribeOption is a functional option for Subscribe
type SubscribeOption func(*SubscribeOptions)

// WithQueueName sets the queue name
func WithQueueName(name string) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.QueueName = name
	}
}

// WithAutoAck enables/disables auto-acknowledgement
func WithAutoAck(autoAck bool) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.AutoAck = autoAck
	}
}

// WithPrefetch sets prefetch count and size
func WithPrefetch(count, size int) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.PrefetchCount = count
		o.PrefetchSize = size
	}
}

// WithRetryPolicy sets retry policy
func WithRetryPolicy(policy *RetryPolicy) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.RetryPolicy = policy
	}
}

// WithDeadLetterQueue sets DLQ configuration
func WithDeadLetterQueue(dlq *DeadLetterQueue) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.DeadLetterQueue = dlq
	}
}

// WithCircuitBreaker sets circuit breaker configuration
func WithCircuitBreaker(cb *CircuitBreakerConfig) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.CircuitBreaker = cb
	}
}

// WithMiddleware adds middleware to the subscription
func WithMiddleware(mw ...Middleware) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.Middleware = append(o.Middleware, mw...)
	}
}

// WithConcurrency sets concurrent workers for message processing
func WithConcurrency(workers int) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.ConcurrentWorkers = workers
	}
}
