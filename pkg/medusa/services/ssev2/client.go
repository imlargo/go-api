package ssev2

import (
	"context"
	"time"
)

// Client represents a connected SSE client
type Client struct {
	ID       string
	Channel  chan Event
	ctx      context.Context
	cancel   context.CancelFunc
	lastSeen time.Time
}
