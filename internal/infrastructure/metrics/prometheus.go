package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type PrometheusMetrics struct {
	httpRequestsTotal   *prometheus.CounterVec
	httpRequestDuration *prometheus.HistogramVec
}

func NewPrometheusMetrics() MetricsService {
	return &PrometheusMetrics{
		httpRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),

		httpRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path", "status"},
		),
	}
}

func (p *PrometheusMetrics) RecordHTTPRequest(method, path, status string) {
	p.httpRequestsTotal.WithLabelValues(method, path, status).Inc()
}

func (p *PrometheusMetrics) RecordHTTPDuration(method, path, status string, duration time.Duration) {
	p.httpRequestDuration.WithLabelValues(method, path, status).Observe(duration.Seconds())
}
