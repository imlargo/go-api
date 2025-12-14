package pubsub

import "time"

// MetricsCollector defines the interface for collecting metrics
type MetricsCollector interface {
	IncMessagesPublished(topic string)
	IncMessagesReceived(topic string)
	IncMessagesProcessed(topic string)
	IncMessagesFailed(topic string)
	ObserveProcessingDuration(topic string, duration time.Duration)
}

// NoOpMetrics is a metrics collector that does nothing
type NoOpMetrics struct{}

func (n *NoOpMetrics) IncMessagesPublished(topic string)                              {}
func (n *NoOpMetrics) IncMessagesReceived(topic string)                               {}
func (n *NoOpMetrics) IncMessagesProcessed(topic string)                              {}
func (n *NoOpMetrics) IncMessagesFailed(topic string)                                 {}
func (n *NoOpMetrics) ObserveProcessingDuration(topic string, duration time.Duration) {}
