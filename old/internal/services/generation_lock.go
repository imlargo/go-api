package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/nicolailuther/butter/internal/enums"
)

var (
	ErrGenerationInProgress = errors.New("content generation already in progress for this account and type")
	ErrLockNotFound         = errors.New("generation lock not found")
)

// GenerationLockService manages generation locks for accounts
type GenerationLockService interface {
	AcquireLock(accountID uint, contentType enums.ContentType) (string, error)
	ReleaseLock(lockID string) error
	HasActiveLock(accountID uint, contentType enums.ContentType) (bool, error)
	ReconcileLocksOnStartup() error
	CleanupExpiredLocks(lockDuration time.Duration) (int, error)
}

type generationLockService struct {
	*Service
}

func NewGenerationLockService(
	container *Service,
) GenerationLockService {
	return &generationLockService{
		Service: container,
	}
}

func (s *generationLockService) AcquireLock(accountID uint, contentType enums.ContentType) (string, error) {
	// Check if lock already exists
	hasLock, err := s.store.GenerationLocks.HasLock(accountID, contentType)
	if err != nil {
		return "", fmt.Errorf("failed to check existing lock: %w", err)
	}

	if hasLock {
		return "", ErrGenerationInProgress
	}

	// Acquire new lock
	lockID, err := s.store.GenerationLocks.AcquireLock(accountID, contentType)
	if err != nil {
		// Check if it's a duplicate key error (race condition)
		// GORM returns a constraint violation for duplicate keys
		errStr := err.Error()
		if strings.Contains(errStr, "duplicate") || strings.Contains(errStr, "unique constraint") || strings.Contains(errStr, "UNIQUE constraint failed") {
			return "", ErrGenerationInProgress
		}
		return "", fmt.Errorf("failed to acquire generation lock: %w", err)
	}

	s.logger.Infow("Generation lock acquired",
		"account_id", accountID,
		"content_type", contentType,
		"lock_id", lockID,
	)

	return lockID, nil
}

func (s *generationLockService) ReleaseLock(lockID string) error {
	if err := s.store.GenerationLocks.ReleaseLock(lockID); err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}

	s.logger.Infow("Generation lock released", "lock_id", lockID)
	return nil
}

func (s *generationLockService) HasActiveLock(accountID uint, contentType enums.ContentType) (bool, error) {
	return s.store.GenerationLocks.HasLock(accountID, contentType)
}

func (s *generationLockService) ReconcileLocksOnStartup() error {
	s.logger.Info("Starting lock reconciliation on startup...")

	// Get all existing locks
	locks, err := s.store.GenerationLocks.GetAllLocks()
	if err != nil {
		return fmt.Errorf("failed to get locks for reconciliation: %w", err)
	}

	if len(locks) == 0 {
		s.logger.Info("No locks found to reconcile")
		return nil
	}

	s.logger.Infow("Found locks to reconcile", "count", len(locks))

	orphanedLockIDs := []string{}

	// Check each lock against generation status
	for _, lock := range locks {
		// Try to find active generation status for this account/content type
		status, err := s.store.GenerationStatus.GetActiveByAccount(lock.AccountID, lock.ContentType)
		if err != nil || status == nil {
			// No active generation status found, lock is orphaned
			s.logger.Warnw("Found lock without active generation status, marking as orphaned",
				"lock_id", lock.LockID,
				"account_id", lock.AccountID,
				"content_type", lock.ContentType,
			)
			orphanedLockIDs = append(orphanedLockIDs, lock.LockID)
			continue
		}

		// Verify the lock ID matches
		if status.LockID == nil || *status.LockID != lock.LockID {
			// Lock ID mismatch, lock is orphaned
			s.logger.Warnw("Found lock with mismatched generation status, marking as orphaned",
				"lock_id", lock.LockID,
				"status_lock_id", status.LockID,
				"account_id", lock.AccountID,
				"content_type", lock.ContentType,
			)
			orphanedLockIDs = append(orphanedLockIDs, lock.LockID)
			continue
		}

		// Lock is valid and associated with active generation
		s.logger.Infow("Lock is valid and associated with active generation",
			"lock_id", lock.LockID,
			"status_id", status.ID,
			"account_id", lock.AccountID,
			"content_type", lock.ContentType,
		)
	}

	// Delete orphaned locks
	if len(orphanedLockIDs) > 0 {
		if err := s.store.GenerationLocks.DeleteOrphanedLocks(orphanedLockIDs); err != nil {
			return fmt.Errorf("failed to delete orphaned locks: %w", err)
		}
		s.logger.Infow("Deleted orphaned locks", "count", len(orphanedLockIDs))
	} else {
		s.logger.Info("No orphaned locks found")
	}

	s.logger.Info("Lock reconciliation completed")
	return nil
}

// CleanupExpiredLocks removes locks that have been held for longer than the specified duration
func (s *generationLockService) CleanupExpiredLocks(lockDuration time.Duration) (int, error) {
	s.logger.Infow("Starting cleanup of expired locks", "lock_duration", lockDuration)

	expirationTime := time.Now().Add(-lockDuration)

	// Get expired locks
	expiredLocks, err := s.store.GenerationLocks.GetExpiredLocks(expirationTime)
	if err != nil {
		return 0, fmt.Errorf("failed to get expired locks: %w", err)
	}

	if len(expiredLocks) == 0 {
		s.logger.Info("No expired locks found")
		return 0, nil
	}

	s.logger.Infow("Found expired locks", "count", len(expiredLocks))

	expiredLockIDs := make([]string, 0, len(expiredLocks))
	for _, lock := range expiredLocks {
		s.logger.Warnw("Found expired lock",
			"lock_id", lock.LockID,
			"account_id", lock.AccountID,
			"content_type", lock.ContentType,
			"locked_at", lock.LockedAt,
			"age", time.Since(lock.LockedAt),
		)
		expiredLockIDs = append(expiredLockIDs, lock.LockID)
	}

	// Delete expired locks
	if err := s.store.GenerationLocks.DeleteOrphanedLocks(expiredLockIDs); err != nil {
		return 0, fmt.Errorf("failed to delete expired locks: %w", err)
	}

	s.logger.Infow("Cleanup complete", "removed_count", len(expiredLocks))
	return len(expiredLocks), nil
}
