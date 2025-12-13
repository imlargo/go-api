package jobs

import (
	"fmt"
	"log"
	"time"

	"github.com/nicolailuther/butter/internal/enums"
	"github.com/nicolailuther/butter/internal/models"
	"github.com/nicolailuther/butter/internal/services"
	"github.com/nicolailuther/butter/pkg/socialmedia"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TrackAccountsAnalytics struct {
	db                 *gorm.DB
	socialMediaGateway socialmedia.SocialMediaService
	fileService        services.FileService
}

type accountUpdate struct {
	account    *models.Account
	updateType updateType
}

// This job tracks the analytics of accounts by fetching their profile data from a social media service
// and updating the account records and analytics data in the database.
func NewTrackAccountsAnalyticsTask(
	db *gorm.DB,
	socialMediaGateway socialmedia.SocialMediaService,
	fileService services.FileService,
) Job {
	return &TrackAccountsAnalytics{
		db:                 db,
		socialMediaGateway: socialMediaGateway,
		fileService:        fileService,
	}
}

func (t *TrackAccountsAnalytics) Execute() error {
	todayDate := time.Now().Format("2006-01-02")
	batchSize := 50

	var accounts []*models.Account
	result := t.db.FindInBatches(&accounts, batchSize, func(tx *gorm.DB, batch int) error {
		fmt.Printf("Processing batch %d with %d records\n", batch, len(accounts))

		var accountUpdates []accountUpdate
		var analyticsData []*models.AccountAnalytic

		for _, account := range accounts {
			accountProfile, err := t.socialMediaGateway.GetProfileData(account.Platform, account.Username)
			if err != nil {
				fmt.Printf("Error fetching profile, account %d - %s: %v\n", account.ID, account.Username, err)
				if t.shouldMarkInactive(account) {
					accountUpdates = append(accountUpdates, accountUpdate{
						account: &models.Account{
							ID:            account.ID,
							UpdatedAt:     time.Now(),
							LastTrackedAt: account.LastTrackedAt,
							Status:        enums.AccountStatusInactive,
						},
						updateType: updateTypeMinimal,
					})
				}
				continue
			}

			// Check if profile is empty or has missing critical fields (username/name)
			// This can happen when an account is suspended, banned, or deleted
			if accountProfile.IsEmpty() || accountProfile.Username == "" {
				fmt.Printf("Empty or incomplete profile data for account %d - %s (username: %s, name: %s)\n",
					account.ID, account.Username, accountProfile.Username, accountProfile.Name)
				if t.shouldMarkInactive(account) {
					accountUpdates = append(accountUpdates, accountUpdate{
						account: &models.Account{
							ID:            account.ID,
							UpdatedAt:     time.Now(),
							LastTrackedAt: account.LastTrackedAt,
							Status:        enums.AccountStatusInactive,
						},
						updateType: updateTypeMinimal,
					})
				}
				continue
			}

			now := time.Now()
			if !hasChanges(account, accountProfile) {
				// Only update tracking metadata
				accountUpdates = append(accountUpdates, accountUpdate{
					account: &models.Account{
						ID:            account.ID,
						UpdatedAt:     now,
						LastTrackedAt: &now,
						Status:        enums.AccountStatusActive,
					},
					updateType: updateTypeMinimal,
				})
				continue
			}

			// Upload profile picture only if not already uploaded
			if account.ProfileImageID == 0 && accountProfile.ProfileImageUrl != "" {
				profileImage, err := t.fileService.UploadFileFromUrl(accountProfile.ProfileImageUrl)
				if err != nil {
					log.Printf("Error uploading profile image for account %d: %v\n", account.ID, err)
				} else {
					account.ProfileImageID = profileImage.ID
				}
			}

			// Full profile update
			updated := &models.Account{
				ID:             account.ID,
				UpdatedAt:      now,
				Username:       accountProfile.Username,
				Name:           accountProfile.Name,
				Followers:      accountProfile.Followers,
				Bio:            accountProfile.Bio,
				BioLink:        accountProfile.BioLink,
				ProfileImageID: account.ProfileImageID,
				LastTrackedAt:  &now,
				Status:         enums.AccountStatusActive,
			}

			accountUpdates = append(accountUpdates, accountUpdate{
				account:    updated,
				updateType: updateTypeFull,
			})

			analyticsData = append(analyticsData, &models.AccountAnalytic{
				Followers: updated.Followers,
				Date:      todayDate,
				AccountID: account.ID,
			})
		}

		// Save account updates
		log.Printf("Updating %d accounts in batch #%d\n", len(accountUpdates), batch)
		if err := t.batchUpdateAccounts(tx, accountUpdates); err != nil {
			return fmt.Errorf("error updating accounts in batch %d: %w", batch, err)
		}

		// Save analytics data
		if len(analyticsData) > 0 {
			log.Printf("Saving %d analytics records in batch #%d\n", len(analyticsData), batch)
			if err := tx.CreateInBatches(analyticsData, batchSize).Error; err != nil {
				return fmt.Errorf("error saving analytics data for batch %d: %w", batch, err)
			}
		}

		return nil
	})

	return result.Error
}

func (t *TrackAccountsAnalytics) GetName() TaskLabel {
	return TaskTrackAccountsAnalytics
}

// shouldMarkInactive checks if an account should be marked as inactive
// due to tracking failures for more than 3 days
func (t *TrackAccountsAnalytics) shouldMarkInactive(account *models.Account) bool {
	if account.LastTrackedAt == nil {
		return false
	}

	daysSinceLastTracked := time.Since(*account.LastTrackedAt).Hours() / 24
	if daysSinceLastTracked > 3 {
		fmt.Printf("Account %d should be marked as inactive - not tracked successfully for %.1f days\n", account.ID, daysSinceLastTracked)
		return true
	}

	return false
}

// batchUpdateAccounts separates accounts by update type and performs efficient batch updates
func (t *TrackAccountsAnalytics) batchUpdateAccounts(tx *gorm.DB, updates []accountUpdate) error {
	if len(updates) == 0 {
		return nil
	}

	var minimalUpdates []*models.Account
	var fullUpdates []*models.Account

	for _, update := range updates {
		switch update.updateType {
		case updateTypeMinimal:
			minimalUpdates = append(minimalUpdates, update.account)
		case updateTypeFull:
			fullUpdates = append(fullUpdates, update.account)
		}
	}

	// Batch update accounts with minimal changes (only tracking metadata)
	if len(minimalUpdates) > 0 {
		log.Printf("Performing minimal update for %d accounts\n", len(minimalUpdates))
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"updated_at",
				"last_tracked_at",
				"status",
			}),
		}).CreateInBatches(minimalUpdates, 500).Error; err != nil {
			return fmt.Errorf("error in minimal batch update: %w", err)
		}
	}

	// Batch update accounts with full profile changes
	if len(fullUpdates) > 0 {
		log.Printf("Performing full update for %d accounts\n", len(fullUpdates))
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"updated_at",
				"username",
				"name",
				"followers",
				"bio",
				"bio_link",
				"profile_image_id",
				"last_tracked_at",
				"status",
			}),
		}).CreateInBatches(fullUpdates, 500).Error; err != nil {
			return fmt.Errorf("error in full batch update: %w", err)
		}
	}

	return nil
}

func hasChanges(account *models.Account, profile *socialmedia.Profile) bool {
	isEqual := account.Username == profile.Username &&
		account.Name == profile.Name &&
		account.Followers == profile.Followers &&
		account.Bio == profile.Bio &&
		account.BioLink == profile.BioLink

	return !isEqual
}
