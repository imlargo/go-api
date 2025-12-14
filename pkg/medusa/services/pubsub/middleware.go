package pubsub

import (
	"context"
	"fmt"
	"time"

	"github.com/imlargo/go-api/pkg/medusa/core/logger"
)

// Middleware wraps a MessageHandler with additional functionality
type Middleware func(MessageHandler) MessageHandler

// ErrorHandler handles errors during message processing
type ErrorHandler func(ctx context.Context, msg *Message, err error) error

// ChainMiddleware combines multiple middleware into one
func ChainMiddleware(middleware ...Middleware) Middleware {
	return func(handler MessageHandler) MessageHandler {
		for i := len(middleware) - 1; i >= 0; i-- {
			handler = middleware[i](handler)
		}
		return handler
	}
}

// LoggingMiddleware logs message processing
func LoggingMiddleware(logger *logger.Logger) Middleware {
	return func(next MessageHandler) MessageHandler {
		return func(ctx context.Context, msg *Message) error {
			logger.Info(fmt.Sprintf("Processing message ID: %s, Topic: %s", msg.ID, msg.Topic))
			err := next(ctx, msg)
			if err != nil {
				logger.Error(fmt.Sprintf("Error processing message ID: %s, error: %v", msg.ID, err))
			}
			return err
		}
	}
}

// RecoveryMiddleware recovers from panics
func RecoveryMiddleware(logger *logger.Logger) Middleware {
	return func(next MessageHandler) MessageHandler {
		return func(ctx context.Context, msg *Message) (err error) {
			defer func() {
				if r := recover(); r != nil {
					logger.Error(fmt.Sprintf("Panic recovered: %v, message ID: %s", r, msg.ID))
					err = fmt.Errorf("panic recovered: %v", r)
				}
			}()
			return next(ctx, msg)
		}
	}
}

// TimeoutMiddleware adds timeout to message processing
func TimeoutMiddleware(timeout time.Duration) Middleware {
	return func(next MessageHandler) MessageHandler {
		return func(ctx context.Context, msg *Message) error {
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			done := make(chan error, 1)
			go func() {
				done <- next(ctx, msg)
			}()

			select {
			case err := <-done:
				return err
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}

// RetryMiddleware implements retry logic with exponential backoff
func RetryMiddleware(policy RetryPolicy) Middleware {
	return func(next MessageHandler) MessageHandler {
		return func(ctx context.Context, msg *Message) error {
			var err error
			delay := policy.InitialDelay

			for attempt := 0; attempt <= policy.MaxRetries; attempt++ {
				err = next(ctx, msg)
				if err == nil {
					return nil
				}

				if attempt < policy.MaxRetries {
					if policy.EnableBackoff {
						time.Sleep(delay)
						delay = time.Duration(float64(delay) * policy.Multiplier)
						if delay > policy.MaxDelay {
							delay = policy.MaxDelay
						}
					} else {
						time.Sleep(policy.InitialDelay)
					}
				}
			}

			return fmt.Errorf("max retries exceeded: %w", err)
		}
	}
}

// MetricsMiddleware tracks message processing metrics
func MetricsMiddleware(metrics MetricsCollector) Middleware {
	return func(next MessageHandler) MessageHandler {
		return func(ctx context.Context, msg *Message) error {
			start := time.Now()
			metrics.IncMessagesReceived(msg.Topic)

			err := next(ctx, msg)

			duration := time.Since(start)
			metrics.ObserveProcessingDuration(msg.Topic, duration)

			if err != nil {
				metrics.IncMessagesFailed(msg.Topic)
			} else {
				metrics.IncMessagesProcessed(msg.Topic)
			}

			return err
		}
	}
}
