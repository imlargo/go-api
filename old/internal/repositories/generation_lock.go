package repositories

import (
	"time"

	"github.com/google/uuid"
	"github.com/nicolailuther/butter/internal/enums"
	"github.com/nicolailuther/butter/internal/models"
)

type GenerationLockRepository interface {
	// Lock operations
	AcquireLock(accountID uint, contentType enums.ContentType) (string, error)
	ReleaseLock(lockID string) error
	HasLock(accountID uint, contentType enums.ContentType) (bool, error)
	GetLock(accountID uint, contentType enums.ContentType) (*models.AccountGenerationLock, error)

	// Reconciliation
	GetAllLocks() ([]*models.AccountGenerationLock, error)
	DeleteOrphanedLocks(lockIDs []string) error
	GetExpiredLocks(expirationTime time.Time) ([]*models.AccountGenerationLock, error)
}

type generationLockRepository struct {
	*Repository
}

func NewGenerationLockRepository(r *Repository) GenerationLockRepository {
	return &generationLockRepository{Repository: r}
}

func (r *generationLockRepository) AcquireLock(accountID uint, contentType enums.ContentType) (string, error) {
	lockID := uuid.New().String()
	now := time.Now()

	lock := &models.AccountGenerationLock{
		AccountID:   accountID,
		ContentType: contentType,
		LockedAt:    now,
		LockID:      lockID,
	}

	// Try to create the lock
	// The unique constraint on (account_id, content_type) will prevent duplicates
	err := r.db.Create(lock).Error
	if err != nil {
		return "", err
	}

	return lockID, nil
}

func (r *generationLockRepository) ReleaseLock(lockID string) error {
	return r.db.Where("lock_id = ?", lockID).
		Delete(&models.AccountGenerationLock{}).Error
}

func (r *generationLockRepository) HasLock(accountID uint, contentType enums.ContentType) (bool, error) {
	var count int64
	err := r.db.Model(&models.AccountGenerationLock{}).
		Where("account_id = ? AND content_type = ?", accountID, contentType).
		Count(&count).Error

	return count > 0, err
}

func (r *generationLockRepository) GetLock(accountID uint, contentType enums.ContentType) (*models.AccountGenerationLock, error) {
	var lock models.AccountGenerationLock
	err := r.db.Where("account_id = ? AND content_type = ?", accountID, contentType).
		First(&lock).Error

	if err != nil {
		return nil, err
	}

	return &lock, nil
}

func (r *generationLockRepository) GetAllLocks() ([]*models.AccountGenerationLock, error) {
	var locks []*models.AccountGenerationLock
	err := r.db.Find(&locks).Error
	return locks, err
}

func (r *generationLockRepository) DeleteOrphanedLocks(lockIDs []string) error {
	if len(lockIDs) == 0 {
		return nil
	}
	return r.db.Where("lock_id IN ?", lockIDs).Delete(&models.AccountGenerationLock{}).Error
}

func (r *generationLockRepository) GetExpiredLocks(expirationTime time.Time) ([]*models.AccountGenerationLock, error) {
	var locks []*models.AccountGenerationLock
	err := r.db.Where("locked_at < ?", expirationTime).Find(&locks).Error
	return locks, err
}
