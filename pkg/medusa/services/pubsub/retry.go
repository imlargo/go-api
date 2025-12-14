package pubsub

import "time"

// RetryStrategy defines how to handle message retries
type RetryStrategy interface {
	ShouldRetry(msg *Message, err error) bool
	NextDelay(msg *Message) time.Duration
}

// ExponentialBackoff implements exponential backoff retry strategy
type ExponentialBackoff struct {
	MaxRetries   int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
}

func (e *ExponentialBackoff) ShouldRetry(msg *Message, err error) bool {
	return msg.RetryCount < e.MaxRetries
}

func (e *ExponentialBackoff) NextDelay(msg *Message) time.Duration {
	delay := float64(e.InitialDelay) * float64(msg.RetryCount+1) * e.Multiplier
	if delay > float64(e.MaxDelay) {
		return e.MaxDelay
	}
	return time.Duration(delay)
}
