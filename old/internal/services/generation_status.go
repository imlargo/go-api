package services

import (
	"fmt"
	"sort"
	"time"

	"github.com/nicolailuther/butter/internal/enums"
	"github.com/nicolailuther/butter/internal/models"
)

// GenerationStatusService manages generation status tracking
type GenerationStatusService interface {
	CreateStatus(accountID uint, contentType enums.ContentType, totalRequested int) (*models.AccountGenerationStatus, error)
	SetLockID(statusID uint, lockID string) error
	SetErrorCode(statusID uint, errorCode enums.GenerationErrorCode, errorMessage string) error
	GetActiveStatus(accountID uint, contentType enums.ContentType) (*models.AccountGenerationStatus, error)
	GetLatestStatus(accountID uint, contentType enums.ContentType) (*models.AccountGenerationStatus, error)
	GetByAccountID(accountID uint, limit int) ([]*models.AccountGenerationStatus, error)
	UpdateTotalRequested(statusID uint, totalRequested int) error
	IncrementQueued(statusID uint) error
	IncrementProcessing(statusID uint) error
	IncrementCompleted(statusID uint) error
	IncrementFailed(statusID uint) error
	CompleteGeneration(statusID uint) error

	// Cleanup and recovery
	ReconcileStuckStatuses(stuckDuration time.Duration) (int, error)

	// Helper methods that find active status by account and content type
	IncrementProcessingByAccount(accountID uint, contentType enums.ContentType) error
}

type generationStatusService struct {
	*Service
	generationLockService GenerationLockService
}

func NewGenerationStatusService(
	container *Service,
	generationLockService GenerationLockService,
) GenerationStatusService {
	return &generationStatusService{
		Service:               container,
		generationLockService: generationLockService,
	}
}

func (s *generationStatusService) CreateStatus(accountID uint, contentType enums.ContentType, totalRequested int) (*models.AccountGenerationStatus, error) {
	status := &models.AccountGenerationStatus{
		AccountID:      accountID,
		ContentType:    contentType,
		TotalRequested: totalRequested,
		Status:         enums.GenerationStatusQueuing,
	}

	if err := s.store.GenerationStatus.Create(status); err != nil {
		return nil, fmt.Errorf("failed to create generation status: %w", err)
	}

	s.logger.Infow("Generation status created",
		"status_id", status.ID,
		"account_id", accountID,
		"content_type", contentType,
		"total_requested", totalRequested,
	)

	return status, nil
}

func (s *generationStatusService) SetLockID(statusID uint, lockID string) error {
	status, err := s.store.GenerationStatus.GetByID(statusID)
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	status.LockID = &lockID
	if err := s.store.GenerationStatus.Update(status); err != nil {
		return fmt.Errorf("failed to set lock ID: %w", err)
	}

	s.logger.Infow("Set lock ID for generation status",
		"status_id", statusID,
		"lock_id", lockID,
	)

	return nil
}

// SetErrorCode sets the error code for a generation status.
// If errorMessage is provided (non-empty), it will be set; otherwise, any existing error message is preserved.
// This allows setting just the error code without overwriting a previously set detailed error message.
func (s *generationStatusService) SetErrorCode(statusID uint, errorCode enums.GenerationErrorCode, errorMessage string) error {
	status, err := s.store.GenerationStatus.GetByID(statusID)
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	status.ErrorCode = errorCode
	if errorMessage != "" {
		status.ErrorMessage = errorMessage
	}

	if err := s.store.GenerationStatus.Update(status); err != nil {
		return fmt.Errorf("failed to set error code: %w", err)
	}

	s.logger.Infow("Set error code for generation status",
		"status_id", statusID,
		"error_code", errorCode,
		"error_message", errorMessage,
	)

	return nil
}

func (s *generationStatusService) GetActiveStatus(accountID uint, contentType enums.ContentType) (*models.AccountGenerationStatus, error) {
	status, err := s.store.GenerationStatus.GetActiveByAccount(accountID, contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to get active status: %w", err)
	}
	return status, nil
}

func (s *generationStatusService) GetLatestStatus(accountID uint, contentType enums.ContentType) (*models.AccountGenerationStatus, error) {
	status, err := s.store.GenerationStatus.GetLatestByAccount(accountID, contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest status: %w", err)
	}
	return status, nil
}

func (s *generationStatusService) UpdateTotalRequested(statusID uint, totalRequested int) error {
	status, err := s.store.GenerationStatus.GetByID(statusID)
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	status.TotalRequested = totalRequested
	if err := s.store.GenerationStatus.Update(status); err != nil {
		return fmt.Errorf("failed to update total requested: %w", err)
	}

	s.logger.Infow("Updated generation status total_requested",
		"status_id", statusID,
		"total_requested", totalRequested,
	)

	return nil
}

func (s *generationStatusService) IncrementQueued(statusID uint) error {
	if err := s.store.GenerationStatus.IncrementQueued(statusID); err != nil {
		return fmt.Errorf("failed to increment queued count: %w", err)
	}

	// Update status to processing if this is the first queued item
	status, err := s.store.GenerationStatus.GetByID(statusID)
	if err == nil {
		status.UpdateProgress()
		s.store.GenerationStatus.Update(status)
	}

	return nil
}

func (s *generationStatusService) IncrementProcessing(statusID uint) error {
	if err := s.store.GenerationStatus.IncrementProcessing(statusID); err != nil {
		return fmt.Errorf("failed to increment processing count: %w", err)
	}

	// Update status to processing and set StartedAt if this is the first processing task
	status, err := s.store.GenerationStatus.GetByID(statusID)
	if err == nil {
		if status.StartedAt == nil && status.TotalProcessing > 0 {
			now := time.Now()
			status.StartedAt = &now
		}
		status.UpdateProgress()
		s.store.GenerationStatus.Update(status)
	}

	return nil
}

func (s *generationStatusService) IncrementCompleted(statusID uint) error {
	if err := s.store.GenerationStatus.IncrementCompleted(statusID); err != nil {
		return fmt.Errorf("failed to increment completed count: %w", err)
	}

	// Check if generation is complete
	status, err := s.store.GenerationStatus.GetByID(statusID)
	if err == nil {
		status.UpdateProgress()
		s.store.GenerationStatus.Update(status)

		if status.IsComplete() {
			s.CompleteGeneration(statusID)
		}
	}

	return nil
}

func (s *generationStatusService) IncrementFailed(statusID uint) error {
	if err := s.store.GenerationStatus.IncrementFailed(statusID); err != nil {
		return fmt.Errorf("failed to increment failed count: %w", err)
	}

	// Check if generation is complete
	status, err := s.store.GenerationStatus.GetByID(statusID)
	if err == nil {
		status.UpdateProgress()
		s.store.GenerationStatus.Update(status)

		if status.IsComplete() {
			s.CompleteGeneration(statusID)
		}
	}

	return nil
}

func (s *generationStatusService) IncrementProcessingByAccount(accountID uint, contentType enums.ContentType) error {
	// Find active generation status for this account and content type
	status, err := s.store.GenerationStatus.GetActiveByAccount(accountID, contentType)
	if err != nil {
		// No active status found - this might be a legacy task
		return nil
	}

	return s.IncrementProcessing(status.ID)
}

func (s *generationStatusService) CompleteGeneration(statusID uint) error {
	status, err := s.store.GenerationStatus.GetByID(statusID)
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	// Release the lock if one exists
	if status.LockID != nil && *status.LockID != "" {
		if err := s.generationLockService.ReleaseLock(*status.LockID); err != nil {
			s.logger.Warnw("Failed to release lock on generation completion",
				"status_id", statusID,
				"lock_id", *status.LockID,
				"error", err,
			)
		} else {
			s.logger.Infow("Released lock on generation completion",
				"status_id", statusID,
				"lock_id", *status.LockID,
			)
		}
	}

	if err := s.store.GenerationStatus.MarkCompleted(statusID); err != nil {
		return fmt.Errorf("failed to mark generation as completed: %w", err)
	}

	s.logger.Infow("Generation completed",
		"status_id", statusID,
		"account_id", status.AccountID,
		"total_completed", status.TotalCompleted,
		"total_failed", status.TotalFailed,
		"final_status", status.Status,
	)

	return nil
}

func (s *generationStatusService) GetByAccountID(accountID uint, limit int) ([]*models.AccountGenerationStatus, error) {
	// Get all statuses for this account (all content types)
	allStatuses := make([]*models.AccountGenerationStatus, 0)

	// Query for video
	videoStatuses, err := s.store.GenerationStatus.GetByAccountAndType(accountID, enums.ContentTypeVideo)
	if err == nil {
		allStatuses = append(allStatuses, videoStatuses...)
	}

	// Query for story
	storyStatuses, err := s.store.GenerationStatus.GetByAccountAndType(accountID, enums.ContentTypeStory)
	if err == nil {
		allStatuses = append(allStatuses, storyStatuses...)
	}

	// Query for slideshow
	slideshowStatuses, err := s.store.GenerationStatus.GetByAccountAndType(accountID, enums.ContentTypeSlideshow)
	if err == nil {
		allStatuses = append(allStatuses, slideshowStatuses...)
	}

	// Sort by started_at DESC
	sort.Slice(allStatuses, func(i, j int) bool {
		if allStatuses[i].StartedAt == nil {
			return false
		}
		if allStatuses[j].StartedAt == nil {
			return true
		}
		return allStatuses[i].StartedAt.After(*allStatuses[j].StartedAt)
	})

	// Apply limit
	if len(allStatuses) > limit {
		allStatuses = allStatuses[:limit]
	}

	return allStatuses, nil
}

// ReconcileStuckStatuses finds and fixes generation statuses that are stuck
// in queuing or processing state for longer than the specified duration
func (s *generationStatusService) ReconcileStuckStatuses(stuckDuration time.Duration) (int, error) {
	s.logger.Infow("Starting reconciliation of stuck generation statuses", "stuck_duration", stuckDuration)

	// Find stuck statuses
	// Note: There's a potential race condition where counters could change between
	// reading the status and applying fixes. This is acceptable because:
	// 1. GetStuckStatuses uses updated_at, so only truly idle statuses are selected
	// 2. Active tasks would update the status, moving it out of stuck state
	// 3. Worst case: the fix is applied based on slightly stale data, but the next
	//    reconciliation (30 min later) will correct any overcorrection
	// 4. The alternative (row-level locking) would block active generation unnecessarily
	stuckStatuses, err := s.store.GenerationStatus.GetStuckStatuses(stuckDuration)
	if err != nil {
		return 0, fmt.Errorf("failed to get stuck statuses: %w", err)
	}

	if len(stuckStatuses) == 0 {
		s.logger.Info("No stuck generation statuses found")
		return 0, nil
	}

	s.logger.Infow("Found stuck generation statuses", "count", len(stuckStatuses))
	fixedCount := 0

	for _, status := range stuckStatuses {
		s.logger.Infow("Analyzing stuck status",
			"status_id", status.ID,
			"account_id", status.AccountID,
			"content_type", status.ContentType,
			"status", status.Status,
			"total_requested", status.TotalRequested,
			"total_queued", status.TotalQueued,
			"total_processing", status.TotalProcessing,
			"total_completed", status.TotalCompleted,
			"total_failed", status.TotalFailed,
			"created_at", status.CreatedAt,
		)

		// Check if counters are out of sync
		totalProcessed := status.TotalCompleted + status.TotalFailed

		// Case 1: All tasks are accounted for but status field not updated to completed/failed/partial
		if totalProcessed >= status.TotalRequested &&
			status.Status != enums.GenerationStatusCompleted &&
			status.Status != enums.GenerationStatusFailed &&
			status.Status != enums.GenerationStatusPartial {
			s.logger.Warnw("Found status with all tasks processed but status field not updated",
				"status_id", status.ID,
				"total_requested", status.TotalRequested,
				"total_processed", totalProcessed,
				"current_status", status.Status,
			)

			// Update progress to ensure status field reflects completion state
			status.UpdateProgress()
			if err := s.store.GenerationStatus.Update(status); err != nil {
				s.logger.Errorw("Failed to update status progress",
					"status_id", status.ID,
					"error", err,
				)
				continue
			}

			// Force completion
			if err := s.CompleteGeneration(status.ID); err != nil {
				s.logger.Errorw("Failed to complete stuck status",
					"status_id", status.ID,
					"error", err,
				)
				continue
			}
			fixedCount++
			continue
		}

		// Case 2: Status has been processing for too long with no tasks in processing state
		if status.TotalProcessing == 0 && status.TotalQueued == 0 && totalProcessed < status.TotalRequested {
			s.logger.Warnw("Found status stuck with no active tasks",
				"status_id", status.ID,
				"total_requested", status.TotalRequested,
				"total_processed", totalProcessed,
			)

			// Mark remaining tasks as failed and complete the generation
			missingCount := status.TotalRequested - totalProcessed
			status.TotalFailed += missingCount
			status.UpdateProgress()

			if err := s.store.GenerationStatus.Update(status); err != nil {
				s.logger.Errorw("Failed to update stuck status counters",
					"status_id", status.ID,
					"error", err,
				)
				continue
			}

			// Mark as completed
			if err := s.CompleteGeneration(status.ID); err != nil {
				s.logger.Errorw("Failed to complete stuck status",
					"status_id", status.ID,
					"error", err,
				)
				continue
			}
			fixedCount++
			continue
		}

		// Case 3: Counters don't add up correctly (processing counter is stuck)
		totalAccounted := status.TotalQueued + status.TotalProcessing + status.TotalCompleted + status.TotalFailed
		if totalAccounted > status.TotalRequested {
			s.logger.Warnw("Found status with invalid counters",
				"status_id", status.ID,
				"total_requested", status.TotalRequested,
				"total_accounted", totalAccounted,
			)

			// Fix the counters by moving excess to failed
			// Handle case where excess spans multiple counters
			excess := totalAccounted - status.TotalRequested

			// First, try to remove from TotalProcessing
			if excess > 0 && status.TotalProcessing > 0 {
				removeFromProcessing := excess
				if status.TotalProcessing < excess {
					removeFromProcessing = status.TotalProcessing
				}
				status.TotalProcessing -= removeFromProcessing
				status.TotalFailed += removeFromProcessing
				excess -= removeFromProcessing
			}

			// Then remove from TotalQueued if needed
			if excess > 0 && status.TotalQueued > 0 {
				removeFromQueued := excess
				if status.TotalQueued < excess {
					removeFromQueued = status.TotalQueued
				}
				status.TotalQueued -= removeFromQueued
				status.TotalFailed += removeFromQueued
				excess -= removeFromQueued
			}

			status.UpdateProgress()

			if err := s.store.GenerationStatus.Update(status); err != nil {
				s.logger.Errorw("Failed to fix status counters",
					"status_id", status.ID,
					"error", err,
				)
				continue
			}

			// Check if it should be completed now
			if status.IsComplete() {
				if err := s.CompleteGeneration(status.ID); err != nil {
					s.logger.Errorw("Failed to complete fixed status",
						"status_id", status.ID,
						"error", err,
					)
					// Don't continue here - the counters were fixed even if completion failed
					// This will be caught in the next reconciliation
				}
			}
			// Only increment if we successfully updated the counters
			fixedCount++
		}
	}

	s.logger.Infow("Reconciliation complete", "fixed_count", fixedCount, "total_stuck", len(stuckStatuses))
	return fixedCount, nil
}
