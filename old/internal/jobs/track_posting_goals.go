package jobs

import (
	"fmt"
	"log"
	"time"

	"github.com/nicolailuther/butter/internal/models"
	"gorm.io/gorm"
)

type TrackPostingGoals struct {
	db *gorm.DB
}

func NewTrackPostingGoalsTask(
	db *gorm.DB,
) Job {
	return &TrackPostingGoals{
		db: db,
	}
}

func (t *TrackPostingGoals) Execute() error {

	colombiaLocation, _ := time.LoadLocation("America/Bogota")
	now := time.Now().In(colombiaLocation)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, colombiaLocation)
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, colombiaLocation)
	batchSize := 100

	type AccountContentCounts struct {
		AccountID      uint  `json:"account_id"`
		VideoCount     int64 `json:"video_count"`
		SlideshowCount int64 `json:"slideshow_count"`
		StoryCount     int64 `json:"story_count"`
	}

	var results []AccountContentCounts

	// Query using the type field directly from posts table
	query := `
		SELECT 
			p.account_id,
			COUNT(CASE WHEN p.type = 'video' THEN 1 END) as video_count,
			COUNT(CASE WHEN p.type = 'slideshow' THEN 1 END) as slideshow_count,
			COUNT(CASE WHEN p.type = 'story' THEN 1 END) as story_count
		FROM posts p
		WHERE p.created_at BETWEEN ? AND ?
		GROUP BY p.account_id
	`

	err := t.db.Raw(query, startOfDay, endOfDay).Scan(&results).Error
	if err != nil {
		return fmt.Errorf("failed to get post counts: %w", err)
	}

	log.Println("Results: ", len(results))

	var accountCounts map[uint]AccountContentCounts
	accountCounts = make(map[uint]AccountContentCounts)
	for _, result := range results {
		accountCounts[result.AccountID] = result
	}

	var accounts []*models.Account
	var analytics []*models.PostingGoal
	result := t.db.FindInBatches(&accounts, batchSize, func(tx *gorm.DB, batch int) error {
		fmt.Printf("Processing batch %d with %d records\n", batch, len(accounts))

		for _, account := range accounts {
			counts := accountCounts[account.ID]

			analytic := &models.PostingGoal{
				PostGoal:      account.PostingGoal,
				SlideshowGoal: account.SlideshowPostingGoal,
				StoryGoal:     account.StoryPostingGoal,

				TotalPosts:      int(counts.VideoCount),
				TotalSlideshows: int(counts.SlideshowCount),
				TotalStories:    int(counts.StoryCount),
				MarketingCost:   account.DailyMarketingCost,

				AccountID: account.ID,
			}

			analytics = append(analytics, analytic)
		}

		return nil
	})

	if result.Error != nil {
		return result.Error
	}

	err = t.db.CreateInBatches(analytics, batchSize).Error
	if err != nil {
		return err
	}

	return result.Error
}

func (t *TrackPostingGoals) GetName() TaskLabel {
	return TaskTrackPostingGoals
}
