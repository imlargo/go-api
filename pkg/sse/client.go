package sse

import (
	"context"
	"time"
)

type Connection interface {
	GetChannel() <-chan *Message
	GetContext() context.Context
	UpdateLastSeen()
}

type clientConn struct {
	ID       string
	UserID   uint
	Channel  chan *Message
	Context  context.Context
	Cancel   context.CancelFunc
	LastSeen time.Time
}

func (c *clientConn) GetChannel() <-chan *Message {
	return c.Channel
}

func (c *clientConn) GetContext() context.Context {
	return c.Context
}

func (c *clientConn) UpdateLastSeen() {
	c.LastSeen = time.Now()
}
