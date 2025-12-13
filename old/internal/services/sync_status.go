package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/nicolailuther/butter/internal/enums"
	"github.com/nicolailuther/butter/internal/models"
)

var (
	ErrSyncInProgress     = errors.New("post sync already in progress for this account")
	ErrSyncStatusNotFound = errors.New("sync status not found")
)

// SyncStatusService manages sync status tracking and acts as lock manager
// The unique index on (account_id, is_active) where is_active=true ensures only one active sync per account
type SyncStatusService interface {
	// Lock-like operations through status
	AcquireSync(accountID uint, totalToProcess int) (*models.AccountSyncStatus, error)

	// Status queries
	GetActiveStatus(accountID uint) (*models.AccountSyncStatus, error)
	GetLatestStatus(accountID uint) (*models.AccountSyncStatus, error)

	// Progress updates
	IncrementProcessed(statusID uint) error
	IncrementSynced(statusID uint) error
	IncrementFailed(statusID uint) error

	// Completion and failure
	CompleteSync(statusID uint) error
	MarkFailed(statusID uint, errorMessage string) error

	// Reconciliation
	ReconcileOnStartup() error
}

type syncStatusService struct {
	*Service
}

func NewSyncStatusService(
	container *Service,
) SyncStatusService {
	return &syncStatusService{
		Service: container,
	}
}

// AcquireSync creates a new sync status, acting as lock acquisition
// The unique index prevents concurrent syncs for the same account
func (s *syncStatusService) AcquireSync(accountID uint, totalToProcess int) (*models.AccountSyncStatus, error) {
	// Check if sync already exists for this account
	existingStatus, err := s.store.SyncStatus.GetActiveByAccount(accountID)
	if err == nil && existingStatus != nil {
		return nil, ErrSyncInProgress
	}

	now := time.Now()
	status := &models.AccountSyncStatus{
		AccountID:      accountID,
		TotalToProcess: totalToProcess,
		Status:         enums.SyncStatusSyncing,
		StartedAt:      &now,
	}

	if err := s.store.SyncStatus.Create(status); err != nil {
		// Check if it's a duplicate key error (race condition)
		errStr := err.Error()
		if strings.Contains(errStr, "duplicate") || strings.Contains(errStr, "unique constraint") || strings.Contains(errStr, "UNIQUE constraint failed") {
			return nil, ErrSyncInProgress
		}
		return nil, fmt.Errorf("failed to create sync status: %w", err)
	}

	s.logger.Infow("Sync acquired (status created)",
		"status_id", status.ID,
		"account_id", accountID,
		"total_to_process", totalToProcess,
	)

	return status, nil
}

func (s *syncStatusService) GetActiveStatus(accountID uint) (*models.AccountSyncStatus, error) {
	status, err := s.store.SyncStatus.GetActiveByAccount(accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active status: %w", err)
	}
	return status, nil
}

func (s *syncStatusService) GetLatestStatus(accountID uint) (*models.AccountSyncStatus, error) {
	status, err := s.store.SyncStatus.GetLatestByAccount(accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest status: %w", err)
	}
	return status, nil
}

func (s *syncStatusService) IncrementProcessed(statusID uint) error {
	if err := s.store.SyncStatus.IncrementProcessed(statusID); err != nil {
		return fmt.Errorf("failed to increment processed count: %w", err)
	}

	return nil
}

func (s *syncStatusService) IncrementSynced(statusID uint) error {
	if err := s.store.SyncStatus.IncrementSynced(statusID); err != nil {
		return fmt.Errorf("failed to increment synced count: %w", err)
	}

	return nil
}

func (s *syncStatusService) IncrementFailed(statusID uint) error {
	if err := s.store.SyncStatus.IncrementFailed(statusID); err != nil {
		return fmt.Errorf("failed to increment failed count: %w", err)
	}

	return nil
}

func (s *syncStatusService) CompleteSync(statusID uint) error {
	// Mark as completed (this also calls UpdateProgress internally and releases the lock)
	if err := s.store.SyncStatus.MarkCompleted(statusID); err != nil {
		return fmt.Errorf("failed to mark sync as completed: %w", err)
	}

	status, err := s.store.SyncStatus.GetByID(statusID)
	if err != nil {
		s.logger.Warnw("Failed to get status for logging", "status_id", statusID, "error", err)
		return nil
	}

	s.logger.Infow("Sync completed",
		"status_id", statusID,
		"account_id", status.AccountID,
		"total_synced", status.TotalSynced,
		"total_failed", status.TotalFailed,
		"final_status", status.Status,
	)

	return nil
}

func (s *syncStatusService) MarkFailed(statusID uint, errorMessage string) error {
	if err := s.store.SyncStatus.MarkFailed(statusID, errorMessage); err != nil {
		return fmt.Errorf("failed to mark sync as failed: %w", err)
	}

	s.logger.Infow("Sync marked as failed",
		"status_id", statusID,
		"error_message", errorMessage,
	)

	return nil
}

// ReconcileOnStartup identifies and marks orphaned sync statuses as failed
// This is called when the API starts to handle statuses left in 'syncing' state from server crashes
func (s *syncStatusService) ReconcileOnStartup() error {
	s.logger.Info("Starting sync status reconciliation on startup...")

	activeSyncs, err := s.store.SyncStatus.GetAllActiveSyncs()
	if err != nil {
		s.logger.Errorw("Failed to fetch active sync statuses for reconciliation", "error", err)
		return fmt.Errorf("failed to fetch active sync statuses: %w", err)
	}

	if len(activeSyncs) == 0 {
		s.logger.Info("No orphaned sync statuses found")
		return nil
	}

	s.logger.Infow("Found orphaned sync statuses to reconcile", "count", len(activeSyncs))

	for _, status := range activeSyncs {
		errMsg := "Sync marked as failed during reconciliation after unexpected shutdown"
		if err := s.MarkFailed(status.ID, errMsg); err != nil {
			s.logger.Errorw("Failed to mark orphaned sync as failed during reconciliation",
				"status_id", status.ID,
				"account_id", status.AccountID,
				"error", err)
			continue
		}
		s.logger.Infow("Orphaned sync status marked as failed during reconciliation",
			"status_id", status.ID,
			"account_id", status.AccountID,
		)
	}

	s.logger.Info("Sync status reconciliation completed")
	return nil
}
