package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/nicolailuther/butter/internal/dto"
	"github.com/nicolailuther/butter/internal/enums"
	"github.com/nicolailuther/butter/internal/models"
	"github.com/nicolailuther/butter/pkg/taskqueue"
	"gorm.io/gorm"
)

// ConcurrentContentGenerationService extends ContentGenerationService with concurrent capabilities
type ConcurrentContentGenerationService interface {
	ContentGenerationService
	GenerateContentConcurrent(request *dto.GenerateContent, taskManager taskqueue.TaskManager) ([]*models.GeneratedContent, error)
}

type concurrentGenerationService struct {
	*Service
	*GenerationServiceImpl
	lockService   GenerationLockService
	statusService GenerationStatusService
}

func NewConcurrentContentGenerationService(
	container *Service,
	baseService *GenerationServiceImpl,
	lockService GenerationLockService,
	statusService GenerationStatusService,
) ConcurrentContentGenerationService {
	return &concurrentGenerationService{
		Service:               container,
		GenerationServiceImpl: baseService,
		lockService:           lockService,
		statusService:         statusService,
	}
}

// GenerateContentConcurrent generates content concurrently using task queue
// Note: This function returns an empty slice as GeneratedContent records are created
// asynchronously when tasks complete. To track the progress and completion of content generation,
// use GenerationStatusService to monitor status updates, or subscribe to task events if available.
func (s *concurrentGenerationService) GenerateContentConcurrent(
	request *dto.GenerateContent,
	taskManager taskqueue.TaskManager,
) ([]*models.GeneratedContent, error) {

	account, err := s.store.Accounts.GetByID(request.AccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	if err := s.validateGenerationLimits(account, request.Type, request.Quantity); err != nil {
		return nil, err
	}

	if err := s.validatePlatform(account.Platform); err != nil {
		return nil, err
	}

	// Acquire generation lock to prevent concurrent generation
	// Lock will be released when generation status is marked as completed
	lockID, err := s.lockService.AcquireLock(request.AccountID, request.Type)
	if err != nil {
		return nil, err
	}

	// Create generation status record
	generationStatus, err := s.statusService.CreateStatus(request.AccountID, request.Type, request.Quantity)
	if err != nil {
		// Release lock if status creation fails
		s.lockService.ReleaseLock(lockID)
		return nil, fmt.Errorf("failed to create generation status: %w", err)
	}

	// Store lock ID in generation status so it can be released when generation completes
	if err := s.statusService.SetLockID(generationStatus.ID, lockID); err != nil {
		// Critical error: lock cannot be tracked, release immediately to avoid orphan
		s.lockService.ReleaseLock(lockID)
		return nil, fmt.Errorf("failed to set lock ID in status: %w", err)
	}

	strategy, err := s.getStrategy(request.Type)
	if err != nil {
		// Set error code based on the error type
		if err == ErrUnsupportedType {
			if setErr := s.statusService.SetErrorCode(generationStatus.ID, enums.GenerationErrorCodeUnsupportedType, err.Error()); setErr != nil {
				log.Printf("[Concurrent Generation] Warning: Failed to set error code: %v", setErr)
			}
		} else {
			if setErr := s.statusService.SetErrorCode(generationStatus.ID, enums.GenerationErrorCodeUnknown, err.Error()); setErr != nil {
				log.Printf("[Concurrent Generation] Warning: Failed to set error code: %v", setErr)
			}
		}

		// Mark all requested tasks as failed before completing generation
		// This ensures UpdateProgress() correctly sets status to "failed"
		s.markAllTasksAsFailed(generationStatus.ID, request.Quantity)

		// Mark generation as failed and release lock
		if completeErr := s.statusService.CompleteGeneration(generationStatus.ID); completeErr != nil {
			log.Printf("[Concurrent Generation] Error: Failed to complete generation after strategy error: %v", completeErr)
		}
		return nil, err
	}

	// Phase 1: Select all content to generate (with atomic updates)
	log.Printf("[Concurrent Generation] Selecting %d pieces of content for account %d", request.Quantity, request.AccountID)
	contentAccountsToGenerate, err := s.selectContentForGeneration(account, request, strategy, generationStatus.ID)
	if err != nil {
		// Mark all requested tasks as failed before completing generation
		// This ensures UpdateProgress() correctly sets status to "failed"
		s.markAllTasksAsFailed(generationStatus.ID, request.Quantity)

		// Mark generation as failed and release lock
		if completeErr := s.statusService.CompleteGeneration(generationStatus.ID); completeErr != nil {
			log.Printf("[Concurrent Generation] Error: Failed to complete generation after selection error: %v", completeErr)
		}
		return nil, fmt.Errorf("failed to select content: %w", err)
	}

	if len(contentAccountsToGenerate) == 0 {
		// Set error code for no content available
		if setErr := s.statusService.SetErrorCode(generationStatus.ID, enums.GenerationErrorCodeNoContentAvailable, ErrNoContentAvailable.Error()); setErr != nil {
			log.Printf("[Concurrent Generation] Warning: Failed to set error code: %v", setErr)
		}

		// Mark all requested tasks as failed before completing generation
		// This ensures UpdateProgress() correctly sets status to "failed"
		s.markAllTasksAsFailed(generationStatus.ID, request.Quantity)

		// Mark generation as failed and release lock
		if completeErr := s.statusService.CompleteGeneration(generationStatus.ID); completeErr != nil {
			log.Printf("[Concurrent Generation] Error: Failed to complete generation when no content available: %v", completeErr)
		}
		return nil, ErrNoContentAvailable
	}

	log.Printf("[Concurrent Generation] Selected %d pieces of content, submitting to task queue", len(contentAccountsToGenerate))

	// Update total_requested to match actual selection if fewer items were available
	if len(contentAccountsToGenerate) < request.Quantity {
		log.Printf("[Concurrent Generation] Warning: Only %d/%d content items available, updating status",
			len(contentAccountsToGenerate), request.Quantity)

		// Update the generation status with actual count
		if err := s.statusService.UpdateTotalRequested(generationStatus.ID, len(contentAccountsToGenerate)); err != nil {
			log.Printf("[Concurrent Generation] Warning: Failed to update generation status with actual count: %v", err)
		}
	}

	// Phase 2: Submit all selected content to task queue concurrently
	return s.submitContentToQueue(account, request, strategy, contentAccountsToGenerate, taskManager, generationStatus.ID)
}

// selectContentForGeneration selects N pieces of content atomically
func (s *concurrentGenerationService) selectContentForGeneration(
	account *models.Account,
	request *dto.GenerateContent,
	strategy contentStrategy,
	statusID uint,
) ([]*models.ContentAccount, error) {

	var contentAccounts []*models.ContentAccount

	for i := 0; i < request.Quantity; i++ {
		// Select next content using strategy
		contentAccount, err := strategy.SelectContent(account, request)
		if err != nil || contentAccount == nil {
			log.Printf("[Concurrent Generation] Could not select content %d/%d: %v", i+1, request.Quantity, err)
			break
		}

		// CRITICAL: Mark as selected immediately to prevent re-selection
		if err := s.markContentAsSelected(contentAccount); err != nil {
			log.Printf("[Concurrent Generation] Failed to mark content as selected: %v", err)
			continue
		}

		contentAccounts = append(contentAccounts, contentAccount)
		log.Printf("[Concurrent Generation] Selected content %d/%d: content_id=%d",
			i+1, request.Quantity, contentAccount.ContentID)
	}

	return contentAccounts, nil
}

// markContentAsSelected atomically marks content as being generated
func (s *concurrentGenerationService) markContentAsSelected(contentAccount *models.ContentAccount) error {
	return s.store.DB.Transaction(func(tx *gorm.DB) error {
		now := time.Now()

		err := tx.Model(&models.ContentAccount{}).
			Where("id = ?", contentAccount.ID).
			Updates(map[string]interface{}{
				"times_generated":   gorm.Expr("times_generated + 1"),
				"last_generated_at": now,
			}).Error
		if err != nil {
			return fmt.Errorf("failed to mark content as selected: %w", err)
		}

		// Also update the Content table
		if err := tx.Model(&models.Content{}).
			Where("id = ?", contentAccount.ContentID).
			Updates(map[string]interface{}{
				"times_generated":   gorm.Expr("times_generated + 1"),
				"last_generated_at": now,
			}).Error; err != nil {
			return fmt.Errorf("failed to update content: %w", err)
		}

		return nil
	})
}

// submitContentToQueue submits all selected content to the task queue sequentially
// For video content, it uses the task queue (repurpose engine).
// For story and slideshow content, it generates immediately using the strategy's GenerateContent method.
func (s *concurrentGenerationService) submitContentToQueue(
	account *models.Account,
	request *dto.GenerateContent,
	strategy contentStrategy,
	contentAccounts []*models.ContentAccount,
	taskManager taskqueue.TaskManager,
	statusID uint,
) ([]*models.GeneratedContent, error) {

	var generatedContent []*models.GeneratedContent
	var submissionErrors []error
	successCount := 0

	// For non-video types (story/slideshow), set StartedAt timestamp since generation begins immediately
	if request.Type != enums.ContentTypeVideo {
		status, err := s.store.GenerationStatus.GetByID(statusID)
		if err == nil && status.StartedAt == nil {
			now := time.Now()
			status.StartedAt = &now
			status.UpdateProgress()
			s.store.GenerationStatus.Update(status)
		}
	}

	for _, contentAccount := range contentAccounts {
		// Get populated content
		content, err := s.store.Contents.GetPopulated(contentAccount.ContentID)
		if err != nil {
			log.Printf("[Concurrent Generation] Failed to get content %d: %v", contentAccount.ContentID, err)
			s.rollbackContentSelection(contentAccount)
			submissionErrors = append(submissionErrors, fmt.Errorf("failed to get content %d: %w", contentAccount.ContentID, err))

			// For story/slideshow, increment failed count to ensure proper status tracking and lock release
			if request.Type != enums.ContentTypeVideo {
				if err := s.statusService.IncrementFailed(statusID); err != nil {
					log.Printf("[Concurrent Generation] Warning: Failed to increment failed count: %v", err)
				}
			}
			continue
		}

		// Get text overlay if needed
		textOverlay, err := s.getTextOverlayIfNeeded(account, content)
		if err != nil {
			log.Printf("[Concurrent Generation] Failed to get text overlay: %v", err)
			s.rollbackContentSelection(contentAccount)
			submissionErrors = append(submissionErrors, fmt.Errorf("failed to get text overlay for content %d: %w", contentAccount.ContentID, err))

			// For story/slideshow, increment failed count to ensure proper status tracking and lock release
			if request.Type != enums.ContentTypeVideo {
				if err := s.statusService.IncrementFailed(statusID); err != nil {
					log.Printf("[Concurrent Generation] Warning: Failed to increment failed count: %v", err)
				}
			}
			continue
		}

		// Different handling based on content type:
		// - Video: Uses repurpose engine (task queue)
		// - Story/Slideshow: Generate immediately (no processing needed)
		if request.Type == enums.ContentTypeVideo {
			// VIDEO: Submit to task queue for repurpose engine processing
			taskRequest := s.prepareTaskRequest(account, request, contentAccount, content, textOverlay)
			if taskRequest == nil {
				log.Printf("[Concurrent Generation] Failed to prepare task request: no content files available")
				s.rollbackContentSelection(contentAccount)
				submissionErrors = append(submissionErrors, fmt.Errorf("no content files available for content %d", contentAccount.ContentID))
				continue
			}

			taskID, err := taskManager.SubmitTask(context.Background(), taskRequest)
			if err != nil {
				log.Printf("[Concurrent Generation] Failed to submit task: %v", err)
				s.rollbackContentSelection(contentAccount)
				submissionErrors = append(submissionErrors, fmt.Errorf("failed to submit task for content %d: %w", contentAccount.ContentID, err))
				continue
			}

			// Increment queued count for video tasks
			if err := s.statusService.IncrementQueued(statusID); err != nil {
				log.Printf("[Concurrent Generation] Warning: Failed to increment queued count: %v", err)
			}

			log.Printf("[Concurrent Generation] Video task submitted: task_id=%s, content_id=%d",
				taskID, contentAccount.ContentID)
		} else {
			// Increment queued count for video tasks
			if err := s.statusService.IncrementQueued(statusID); err != nil {
				log.Printf("[Concurrent Generation] Warning: Failed to increment queued count: %v", err)
			}

			// Increment processing count for video tasks
			if err := s.statusService.IncrementProcessing(statusID); err != nil {
				log.Printf("[Concurrent Generation] Warning: Failed to increment queued count: %v", err)
			}

			// STORY/SLIDESHOW: Generate immediately using strategy
			generated, err := strategy.GenerateContent(account, request, contentAccount, content, textOverlay)
			if err != nil {
				log.Printf("[Concurrent Generation] Failed to generate %s content %d: %v", request.Type, contentAccount.ContentID, err)
				s.rollbackContentSelection(contentAccount)
				submissionErrors = append(submissionErrors, fmt.Errorf("failed to generate %s content %d: %w", request.Type, contentAccount.ContentID, err))

				// Increment failed count
				if err := s.statusService.IncrementFailed(statusID); err != nil {
					log.Printf("[Concurrent Generation] Warning: Failed to increment failed count: %v", err)
				}
				continue
			}

			// Increment completed count for immediate generation
			if err := s.statusService.IncrementCompleted(statusID); err != nil {
				log.Printf("[Concurrent Generation] Warning: Failed to increment completed count: %v", err)
			}

			generatedContent = append(generatedContent, generated)
			log.Printf("[Concurrent Generation] %s content generated immediately: content_id=%d, generated_id=%d",
				request.Type, contentAccount.ContentID, generated.ID)
		}

		successCount++
	}

	log.Printf("[Concurrent Generation] Submission complete for account %d: %d succeeded, %d failed",
		request.AccountID, successCount, len(submissionErrors))

	// If all submissions failed, return an error
	if successCount == 0 && len(submissionErrors) > 0 {
		return nil, fmt.Errorf("all task submissions failed: first error: %w", submissionErrors[0])
	}

	// Log any partial failures
	if len(submissionErrors) > 0 {
		log.Printf("[Concurrent Generation] Warning: %d/%d tasks failed to submit", len(submissionErrors), len(contentAccounts))
	}

	return generatedContent, nil
}

// rollbackContentSelection reverts the selection mark if task submission fails
func (s *concurrentGenerationService) rollbackContentSelection(contentAccount *models.ContentAccount) {
	tx := s.store.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Use GREATEST to prevent times_generated from going negative
	err := tx.Model(&models.ContentAccount{}).
		Where("id = ?", contentAccount.ID).
		Updates(map[string]interface{}{
			"times_generated": gorm.Expr("GREATEST(times_generated - ?, 0)", 1), // Safe atomic decrement
		}).Error

	if err != nil {
		log.Printf("[Concurrent Generation] Failed to rollback content account: %v", err)
		tx.Rollback()
		return
	}

	// Also rollback the Content table increment with safety check
	if err := tx.Model(&models.Content{}).
		Where("id = ?", contentAccount.ContentID).
		Update("times_generated", gorm.Expr("GREATEST(times_generated - ?, 0)", 1)).Error; err != nil {
		log.Printf("[Concurrent Generation] Failed to rollback content: %v", err)
		tx.Rollback()
		return
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("[Concurrent Generation] Failed to commit rollback transaction: %v", err)
		return
	}

	log.Printf("[Concurrent Generation] Rolled back selection for content_account_id=%d", contentAccount.ID)
}

// markAllTasksAsFailed marks all requested tasks as failed to ensure proper status tracking
// This is necessary before calling CompleteGeneration() when an error occurs early in the process
func (s *concurrentGenerationService) markAllTasksAsFailed(statusID uint, quantity int) {
	for i := 0; i < quantity; i++ {
		if err := s.statusService.IncrementFailed(statusID); err != nil {
			log.Printf("[Concurrent Generation] Warning: Failed to increment failed count: %v", err)
		}
	}
}

// prepareTaskRequest creates a task request from generation parameters
func (s *concurrentGenerationService) prepareTaskRequest(
	account *models.Account,
	request *dto.GenerateContent,
	contentAccount *models.ContentAccount,
	content *models.Content,
	textOverlay *models.TextOverlay,
) *dto.ReporpuseVideo {

	if len(content.ContentFiles) == 0 {
		return nil
	}

	sourceFile := content.ContentFiles[0].File
	isMainAccount := account.AccountRole == enums.AccountRoleMain
	useMirror := shouldMirror(content.UseMirror, contentAccount.TimesGenerated)

	var overlayContent string
	var useOverlay bool
	var textOverlayID uint
	if textOverlay != nil {
		overlayContent = textOverlay.Content
		useOverlay = textOverlay.Content != ""
		textOverlayID = textOverlay.ID
	}

	return &dto.ReporpuseVideo{
		FileID:           sourceFile.ID,
		AccountID:        account.ID,
		ContentID:        content.ID,
		ContentAccountID: contentAccount.ID,
		ContentType:      string(request.Type), // CRITICAL: Set content type for proper status updates
		TextOverlay:      overlayContent,
		MainAccount:      isMainAccount,
		TextOverlayID:    textOverlayID,
		// Set legacy fields for compatibility
		UseMirror:   useMirror,
		UseOverlays: useOverlay,
		IsMain:      isMainAccount,
	}
}
