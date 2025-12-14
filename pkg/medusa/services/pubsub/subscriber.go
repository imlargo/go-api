package pubsub

// Subscriber defines the interface for subscribing to topics
type Subscriber interface {
	Subscribe(topic string, handler HandlerFunc, opts ...SubscribeOption) error
	Unsubscribe(topic string) error
	Close() error
}

// SubscribeOption configures subscribe behavior
type SubscribeOption func(*subscribeOptions)

type subscribeOptions struct {
	queueName     string
	routingKey    string
	autoAck       bool
	exclusive     bool
	concurrency   int
	retryStrategy RetryStrategy
}

func WithQueueName(name string) SubscribeOption {
	return func(o *subscribeOptions) {
		o.queueName = name
	}
}

func WithRoutingKey(key string) SubscribeOption {
	return func(o *subscribeOptions) {
		o.routingKey = key
	}
}

func WithAutoAck(autoAck bool) SubscribeOption {
	return func(o *subscribeOptions) {
		o.autoAck = autoAck
	}
}

func WithConcurrency(n int) SubscribeOption {
	return func(o *subscribeOptions) {
		o.concurrency = n
	}
}

func WithRetryStrategy(strategy RetryStrategy) SubscribeOption {
	return func(o *subscribeOptions) {
		o.retryStrategy = strategy
	}
}
