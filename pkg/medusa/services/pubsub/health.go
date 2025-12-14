package pubsub

import "context"

// HealthChecker provides health check capabilities
type HealthChecker interface {
	// HealthCheck returns the current health status
	HealthCheck(ctx context.Context) error
}
