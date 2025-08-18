package sse2

import (
	"context"
	"time"
)

type SSENotificationConnection interface {
	GetChannel() <-chan *Message
	GetContext() context.Context
	UpdateLastSeen()
}

type ConnectionClient struct {
	ID       string
	UserID   uint
	Channel  chan *Message
	Context  context.Context
	Cancel   context.CancelFunc
	LastSeen time.Time
}

func (c *ConnectionClient) GetChannel() <-chan *Message {
	return c.Channel
}

func (c *ConnectionClient) GetContext() context.Context {
	return c.Context
}

func (c *ConnectionClient) UpdateLastSeen() {
	c.LastSeen = time.Now()
}
