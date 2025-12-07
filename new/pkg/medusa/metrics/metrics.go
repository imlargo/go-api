package metrics

import "time"

type MetricsService interface {
	RecordHTTPRequest(method, path, status string)
	RecordHTTPDuration(method, path, status string, duration time.Duration)
}
