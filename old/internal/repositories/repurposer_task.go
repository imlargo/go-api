package repositories

import (
	"time"

	"github.com/nicolailuther/butter/internal/enums"
	"github.com/nicolailuther/butter/internal/models"
)

type RepurposerTaskRepository interface {
	// Basic CRUD
	Create(task *models.RepurposerTask) error
	GetByID(id uint) (*models.RepurposerTask, error)
	GetByTaskID(taskID string) (*models.RepurposerTask, error)
	Update(task *models.RepurposerTask) error
	Delete(id uint) error

	// Status queries
	GetByStatus(status enums.TaskStatus, limit int, offset int) ([]*models.RepurposerTask, error)
	CountByStatus(status enums.TaskStatus) (int64, error)
	GetByAccountID(accountID uint, limit int, offset int) ([]*models.RepurposerTask, error)

	// Status updates
	UpdateStatus(taskID string, status enums.TaskStatus, errorMessage string) error
	UpdateWorkerInfo(taskID string, workerID string) error
	UpdateHeartbeat(taskID string) error
	UpdateMetrics(taskID string, processingTimeMs int64, queueTimeMs int64) error

	// Task recovery
	FindOrphanedTasks(timeout time.Duration) ([]*models.RepurposerTask, error)
	FindExpiredProcessingTasks(maxProcessingTime time.Duration) ([]*models.RepurposerTask, error)

	// Statistics
	GetRecentTasks(limit int) ([]*models.RepurposerTask, error)
	GetAverageProcessingTime(since time.Time) (float64, error)
	GetAverageQueueTime(since time.Time) (float64, error)
	CountCompletedSince(since time.Time) (int64, error)
	CountFailedSince(since time.Time) (int64, error)
}

type repurposerTaskRepository struct {
	*Repository
}

func NewRepurposerTaskRepository(r *Repository) RepurposerTaskRepository {
	return &repurposerTaskRepository{Repository: r}
}

func (r *repurposerTaskRepository) Create(task *models.RepurposerTask) error {
	return r.db.Create(task).Error
}

func (r *repurposerTaskRepository) GetByID(id uint) (*models.RepurposerTask, error) {
	var task models.RepurposerTask
	err := r.db.First(&task, id).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *repurposerTaskRepository) GetByTaskID(taskID string) (*models.RepurposerTask, error) {
	var task models.RepurposerTask
	err := r.db.Where("task_id = ?", taskID).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *repurposerTaskRepository) Update(task *models.RepurposerTask) error {
	return r.db.Save(task).Error
}

func (r *repurposerTaskRepository) Delete(id uint) error {
	return r.db.Delete(&models.RepurposerTask{}, id).Error
}

func (r *repurposerTaskRepository) GetByStatus(status enums.TaskStatus, limit int, offset int) ([]*models.RepurposerTask, error) {
	var tasks []*models.RepurposerTask
	err := r.db.
		Where("status = ?", status).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&tasks).Error
	return tasks, err
}

func (r *repurposerTaskRepository) CountByStatus(status enums.TaskStatus) (int64, error) {
	var count int64
	err := r.db.
		Model(&models.RepurposerTask{}).
		Where("status = ?", status).
		Count(&count).Error
	return count, err
}

func (r *repurposerTaskRepository) GetByAccountID(accountID uint, limit int, offset int) ([]*models.RepurposerTask, error) {
	var tasks []*models.RepurposerTask
	err := r.db.
		Where("account_id = ?", accountID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&tasks).Error
	return tasks, err
}

func (r *repurposerTaskRepository) UpdateStatus(taskID string, status enums.TaskStatus, errorMessage string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	switch status {
	case enums.TaskStatusQueued:
		updates["queued_at"] = time.Now()
	case enums.TaskStatusProcessing:
		updates["started_at"] = time.Now()
	case enums.TaskStatusCompleted:
		updates["completed_at"] = time.Now()
	case enums.TaskStatusFailed:
		updates["failed_at"] = time.Now()
		if errorMessage != "" {
			updates["error_message"] = errorMessage
		}
	}

	return r.db.
		Model(&models.RepurposerTask{}).
		Where("task_id = ?", taskID).
		Updates(updates).Error
}

func (r *repurposerTaskRepository) UpdateWorkerInfo(taskID string, workerID string) error {
	return r.db.
		Model(&models.RepurposerTask{}).
		Where("task_id = ?", taskID).
		Updates(map[string]interface{}{
			"worker_id":         workerID,
			"last_heartbeat_at": time.Now(),
		}).Error
}

func (r *repurposerTaskRepository) UpdateHeartbeat(taskID string) error {
	return r.db.
		Model(&models.RepurposerTask{}).
		Where("task_id = ?", taskID).
		Update("last_heartbeat_at", time.Now()).Error
}

func (r *repurposerTaskRepository) UpdateMetrics(taskID string, processingTimeMs int64, queueTimeMs int64) error {
	return r.db.
		Model(&models.RepurposerTask{}).
		Where("task_id = ?", taskID).
		Updates(map[string]interface{}{
			"processing_time_ms": processingTimeMs,
			"queue_time_ms":      queueTimeMs,
		}).Error
}

func (r *repurposerTaskRepository) FindOrphanedTasks(timeout time.Duration) ([]*models.RepurposerTask, error) {
	var tasks []*models.RepurposerTask
	cutoff := time.Now().Add(-timeout)

	err := r.db.
		Where("status = ?", enums.TaskStatusProcessing).
		Where("last_heartbeat_at < ? OR last_heartbeat_at IS NULL", cutoff).
		Find(&tasks).Error

	return tasks, err
}

func (r *repurposerTaskRepository) FindExpiredProcessingTasks(maxProcessingTime time.Duration) ([]*models.RepurposerTask, error) {
	var tasks []*models.RepurposerTask
	cutoff := time.Now().Add(-maxProcessingTime)

	err := r.db.
		Where("status = ?", enums.TaskStatusProcessing).
		Where("started_at < ?", cutoff).
		Find(&tasks).Error

	return tasks, err
}

func (r *repurposerTaskRepository) GetRecentTasks(limit int) ([]*models.RepurposerTask, error) {
	var tasks []*models.RepurposerTask
	err := r.db.
		Order("created_at DESC").
		Limit(limit).
		Find(&tasks).Error
	return tasks, err
}

func (r *repurposerTaskRepository) GetAverageProcessingTime(since time.Time) (float64, error) {
	var avg float64
	err := r.db.
		Model(&models.RepurposerTask{}).
		Where("status = ?", enums.TaskStatusCompleted).
		Where("completed_at >= ?", since).
		Where("processing_time_ms > 0").
		Select("AVG(processing_time_ms)").
		Scan(&avg).Error
	return avg, err
}

func (r *repurposerTaskRepository) GetAverageQueueTime(since time.Time) (float64, error) {
	var avg float64
	err := r.db.
		Model(&models.RepurposerTask{}).
		Where("completed_at >= ?", since).
		Where("queue_time_ms > 0").
		Select("AVG(queue_time_ms)").
		Scan(&avg).Error
	return avg, err
}

func (r *repurposerTaskRepository) CountCompletedSince(since time.Time) (int64, error) {
	var count int64
	err := r.db.
		Model(&models.RepurposerTask{}).
		Where("status = ?", enums.TaskStatusCompleted).
		Where("completed_at >= ?", since).
		Count(&count).Error
	return count, err
}

func (r *repurposerTaskRepository) CountFailedSince(since time.Time) (int64, error) {
	var count int64
	err := r.db.
		Model(&models.RepurposerTask{}).
		Where("status = ?", enums.TaskStatusFailed).
		Where("failed_at >= ?", since).
		Count(&count).Error
	return count, err
}
