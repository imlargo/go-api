package pubsub

import "time"

// ExchangeType defines message routing pattern
type ExchangeType string

const (
	ExchangeDirect  ExchangeType = "direct"
	ExchangeFanout  ExchangeType = "fanout"
	ExchangeTopic   ExchangeType = "topic"
	ExchangeHeaders ExchangeType = "headers"
)

// RetryPolicy defines retry behavior
type RetryPolicy struct {
	MaxAttempts     int
	InitialInterval time.Duration
	MaxInterval     time.Duration
	Multiplier      float64
	RandomFactor    float64
}

// DeadLetterQueue configuration
type DeadLetterQueue struct {
	Enabled    bool
	QueueName  string
	MaxRetries int
}

// CircuitBreakerConfig defines circuit breaker settings
type CircuitBreakerConfig struct {
	Enabled          bool
	MaxFailures      int
	ResetTimeout     time.Duration
	HalfOpenRequests int
}

// TopologyConfig defines broker topology
type TopologyConfig struct {
	Exchanges []ExchangeConfig
	Queues    []QueueConfig
	Bindings  []BindingConfig
}

// ExchangeConfig defines an exchange
type ExchangeConfig struct {
	Name       string
	Type       ExchangeType
	Durable    bool
	AutoDelete bool
	Internal   bool
	NoWait     bool
	Args       map[string]interface{}
}

// QueueConfig defines a queue
type QueueConfig struct {
	Name       string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Args       map[string]interface{}
}

// BindingConfig defines a binding between exchange and queue
type BindingConfig struct {
	Queue      string
	Exchange   string
	RoutingKey string
	NoWait     bool
	Args       map[string]interface{}
}

// HealthCheck represents health status
type HealthCheck struct {
	Status  string
	Details map[string]interface{}
}
