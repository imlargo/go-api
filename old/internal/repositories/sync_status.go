package repositories

import (
	"time"

	"github.com/nicolailuther/butter/internal/enums"
	"github.com/nicolailuther/butter/internal/models"
	"gorm.io/gorm"
)

type SyncStatusRepository interface {
	// Basic CRUD
	Create(status *models.AccountSyncStatus) error
	GetByID(id uint) (*models.AccountSyncStatus, error)
	Update(status *models.AccountSyncStatus) error
	Delete(id uint) error

	// Queries
	GetActiveByAccount(accountID uint) (*models.AccountSyncStatus, error)
	GetLatestByAccount(accountID uint) (*models.AccountSyncStatus, error)
	GetAllActiveSyncs() ([]*models.AccountSyncStatus, error)

	// Progress updates
	IncrementProcessed(id uint) error
	IncrementSynced(id uint) error
	IncrementFailed(id uint) error

	// Status updates
	UpdateStatus(id uint, status enums.SyncStatus) error
	MarkCompleted(id uint) error
	MarkFailed(id uint, errorMessage string) error
}

type syncStatusRepository struct {
	*Repository
}

func NewSyncStatusRepository(r *Repository) SyncStatusRepository {
	return &syncStatusRepository{Repository: r}
}

func (r *syncStatusRepository) Create(status *models.AccountSyncStatus) error {
	return r.db.Create(status).Error
}

func (r *syncStatusRepository) GetByID(id uint) (*models.AccountSyncStatus, error) {
	var status models.AccountSyncStatus
	err := r.db.First(&status, id).Error
	if err != nil {
		return nil, err
	}
	return &status, nil
}

func (r *syncStatusRepository) Update(status *models.AccountSyncStatus) error {
	return r.db.Save(status).Error
}

func (r *syncStatusRepository) Delete(id uint) error {
	return r.db.Delete(&models.AccountSyncStatus{}, id).Error
}

func (r *syncStatusRepository) GetActiveByAccount(accountID uint) (*models.AccountSyncStatus, error) {
	var status models.AccountSyncStatus
	err := r.db.Where("account_id = ?", accountID).
		Where("is_active = ?", true).
		Order("started_at DESC").
		First(&status).Error

	if err != nil {
		return nil, err
	}
	return &status, nil
}

func (r *syncStatusRepository) GetLatestByAccount(accountID uint) (*models.AccountSyncStatus, error) {
	var status models.AccountSyncStatus
	err := r.db.Where("account_id = ?", accountID).
		Order("created_at DESC").
		First(&status).Error

	if err != nil {
		return nil, err
	}
	return &status, nil
}

func (r *syncStatusRepository) IncrementProcessed(id uint) error {
	return r.db.Model(&models.AccountSyncStatus{}).
		Where("id = ?", id).
		UpdateColumn("total_processed", gorm.Expr("total_processed + ?", 1)).Error
}

func (r *syncStatusRepository) IncrementSynced(id uint) error {
	return r.db.Model(&models.AccountSyncStatus{}).
		Where("id = ?", id).
		UpdateColumn("total_synced", gorm.Expr("total_synced + ?", 1)).Error
}

func (r *syncStatusRepository) IncrementFailed(id uint) error {
	return r.db.Model(&models.AccountSyncStatus{}).
		Where("id = ?", id).
		UpdateColumn("total_failed", gorm.Expr("total_failed + ?", 1)).Error
}

func (r *syncStatusRepository) UpdateStatus(id uint, status enums.SyncStatus) error {
	return r.db.Model(&models.AccountSyncStatus{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *syncStatusRepository) MarkCompleted(id uint) error {
	var status models.AccountSyncStatus
	if err := r.db.First(&status, id).Error; err != nil {
		return err
	}
	status.UpdateProgress()
	now := time.Now()
	return r.db.Model(&models.AccountSyncStatus{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"completed_at": now,
			"status":       status.Status,
			"is_active":    status.IsActive,
		}).Error
}

func (r *syncStatusRepository) GetAllActiveSyncs() ([]*models.AccountSyncStatus, error) {
	var statuses []*models.AccountSyncStatus
	err := r.db.Where("is_active = ?", true).
		Find(&statuses).Error
	return statuses, err
}

func (r *syncStatusRepository) MarkFailed(id uint, errorMessage string) error {
	now := time.Now()
	return r.db.Model(&models.AccountSyncStatus{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":        enums.SyncStatusFailed,
			"error_message": errorMessage,
			"completed_at":  now,
			"is_active":     false,
		}).Error
}
