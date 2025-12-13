package taskqueue

import (
	"context"
	"time"

	"github.com/nicolailuther/butter/internal/enums"
)

// GetStats returns queue statistics
func (tm *taskManager) GetStats(ctx context.Context) (*QueueStats, error) {
	stats := &QueueStats{}

	keys := tm.config.GetQueueKeys()

	// Get queue sizes
	highSize, _ := tm.redis.LLen(ctx, keys.HighPriority).Result()
	normalSize, _ := tm.redis.LLen(ctx, keys.NormalPriority).Result()
	lowSize, _ := tm.redis.LLen(ctx, keys.LowPriority).Result()
	dlqSize, _ := tm.redis.LLen(ctx, keys.DLQ).Result()

	stats.TotalQueued = int(highSize + normalSize + lowSize)
	stats.TotalDLQ = int(dlqSize)

	// Get counts by status
	pendingCount, _ := tm.taskRepo.CountByStatus(enums.TaskStatusPending)
	processingCount, _ := tm.taskRepo.CountByStatus(enums.TaskStatusProcessing)

	stats.TotalPending = int(pendingCount)
	stats.TotalProcessing = int(processingCount)

	// Get historical counts (last 24 hours)
	since := time.Now().Add(-24 * time.Hour)
	completedCount, _ := tm.taskRepo.CountCompletedSince(since)
	failedCount, _ := tm.taskRepo.CountFailedSince(since)

	stats.TotalCompleted = completedCount
	stats.TotalFailed = failedCount

	// Get average times
	avgProcessing, _ := tm.taskRepo.GetAverageProcessingTime(since)
	avgQueue, _ := tm.taskRepo.GetAverageQueueTime(since)

	stats.AvgProcessingTime = time.Duration(avgProcessing) * time.Millisecond
	stats.AvgQueueTime = time.Duration(avgQueue) * time.Millisecond

	// Calculate tasks per hour
	if completedCount > 0 {
		stats.TasksPerHour = float64(completedCount) / 24.0
	}

	// Get worker stats
	tm.workersMux.RLock()
	stats.ActiveWorkers = len(tm.workers)
	stats.IdleWorkers = 0 // This would need more sophisticated tracking
	tm.workersMux.RUnlock()

	return stats, nil
}

// GetWorkerStats returns statistics for all workers
func (tm *taskManager) GetWorkerStats(ctx context.Context) ([]*WorkerStats, error) {
	tm.workersMux.RLock()
	defer tm.workersMux.RUnlock()

	stats := make([]*WorkerStats, len(tm.workers))
	for i, worker := range tm.workers {
		stats[i] = &WorkerStats{
			WorkerID: worker.id,
			IsActive: true, // Worker is active if it exists
			// Additional stats would require more tracking
		}
	}

	return stats, nil
}
