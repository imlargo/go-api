package pubsub

import "errors"

var (
	ErrNotConnected     = errors.New("not connected to messaging backend")
	ErrAlreadyConnected = errors.New("already connected")
	ErrPublishFailed    = errors.New("failed to publish message")
	ErrSubscribeFailed  = errors.New("failed to subscribe to topic")
	ErrInvalidConfig    = errors.New("invalid configuration")
	ErrTimeout          = errors.New("operation timeout")
	ErrClosed           = errors.New("pubsub is closed")
)
