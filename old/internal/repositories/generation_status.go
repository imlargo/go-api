package repositories

import (
	"time"

	"github.com/nicolailuther/butter/internal/enums"
	"github.com/nicolailuther/butter/internal/models"
	"gorm.io/gorm"
)

type GenerationStatusRepository interface {
	// Basic CRUD
	Create(status *models.AccountGenerationStatus) error
	GetByID(id uint) (*models.AccountGenerationStatus, error)
	Update(status *models.AccountGenerationStatus) error
	Delete(id uint) error

	// Queries
	GetActiveByAccount(accountID uint, contentType enums.ContentType) (*models.AccountGenerationStatus, error)
	GetLatestByAccount(accountID uint, contentType enums.ContentType) (*models.AccountGenerationStatus, error)
	GetByAccountAndType(accountID uint, contentType enums.ContentType) ([]*models.AccountGenerationStatus, error)

	// Progress updates
	IncrementQueued(id uint) error
	IncrementProcessing(id uint) error
	IncrementCompleted(id uint) error
	IncrementFailed(id uint) error

	// Status updates
	UpdateStatus(id uint, status enums.GenerationStatus) error
	MarkCompleted(id uint) error

	// Cleanup and recovery
	GetStuckStatuses(stuckDuration time.Duration) ([]*models.AccountGenerationStatus, error)
}

type generationStatusRepository struct {
	*Repository
}

func NewGenerationStatusRepository(r *Repository) GenerationStatusRepository {
	return &generationStatusRepository{Repository: r}
}

func (r *generationStatusRepository) Create(status *models.AccountGenerationStatus) error {
	return r.db.Create(status).Error
}

func (r *generationStatusRepository) GetByID(id uint) (*models.AccountGenerationStatus, error) {
	var status models.AccountGenerationStatus
	err := r.db.First(&status, id).Error
	if err != nil {
		return nil, err
	}
	return &status, nil
}

func (r *generationStatusRepository) Update(status *models.AccountGenerationStatus) error {
	return r.db.Save(status).Error
}

func (r *generationStatusRepository) Delete(id uint) error {
	return r.db.Delete(&models.AccountGenerationStatus{}, id).Error
}

func (r *generationStatusRepository) GetActiveByAccount(accountID uint, contentType enums.ContentType) (*models.AccountGenerationStatus, error) {
	var status models.AccountGenerationStatus
	err := r.db.Where("account_id = ? AND content_type = ?", accountID, contentType).
		Where("status IN ?", []enums.GenerationStatus{
			enums.GenerationStatusQueuing,
			enums.GenerationStatusProcessing,
		}).
		Order("started_at DESC").
		First(&status).Error

	if err != nil {
		return nil, err
	}
	return &status, nil
}

func (r *generationStatusRepository) GetLatestByAccount(accountID uint, contentType enums.ContentType) (*models.AccountGenerationStatus, error) {
	var status models.AccountGenerationStatus
	err := r.db.Where("account_id = ? AND content_type = ?", accountID, contentType).
		Order("created_at DESC").
		First(&status).Error

	if err != nil {
		return nil, err
	}
	return &status, nil
}

func (r *generationStatusRepository) GetByAccountAndType(accountID uint, contentType enums.ContentType) ([]*models.AccountGenerationStatus, error) {
	var statuses []*models.AccountGenerationStatus
	err := r.db.Where("account_id = ? AND content_type = ?", accountID, contentType).
		Order("started_at DESC").
		Find(&statuses).Error

	return statuses, err
}

func (r *generationStatusRepository) IncrementQueued(id uint) error {
	return r.db.Model(&models.AccountGenerationStatus{}).
		Where("id = ?", id).
		UpdateColumn("total_queued", gorm.Expr("total_queued + ?", 1)).Error
}

func (r *generationStatusRepository) IncrementProcessing(id uint) error {
	// Use conditional logic to only decrement total_queued if it's greater than 0
	// This handles retry scenarios where total_queued might already be 0
	return r.db.Model(&models.AccountGenerationStatus{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"total_processing": gorm.Expr("total_processing + ?", 1),
			"total_queued":     gorm.Expr("CASE WHEN total_queued > 0 THEN total_queued - 1 ELSE total_queued END"),
		}).Error
}

func (r *generationStatusRepository) IncrementCompleted(id uint) error {
	return r.db.Model(&models.AccountGenerationStatus{}).
		Where("id = ? AND total_processing > 0", id).
		Updates(map[string]interface{}{
			"total_completed":  gorm.Expr("total_completed + ?", 1),
			"total_processing": gorm.Expr("total_processing - ?", 1),
		}).Error
}

func (r *generationStatusRepository) IncrementFailed(id uint) error {
	// Atomically increment failed and decrement processing if > 0
	// Use CASE statement to handle conditional decrement without race condition
	return r.db.Model(&models.AccountGenerationStatus{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"total_failed":     gorm.Expr("total_failed + ?", 1),
			"total_processing": gorm.Expr("CASE WHEN total_processing > 0 THEN total_processing - 1 ELSE total_processing END"),
		}).Error
}

func (r *generationStatusRepository) UpdateStatus(id uint, status enums.GenerationStatus) error {
	return r.db.Model(&models.AccountGenerationStatus{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *generationStatusRepository) MarkCompleted(id uint) error {
	now := time.Now()
	return r.db.Model(&models.AccountGenerationStatus{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"completed_at": now,
		}).Error
}

// GetStuckStatuses finds generation statuses that have been in queuing or processing state
// for longer than the specified duration and may be stuck
func (r *generationStatusRepository) GetStuckStatuses(stuckDuration time.Duration) ([]*models.AccountGenerationStatus, error) {
	var statuses []*models.AccountGenerationStatus
	cutoffTime := time.Now().Add(-stuckDuration)

	err := r.db.Where("status IN ?", []enums.GenerationStatus{
		enums.GenerationStatusQueuing,
		enums.GenerationStatusProcessing,
	}).
		Where("updated_at < ?", cutoffTime).
		Find(&statuses).Error

	return statuses, err
}
