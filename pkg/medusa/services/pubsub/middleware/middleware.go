// Package middleware provides common middleware implementations
package middleware

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/imlargo/go-api/pkg/medusa/core/logger"
	"github.com/imlargo/go-api/pkg/medusa/services/pubsub"
)

// LoggingMiddleware logs message processing
func LoggingMiddleware(logger *logger.Logger) pubsub.Middleware {
	return func(next pubsub.MessageHandler) pubsub.MessageHandler {
		return func(ctx context.Context, msg *pubsub.Message) error {
			start := time.Now()
			logger.Sugar().Info("Processing message", map[string]interface{}{
				"messageId":     msg.ID,
				"topic":         msg.Topic,
				"correlationId": msg.CorrelationID,
			})

			err := next(ctx, msg)

			fields := map[string]interface{}{
				"messageId": msg.ID,
				"topic":     msg.Topic,
				"duration":  time.Since(start).String(),
			}

			if err != nil {
				fields["error"] = err.Error()
				logger.Sugar().Error("Failed to process message", fields)
			} else {
				logger.Sugar().Info("Message processed successfully", fields)
			}

			return err
		}
	}
}

// RecoveryMiddleware recovers from panics
func RecoveryMiddleware(logger *logger.Logger) pubsub.Middleware {
	return func(next pubsub.MessageHandler) pubsub.MessageHandler {
		return func(ctx context.Context, msg *pubsub.Message) (err error) {
			defer func() {
				if r := recover(); r != nil {
					logger.Sugar().Error("Panic recovered", map[string]interface{}{
						"messageId": msg.ID,
						"panic":     fmt.Sprintf("%v", r),
					})
					err = fmt.Errorf("panic recovered: %v", r)
				}
			}()

			return next(ctx, msg)
		}
	}
}

// TimeoutMiddleware enforces message processing timeout
func TimeoutMiddleware(timeout time.Duration) pubsub.Middleware {
	return func(next pubsub.MessageHandler) pubsub.MessageHandler {
		return func(ctx context.Context, msg *pubsub.Message) error {
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			done := make(chan error, 1)
			go func() {
				done <- next(ctx, msg)
			}()

			select {
			case <-ctx.Done():
				return fmt.Errorf("message processing timeout after %v", timeout)
			case err := <-done:
				return err
			}
		}
	}
}

// ValidationMiddleware validates messages
func ValidationMiddleware(validator MessageValidator) pubsub.Middleware {
	return func(next pubsub.MessageHandler) pubsub.MessageHandler {
		return func(ctx context.Context, msg *pubsub.Message) error {
			if err := validator.Validate(msg); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}
			return next(ctx, msg)
		}
	}
}

// MessageValidator validates messages
type MessageValidator interface {
	Validate(msg *pubsub.Message) error
}

// DeduplicationMiddleware prevents processing duplicate messages
func DeduplicationMiddleware(store DeduplicationStore, ttl time.Duration) pubsub.Middleware {
	return func(next pubsub.MessageHandler) pubsub.MessageHandler {
		return func(ctx context.Context, msg *pubsub.Message) error {
			if msg.ID == "" {
				return errors.New("message ID is required for deduplication")
			}

			// Check if message was already processed
			if store.Exists(msg.ID) {
				return nil // Skip duplicate
			}

			// Process message
			err := next(ctx, msg)
			if err == nil {
				// Mark as processed
				store.Add(msg.ID, ttl)
			}

			return err
		}
	}
}

// DeduplicationStore stores processed message IDs
type DeduplicationStore interface {
	Add(messageID string, ttl time.Duration) error
	Exists(messageID string) bool
	Remove(messageID string) error
}

// InMemoryDeduplicationStore is a simple in-memory implementation
type InMemoryDeduplicationStore struct {
	store map[string]time.Time
	mu    sync.RWMutex
}

func NewInMemoryDeduplicationStore() *InMemoryDeduplicationStore {
	store := &InMemoryDeduplicationStore{
		store: make(map[string]time.Time),
	}
	go store.cleanup()
	return store
}

func (s *InMemoryDeduplicationStore) Add(messageID string, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store[messageID] = time.Now().Add(ttl)
	return nil
}

func (s *InMemoryDeduplicationStore) Exists(messageID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	expiry, exists := s.store[messageID]
	if !exists {
		return false
	}
	return time.Now().Before(expiry)
}

func (s *InMemoryDeduplicationStore) Remove(messageID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.store, messageID)
	return nil
}

func (s *InMemoryDeduplicationStore) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for id, expiry := range s.store {
			if now.After(expiry) {
				delete(s.store, id)
			}
		}
		s.mu.Unlock()
	}
}

// MetricsMiddleware collects metrics
func MetricsMiddleware(metrics pubsub.Metrics) pubsub.Middleware {
	return func(next pubsub.MessageHandler) pubsub.MessageHandler {
		return func(ctx context.Context, msg *pubsub.Message) error {
			start := time.Now()
			err := next(ctx, msg)
			duration := time.Since(start)

			if err != nil {
				metrics.IncrementFailed(msg.Topic)
			} else {
				metrics.IncrementConsumed(msg.Topic)
				metrics.RecordProcessingLatency(msg.Topic, duration)
			}

			return err
		}
	}
}

// TracingMiddleware adds distributed tracing
func TracingMiddleware(tracer Tracer) pubsub.Middleware {
	return func(next pubsub.MessageHandler) pubsub.MessageHandler {
		return func(ctx context.Context, msg *pubsub.Message) error {
			span := tracer.StartSpan(ctx, "message.process", map[string]interface{}{
				"message.id":    msg.ID,
				"message.topic": msg.Topic,
			})
			defer span.Finish()

			ctx = tracer.ContextWithSpan(ctx, span)

			err := next(ctx, msg)
			if err != nil {
				span.SetError(err)
			}

			return err
		}
	}
}

// Tracer interface for distributed tracing
type Tracer interface {
	StartSpan(ctx context.Context, operation string, tags map[string]interface{}) Span
	ContextWithSpan(ctx context.Context, span Span) context.Context
}

// Span represents a tracing span
type Span interface {
	SetError(err error)
	Finish()
}

// CircuitBreakerMiddleware implements circuit breaker pattern
func CircuitBreakerMiddleware(config *pubsub.CircuitBreakerConfig) pubsub.Middleware {
	cb := NewCircuitBreaker(config)

	return func(next pubsub.MessageHandler) pubsub.MessageHandler {
		return func(ctx context.Context, msg *pubsub.Message) error {
			return cb.Execute(func() error {
				return next(ctx, msg)
			})
		}
	}
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	config       *pubsub.CircuitBreakerConfig
	failures     int
	lastFailure  time.Time
	state        CircuitBreakerState
	halfOpenReqs int
	mu           sync.RWMutex
}

type CircuitBreakerState int

const (
	StateClosed CircuitBreakerState = iota
	StateOpen
	StateHalfOpen
)

func NewCircuitBreaker(config *pubsub.CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		config: config,
		state:  StateClosed,
	}
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == StateOpen {
		if time.Since(cb.lastFailure) > cb.config.ResetTimeout {
			cb.state = StateHalfOpen
			cb.halfOpenReqs = 0
		} else {
			return errors.New("circuit breaker is open")
		}
	}

	err := fn()

	if err != nil {
		cb.failures++
		cb.lastFailure = time.Now()

		if cb.state == StateHalfOpen {
			cb.state = StateOpen
		} else if cb.failures >= cb.config.MaxFailures {
			cb.state = StateOpen
		}

		return err
	}

	// Success
	if cb.state == StateHalfOpen {
		cb.halfOpenReqs++
		if cb.halfOpenReqs >= cb.config.HalfOpenRequests {
			cb.state = StateClosed
			cb.failures = 0
		}
	} else if cb.state == StateClosed {
		cb.failures = 0
	}

	return nil
}

// RateLimitMiddleware limits message processing rate
func RateLimitMiddleware(limiter RateLimiter) pubsub.Middleware {
	return func(next pubsub.MessageHandler) pubsub.MessageHandler {
		return func(ctx context.Context, msg *pubsub.Message) error {
			if !limiter.Allow() {
				return errors.New("rate limit exceeded")
			}
			return next(ctx, msg)
		}
	}
}

// RateLimiter interface
type RateLimiter interface {
	Allow() bool
}

// TokenBucketRateLimiter implements token bucket algorithm
type TokenBucketRateLimiter struct {
	rate     float64
	capacity int
	tokens   float64
	lastTime time.Time
	mu       sync.Mutex
}

func NewTokenBucketRateLimiter(rate float64, capacity int) *TokenBucketRateLimiter {
	return &TokenBucketRateLimiter{
		rate:     rate,
		capacity: capacity,
		tokens:   float64(capacity),
		lastTime: time.Now(),
	}
}

func (rl *TokenBucketRateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastTime).Seconds()
	rl.tokens = min(float64(rl.capacity), rl.tokens+elapsed*rl.rate)
	rl.lastTime = now

	if rl.tokens >= 1 {
		rl.tokens--
		return true
	}

	return false
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// TransformMiddleware transforms messages
func TransformMiddleware(transformer MessageTransformer) pubsub.Middleware {
	return func(next pubsub.MessageHandler) pubsub.MessageHandler {
		return func(ctx context.Context, msg *pubsub.Message) error {
			transformed, err := transformer.Transform(msg)
			if err != nil {
				return fmt.Errorf("message transformation failed: %w", err)
			}
			return next(ctx, transformed)
		}
	}
}

// MessageTransformer transforms messages
type MessageTransformer interface {
	Transform(msg *pubsub.Message) (*pubsub.Message, error)
}

// FilterMiddleware filters messages
func FilterMiddleware(filter MessageFilter) pubsub.Middleware {
	return func(next pubsub.MessageHandler) pubsub.MessageHandler {
		return func(ctx context.Context, msg *pubsub.Message) error {
			if !filter.ShouldProcess(msg) {
				return nil // Skip message
			}
			return next(ctx, msg)
		}
	}
}

// MessageFilter determines if a message should be processed
type MessageFilter interface {
	ShouldProcess(msg *pubsub.Message) bool
}

// EnrichmentMiddleware enriches messages with additional data
func EnrichmentMiddleware(enricher MessageEnricher) pubsub.Middleware {
	return func(next pubsub.MessageHandler) pubsub.MessageHandler {
		return func(ctx context.Context, msg *pubsub.Message) error {
			if err := enricher.Enrich(msg); err != nil {
				return fmt.Errorf("message enrichment failed: %w", err)
			}
			return next(ctx, msg)
		}
	}
}

// MessageEnricher enriches messages
type MessageEnricher interface {
	Enrich(msg *pubsub.Message) error
}
