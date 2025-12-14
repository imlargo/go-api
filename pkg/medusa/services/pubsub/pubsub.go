package pubsub

import (
	"context"
)

// Bus combines Publisher and Subscriber
type Bus interface {
	Publisher
	Subscriber
}

// MessageBroker is the main interface for message broker implementations
type MessageBroker interface {
	Bus

	// Connect establishes connection to the message broker
	Connect(ctx context.Context) error

	// Disconnect closes the connection
	Disconnect() error

	// IsConnected returns the connection status
	IsConnected() bool

	// Health returns health status of the broker
	Health(ctx context.Context) error
}
