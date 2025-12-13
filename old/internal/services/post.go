package services

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"strings"
	"time"

	"github.com/nicolailuther/butter/internal/enums"
	"github.com/nicolailuther/butter/internal/models"
	"github.com/nicolailuther/butter/pkg/socialmedia"
	"github.com/nicolailuther/butter/pkg/utils"
	"gorm.io/gorm"
)

type PostService interface {
	GetPostedContent(accountID uint, contentType enums.ContentType, limit int) ([]*models.Post, error)
	GetAllPostedContent(accountID uint, limit int) ([]*models.Post, error)

	PostContent(accountID uint, generatedContentID uint, url string, contentType enums.ContentType) (*models.Post, error)
	TrackPost(accountID uint, url string) (*models.Post, error)

	SyncPosts(accountID uint) error
}

type postServiceImpl struct {
	*Service
	fileService        FileService
	contentService     ContentService
	socialMediaService socialmedia.SocialMediaService
	syncStatusService  SyncStatusService
}

func NewPostService(
	container *Service,
	fileService FileService,
	contentService ContentService,
	socialMediaService socialmedia.SocialMediaService,
	syncStatusService SyncStatusService,
) PostService {
	return &postServiceImpl{
		container,
		fileService,
		contentService,
		socialMediaService,
		syncStatusService,
	}
}

const (
	// fingerprintCacheTTL is the time-to-live for cached fingerprints (24 hours)
	fingerprintCacheTTL = 24 * time.Hour
)

// getCachedFingerprint retrieves a fingerprint from cache if it exists
func (s *postServiceImpl) getCachedFingerprint(videoURL string) (*utils.VideoFingerprint, error) {
	key := s.cacheKeys.VideoFingerprint(videoURL)

	var fingerprint utils.VideoFingerprint
	err := s.cache.GetJSON(key, &fingerprint)
	if err != nil {
		return nil, err
	}

	return &fingerprint, nil
}

// setCachedFingerprint stores a fingerprint in cache
func (s *postServiceImpl) setCachedFingerprint(videoURL string, fingerprint *utils.VideoFingerprint) error {
	if fingerprint == nil {
		return fmt.Errorf("fingerprint cannot be nil")
	}

	key := s.cacheKeys.VideoFingerprint(videoURL)
	return s.cache.Set(key, fingerprint, fingerprintCacheTTL)
}

// generateVideoFingerprintWithCache generates or retrieves a cached fingerprint
func (s *postServiceImpl) generateVideoFingerprintWithCache(videoURL string) (*utils.VideoFingerprint, error) {
	// Try to get from cache first
	cached, err := s.getCachedFingerprint(videoURL)
	if err == nil && cached != nil {
		return cached, nil
	}

	// Generate new fingerprint if not in cache
	fingerprint, err := utils.GenerateVideoFingerprint(videoURL)
	if err != nil {
		return nil, err
	}

	// Cache the generated fingerprint
	if err := s.setCachedFingerprint(videoURL, fingerprint); err != nil {
		log.Printf("[FingerprintCache] Failed to cache fingerprint: %v", err)
	}

	return fingerprint, nil
}

func (s *postServiceImpl) GetPostedContent(accountID uint, contentType enums.ContentType, limit int) ([]*models.Post, error) {
	if accountID == 0 {
		return nil, errors.New("account ID cannot be zero")
	}

	posts, err := s.store.Posts.GetByAccountAndType(accountID, string(contentType), limit)
	if err != nil {
		return nil, err
	}

	return posts, nil
}

func (s *postServiceImpl) GetAllPostedContent(accountID uint, limit int) ([]*models.Post, error) {
	if accountID == 0 {
		return nil, errors.New("account ID cannot be zero")
	}

	posts, err := s.store.Posts.GetByAccount(accountID, limit)
	if err != nil {
		return nil, err
	}

	return posts, nil
}

func (s *postServiceImpl) PostContent(accountID uint, generatedContentID uint, url string, contentType enums.ContentType) (*models.Post, error) {

	if accountID == 0 {
		return nil, errors.New("account ID cannot be zero")
	}

	if generatedContentID == 0 {
		return nil, errors.New("generated content ID cannot be zero")
	}

	if url == "" {
		return nil, errors.New("URL cannot be empty")
	}

	// Check for duplicate URL for this account
	existingPost, err := s.store.Posts.GetByUrl(url)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existingPost != nil {
		return nil, errors.New("a post with this URL already exists for this account")
	}

	account, err := s.store.Accounts.GetByID(accountID)
	if err != nil {
		return nil, err
	}

	generatedContent, err := s.store.GeneratedContents.Get(generatedContentID)
	if err != nil {
		return nil, err
	}

	post := &models.Post{
		Platform:  account.Platform,
		Type:      contentType,
		Url:       url,
		IsTracked: false,

		ThumbnailID:        generatedContent.Files[0].ThumbnailID,
		AccountContentID:   generatedContent.ContentAccountID,
		ContentID:          generatedContent.ContentID,
		AccountID:          accountID,
		GeneratedContentID: generatedContentID,
		TextOverlayID:      generatedContent.TextOverlayID,
	}

	if err := s.store.Posts.Create(post); err != nil {
		return nil, err
	}

	// Mark generated content as posted
	if err := s.store.GeneratedContents.MarkAsPosted(generatedContentID); err != nil {
		return nil, err
	}

	// Update content and assignation as posted
	if err := s.store.Contents.UpdateVideoAsPosted(generatedContent.ContentID); err != nil {
		return nil, err
	}

	if err := s.store.ContentAccounts.UpdateVideoAsPosted(generatedContent.ContentAccountID); err != nil {
		return nil, err
	}

	return post, nil
}

func (s *postServiceImpl) TrackPost(accountID uint, url string) (*models.Post, error) {

	if accountID == 0 {
		return nil, errors.New("account ID cannot be zero")
	}

	if url == "" {
		return nil, errors.New("URL cannot be empty")
	}

	// Check for duplicate URL
	existingPost, err := s.store.Posts.GetByUrl(url)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existingPost != nil {
		return nil, errors.New("a post with this URL already exists for this account")
	}

	account, err := s.store.Accounts.GetByID(accountID)
	if err != nil {
		return nil, err
	}

	postInfo, err := s.socialMediaService.GetPostData(account.Platform, url)
	if err != nil {
		return nil, errors.New("failed to retrieve post data: " + err.Error())
	}

	if postInfo == nil {
		return nil, errors.New("post data is empty")
	}

	thumbnail, err := s.fileService.UploadFileFromUrl(postInfo.ThumbnailUrl)
	if err != nil {
		return nil, errors.New("failed to upload thumbnail: " + err.Error())
	}

	post := &models.Post{
		Platform:    account.Platform,
		Url:         url,
		IsTracked:   true,
		AccountID:   accountID,
		ThumbnailID: thumbnail.ID,
		TotalViews:  postInfo.Views,
	}

	if err := s.store.Posts.Create(post); err != nil {
		s.fileService.DeleteFile(thumbnail.ID)
		return nil, err
	}

	post.Thumbnail = thumbnail

	return post, nil
}

func (s *postServiceImpl) SyncPosts(accountID uint) error {
	if accountID == 0 {
		return errors.New("account ID cannot be zero")
	}

	// Step 1: Check if sync is already in progress for this account
	// This early check prevents unnecessary database queries if sync is already active
	existingStatus, err := s.syncStatusService.GetActiveStatus(accountID)
	if err == nil && existingStatus != nil {
		return errors.New("sync is already in progress for this account")
	}

	// Step 2: Get account and validate
	account, err := s.store.Accounts.GetByID(accountID)
	if err != nil {
		return fmt.Errorf("get account: %w", err)
	}

	if account == nil {
		return errors.New("account not found")
	}

	if account.Platform != enums.PlatformInstagram {
		return errors.New("sync posts is only supported for Instagram accounts")
	}

	// Step 3: Count pending generated contents (not posted yet)
	videoType := enums.ContentTypeVideo
	totalToProcess, err := s.store.GeneratedContents.CountByAccountAndType(accountID, videoType)
	if err != nil {
		return fmt.Errorf("count pending contents: %w", err)
	}

	if totalToProcess == 0 {
		log.Printf("[SyncPosts] No pending contents found for account %d", accountID)
		return nil
	}

	// Step 4: Create sync status record immediately to acquire lock
	syncStatus, err := s.syncStatusService.AcquireSync(accountID, totalToProcess)
	if err != nil {
		if err == ErrSyncInProgress {
			return errors.New("sync is already in progress for this account")
		}
		return fmt.Errorf("failed to acquire sync: %w", err)
	}

	log.Printf("[SyncPosts] Starting sync: account=%d, user=%s, status_id=%d, total_to_process=%d",
		accountID, account.Username, syncStatus.ID, totalToProcess)

	// Step 5: Launch background processing
	go s.syncPostsBackground(accountID, syncStatus.ID)

	return nil
}

// syncPostsBackground performs the actual sync processing in the background
func (s *postServiceImpl) syncPostsBackground(accountID uint, syncStatusID uint) {
	// Recover from panics to prevent crashes
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[SyncPostsBackground] Panic recovered: %v", r)
			if markErr := s.syncStatusService.MarkFailed(syncStatusID, fmt.Sprintf("Panic during sync: %v", r)); markErr != nil {
				log.Printf("[SyncPostsBackground] Failed to mark sync as failed after panic: %v", markErr)
			}
		}
	}()

	// Get account info
	account, err := s.store.Accounts.GetByID(accountID)
	if err != nil {
		log.Printf("[SyncPostsBackground] Failed to get account: %v", err)
		if markErr := s.syncStatusService.MarkFailed(syncStatusID, "Failed to get account: "+err.Error()); markErr != nil {
			log.Printf("[SyncPostsBackground] Failed to mark sync as failed: %v", markErr)
		}
		return
	}

	// Step 1: Get unposted video contents (only videos are supported for sync)
	videoType := enums.ContentTypeVideo
	generatedContents, err := s.store.GeneratedContents.GetByAccountAndType(accountID, &videoType)
	if err != nil {
		log.Printf("[SyncPostsBackground] Failed to get contents: %v", err)
		if markErr := s.syncStatusService.MarkFailed(syncStatusID, "Failed to get contents: "+err.Error()); markErr != nil {
			log.Printf("[SyncPostsBackground] Failed to mark sync as failed: %v", markErr)
		}
		return
	}

	if len(generatedContents) == 0 {
		log.Printf("[SyncPostsBackground] No contents found")
		if completeErr := s.syncStatusService.CompleteSync(syncStatusID); completeErr != nil {
			log.Printf("[SyncPostsBackground] Failed to complete sync status: %v", completeErr)
		}
		return
	}

	// Step 2: Build fingerprint index (inline, without separate maps)
	contentsFingerPrints := make(map[uint]*utils.VideoFingerprint)

	for _, gc := range generatedContents {
		// Validate content
		if len(gc.Files) == 0 || len(gc.Files) > 1 || gc.Files[0].File.Url == "" {
			log.Printf("[SyncPostsBackground] Skipping generated content ID %d: invalid data", gc.ID)
			continue
		}

		fingerprint, err := s.generateVideoFingerprintWithCache(gc.Files[0].File.Url)
		if err != nil {
			log.Printf("[SyncPostsBackground] Failed to generate fingerprint for generated content ID %d: %v", gc.ID, err)
			continue
		}

		contentsFingerPrints[gc.ID] = fingerprint

		// Shorter delays: 3-5 seconds
		time.Sleep(time.Duration(3+rand.Intn(2)) * time.Second)
	}

	if len(contentsFingerPrints) == 0 {
		log.Printf("[SyncPostsBackground] No valid contents to sync")
		if completeErr := s.syncStatusService.CompleteSync(syncStatusID); completeErr != nil {
			log.Printf("[SyncPostsBackground] Failed to complete sync status: %v", completeErr)
		}
		return
	}

	log.Printf("[SyncPostsBackground] Generated %d fingerprints from %d contents", len(contentsFingerPrints), len(generatedContents))

	// Calculate fetch amount based on generated contents
	fetchAmount := min(len(contentsFingerPrints)*3, 50)
	if fetchAmount < 15 {
		fetchAmount = 15
	}

	// Step 3: Fetch Instagram posts (optimal amount)
	log.Printf("[SyncPostsBackground] Fetching %d posts from Instagram to match against %d contents", fetchAmount, len(contentsFingerPrints))

	instagramPosts, err := s.socialMediaService.GetUserReels(account.Platform, account.Username, fetchAmount)
	if err != nil {
		// Complete sync with error before returning
		if markErr := s.syncStatusService.MarkFailed(syncStatusID, "Failed to fetch Instagram posts: "+err.Error()); markErr != nil {
			log.Printf("[SyncPostsBackground] Failed to mark sync as failed: %v", markErr)
		}
		return
	}

	// Step 4: Build existing posts map (single pass)
	existingCodes := make(map[string]bool, len(instagramPosts)/2)
	existingPosts, err := s.store.Posts.GetLastNPosts(accountID, fetchAmount)
	if err == nil {
		for _, post := range existingPosts {
			if code := extractInstagramCode(post.Url); code != "" {
				existingCodes[code] = true
			}
		}
	}

	log.Printf("[SyncPostsBackground] Processing %d posts, %d already cached", len(instagramPosts), len(existingCodes))

	// Step 5: Match generated contents to Instagram posts
	bestMatches := make([]bestMatch, 0, 5) // Store top matches in case there are multiple
	skippedPosts := 0

	for _, igPost := range instagramPosts {
		if !igPost.IsVideo || igPost.Code == "" {
			continue
		}

		if existingCodes[igPost.Code] {
			log.Printf("[SyncPostsBackground] Skipping Instagram post code %s: already exists in database", igPost.Code)
			skippedPosts++
			continue
		}

		log.Printf("[SyncPostsBackground] Processing Instagram post code: %s", igPost.Code)

		// Get post data
		postURL := fmt.Sprintf("https://www.instagram.com/p/%s/", igPost.Code)
		post, err := s.socialMediaService.GetPostData(account.Platform, postURL)
		if err != nil || post.VideoURL == "" {
			log.Printf("[SyncPostsBackground] Failed to get post data for Instagram post %s: %v", igPost.Code, err)
			continue
		}

		// Generate fingerprint
		postFP, err := s.generateVideoFingerprintWithCache(post.VideoURL)
		if err != nil {
			continue
		}

		// Find best match (inline, without separate function)
		bestMatchID := uint(0)
		bestSim := 0.0

		for gcID, gcFP := range contentsFingerPrints {
			sim := utils.CompareFingerprints(gcFP, postFP)

			if sim > bestSim && sim >= 0.82 {
				bestMatchID = gcID
				bestSim = sim
			}
		}

		if bestMatchID == 0 {
			// No match found for this Instagram post, continue to next post
			continue
		}

		log.Printf("[Match] Post %s -> Content %d (%.1f%%)", igPost.Code, bestMatchID, bestSim*100)

		// Store best match for later processing
		bestMatches = append(bestMatches, bestMatch{
			contentID: bestMatchID,
			postURL:   postURL,
			igCode:    igPost.Code,
		})

		delete(contentsFingerPrints, bestMatchID)

		// Early exit if all contents are matched
		if len(contentsFingerPrints) == 0 {
			break
		}

		time.Sleep(time.Duration(5+rand.Intn(5)) * time.Second)
	}

	// Step 6: Create posts for matched contents
	matchedCount := 0
	for _, bm := range bestMatches {
		// Mark this content as processed (attempting to post)
		if err := s.syncStatusService.IncrementProcessed(syncStatusID); err != nil {
			log.Printf("[SyncPostsBackground] Failed to increment processed count for sync status %d: %v", syncStatusID, err)
			// Continue processing but log the error
		}

		_, err := s.PostContent(accountID, bm.contentID, bm.postURL, enums.ContentTypeVideo)
		if err != nil {
			log.Printf("[Match] Failed to create post: %v", err)
			if err := s.syncStatusService.IncrementFailed(syncStatusID); err != nil {
				log.Printf("[SyncPostsBackground] Failed to increment failed count for sync status %d: %v", syncStatusID, err)
			}
			continue
		}
		if err := s.syncStatusService.IncrementSynced(syncStatusID); err != nil {
			log.Printf("[SyncPostsBackground] Failed to increment synced count for sync status %d: %v", syncStatusID, err)
		}
		matchedCount++
	}

	// Step 7: Mark remaining unmatched contents as failed (not found)
	unmatchedCount := len(contentsFingerPrints)
	if unmatchedCount > 0 {
		log.Printf("[SyncPostsBackground] %d generated contents were not found in Instagram posts", unmatchedCount)
		for contentID := range contentsFingerPrints {
			log.Printf("[SyncPostsBackground] Content ID %d not found in Instagram posts", contentID)
			if err := s.syncStatusService.IncrementProcessed(syncStatusID); err != nil {
				log.Printf("[SyncPostsBackground] Failed to increment processed count for sync status %d: %v", syncStatusID, err)
			}
			if err := s.syncStatusService.IncrementFailed(syncStatusID); err != nil {
				log.Printf("[SyncPostsBackground] Failed to increment failed count for sync status %d: %v", syncStatusID, err)
			}
		}
	}

	log.Printf("[SyncPostsBackground] Completed: synced=%d, not_found=%d, skipped_posts=%d",
		matchedCount, unmatchedCount, skippedPosts)

	// Complete the sync status
	if err := s.syncStatusService.CompleteSync(syncStatusID); err != nil {
		log.Printf("[SyncPostsBackground] Failed to complete sync status: %v", err)
	}
}

// Estructura auxiliar para guardar matches
type bestMatch struct {
	contentID uint
	postURL   string
	igCode    string
}

// extractInstagramCode extrae el código de una URL (sin parsing innecesario)
func extractInstagramCode(igURL string) string {
	if igURL == "" {
		return ""
	}

	// Parse simple
	u, err := url.Parse(igURL)
	if err != nil {
		return ""
	}

	path := strings.Trim(u.Path, "/")
	if path == "" {
		return ""
	}

	parts := strings.Split(path, "/")

	// Buscar en las últimas posiciones (más probable)
	for i := len(parts) - 1; i > 0; i-- {
		if (parts[i-1] == "p" || parts[i-1] == "reel") && parts[i] != "" {
			return parts[i]
		}
	}

	return ""
}

// calculateFetchAmount ahora inline y más simple
func calculateFetchAmount(contentCount int) int {
	return min(contentCount*3, 50)
}

// Utilidades simples
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
