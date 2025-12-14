package pubsub

import "context"

// Connection represents a connection to the messaging backend
type Connection interface {
	Connect(ctx context.Context) error
	Disconnect() error
	IsConnected() bool
	Reconnect(ctx context.Context) error
}
