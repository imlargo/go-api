package pubsub

import "context"

// Connector manages connections to the messaging backend
type Connector interface {
	// Connect establishes connection to the messaging system
	Connect(ctx context.Context) error
	// Disconnect closes the connection
	Disconnect() error
	// IsConnected returns current connection status
	IsConnected() bool
}
