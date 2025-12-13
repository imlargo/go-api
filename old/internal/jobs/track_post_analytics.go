package jobs

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/nicolailuther/butter/internal/models"
	"github.com/nicolailuther/butter/internal/services"

	"github.com/nicolailuther/butter/pkg/socialmedia"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TrackPostAnalytics struct {
	db                 *gorm.DB
	socialMediaGateway socialmedia.SocialMediaService
	fileService        services.FileService
}

func NewTrackPostAnalyticsTask(
	db *gorm.DB,
	socialMediaGateway socialmedia.SocialMediaService,
	fileService services.FileService,
) Job {
	return &TrackPostAnalytics{
		db:                 db,
		socialMediaGateway: socialMediaGateway,
		fileService:        fileService,
	}
}

type PostResult struct {
	Analytic    *models.PostAnalytic
	UpdatedPost *models.Post
	Error       error
}

var AccountViews []struct {
	AccountID  uint `json:"account_id"`
	TotalViews int  `json:"total_views"`
}

type ContentAccountViews struct {
	ContentID        uint `json:"content_id"`
	AccountID        uint `json:"account_id"`
	ContentAccountID uint `json:"content_account_id"`
	TotalViews       int  `json:"total_views"`   // Sum of all views for this content-account
	AverageViews     int  `json:"average_views"` // Average views per post for this content-account
	TimesPosted      int  `json:"times_posted"`
}

// Structure for worker jobs
type postJob struct {
	index int
	post  *models.Post
}

func (t *TrackPostAnalytics) Execute() error {
	log.Println("[Job Start] Starting analytics collection task")

	// Define time periods based on calendar days (not exact hours)
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	twoHoursAgo := now.Add(-2 * time.Hour)

	// Deadline dates to categorize posts by age
	tenDaysAgo := today.AddDate(0, 0, -10)
	twentyDaysAgo := today.AddDate(0, 0, -20)
	thirtyDaysAgo := today.AddDate(0, 0, -30)

	// Deadline dates to check last scraping (start of the day)
	threeDaysAgoStart := today.AddDate(0, 0, -3)
	fourDaysAgoStart := today.AddDate(0, 0, -4)
	fiveDaysAgoStart := today.AddDate(0, 0, -5)

	query := t.db.
		Joins("JOIN accounts ON posts.account_id = accounts.id").
		Where("accounts.deleted_at IS NULL").
		Where("posts.updated_at < ? OR posts.created_at >= ?", twoHoursAgo, twoHoursAgo).
		Where("posts.track = ?", true).
		Where(`(
		-- Posts less than 10 days old: normal scraping (every day)
		posts.created_at > ?
		OR
		-- Posts between 10-20 days old: scraping every 3 days
		(posts.created_at <= ? AND posts.created_at > ? AND posts.updated_at <= ?)
		OR
		-- Posts between 20-30 days old: scraping every 4 days  
		(posts.created_at <= ? AND posts.created_at > ? AND posts.updated_at <= ?)
		OR
		-- Posts older than 30 days: scraping every 5 days
		(posts.created_at <= ? AND posts.updated_at <= ?)
	)`,
			tenDaysAgo,                                   // Posts less than 10 days old
			tenDaysAgo, twentyDaysAgo, threeDaysAgoStart, // Posts 10-20 days old, last update 3+ days ago
			twentyDaysAgo, thirtyDaysAgo, fourDaysAgoStart, // Posts 20-30 days old, last update 4+ days ago
			thirtyDaysAgo, fiveDaysAgoStart) // Posts 30+ days old, last update 5+ days ago

	var tempPosts []*models.Post
	batchCount := 0
	totalProcessed := 0

	result := query.
		FindInBatches(&tempPosts, DefaultBatchSize, func(tx *gorm.DB, batch int) error {

			startTime := time.Now()
			batchCount++
			totalProcessed += len(tempPosts)

			log.Printf("[Batch %d] Starting batch with %d posts", batch, len(tempPosts))

			analytics, postToUpdate := t.ProcessPosts(tempPosts, batch)

			err := t.db.Transaction(func(tx2 *gorm.DB) error {

				if SaveAnalytics && len(analytics) > 0 {
					log.Printf("[Batch %d] Saving %d analytics records", batch, len(analytics))
					if err := tx2.CreateInBatches(analytics, 500).Error; err != nil {
						log.Printf("[Error] Failed to save analytics for batch %d: %v", batch, err)
						return fmt.Errorf("failed to save analytics: %w", err)
					}
				}

				// Update all posts
				if UpdatePosts && len(postToUpdate) > 0 {
					log.Printf("[Batch %d] Updating %d posts", batch, len(postToUpdate))

					now := time.Now()
					for i := range postToUpdate {
						postToUpdate[i].UpdatedAt = now
					}

					if err := tx2.Clauses(clause.OnConflict{
						Columns: []clause.Column{{Name: "id"}},
						DoUpdates: clause.AssignmentColumns([]string{
							"updated_at",
							"total_views",
							"is_deleted",
							"thumbnail_id",
						}),
					}).CreateInBatches(postToUpdate, 500).Error; err != nil {
						log.Printf("[Error] Failed to update posts for batch %d: %v", batch, err)
						return fmt.Errorf("failed to update posts: %w", err)
					}
				}

				return nil
			})
			if err != nil {
				log.Printf("[Error] Transaction failed for batch %d: %v", batch, err)
				return err
			}

			elapsed := time.Since(startTime)
			log.Printf("[Batch %d] Completed in %v", batch, elapsed)

			if elapsed < time.Minute {
				time.Sleep(time.Minute - elapsed)
			}

			return nil
		})

	if result.Error != nil {
		log.Printf("[Error] Batch processing failed: %v", result.Error)
		return fmt.Errorf("batch processing error: %w", result.Error)
	}

	log.Printf("[Job Progress] Processed %d batches with %d total posts", batchCount, totalProcessed)

	// Update uploaded content and accounts
	log.Println("[Job Progress] Updating content views...")
	if err := t.updateContentViews(); err != nil {
		log.Printf("[Error] Failed to update content views: %v", err)
		return fmt.Errorf("failed to update content views: %w", err)
	}

	log.Println("[Job Progress] Updating account views...")
	if err := t.updateAccountViews(); err != nil {
		log.Printf("[Error] Failed to update account views: %v", err)
		return fmt.Errorf("failed to update account views: %w", err)
	}

	log.Println("[Job Complete] Analytics collection task finished successfully")
	return nil
}

func (t *TrackPostAnalytics) ProcessPosts(posts []*models.Post, batchNum int) ([]*models.PostAnalytic, []*models.Post) {
	numPosts := len(posts)

	if numPosts == 0 {
		return []*models.PostAnalytic{}, []*models.Post{}
	}

	jobs := make(chan postJob, DefaultNumWorkers)
	results := make(chan PostResult, DefaultNumWorkers)

	var wg sync.WaitGroup
	today := time.Now().Format("2006-01-02")

	// Create workers
	numWorkers := DefaultNumWorkers
	if numPosts < DefaultNumWorkers {
		numWorkers = numPosts // Do not create more workers than posts
	}

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for job := range jobs {
				result := t.ProcessPost(job.post, today)
				results <- result
			}
		}(i)
	}

	// Goroutine to close the results channel when all workers are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Send jobs in another goroutine to avoid blocking
	go func() {
		defer close(jobs)
		for index, post := range posts {
			jobs <- postJob{
				index: index,
				post:  post,
			}
		}
	}()

	batchAnalytics := make([]*models.PostAnalytic, 0, numPosts)
	batchUpdatedPosts := make([]*models.Post, 0, numPosts)

	// Track error statistics
	errorCount := 0
	timeoutCount := 0
	rateLimitCount := 0
	notFoundCount := 0

	log.Printf("[Batch %d] Processing %d posts with %d workers...", batchNum, numPosts, numWorkers)

	// Procesar resultados mientras los workers trabajan
	for result := range results {
		if result.Error != nil {
			errorCount++

			// Categorize error types for better reporting
			errStr := strings.ToLower(result.Error.Error())
			if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "deadline exceeded") {
				timeoutCount++
			} else if strings.Contains(errStr, "rate limit") {
				rateLimitCount++
			} else if strings.Contains(errStr, "not found") {
				notFoundCount++
			}

			log.Printf("[Batch %d] Error processing post: %v", batchNum, result.Error)
			continue
		}

		if result.Analytic != nil {
			batchAnalytics = append(batchAnalytics, result.Analytic)
		}

		if result.UpdatedPost != nil {
			batchUpdatedPosts = append(batchUpdatedPosts, result.UpdatedPost)
		}
	}

	successCount := numPosts - errorCount
	log.Printf("[Batch %d] Completed: %d/%d successful, %d errors (timeouts: %d, rate limits: %d, not found: %d), %d analytics, %d posts updated",
		batchNum, successCount, numPosts, errorCount, timeoutCount, rateLimitCount, notFoundCount,
		len(batchAnalytics), len(batchUpdatedPosts))

	return batchAnalytics, batchUpdatedPosts
}

func (t *TrackPostAnalytics) ProcessPost(post *models.Post, todayDate string) PostResult {
	log.Printf("[Info] Processing post ID=%d, URL=%s, Platform=%s", post.ID, post.Url, post.Platform)

	postData, err := t.socialMediaGateway.GetPostData(post.Platform, post.Url)
	if err != nil {
		// Handle rate limiting
		if errors.Is(err, socialmedia.ErrRateLimited) {
			log.Printf("[Warning] Rate limit reached for post ID=%d, waiting 10s before continuing...", post.ID)
			time.Sleep(10 * time.Second)
			return PostResult{
				Error: fmt.Errorf("rate limited for post ID %d, will retry later", post.ID),
			}
		}

		// Handle post not found
		if errors.Is(err, socialmedia.ErrPostNotFound) {
			if post.IsDeleted {
				log.Printf("[Info] Post ID=%d already marked as deleted, skipping", post.ID)
				return PostResult{}
			}

			log.Printf("[Info] Post ID=%d not found on platform, marking as deleted", post.ID)
			updatedPost := *post
			updatedPost.IsDeleted = true
			return PostResult{
				UpdatedPost: &updatedPost,
			}
		}

		// Handle timeout errors specifically
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "deadline exceeded") {
			log.Printf("[Error] Timeout fetching data for post ID=%d: %v", post.ID, err)
			return PostResult{
				Error: fmt.Errorf("timeout for post ID %d: %w", post.ID, err),
			}
		}

		// Generic error
		log.Printf("[Error] Failed to fetch data for post ID=%d: %v", post.ID, err)
		return PostResult{
			Error: fmt.Errorf("error fetching post data for post ID %d: %w", post.ID, err),
		}
	}

	if postData == nil {
		log.Printf("[Error] No data returned for post ID=%d despite no error", post.ID)
		return PostResult{
			Error: fmt.Errorf("no data returned for post ID %d", post.ID),
		}
	}

	log.Printf("[Success] Fetched data for post ID=%d: views=%d, likes=%d, comments=%d",
		post.ID, postData.Views, postData.LikeCount, postData.CommentCount)

	updatedPost := *post

	// Upload thumbnail if it doesn't exist
	if post.ThumbnailID == 0 && postData.ThumbnailUrl != "" {
		thumbnail, err := t.fileService.UploadFileFromUrl(postData.ThumbnailUrl)
		if err != nil {
			log.Printf("[Error] uploading thumbnail for post %d: %v\n", post.ID, err)
		} else {
			updatedPost.ThumbnailID = thumbnail.ID
			log.Printf("[Info] uploaded thumbnail for post %d\n", post.ID)
		}
	}

	if post.IsDeleted {
		updatedPost.IsDeleted = false
	}

	if postData.Views != 0 {
		updatedPost.TotalViews = postData.Views
	}

	analytic := &models.PostAnalytic{
		TotalViews: updatedPost.TotalViews,
		Date:       todayDate,
		PostID:     post.ID,
		AccountID:  post.AccountID,
	}

	// Rate limiting at the end to prevent delaying the first request in each worker
	// while still maintaining the rate limit between subsequent requests
	time.Sleep(1 * time.Second)

	return PostResult{
		Analytic:    analytic,
		UpdatedPost: &updatedPost,
	}
}

func (t *TrackPostAnalytics) updateContentViews() error {
	// First, update Content table with aggregated views across all accounts
	if err := t.updateContentTotalViews(); err != nil {
		return err
	}

	// Then, update ContentAccount table with account-specific views
	if err := t.updateContentAccountViews(); err != nil {
		return err
	}

	return nil
}

// ContentViewsDetailed represents the detailed aggregated views for a content
type ContentViewsDetailed struct {
	ContentID    uint `json:"content_id"`
	TotalViews   int  `json:"total_views"`   // Sum of all views
	AverageViews int  `json:"average_views"` // Average views per post
	TimesPosted  int  `json:"times_posted"`
}

// updateContentTotalViews updates the main Content table with aggregated analytics
func (t *TrackPostAnalytics) updateContentTotalViews() error {
	var contentViews []ContentViewsDetailed

	// Get aggregated views per content across all posts
	// total_views = SUM, average_views = AVG
	err := t.db.Model(&models.Post{}).
		Select(`
			content_id, 
			COALESCE(SUM(total_views), 0) as total_views,
			CAST(ROUND(AVG(total_views)) AS INTEGER) as average_views,
			COUNT(*) as times_posted
		`).
		Where("content_id IS NOT NULL AND content_id > 0").
		Group("content_id").
		Scan(&contentViews).Error

	if err != nil {
		return err
	}

	if len(contentViews) == 0 {
		log.Println("No content views to update")
		return nil
	}

	// Prepare batch update for Content table
	var updatedContents []models.Content
	for _, view := range contentViews {
		updatedContents = append(updatedContents, models.Content{
			ID:           view.ContentID,
			TotalViews:   view.TotalViews,   // Sum of all post views
			AverageViews: view.AverageViews, // Average views per post
			TimesPosted:  view.TimesPosted,
		})
	}

	// Update content records
	err = t.db.Model(&models.Content{}).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"total_views", "average_views", "times_posted"}),
		}).CreateInBatches(updatedContents, 500).Error

	if err != nil {
		return err
	}

	log.Printf("Updated %d content records with analytics data\n", len(updatedContents))
	return nil
}

// updateContentAccountViews updates ContentAccount table with account-specific analytics
func (t *TrackPostAnalytics) updateContentAccountViews() error {
	var contentAccountViews []ContentAccountViews

	// Get aggregated views per content-account combination
	// We need to join with ContentAccount to get the content_account_id
	err := t.db.Table("posts p").
		Select(`
			p.content_id,
			p.account_id,
			ca.id as content_account_id,
			COALESCE(SUM(p.total_views), 0) as total_views,
			CAST(ROUND(AVG(p.total_views)) AS INTEGER) as average_views,
			COUNT(*) as times_posted
		`).
		Joins("JOIN content_accounts ca ON ca.content_id = p.content_id AND ca.account_id = p.account_id").
		Where("p.content_id IS NOT NULL AND p.content_id > 0 AND p.account_id IS NOT NULL").
		Group("p.content_id, p.account_id, ca.id").
		Scan(&contentAccountViews).Error

	if err != nil {
		return err
	}

	if len(contentAccountViews) == 0 {
		log.Println("No content-account views to update")
		return nil
	}

	// Prepare batch update for ContentAccount table
	var updatedContentAccounts []models.ContentAccount
	for _, view := range contentAccountViews {
		updatedContentAccounts = append(updatedContentAccounts, models.ContentAccount{
			ID:                  view.ContentAccountID,
			TimesPosted:         view.TimesPosted,
			AccountTotalViews:   view.TotalViews,   // Sum of views for this content-account
			AccountAverageViews: view.AverageViews, // Average views per post for this content-account
		})
	}

	// Update content account records
	err = t.db.Model(&models.ContentAccount{}).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"times_posted", "account_total_views", "account_average_views"}),
		}).CreateInBatches(updatedContentAccounts, 500).Error

	if err != nil {
		return err
	}

	log.Printf("Updated %d content-account records with analytics data\n", len(updatedContentAccounts))
	return nil
}

func (t *TrackPostAnalytics) updateAccountViews() error {
	// Get avg views per account
	err := t.db.Model(&models.Post{}).
		Select("account_id, CAST(ROUND(AVG(total_views)) AS INTEGER) as total_views").
		Where("account_id IS NOT NULL AND account_id > 0").
		Group("account_id").
		Scan(&AccountViews).Error

	if err != nil {
		return err
	}

	if len(AccountViews) == 0 {
		log.Println("No account views to update")
		return nil
	}

	var updatedAccounts []models.Account
	for _, view := range AccountViews {
		updatedAccounts = append(updatedAccounts, models.Account{
			ID:           view.AccountID,
			AverageViews: view.TotalViews,
		})
	}

	// Update accounts
	err = t.db.Model(&models.Account{}).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"average_views", "updated_at"}),
		}).CreateInBatches(updatedAccounts, 500).Error

	if err != nil {
		return err
	}

	log.Printf("Updated %d account records with average views\n", len(updatedAccounts))
	return nil
}

func (t *TrackPostAnalytics) GetName() TaskLabel {
	return TaskTrackPostAnalytics
}
