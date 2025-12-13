package taskqueue

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nicolailuther/butter/internal/dto"
	"github.com/nicolailuther/butter/internal/enums"
	"github.com/nicolailuther/butter/internal/models"
	"github.com/nicolailuther/butter/internal/repositories"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// TaskManager manages the task queue system
type TaskManager interface {
	// Task submission
	SubmitTask(ctx context.Context, request *dto.ReporpuseVideo) (string, error)
	SubmitTaskWithPriority(ctx context.Context, request *dto.ReporpuseVideo, priority enums.TaskPriority) (string, error)

	// Task queries
	GetTask(ctx context.Context, taskID string) (*TaskInfo, error)
	GetTasksByAccount(ctx context.Context, accountID uint, limit int, offset int) ([]*TaskInfo, error)
	GetTaskHistory(ctx context.Context, filter TaskFilter) ([]*TaskInfo, error)

	// Task control
	CancelTask(ctx context.Context, taskID string) error
	RetryTask(ctx context.Context, taskID string) error

	// Statistics
	GetStats(ctx context.Context) (*QueueStats, error)
	GetWorkerStats(ctx context.Context) ([]*WorkerStats, error)

	// Lifecycle
	Start(ctx context.Context) error
	Shutdown(ctx context.Context) error
	RecoverOrphanedTasks(ctx context.Context) error
}

// taskManager implements TaskManager
type taskManager struct {
	config   Config
	redis    *redis.Client
	taskRepo repositories.RepurposerTaskRepository
	logger   *zap.SugaredLogger

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	workers    []*worker
	workersMux sync.RWMutex

	taskHandler TaskHandler
}

// TaskHandler is the function type that processes tasks
type TaskHandler func(ctx context.Context, request *dto.ReporpuseVideo) (*dto.ReporpuseContentResult, error)

// NewTaskManager creates a new task manager
func NewTaskManager(
	config Config,
	redisClient *redis.Client,
	taskRepo repositories.RepurposerTaskRepository,
	logger *zap.SugaredLogger,
	taskHandler TaskHandler,
) TaskManager {
	ctx, cancel := context.WithCancel(context.Background())

	tm := &taskManager{
		config:      config,
		redis:       redisClient,
		taskRepo:    taskRepo,
		logger:      logger,
		ctx:         ctx,
		cancel:      cancel,
		taskHandler: taskHandler,
	}

	return tm
}

// Start starts the task manager and workers
func (tm *taskManager) Start(ctx context.Context) error {
	tm.logger.Infow("Starting task manager",
		"worker_count", tm.config.WorkerCount,
		"task_timeout", tm.config.TaskTimeout,
	)

	// Start workers
	tm.workersMux.Lock()
	tm.workers = make([]*worker, tm.config.WorkerCount)
	for i := 0; i < tm.config.WorkerCount; i++ {
		w := newWorker(fmt.Sprintf("worker-%d", i+1), tm)
		tm.workers[i] = w
		tm.wg.Add(1)
		go w.run()
	}
	tm.workersMux.Unlock()

	// Start retry scheduler
	tm.wg.Add(1)
	go tm.retryScheduler()

	// Start cleanup job
	tm.wg.Add(1)
	go tm.cleanupJob()

	tm.logger.Info("Task manager started successfully")
	return nil
}

// Shutdown gracefully shuts down the task manager
func (tm *taskManager) Shutdown(ctx context.Context) error {
	tm.logger.Info("Starting graceful shutdown...")

	// Signal workers to stop
	tm.cancel()

	// Wait for workers with timeout
	done := make(chan struct{})
	go func() {
		tm.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		tm.logger.Info("All workers finished gracefully")
		// Close Redis connection only after all workers have finished
		if err := tm.redis.Close(); err != nil {
			tm.logger.Errorw("Error closing Redis connection", "error", err)
		}
	case <-ctx.Done():
		tm.logger.Warn("Shutdown timeout reached, some workers may not have finished")
		// Do not close Redis connection here, as workers may still be using it
	}

	tm.logger.Info("Shutdown complete")
	return nil
}

// SubmitTask submits a new task with default priority
func (tm *taskManager) SubmitTask(ctx context.Context, request *dto.ReporpuseVideo) (string, error) {
	return tm.SubmitTaskWithPriority(ctx, request, enums.TaskPriorityNormal)
}

// SubmitTaskWithPriority submits a new task with specified priority
func (tm *taskManager) SubmitTaskWithPriority(ctx context.Context, request *dto.ReporpuseVideo, priority enums.TaskPriority) (string, error) {
	// Create task record
	taskID := uuid.New().String()
	accountID := request.AccountID
	task := &models.RepurposerTask{
		TaskID:           taskID,
		FileID:           request.FileID,
		AccountID:        &accountID,
		ContentID:        &request.ContentID,
		ContentAccountID: &request.ContentAccountID,
		Status:           enums.TaskStatusPending,
		Priority:         priority,
		MaxRetries:       tm.config.MaxRetries,
	}

	// Store request data
	if err := task.MarshalRequestData(request); err != nil {
		return "", fmt.Errorf("failed to marshal request data: %w", err)
	}

	// Save to database
	if err := tm.taskRepo.Create(task); err != nil {
		return "", fmt.Errorf("failed to create task: %w", err)
	}

	// Add to appropriate Redis queue
	queueKey := tm.getQueueKeyForPriority(priority)
	if err := tm.redis.LPush(ctx, queueKey, taskID).Err(); err != nil {
		// Rollback: mark as failed
		tm.taskRepo.UpdateStatus(taskID, enums.TaskStatusFailed, "Failed to queue task")
		return "", fmt.Errorf("failed to queue task: %w", err)
	}

	// Update status to queued
	if err := tm.taskRepo.UpdateStatus(taskID, enums.TaskStatusQueued, ""); err != nil {
		tm.logger.Warnw("Failed to update task status to queued", "task_id", taskID, "error", err)
	}

	// Publish event
	tm.publishEvent(ctx, EventTaskQueued, taskID, request.AccountID, enums.TaskStatusQueued, nil)

	tm.logger.Infow("Task submitted", "task_id", taskID, "priority", priority)
	return taskID, nil
}

// GetTask retrieves task information
func (tm *taskManager) GetTask(ctx context.Context, taskID string) (*TaskInfo, error) {
	task, err := tm.taskRepo.GetByTaskID(taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return tm.modelToTaskInfo(task), nil
}

// GetTasksByAccount retrieves tasks for an account
func (tm *taskManager) GetTasksByAccount(ctx context.Context, accountID uint, limit int, offset int) ([]*TaskInfo, error) {
	tasks, err := tm.taskRepo.GetByAccountID(accountID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks by account: %w", err)
	}

	result := make([]*TaskInfo, len(tasks))
	for i, task := range tasks {
		result[i] = tm.modelToTaskInfo(task)
	}

	return result, nil
}

// GetTaskHistory retrieves task history with filters
func (tm *taskManager) GetTaskHistory(ctx context.Context, filter TaskFilter) ([]*TaskInfo, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}

	var tasks []*models.RepurposerTask
	var err error

	if filter.Status != nil {
		tasks, err = tm.taskRepo.GetByStatus(*filter.Status, limit, filter.Offset)
	} else if filter.AccountID != nil {
		tasks, err = tm.taskRepo.GetByAccountID(*filter.AccountID, limit, filter.Offset)
	} else {
		tasks, err = tm.taskRepo.GetRecentTasks(limit)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get task history: %w", err)
	}

	result := make([]*TaskInfo, len(tasks))
	for i, task := range tasks {
		result[i] = tm.modelToTaskInfo(task)
	}

	return result, nil
}

// CancelTask cancels a task
func (tm *taskManager) CancelTask(ctx context.Context, taskID string) error {
	task, err := tm.taskRepo.GetByTaskID(taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Can only cancel pending or queued tasks
	if task.Status != enums.TaskStatusPending && task.Status != enums.TaskStatusQueued {
		return fmt.Errorf("cannot cancel task in status: %s", task.Status)
	}

	// Remove from Redis queue
	keys := tm.config.GetQueueKeys()
	removedHigh, err := tm.redis.LRem(ctx, keys.HighPriority, 0, taskID).Result()
	if err != nil {
		return fmt.Errorf("failed to remove task from high priority queue: %w", err)
	}
	removedNormal, err := tm.redis.LRem(ctx, keys.NormalPriority, 0, taskID).Result()
	if err != nil {
		return fmt.Errorf("failed to remove task from normal priority queue: %w", err)
	}
	removedLow, err := tm.redis.LRem(ctx, keys.LowPriority, 0, taskID).Result()
	if err != nil {
		return fmt.Errorf("failed to remove task from low priority queue: %w", err)
	}

	totalRemoved := removedHigh + removedNormal + removedLow
	if totalRemoved == 0 {
		tm.logger.Warnw("Task was not found in any queue during cancellation", "task_id", taskID)
		// Still update status as it may have just been picked up by a worker
	}

	// Update status
	if err := tm.taskRepo.UpdateStatus(taskID, enums.TaskStatusCanceled, "Canceled by user"); err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	// Publish event
	if task.AccountID != nil {
		tm.publishEvent(ctx, EventTaskCanceled, taskID, *task.AccountID, enums.TaskStatusCanceled, nil)
	}

	tm.logger.Infow("Task canceled", "task_id", taskID)
	return nil
}

// RetryTask manually retries a failed task
func (tm *taskManager) RetryTask(ctx context.Context, taskID string) error {
	task, err := tm.taskRepo.GetByTaskID(taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Can only retry failed tasks
	if task.Status != enums.TaskStatusFailed {
		return fmt.Errorf("can only retry failed tasks, current status: %s", task.Status)
	}

	// Reset attempts
	task.Attempts = 0
	task.Status = enums.TaskStatusQueued
	task.ErrorMessage = ""
	task.FailedAt = nil
	now := time.Now()
	task.QueuedAt = &now

	if err := tm.taskRepo.Update(task); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	// Re-queue
	queueKey := tm.getQueueKeyForPriority(task.Priority)
	if err := tm.redis.LPush(ctx, queueKey, taskID).Err(); err != nil {
		return fmt.Errorf("failed to re-queue task: %w", err)
	}

	// Publish event
	if task.AccountID != nil {
		tm.publishEvent(ctx, EventTaskRetry, taskID, *task.AccountID, enums.TaskStatusQueued, nil)
	}

	tm.logger.Infow("Task manually retried", "task_id", taskID)
	return nil
}

// Helper methods

func (tm *taskManager) getQueueKeyForPriority(priority enums.TaskPriority) string {
	keys := tm.config.GetQueueKeys()

	if priority >= tm.config.PriorityHighThreshold {
		return keys.HighPriority
	} else if priority >= tm.config.PriorityNormalThreshold {
		return keys.NormalPriority
	}
	return keys.LowPriority
}

func (tm *taskManager) modelToTaskInfo(task *models.RepurposerTask) *TaskInfo {
	info := &TaskInfo{
		ID:           task.TaskID,
		TaskID:       task.TaskID,
		Status:       task.Status,
		FileID:       task.FileID,
		AccountID:    task.AccountID,
		ContentID:    task.ContentID,
		Priority:     task.Priority,
		Attempts:     task.Attempts,
		MaxRetries:   task.MaxRetries,
		CreatedAt:    task.CreatedAt,
		QueuedAt:     task.QueuedAt,
		StartedAt:    task.StartedAt,
		CompletedAt:  task.CompletedAt,
		FailedAt:     task.FailedAt,
		ErrorMessage: task.ErrorMessage,
		WorkerID:     task.WorkerID,
		RequestData:  task.RequestData,
		ResultData:   task.ResultData,
	}

	if task.ProcessingTimeMs > 0 {
		duration := time.Duration(task.ProcessingTimeMs) * time.Millisecond
		info.ProcessingTime = &duration
	}

	if task.QueueTimeMs > 0 {
		duration := time.Duration(task.QueueTimeMs) * time.Millisecond
		info.QueueTime = &duration
	}

	return info
}

func (tm *taskManager) calculateRetryDelay(attempt int) time.Duration {
	delay := time.Duration(float64(tm.config.InitialRetryDelay) * math.Pow(tm.config.BackoffFactor, float64(attempt-1)))

	if delay > tm.config.MaxRetryDelay {
		delay = tm.config.MaxRetryDelay
	}

	return delay
}

func (tm *taskManager) publishEvent(ctx context.Context, eventType string, taskID string, accountID uint, status enums.TaskStatus, data map[string]interface{}) {
	event := TaskEvent{
		EventType: eventType,
		TaskID:    taskID,
		AccountID: &accountID,
		Status:    status,
		Timestamp: time.Now(),
		Data:      data,
	}

	// Publish to Redis Pub/Sub
	keys := tm.config.GetQueueKeys()
	if err := tm.redis.Publish(ctx, keys.Events, event).Err(); err != nil {
		tm.logger.Warnw("Failed to publish event", "error", err)
	}
}

// Background jobs

func (tm *taskManager) retryScheduler() {
	defer tm.wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-tm.ctx.Done():
			return
		case <-ticker.C:
			tm.processScheduledRetries()
		}
	}
}

func (tm *taskManager) processScheduledRetries() {
	ctx := tm.ctx
	now := time.Now()

	// Get tasks scheduled for retry
	retryKey := tm.config.GetRetryScheduleKey()
	results, err := tm.redis.ZRangeByScore(ctx, retryKey, &redis.ZRangeBy{
		Min: "0",
		Max: fmt.Sprintf("%d", now.Unix()),
	}).Result()

	if err != nil {
		tm.logger.Warnw("Failed to get scheduled retries", "error", err)
		return
	}

	for _, taskID := range results {
		// Remove from retry schedule
		tm.redis.ZRem(ctx, retryKey, taskID)

		// Get task
		task, err := tm.taskRepo.GetByTaskID(taskID)
		if err != nil {
			tm.logger.Warnw("Failed to get task for retry", "task_id", taskID, "error", err)
			continue
		}

		// Re-queue
		queueKey := tm.getQueueKeyForPriority(task.Priority)
		if err := tm.redis.LPush(ctx, queueKey, taskID).Err(); err != nil {
			tm.logger.Warnw("Failed to re-queue task", "task_id", taskID, "error", err)
			continue
		}

		tm.logger.Infow("Task re-queued after retry delay", "task_id", taskID, "attempt", task.Attempts)
	}
}

func (tm *taskManager) cleanupJob() {
	defer tm.wg.Done()

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-tm.ctx.Done():
			return
		case <-ticker.C:
			// Check DLQ size
			tm.checkDLQSize()
		}
	}
}

func (tm *taskManager) checkDLQSize() {
	ctx := tm.ctx
	keys := tm.config.GetQueueKeys()

	size, err := tm.redis.LLen(ctx, keys.DLQ).Result()
	if err != nil {
		tm.logger.Warnw("Failed to get DLQ size", "error", err)
		return
	}

	if int(size) >= tm.config.DLQAlertThreshold {
		tm.logger.Warnw("DLQ size exceeds threshold",
			"size", size,
			"threshold", tm.config.DLQAlertThreshold,
		)
	}
}

func (tm *taskManager) RecoverOrphanedTasks(ctx context.Context) error {
	tm.logger.Info("Recovering orphaned tasks...")

	orphanedTasks, err := tm.taskRepo.FindOrphanedTasks(tm.config.OrphanTimeout)
	if err != nil {
		return fmt.Errorf("failed to find orphaned tasks: %w", err)
	}

	tm.logger.Infow("Found orphaned tasks", "count", len(orphanedTasks))

	for _, task := range orphanedTasks {
		if task.Attempts >= task.MaxRetries {
			// Move to failed
			tm.taskRepo.UpdateStatus(task.TaskID, enums.TaskStatusFailed, "Task orphaned after max retries")
			tm.logger.Infow("Orphaned task marked as failed", "task_id", task.TaskID)
		} else {
			// Re-queue for retry
			task.Attempts++
			task.Status = enums.TaskStatusQueued
			if err := tm.taskRepo.Update(task); err != nil {
				tm.logger.Warnw("Failed to update orphaned task", "task_id", task.TaskID, "error", err)
				continue
			}

			queueKey := tm.getQueueKeyForPriority(task.Priority)
			if err := tm.redis.LPush(ctx, queueKey, task.TaskID).Err(); err != nil {
				tm.logger.Warnw("Failed to re-queue orphaned task", "task_id", task.TaskID, "error", err)
				continue
			}

			tm.logger.Infow("Orphaned task re-queued", "task_id", task.TaskID)
		}
	}

	return nil
}
