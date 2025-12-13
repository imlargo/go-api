package jobs

import (
	"fmt"
	"log"
	"time"

	"github.com/nicolailuther/butter/internal/services"
	"github.com/nicolailuther/butter/internal/store"
)

type CleanupStuckGeneration struct {
	store                   *store.Store
	generationStatusService services.GenerationStatusService
	generationLockService   services.GenerationLockService
}

func NewCleanupStuckGenerationTask(
	store *store.Store,
	generationStatusService services.GenerationStatusService,
	generationLockService services.GenerationLockService,
) Job {
	return &CleanupStuckGeneration{
		store:                   store,
		generationStatusService: generationStatusService,
		generationLockService:   generationLockService,
	}
}

// CleanupStuckGeneration is a job responsible for detecting and fixing stuck generation statuses
// and cleaning up expired locks. It should run periodically (e.g., every 30 minutes) to ensure
// that generation processes don't get permanently stuck due to counter inconsistencies or
// expired locks.
func (c *CleanupStuckGeneration) Execute() error {
	log.Println("Starting cleanup of stuck generation statuses and expired locks")

	hasErrors := false

	// Step 1: Reconcile stuck statuses (statuses stuck for more than 3 hours)
	stuckDuration := 3 * time.Hour
	fixedStatuses, err := c.generationStatusService.ReconcileStuckStatuses(stuckDuration)
	if err != nil {
		log.Printf("Error reconciling stuck statuses: %v\n", err)
		hasErrors = true
		// Continue to try cleanup, but track the error
	} else {
		log.Printf("Fixed %d stuck generation statuses\n", fixedStatuses)
	}

	// Step 2: Cleanup expired locks (locks held for more than 3 hours)
	lockDuration := 3 * time.Hour
	removedLocks, err := c.generationLockService.CleanupExpiredLocks(lockDuration)
	if err != nil {
		log.Printf("Error cleaning up expired locks: %v\n", err)
		hasErrors = true
	} else {
		log.Printf("Removed %d expired locks\n", removedLocks)
	}

	if hasErrors {
		log.Println("Cleanup completed with errors")
		return fmt.Errorf("cleanup completed with errors")
	}

	log.Println("Cleanup completed successfully")
	return nil
}

func (c *CleanupStuckGeneration) GetName() TaskLabel {
	return TaskCleanupStuckGeneration
}
