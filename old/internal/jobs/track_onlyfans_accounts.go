package jobs

import (
	"fmt"
	"log"

	"github.com/nicolailuther/butter/internal/models"
	"github.com/nicolailuther/butter/internal/store"
	"github.com/nicolailuther/butter/pkg/onlyfans"
)

type TrackOnlyfansAccounts struct {
	store                  *store.Store
	onlyfansServiceGateway onlyfans.OnlyfansServiceGateway
}

func NewTrackOnlyfansAccountsTask(
	store *store.Store,
	onlyfansServiceGateway onlyfans.OnlyfansServiceGateway,
) Job {
	return &TrackOnlyfansAccounts{
		store:                  store,
		onlyfansServiceGateway: onlyfansServiceGateway,
	}
}

// TrackOnlyfansAccounts is a job responsible for synchronizing OnlyFans account data.
// It retrieves all OnlyFans accounts from the repository, fetches the latest account data
// from the external OnlyFans service, and updates the local database to reflect any changes.
// The job processes each account individually, updating fields such as username, name, and
// subscriber count. If an error occurs while fetching data for a specific account, it logs
// the error and continues processing the remaining accounts. After collecting all updated
// accounts, it performs a batch update in the repository. This job ensures that the local
// representation of OnlyFans accounts remains consistent with the external service.
func (t *TrackOnlyfansAccounts) Execute() error {
	accounts, err := t.store.OnlyfansAccounts.GetAll()
	if err != nil {
		return err
	}

	log.Println("Found", len(accounts), "OnlyFans accounts to track data for")

	var toUpdateAccounts []*models.OnlyfansAccount
	for i, account := range accounts {
		fmt.Printf("Processing account %d/%d: ID#%d\n", i+1, len(accounts), account.ID)

		accountData, err := t.onlyfansServiceGateway.GetAccountData(account.ExternalID)
		if err != nil {
			// Continue processing other accounts even if one fails
			fmt.Printf("Error fetching data for account ID#%d: %v\n", account.ID, err)
			continue
		}

		updated := *account
		updated.Username = accountData.Username
		updated.Name = accountData.Name
		updated.Subscribers = accountData.Subscribers

		toUpdateAccounts = append(toUpdateAccounts, &updated)
	}

	log.Println("Found", len(toUpdateAccounts), "accounts to update")
	if len(toUpdateAccounts) > 0 {
		err := t.store.OnlyfansAccounts.UpdateMany(toUpdateAccounts)
		if err != nil {
			return fmt.Errorf("error updating accounts: %w", err)
		}
	}

	return nil
}

func (t *TrackOnlyfansAccounts) GetName() TaskLabel {
	return TaskTrackOnlyfansAccounts
}
