package observability

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/imlargo/go-api/pkg/medusa/core/logger"
	"github.com/imlargo/go-api/pkg/medusa/services/pubsub"
)

type Span interface {
	SetError(err error)
	Finish()
}

// NoopMetrics is a no-op implementation of Metrics
type NoopMetrics struct{}

func (m *NoopMetrics) IncrementPublished(topic string)                              {}
func (m *NoopMetrics) IncrementConsumed(topic string)                               {}
func (m *NoopMetrics) IncrementFailed(topic string)                                 {}
func (m *NoopMetrics) IncrementRetry(topic string)                                  {}
func (m *NoopMetrics) RecordPublishLatency(topic string, duration time.Duration)    {}
func (m *NoopMetrics) RecordProcessingLatency(topic string, duration time.Duration) {}

// PrometheusMetrics implements Metrics for Prometheus
type PrometheusMetrics struct {
	publishedCounter  map[string]int64
	consumedCounter   map[string]int64
	failedCounter     map[string]int64
	retryCounter      map[string]int64
	publishLatency    map[string][]time.Duration
	processingLatency map[string][]time.Duration
	mu                sync.RWMutex
}

func NewPrometheusMetrics() *PrometheusMetrics {
	return &PrometheusMetrics{
		publishedCounter:  make(map[string]int64),
		consumedCounter:   make(map[string]int64),
		failedCounter:     make(map[string]int64),
		retryCounter:      make(map[string]int64),
		publishLatency:    make(map[string][]time.Duration),
		processingLatency: make(map[string][]time.Duration),
	}
}

func (m *PrometheusMetrics) IncrementPublished(topic string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.publishedCounter[topic]++
}

func (m *PrometheusMetrics) IncrementConsumed(topic string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.consumedCounter[topic]++
}

func (m *PrometheusMetrics) IncrementFailed(topic string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failedCounter[topic]++
}

func (m *PrometheusMetrics) IncrementRetry(topic string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.retryCounter[topic]++
}

func (m *PrometheusMetrics) RecordPublishLatency(topic string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.publishLatency[topic] = append(m.publishLatency[topic], duration)
}

func (m *PrometheusMetrics) RecordProcessingLatency(topic string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.processingLatency[topic] = append(m.processingLatency[topic], duration)
}

func (m *PrometheusMetrics) GetMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"published_messages": m.publishedCounter,
		"consumed_messages":  m.consumedCounter,
		"failed_messages":    m.failedCounter,
		"retry_attempts":     m.retryCounter,
		"publish_latency":    m.calculateLatencyStats(m.publishLatency),
		"processing_latency": m.calculateLatencyStats(m.processingLatency),
	}
}

func (m *PrometheusMetrics) calculateLatencyStats(latencies map[string][]time.Duration) map[string]interface{} {
	stats := make(map[string]interface{})

	for topic, durations := range latencies {
		if len(durations) == 0 {
			continue
		}

		var sum time.Duration
		for _, d := range durations {
			sum += d
		}

		avg := sum / time.Duration(len(durations))
		stats[topic] = map[string]interface{}{
			"count":   len(durations),
			"average": avg.String(),
		}
	}

	return stats
}

// HealthChecker monitors broker health
type HealthChecker struct {
	broker   pubsub.MessageBroker
	interval time.Duration
	ctx      context.Context
	cancel   context.CancelFunc
	status   *HealthStatus
	mu       sync.RWMutex
}

type HealthStatus struct {
	Healthy      bool
	LastCheck    time.Time
	LastError    error
	CheckCount   int
	FailureCount int
}

func NewHealthChecker(broker pubsub.MessageBroker, interval time.Duration) *HealthChecker {
	ctx, cancel := context.WithCancel(context.Background())
	return &HealthChecker{
		broker:   broker,
		interval: interval,
		ctx:      ctx,
		cancel:   cancel,
		status: &HealthStatus{
			Healthy: true,
		},
	}
}

func (hc *HealthChecker) Start() {
	go hc.monitor()
}

func (hc *HealthChecker) Stop() {
	hc.cancel()
}

func (hc *HealthChecker) GetStatus() *HealthStatus {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return hc.status
}

func (hc *HealthChecker) monitor() {
	ticker := time.NewTicker(hc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-hc.ctx.Done():
			return
		case <-ticker.C:
			hc.check()
		}
	}
}

func (hc *HealthChecker) check() {
	ctx, cancel := context.WithTimeout(hc.ctx, 5*time.Second)
	defer cancel()

	err := hc.broker.Health(ctx)

	hc.mu.Lock()
	defer hc.mu.Unlock()

	hc.status.CheckCount++
	hc.status.LastCheck = time.Now()

	if err != nil {
		hc.status.Healthy = false
		hc.status.LastError = err
		hc.status.FailureCount++
	} else {
		hc.status.Healthy = true
		hc.status.LastError = nil
	}
}

// OpenTelemetryTracer implements distributed tracing
type OpenTelemetryTracer struct {
	serviceName string
}

func NewOpenTelemetryTracer(serviceName string) *OpenTelemetryTracer {
	return &OpenTelemetryTracer{
		serviceName: serviceName,
	}
}

func (t *OpenTelemetryTracer) StartSpan(ctx context.Context, operation string, tags map[string]interface{}) Span {
	span := &openTelemetrySpan{
		operation: operation,
		startTime: time.Now(),
		tags:      tags,
	}
	return span
}

func (t *OpenTelemetryTracer) ContextWithSpan(ctx context.Context, span Span) context.Context {
	return context.WithValue(ctx, "span", span)
}

type openTelemetrySpan struct {
	operation string
	startTime time.Time
	endTime   time.Time
	tags      map[string]interface{}
	error     error
}

func (s *openTelemetrySpan) SetError(err error) {
	s.error = err
	if s.tags == nil {
		s.tags = make(map[string]interface{})
	}
	s.tags["error"] = true
	s.tags["error.message"] = err.Error()
}

func (s *openTelemetrySpan) Finish() {
	s.endTime = time.Now()
	duration := s.endTime.Sub(s.startTime)

	// In a real implementation, this would export to OpenTelemetry
	log.Printf("[TRACE] operation=%s duration=%v tags=%v",
		s.operation, duration, s.tags)
}

// MessageAuditor logs all message operations for compliance
type MessageAuditor struct {
	store  AuditStore
	logger *logger.Logger
}

type AuditEntry struct {
	ID        string
	Timestamp time.Time
	Operation string
	Topic     string
	MessageID string
	UserID    string
	Success   bool
	Error     string
	Metadata  map[string]interface{}
}

type AuditStore interface {
	Store(entry *AuditEntry) error
	Query(filter map[string]interface{}) ([]*AuditEntry, error)
}

func NewMessageAuditor(store AuditStore, logger *logger.Logger) *MessageAuditor {
	return &MessageAuditor{
		store:  store,
		logger: logger,
	}
}

func (ma *MessageAuditor) AuditPublish(ctx context.Context, msg *pubsub.Message, success bool, err error) {
	entry := &AuditEntry{
		ID:        fmt.Sprintf("audit-%d", time.Now().UnixNano()),
		Timestamp: time.Now(),
		Operation: "publish",
		Topic:     msg.Topic,
		MessageID: msg.ID,
		Success:   success,
	}

	if err != nil {
		entry.Error = err.Error()
	}

	if err := ma.store.Store(entry); err != nil {
		ma.logger.Sugar().Error("Failed to store audit entry", map[string]interface{}{
			"error": err.Error(),
		})
	}
}

func (ma *MessageAuditor) AuditConsume(ctx context.Context, msg *pubsub.Message, success bool, err error) {
	entry := &AuditEntry{
		ID:        fmt.Sprintf("audit-%d", time.Now().UnixNano()),
		Timestamp: time.Now(),
		Operation: "consume",
		Topic:     msg.Topic,
		MessageID: msg.ID,
		Success:   success,
	}

	if err != nil {
		entry.Error = err.Error()
	}

	if err := ma.store.Store(entry); err != nil {
		ma.logger.Sugar().Error("Failed to store audit entry", map[string]interface{}{
			"error": err.Error(),
		})
	}
}

// InMemoryAuditStore stores audit entries in memory
type InMemoryAuditStore struct {
	entries []*AuditEntry
	mu      sync.RWMutex
}

func NewInMemoryAuditStore() *InMemoryAuditStore {
	return &InMemoryAuditStore{
		entries: make([]*AuditEntry, 0),
	}
}

func (s *InMemoryAuditStore) Store(entry *AuditEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries = append(s.entries, entry)
	return nil
}

func (s *InMemoryAuditStore) Query(filter map[string]interface{}) ([]*AuditEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]*AuditEntry, 0)
	for _, entry := range s.entries {
		if s.matchesFilter(entry, filter) {
			results = append(results, entry)
		}
	}
	return results, nil
}

func (s *InMemoryAuditStore) matchesFilter(entry *AuditEntry, filter map[string]interface{}) bool {
	for k, v := range filter {
		switch k {
		case "operation":
			if entry.Operation != v.(string) {
				return false
			}
		case "topic":
			if entry.Topic != v.(string) {
				return false
			}
		case "messageId":
			if entry.MessageID != v.(string) {
				return false
			}
		}
	}
	return true
}

// PerformanceMonitor tracks performance metrics
type PerformanceMonitor struct {
	metrics         *PrometheusMetrics
	alertThresholds map[string]time.Duration
	alertHandlers   []AlertHandler
	mu              sync.RWMutex
}

type AlertHandler func(alert *Alert)

type Alert struct {
	Type      string
	Severity  string
	Message   string
	Metric    string
	Value     interface{}
	Threshold interface{}
	Timestamp time.Time
}

func NewPerformanceMonitor(metrics *PrometheusMetrics) *PerformanceMonitor {
	return &PerformanceMonitor{
		metrics:         metrics,
		alertThresholds: make(map[string]time.Duration),
		alertHandlers:   make([]AlertHandler, 0),
	}
}

func (pm *PerformanceMonitor) SetThreshold(metric string, threshold time.Duration) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.alertThresholds[metric] = threshold
}

func (pm *PerformanceMonitor) AddAlertHandler(handler AlertHandler) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.alertHandlers = append(pm.alertHandlers, handler)
}

func (pm *PerformanceMonitor) CheckThresholds() {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// This would be implemented based on actual metric collection
	// and comparison with thresholds
}

func (pm *PerformanceMonitor) triggerAlert(alert *Alert) {
	for _, handler := range pm.alertHandlers {
		handler(alert)
	}
}
