package sse

import (
	"context"
	"time"

	"github.com/imlargo/go-api-template/internal/domain/models"
)

type ConnectionClient struct {
	ID       string
	UserID   uint
	Channel  chan *models.Notification
	Context  context.Context
	Cancel   context.CancelFunc
	LastSeen time.Time
}

func (c *ConnectionClient) GetChannel() <-chan *models.Notification {
	return c.Channel
}

func (c *ConnectionClient) GetContext() context.Context {
	return c.Context
}

func (c *ConnectionClient) UpdateLastSeen() {
	c.LastSeen = time.Now()
}
