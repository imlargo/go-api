package taskqueue

import (
	"context"
	"fmt"
	"time"

	"github.com/nicolailuther/butter/internal/dto"
	"github.com/nicolailuther/butter/internal/enums"
	"github.com/nicolailuther/butter/internal/models"
	"github.com/redis/go-redis/v9"
)

// worker processes tasks from the queue
type worker struct {
	id      string
	manager *taskManager
}

func newWorker(id string, manager *taskManager) *worker {
	return &worker{
		id:      id,
		manager: manager,
	}
}

func (w *worker) run() {
	defer w.manager.wg.Done()

	w.manager.logger.Infow("Worker started", "worker_id", w.id)

	backoff := 100 * time.Millisecond // Start with 100ms
	maxBackoff := 5 * time.Second     // Cap at 5 seconds

	for {
		select {
		case <-w.manager.ctx.Done():
			w.manager.logger.Infow("Worker stopping", "worker_id", w.id)
			return
		default:
			task := w.fetchTask()
			if task == nil {
				// Use exponential backoff when no tasks available
				time.Sleep(backoff)
				backoff = backoff * 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
				continue
			}

			// Reset backoff when we successfully get a task
			backoff = 100 * time.Millisecond
			w.processTask(task)
		}
	}
}

func (w *worker) fetchTask() *taskFetch {
	ctx := w.manager.ctx
	keys := w.manager.config.GetQueueKeys()

	// Try high priority first, then normal, then low
	queues := []string{
		keys.HighPriority,
		keys.NormalPriority,
		keys.LowPriority,
	}

	for _, queueKey := range queues {
		taskID, err := w.manager.redis.RPop(ctx, queueKey).Result()
		if err == redis.Nil {
			continue // Queue empty, try next
		}
		if err != nil {
			w.manager.logger.Warnw("Error fetching from queue", "queue", queueKey, "error", err)
			continue
		}

		// Fetch task from database
		task, err := w.manager.taskRepo.GetByTaskID(taskID)
		if err != nil {
			w.manager.logger.Warnw("Error fetching task", "task_id", taskID, "error", err)
			continue
		}

		// Acquire lock
		lockKey := w.manager.config.GetTaskLockKey(taskID)
		locked, err := w.manager.redis.SetNX(ctx, lockKey, w.id, w.manager.config.TaskTimeout).Result()
		if err != nil || !locked {
			w.manager.logger.Warnw("Failed to acquire lock for task", "task_id", taskID)
			// Re-queue task
			if err := w.manager.redis.LPush(ctx, queueKey, taskID).Err(); err != nil {
				w.manager.logger.Errorw("Failed to re-queue task after lock acquisition failure", "task_id", taskID, "queue", queueKey, "error", err)
			}
			continue
		}

		return &taskFetch{
			task:     task,
			queueKey: queueKey,
			lockKey:  lockKey,
		}
	}

	return nil
}

type taskFetch struct {
	task     *models.RepurposerTask
	queueKey string
	lockKey  string
}

func (w *worker) processTask(fetch *taskFetch) {
	ctx := w.manager.ctx
	taskID := fetch.task.TaskID
	queuedAt := fetch.task.QueuedAt
	if queuedAt == nil {
		now := time.Now()
		queuedAt = &now
	}

	// Update status to processing
	if err := w.manager.taskRepo.UpdateStatus(taskID, enums.TaskStatusProcessing, ""); err != nil {
		w.manager.logger.Warnw("Failed to update task status to processing", "task_id", taskID, "error", err)
	}

	if err := w.manager.taskRepo.UpdateWorkerInfo(taskID, w.id); err != nil {
		w.manager.logger.Warnw("Failed to update worker info", "task_id", taskID, "error", err)
	}

	// Publish event
	if fetch.task.AccountID != nil {
		w.manager.publishEvent(ctx, EventTaskStarted, taskID, *fetch.task.AccountID, enums.TaskStatusProcessing, nil)
	}

	// Start heartbeat
	heartbeatDone := make(chan struct{})
	go w.sendHeartbeat(taskID, heartbeatDone)

	// Process task with timeout
	taskCtx, cancel := context.WithTimeout(ctx, w.manager.config.TaskTimeout)
	defer cancel()

	startTime := time.Now()
	result, err := w.executeTask(taskCtx, fetch.task)
	processingTime := time.Since(startTime)
	queueTime := startTime.Sub(*queuedAt)

	// Stop heartbeat
	close(heartbeatDone)

	// Release lock
	w.manager.redis.Del(ctx, fetch.lockKey)

	// Update metrics
	w.manager.taskRepo.UpdateMetrics(taskID, processingTime.Milliseconds(), queueTime.Milliseconds())

	// Handle result
	if err != nil {
		w.handleTaskFailure(fetch.task, err)
	} else {
		w.handleTaskSuccess(fetch.task, result)
	}
}

func (w *worker) executeTask(ctx context.Context, task *models.RepurposerTask) (*dto.ReporpuseContentResult, error) {
	// Unmarshal request data
	var request dto.ReporpuseVideo
	if err := task.UnmarshalRequestData(&request); err != nil {
		return nil, fmt.Errorf("failed to unmarshal request data: %w", err)
	}

	// Execute the task handler
	return w.manager.taskHandler(ctx, &request)
}

func (w *worker) sendHeartbeat(taskID string, done chan struct{}) {
	ticker := time.NewTicker(w.manager.config.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			if err := w.manager.taskRepo.UpdateHeartbeat(taskID); err != nil {
				w.manager.logger.Warnw("Failed to send heartbeat", "task_id", taskID, "error", err)
			}
		}
	}
}

func (w *worker) handleTaskSuccess(task *models.RepurposerTask, result *dto.ReporpuseContentResult) {
	ctx := w.manager.ctx
	taskID := task.TaskID

	// Get task from DB to update
	dbTask, err := w.manager.taskRepo.GetByTaskID(taskID)
	if err != nil {
		w.manager.logger.Errorw("Failed to get task for success handling", "task_id", taskID, "error", err)
		return
	}

	// Store result
	if err := dbTask.MarshalResultData(result); err != nil {
		w.manager.logger.Warnw("Failed to marshal result data", "task_id", taskID, "error", err)
	}

	dbTask.Status = enums.TaskStatusCompleted
	now := time.Now()
	dbTask.CompletedAt = &now

	if err := w.manager.taskRepo.Update(dbTask); err != nil {
		w.manager.logger.Errorw("Failed to update task as completed", "task_id", taskID, "error", err)
		return
	}

	// Publish event
	if task.AccountID != nil {
		w.manager.publishEvent(ctx, EventTaskCompleted, taskID, *task.AccountID, enums.TaskStatusCompleted, nil)
	}

	w.manager.logger.Infow("Task completed successfully",
		"task_id", taskID,
		"worker_id", w.id,
	)
}

func (w *worker) handleTaskFailure(task *models.RepurposerTask, taskErr error) {
	ctx := w.manager.ctx
	taskID := task.TaskID

	w.manager.logger.Warnw("Task failed",
		"task_id", taskID,
		"attempt", task.Attempts+1,
		"error", taskErr,
	)

	// Get task from DB to update
	dbTask, err := w.manager.taskRepo.GetByTaskID(taskID)
	if err != nil {
		w.manager.logger.Errorw("Failed to get task for failure handling", "task_id", taskID, "error", err)
		return
	}

	dbTask.Attempts++

	// Check if we should retry
	if dbTask.Attempts >= dbTask.MaxRetries {
		// Max retries reached, move to DLQ
		w.moveToDLQ(dbTask, taskErr)
		return
	}

	// Schedule retry with exponential backoff
	delay := w.manager.calculateRetryDelay(dbTask.Attempts)
	retryAt := time.Now().Add(delay)

	dbTask.Status = enums.TaskStatusQueued
	if err := w.manager.taskRepo.Update(dbTask); err != nil {
		w.manager.logger.Errorw("Failed to update task for retry", "task_id", taskID, "error", err)
		return
	}

	// Add to retry schedule (Redis sorted set)
	retryKey := w.manager.config.GetRetryScheduleKey()
	score := float64(retryAt.Unix())
	if err := w.manager.redis.ZAdd(ctx, retryKey, redis.Z{
		Score:  score,
		Member: taskID,
	}).Err(); err != nil {
		w.manager.logger.Errorw("Failed to schedule retry", "task_id", taskID, "error", err)
		return
	}

	// Publish event
	if task.AccountID != nil {
		w.manager.publishEvent(ctx, EventTaskRetry, taskID, *task.AccountID, enums.TaskStatusQueued, map[string]interface{}{
			"attempt":  dbTask.Attempts,
			"retry_at": retryAt,
		})
	}

	w.manager.logger.Infow("Task scheduled for retry",
		"task_id", taskID,
		"attempt", dbTask.Attempts,
		"retry_at", retryAt,
		"delay", delay,
	)
}

func (w *worker) moveToDLQ(task *models.RepurposerTask, taskErr error) {
	ctx := w.manager.ctx
	taskID := task.TaskID

	// Update status to failed
	if err := w.manager.taskRepo.UpdateStatus(taskID, enums.TaskStatusFailed, taskErr.Error()); err != nil {
		w.manager.logger.Errorw("Failed to update task as failed", "task_id", taskID, "error", err)
		return
	}

	// Add to DLQ
	keys := w.manager.config.GetQueueKeys()
	if err := w.manager.redis.LPush(ctx, keys.DLQ, taskID).Err(); err != nil {
		w.manager.logger.Errorw("Failed to add task to DLQ", "task_id", taskID, "error", err)
		return
	}

	// Publish event
	if task.AccountID != nil {
		w.manager.publishEvent(ctx, EventTaskDLQ, taskID, *task.AccountID, enums.TaskStatusFailed, map[string]interface{}{
			"error": taskErr.Error(),
		})
	}

	w.manager.logger.Errorw("Task moved to DLQ after max retries",
		"task_id", taskID,
		"attempts", task.Attempts,
		"error", taskErr,
	)
}
