package jobs

import (
	"fmt"
	"log"

	"github.com/nicolailuther/butter/internal/models"
	"github.com/nicolailuther/butter/internal/store"
	"github.com/nicolailuther/butter/pkg/onlyfans"
)

type TrackOnlyfansLinks struct {
	store                  *store.Store
	onlyfansServiceGateway onlyfans.OnlyfansServiceGateway
}

func NewTrackOnlyfansLinksTask(
	store *store.Store,
	onlyfansServiceGateway onlyfans.OnlyfansServiceGateway,
) Job {
	return &TrackOnlyfansLinks{
		store,
		onlyfansServiceGateway,
	}
}

// Execute synchronizes tracking links for all OnlyFans accounts.
// It retrieves all OnlyFans accounts from the repository, fetches their current tracking links
// from the external OnlyFans service, and compares them with the tracking links stored in the database.
// For each link, it determines whether to insert a new record or update an existing one based on the presence
// of the link in the database. The method prepares a list of links to upsert, ensuring that the database reflects
// the latest state from the OnlyFans service. Returns an error if any step in the process fails.
func (t *TrackOnlyfansLinks) Execute() error {

	accounts, err := t.store.OnlyfansAccounts.GetAll()
	if err != nil {
		return err
	}

	log.Println("Found", len(accounts), "OnlyFans accounts to track links for")

	for i, account := range accounts {

		fmt.Printf("Processing account %d/%d: ID#%d\n", i+1, len(accounts), account.ID)

		links, err := t.onlyfansServiceGateway.GetAccountTrackingLinks(account.ExternalID)
		if err != nil {
			// Continue processing other accounts even if one fails
			fmt.Printf("Error fetching tracking links for account ID#%d: %v\n", account.ID, err)
			continue
		}

		client, err := t.store.Clients.GetByID(account.ClientID)
		if err != nil {
			fmt.Printf("failed to get client for account ID#%d: %v\n", account.ID, err)
			continue
		}

		if client == nil {
			fmt.Printf("No client found for account ID#%d, skipping...\n", account.ID)
			continue
		}

		var trackingLinks []*models.OnlyfansTrackingLink
		for _, link := range links {
			trackingLinks = append(trackingLinks, &models.OnlyfansTrackingLink{
				ExternalID:        link.ID,
				Name:              link.Name,
				Url:               link.Url,
				Clicks:            link.Clicks,
				Subscribers:       link.Subscribers,
				Revenue:           link.Revenue,
				ClientID:          client.ID,
				OnlyfansAccountID: account.ID,
			})
		}

		fmt.Println("Upserting", len(trackingLinks), "tracking links for account ID#", account.ID)
		if err := t.store.OnlyfansLinks.UpsertLinks(trackingLinks); err != nil {
			fmt.Printf("Error upserting tracking links for account ID#%d: %v\n", account.ID, err)
			continue
		}
	}

	return nil
}

func (t *TrackOnlyfansLinks) GetName() TaskLabel {
	return TaskTrackOnlyfansLinks
}
