package taskqueue

import (
	"encoding/json"
	"time"

	"github.com/nicolailuther/butter/internal/enums"
)

// TaskInfo represents information about a task
type TaskInfo struct {
	ID             string                 `json:"id"`
	TaskID         string                 `json:"task_id"`
	Status         enums.TaskStatus       `json:"status"`
	FileID         uint                   `json:"file_id"`
	AccountID      *uint                  `json:"account_id,omitempty"`
	ContentID      *uint                  `json:"content_id,omitempty"`
	Priority       enums.TaskPriority     `json:"priority"`
	Attempts       int                    `json:"attempts"`
	MaxRetries     int                    `json:"max_retries"`
	CreatedAt      time.Time              `json:"created_at"`
	QueuedAt       *time.Time             `json:"queued_at,omitempty"`
	StartedAt      *time.Time             `json:"started_at,omitempty"`
	CompletedAt    *time.Time             `json:"completed_at,omitempty"`
	FailedAt       *time.Time             `json:"failed_at,omitempty"`
	ProcessingTime *time.Duration         `json:"processing_time,omitempty"`
	QueueTime      *time.Duration         `json:"queue_time,omitempty"`
	ErrorMessage   string                 `json:"error_message,omitempty"`
	WorkerID       string                 `json:"worker_id,omitempty"`
	RequestData    map[string]interface{} `json:"request_data,omitempty"`
	ResultData     map[string]interface{} `json:"result_data,omitempty"`
}

// QueueStats represents statistics about the task queue
type QueueStats struct {
	TotalPending      int           `json:"total_pending"`
	TotalQueued       int           `json:"total_queued"`
	TotalProcessing   int           `json:"total_processing"`
	TotalCompleted    int64         `json:"total_completed"`
	TotalFailed       int64         `json:"total_failed"`
	TotalDLQ          int           `json:"total_dlq"`
	ActiveWorkers     int           `json:"active_workers"`
	IdleWorkers       int           `json:"idle_workers"`
	AvgProcessingTime time.Duration `json:"avg_processing_time"`
	AvgQueueTime      time.Duration `json:"avg_queue_time"`
	TasksPerHour      float64       `json:"tasks_per_hour"`
}

// WorkerStats represents statistics about a worker
type WorkerStats struct {
	WorkerID       string    `json:"worker_id"`
	IsActive       bool      `json:"is_active"`
	CurrentTaskID  string    `json:"current_task_id,omitempty"`
	TasksProcessed int       `json:"tasks_processed"`
	LastHeartbeat  time.Time `json:"last_heartbeat,omitempty"`
}

// TaskFilter represents filter criteria for querying tasks
type TaskFilter struct {
	AccountID *uint             `json:"account_id,omitempty"`
	Status    *enums.TaskStatus `json:"status,omitempty"`
	Since     *time.Time        `json:"since,omitempty"`
	Limit     int               `json:"limit,omitempty"`
	Offset    int               `json:"offset,omitempty"`
}

// TaskEvent represents an event published about a task
type TaskEvent struct {
	EventType string                 `json:"event_type"`
	TaskID    string                 `json:"task_id"`
	AccountID *uint                  `json:"account_id,omitempty"`
	Status    enums.TaskStatus       `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// MarshalBinary implements encoding.BinaryMarshaler for Redis publishing
func (e TaskEvent) MarshalBinary() ([]byte, error) {
	return json.Marshal(e)
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler for Redis subscribing
func (e *TaskEvent) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, e)
}

// Event type constants
const (
	EventTaskQueued    = "task_queued"
	EventTaskStarted   = "task_started"
	EventTaskCompleted = "task_completed"
	EventTaskFailed    = "task_failed"
	EventTaskRetry     = "task_retry"
	EventTaskCanceled  = "task_canceled"
	EventTaskDLQ       = "task_dlq"
)
