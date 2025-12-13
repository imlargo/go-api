package taskqueue

import (
	"fmt"
	"strings"
	"time"

	"github.com/nicolailuther/butter/internal/enums"
)

// Config holds configuration for the task queue system
type Config struct {
	// Worker configuration
	WorkerCount int           // Number of concurrent workers
	TaskTimeout time.Duration // Maximum time for a task to complete

	// Retry configuration
	MaxRetries        int           // Maximum number of retry attempts
	InitialRetryDelay time.Duration // Initial delay before retry
	MaxRetryDelay     time.Duration // Maximum delay between retries
	BackoffFactor     float64       // Exponential backoff multiplier

	// Heartbeat configuration
	HeartbeatInterval time.Duration // How often workers send heartbeat
	OrphanTimeout     time.Duration // Time before task is considered orphaned

	// Priority configuration
	PriorityHighThreshold   enums.TaskPriority // Priority >= this goes to high queue
	PriorityNormalThreshold enums.TaskPriority // Priority >= this goes to normal queue

	// Queue configuration
	DLQAlertThreshold int // Alert when DLQ has this many tasks

	// Redis configuration
	RedisKeyPrefix string // Prefix for all Redis keys. Must not be empty or contain invalid characters (colons, spaces).
}

// DefaultConfig returns default configuration
func DefaultConfig() Config {
	return Config{
		WorkerCount:             7,
		TaskTimeout:             30 * time.Minute,
		MaxRetries:              3,
		InitialRetryDelay:       30 * time.Second,
		MaxRetryDelay:           30 * time.Minute,
		BackoffFactor:           2.0,
		HeartbeatInterval:       30 * time.Second,
		OrphanTimeout:           30 * time.Minute,
		PriorityHighThreshold:   10,
		PriorityNormalThreshold: 5,
		DLQAlertThreshold:       10,
		RedisKeyPrefix:          "repurposer",
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.RedisKeyPrefix == "" {
		return fmt.Errorf("RedisKeyPrefix cannot be empty")
	}
	// Check for invalid characters that could cause Redis key issues
	if strings.ContainsAny(c.RedisKeyPrefix, ": \t\n\r") {
		return fmt.Errorf("RedisKeyPrefix contains invalid characters (colons, spaces, or whitespace)")
	}
	return nil
}

// QueueKeys returns the Redis keys for queues
type QueueKeys struct {
	HighPriority   string
	NormalPriority string
	LowPriority    string
	DLQ            string
	Events         string
}

func (c *Config) GetQueueKeys() QueueKeys {
	prefix := c.RedisKeyPrefix
	return QueueKeys{
		HighPriority:   prefix + ":queue:priority:high",
		NormalPriority: prefix + ":queue:priority:normal",
		LowPriority:    prefix + ":queue:priority:low",
		DLQ:            prefix + ":dlq",
		Events:         prefix + ":events",
	}
}

// GetProcessingKey returns the Redis key for a worker's processing set
func (c *Config) GetProcessingKey(workerID string) string {
	return c.RedisKeyPrefix + ":processing:" + workerID
}

// GetTaskLockKey returns the Redis key for a task's lock
func (c *Config) GetTaskLockKey(taskID string) string {
	return c.RedisKeyPrefix + ":task:" + taskID + ":lock"
}

// GetRetryScheduleKey returns the Redis key for scheduled retries
func (c *Config) GetRetryScheduleKey() string {
	return c.RedisKeyPrefix + ":retry:scheduled"
}
