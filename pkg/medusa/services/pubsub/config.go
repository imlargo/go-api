package pubsub

import (
	"time"

	"github.com/imlargo/go-api/pkg/medusa/core/logger"
)

// Config holds the configuration for the pub/sub system
type Config struct {
	// Connection settings
	URL               string
	ConnectionName    string
	MaxReconnect      int
	ReconnectDelay    time.Duration
	ConnectionTimeout time.Duration

	// Message settings
	DefaultRetries int
	RetryDelay     time.Duration
	MessageTTL     time.Duration

	// Consumer settings
	PrefetchCount int
	PrefetchSize  int

	// Exchange settings
	ExchangeName string
	ExchangeType string
	Durable      bool
	AutoDelete   bool

	// Queue settings
	QueueDurable    bool
	QueueAutoDelete bool
	QueueExclusive  bool

	// Logging
	Logger *logger.Logger

	// Graceful shutdown
	ShutdownTimeout time.Duration
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		URL:               "amqp://guest:guest@localhost:5672/",
		ConnectionName:    "pubsub-client",
		MaxReconnect:      10,
		ReconnectDelay:    5 * time.Second,
		ConnectionTimeout: 30 * time.Second,
		DefaultRetries:    3,
		RetryDelay:        1 * time.Second,
		MessageTTL:        24 * time.Hour,
		PrefetchCount:     10,
		PrefetchSize:      0,
		ExchangeName:      "pubsub.events",
		ExchangeType:      "topic",
		Durable:           true,
		AutoDelete:        false,
		QueueDurable:      true,
		QueueAutoDelete:   false,
		QueueExclusive:    false,
		Logger:            logger.NewLogger(),
		ShutdownTimeout:   30 * time.Second,
	}
}
