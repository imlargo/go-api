package jobs

import (
	"fmt"
	"log"
	"time"

	"github.com/nicolailuther/butter/internal/dto"
	"github.com/nicolailuther/butter/internal/enums"
	"github.com/nicolailuther/butter/internal/services"
	"github.com/nicolailuther/butter/internal/store"
)

type AutoGenerateContentTask struct {
	store          *store.Store
	contentService services.ContentService
}

func NewAutoGenerateContentTask(
	store *store.Store,
	contentService services.ContentService,
) Job {
	return &AutoGenerateContentTask{
		store:          store,
		contentService: contentService,
	}
}

// Execute runs the auto content generation task.
// It checks for accounts with auto-generation enabled and scheduled for the current hour,
// then triggers content generation for each matching account.
//
// NOTE: Uses server time (timezone) for hour matching. The `auto_generate_hour` field is compared
// against the current hour in the server's local time zone. This means all scheduling is implicitly
// done in the server's timezone, not per-account timezone.
// Future enhancement: add timezone support per account.
func (a *AutoGenerateContentTask) Execute() error {
	currentHour := time.Now().Hour()
	log.Printf("Auto-generate content task started (current hour: %d)\n", currentHour)

	// Get all accounts with auto-generation enabled and scheduled for this hour
	accounts, err := a.store.Accounts.GetAccountsForAutoGeneration(currentHour)
	if err != nil {
		return fmt.Errorf("error fetching accounts for auto-generation: %w", err)
	}

	log.Printf("Found %d accounts scheduled for auto-generation at hour %d\n", len(accounts), currentHour)

	if len(accounts) == 0 {
		log.Println("No accounts require auto-generation at this time")
		return nil
	}

	successCount := 0
	errorCount := 0

	for i, account := range accounts {
		log.Printf("[%d/%d] Processing account ID#%d (%s - %s)\n",
			i+1, len(accounts), account.ID, account.Username, account.Platform)

		// Count existing generated content to calculate remaining quantity needed
		existingCount, err := a.store.GeneratedContents.CountByAccountAndType(account.ID, enums.ContentTypeVideo)
		if err != nil {
			log.Printf("Error counting generated content for account ID#%d (%s): %v\n", account.ID, account.Username, err)
			errorCount++
			continue
		}

		// Calculate remaining quantity needed to reach posting goal
		remainingQuantity := account.PostingGoal - existingCount
		if remainingQuantity <= 0 {
			log.Printf("Account ID#%d (%s) already has %d generated contents, posting goal (%d) already met. Skipping.\n",
				account.ID, account.Username, existingCount, account.PostingGoal)
			continue
		}

		log.Printf("Account ID#%d (%s): existing=%d, posting_goal=%d, generating=%d\n",
			account.ID, account.Username, existingCount, account.PostingGoal, remainingQuantity)

		// Generate content for the account
		// Only generate the remaining quantity needed to complete the posting goal
		generateRequest := &dto.GenerateContent{
			AccountID: account.ID,
			Type:      enums.ContentTypeVideo, // Default to video content
			Quantity:  remainingQuantity,
		}

		_, err = a.contentService.GenerateContent(generateRequest)
		if err != nil {
			log.Printf("Error generating content for account ID#%d (%s): %v\n", account.ID, account.Username, err)
			errorCount++
			continue
		}

		log.Printf("Successfully triggered content generation for account ID#%d (%s)\n", account.ID, account.Username)
		successCount++
	}

	log.Printf("Auto-generate content task completed: %d successful, %d errors\n", successCount, errorCount)

	if errorCount > 0 {
		return fmt.Errorf("completed with %d errors out of %d accounts", errorCount, len(accounts))
	}

	return nil
}

func (a *AutoGenerateContentTask) GetName() TaskLabel {
	return TaskAutoGenerateContent
}
