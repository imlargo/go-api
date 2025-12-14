package pubsub

import (
	"time"

	"github.com/imlargo/go-api/pkg/medusa/core/logger"
)

// Config holds the configuration for the pub/sub system
type Config struct {
	// Connection settings
	URL            string
	MaxReconnects  int
	ReconnectDelay time.Duration
	ConnectionName string

	// Publisher settings
	PublisherConfirm bool
	PublishTimeout   time.Duration

	// Subscriber settings
	PrefetchCount    int
	AutoAck          bool
	ExclusiveConsume bool

	// Quality of Service
	QoS QoSConfig

	// Retry policy
	RetryPolicy RetryPolicy

	// Dead Letter configuration
	DeadLetter DeadLetterConfig

	// Logging
	Logger *logger.Logger
}

// QoSConfig defines Quality of Service settings
type QoSConfig struct {
	PrefetchCount int
	PrefetchSize  int
	Global        bool
}

// RetryPolicy defines retry behavior for failed messages
type RetryPolicy struct {
	MaxRetries    int
	InitialDelay  time.Duration
	MaxDelay      time.Duration
	Multiplier    float64
	EnableBackoff bool
}

// DeadLetterConfig defines dead letter queue settings
type DeadLetterConfig struct {
	Enabled      bool
	ExchangeName string
	QueueName    string
	RoutingKey   string
	TTL          time.Duration
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		URL:              "amqp://guest:guest@localhost:5672/",
		MaxReconnects:    10,
		ReconnectDelay:   5 * time.Second,
		PublisherConfirm: true,
		PublishTimeout:   30 * time.Second,
		PrefetchCount:    10,
		AutoAck:          false,
		QoS: QoSConfig{
			PrefetchCount: 10,
			PrefetchSize:  0,
			Global:        false,
		},
		RetryPolicy: RetryPolicy{
			MaxRetries:    3,
			InitialDelay:  1 * time.Second,
			MaxDelay:      30 * time.Second,
			Multiplier:    2.0,
			EnableBackoff: true,
		},
		DeadLetter: DeadLetterConfig{
			Enabled:      true,
			ExchangeName: "dlx",
			QueueName:    "dlq",
		},
		Logger: logger.NewLogger(),
	}
}
