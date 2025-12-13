package services

import (
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
	"slices"
	"time"

	"github.com/nicolailuther/butter/internal/dto"
	"github.com/nicolailuther/butter/internal/enums"
	"github.com/nicolailuther/butter/internal/models"
	"github.com/nicolailuther/butter/pkg/content"
)

var (
	ErrLimitExceeded       = errors.New("generated content limit exceeded, please post the generated content before generating more or increase your posting goal")
	ErrNoContentAvailable  = errors.New("no content available for generation")
	ErrUnsupportedPlatform = errors.New("content generation is only supported for TikTok and Instagram accounts")
	ErrUnsupportedType     = errors.New("unsupported content type")
	ErrNoContentFiles      = errors.New("no content files available for the selected content")
)

type ContentGenerationService interface {
	GenerateContentFromV2(request *dto.GenerateContent) ([]*models.GeneratedContent, error)
}

type GenerationServiceImpl struct {
	*Service
	mediaService     content.MediaService
	repurposeService content.MediaService
	fileService      FileService
	strategies       map[enums.ContentType]contentStrategy
}

func NewContentGenerationService(
	container *Service,
	mediaService content.MediaService,
	repurposeService content.MediaService,
	fileService FileService,
) *GenerationServiceImpl {
	service := &GenerationServiceImpl{
		Service:          container,
		mediaService:     mediaService,
		repurposeService: repurposeService,
		fileService:      fileService,
	}

	// Register strategies for each content type
	service.strategies = map[enums.ContentType]contentStrategy{
		enums.ContentTypeVideo:     newVideoStrategy(service),
		enums.ContentTypeStory:     newStoryStrategy(service),
		enums.ContentTypeSlideshow: newSlideshowStrategy(service),
	}

	return service
}

// GenerateContentFromV2 is deprecated and no longer supported.
// This method implemented the old sequential generation system which has been replaced
// by the concurrent generation system using Redis task queue.
// Use ConcurrentContentGenerationService.GenerateContentConcurrent() instead.
func (s *GenerationServiceImpl) GenerateContentFromV2(request *dto.GenerateContent) ([]*models.GeneratedContent, error) {
	return nil, errors.New("sequential content generation is deprecated. The system now requires Redis task queue for concurrent generation. Please ensure REDIS_URL is configured")
}

func (s *GenerationServiceImpl) validateGenerationLimits(account *models.Account, contentType enums.ContentType, quantity int) error {

	generatedCount, err := s.store.GeneratedContents.CountByAccountAndType(account.ID, contentType)
	if err != nil {
		return fmt.Errorf("failed to count generated content: %w", err)
	}

	var postingGoal int
	switch contentType {
	case enums.ContentTypeVideo:
		postingGoal = account.PostingGoal
	case enums.ContentTypeSlideshow:
		postingGoal = account.SlideshowPostingGoal
	case enums.ContentTypeStory:
		postingGoal = account.StoryPostingGoal
	default:
		return nil
	}

	if postingGoal > 0 && generatedCount+quantity > postingGoal {
		return ErrLimitExceeded
	}

	return nil
}

func (s *GenerationServiceImpl) validatePlatform(platform enums.Platform) error {
	supportedPlatforms := []enums.Platform{
		enums.PlatformTikTok,
		enums.PlatformInstagram,
	}

	if !slices.Contains(supportedPlatforms, platform) {
		return ErrUnsupportedPlatform
	}

	return nil
}

func (s *GenerationServiceImpl) getStrategy(contentType enums.ContentType) (contentStrategy, error) {
	strategy, exists := s.strategies[contentType]
	if !exists {
		return nil, ErrUnsupportedType
	}
	return strategy, nil
}

// DEPRECATED: These methods implement the old sequential generation system.
// They are no longer used in production but kept for reference.
// The concurrent generation system (generator_concurrent.go) replaces this functionality.

func (s *GenerationServiceImpl) generateMultipleContent(
	account *models.Account,
	request *dto.GenerateContent,
	strategy contentStrategy,
) ([]*models.GeneratedContent, error) {
	return nil, errors.New("sequential generation is deprecated - use concurrent generation with Redis task queue")
}

func (s *GenerationServiceImpl) generateSingleContent(
	account *models.Account,
	request *dto.GenerateContent,
	strategy contentStrategy,
) (*models.GeneratedContent, error) {
	contentAccount, err := strategy.SelectContent(account, request)
	if err != nil {
		return nil, err
	}

	if contentAccount == nil {
		return nil, ErrNoContentAvailable
	}

	content, err := s.store.Contents.GetPopulated(contentAccount.ContentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get populated content: %w", err)
	}

	textOverlay, err := s.getTextOverlayIfNeeded(account, content)
	if err != nil {
		return nil, err
	}

	generated, err := strategy.GenerateContent(account, request, contentAccount, content, textOverlay)
	if err != nil {
		return nil, err
	}

	if err := s.markContentAsGenerated(contentAccount); err != nil {
		log.Printf("[Warning] Failed to mark content as generated: %v", err)
	}

	return generated, nil
}

func (s *GenerationServiceImpl) getTextOverlayIfNeeded(account *models.Account, content *models.Content) (*models.TextOverlay, error) {
	if !content.UseOverlays || account.AccountRole == enums.AccountRoleMain {
		return &models.TextOverlay{}, nil
	}

	overlays, err := s.store.TextOverlays.GetByAccount(account.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get text overlays: %w", err)
	}

	if len(overlays) == 0 {
		return &models.TextOverlay{}, nil
	}

	return overlays[rand.IntN(len(overlays))], nil
}

func (s *GenerationServiceImpl) markContentAsGenerated(contentAccount *models.ContentAccount) error {
	if err := s.store.Contents.UpdateContentAsGenerated(contentAccount.ContentID); err != nil {
		return err
	}
	return s.store.ContentAccounts.UpdateContentAsGenerated(contentAccount.ID)
}

// contentStrategy defines the interface for different content type strategies
type contentStrategy interface {
	SelectContent(account *models.Account, request *dto.GenerateContent) (*models.ContentAccount, error)
	GenerateContent(
		account *models.Account,
		request *dto.GenerateContent,
		contentAccount *models.ContentAccount,
		content *models.Content,
		textOverlay *models.TextOverlay,
	) (*models.GeneratedContent, error)
}

// videoStrategy implements video content generation
type videoStrategy struct {
	service *GenerationServiceImpl
}

func newVideoStrategy(service *GenerationServiceImpl) contentStrategy {
	return &videoStrategy{service: service}
}

func (s *videoStrategy) SelectContent(account *models.Account, request *dto.GenerateContent) (*models.ContentAccount, error) {
	// Main accounts and TikTok use different selection logic
	if account.AccountRole == enums.AccountRoleMain || account.Platform == enums.PlatformTikTok {
		video, err := s.service.store.ContentAccounts.GetNextToGenerateInSequence(account.ID, request.Type, 2)
		if err != nil || video == nil {
			return nil, ErrNoContentAvailable
		}
		return video, nil
	}

	// 60% chance to prioritize new/popular content
	if rand.Float64() < 0.6 {
		if video := s.tryGetNewVideos(account.ID, request.Type); video != nil {
			return video, nil
		}

		if video := s.tryGetPopularVideos(account.ID, request.Type); video != nil {
			return video, nil
		}

		return s.getBackupVideo(account.ID, request.Type)
	}

	// 40% chance to get least recently generated
	return s.getBackupVideo(account.ID, request.Type)
}

func (s *videoStrategy) tryGetNewVideos(accountID uint, contentType enums.ContentType) *models.ContentAccount {
	videos, err := s.service.store.ContentAccounts.FindNeverGenerated(
		accountID,
		contentType,
		10,
	)
	if err != nil || len(videos) == 0 {
		return nil
	}
	return videos[0]
}

func (s *videoStrategy) tryGetPopularVideos(accountID uint, contentType enums.ContentType) *models.ContentAccount {
	videos, err := s.service.store.ContentAccounts.FindByMostPopularPosts(
		accountID,
		contentType,
		time.Now().AddDate(0, 0, -30),
		10,
	)
	if err != nil || len(videos) < 5 {
		return nil
	}
	return findLeastRecentlyGenerated(videos)
}

func (s *videoStrategy) getBackupVideo(accountID uint, contentType enums.ContentType) (*models.ContentAccount, error) {
	videos, err := s.service.store.ContentAccounts.FindLeastRecentlyGenerated(
		accountID,
		contentType,
		10,
	)
	if err != nil {
		return nil, err
	}

	if len(videos) == 0 {
		return nil, ErrNoContentAvailable
	}

	return videos[0], nil
}

func (s *videoStrategy) GenerateContent(
	account *models.Account,
	request *dto.GenerateContent,
	contentAccount *models.ContentAccount,
	content *models.Content,
	textOverlay *models.TextOverlay,
) (*models.GeneratedContent, error) {
	if len(content.ContentFiles) == 0 {
		return nil, ErrNoContentFiles
	}

	sourceFile := content.ContentFiles[0].File
	isMainAccount := account.AccountRole == enums.AccountRoleMain
	useMirror := shouldMirror(content.UseMirror, contentAccount.TimesGenerated)
	useOverlay := textOverlay.Content != ""

	log.Printf("[Video Generation] account=%d, content=%d, source=%d, mirror=%v, overlay=%v",
		request.AccountID, contentAccount.ContentID, sourceFile.ID, useMirror, useOverlay)

	renderResult, err := s.service.repurposeService.RenderVideoWithThumbnailV2(
		int(sourceFile.ID),
		textOverlay.Content,
		useMirror,
		useOverlay,
		isMainAccount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to render video: %w", err)
	}

	return s.createGeneratedContent(
		account.ID,
		content,
		contentAccount,
		textOverlay,
		renderResult,
		useMirror,
		useOverlay,
	)
}

func (s *videoStrategy) createGeneratedContent(
	accountID uint,
	content *models.Content,
	contentAccount *models.ContentAccount,
	textOverlay *models.TextOverlay,
	renderResult *content.ButterRepurposerTaskResponse,
	useMirror bool,
	useOverlay bool,
) (*models.GeneratedContent, error) {
	// Extract IDs from render result
	videoFileID := uint(renderResult.Result.VideoID)
	thumbnailFileID := uint(renderResult.Result.ThumbnailID)

	log.Printf("[Video Generation] Generated files - video=%d, thumbnail=%d, account=%d",
		videoFileID, thumbnailFileID, accountID)

	videoFile, thumbnailFile, err := s.validateAndGetFiles(videoFileID, thumbnailFileID)
	if err != nil {
		return nil, err
	}

	generatedContent := &models.GeneratedContent{
		ContentID:        content.ID,
		ContentAccountID: contentAccount.ID,
		TextOverlayID:    textOverlay.ID,
		AccountID:        accountID,
		Type:             enums.ContentTypeVideo,
		UsedMirror:       useMirror,
		UsedOverlay:      useOverlay,
		Files: []*models.GeneratedContentFile{{
			FileID:        videoFile.ID,
			ThumbnailID:   thumbnailFile.ID,
			FileHash:      renderResult.Result.VideoHash,
			ThumbnailHash: renderResult.Result.ThumbnailHash,
		}},
	}

	if err := s.service.store.GeneratedContents.Create(generatedContent); err != nil {
		s.cleanupFiles(videoFile.ID, thumbnailFile.ID)
		return nil, fmt.Errorf("failed to create generated content: %w", err)
	}

	return s.service.store.GeneratedContents.Get(generatedContent.ID)
}

func (s *videoStrategy) validateAndGetFiles(videoID, thumbnailID uint) (*models.File, *models.File, error) {
	videoFile, err := s.service.fileService.GetFile(videoID)
	if err != nil {
		s.service.fileService.DeleteFile(videoID)
		return nil, nil, fmt.Errorf("failed to get video file: %w", err)
	}

	thumbnailFile, err := s.service.fileService.GetFile(thumbnailID)
	if err != nil {
		s.cleanupFiles(videoID, thumbnailID)
		return nil, nil, fmt.Errorf("failed to get thumbnail file: %w", err)
	}

	return videoFile, thumbnailFile, nil
}

func (s *videoStrategy) cleanupFiles(fileIDs ...uint) {
	for _, id := range fileIDs {
		if err := s.service.fileService.DeleteFile(id); err != nil {
			log.Printf("[Warning] Failed to cleanup file %d: %v", id, err)
		}
	}
}

type storyStrategy struct {
	service *GenerationServiceImpl
}

func newStoryStrategy(service *GenerationServiceImpl) contentStrategy {
	return &storyStrategy{service: service}
}

func (s *storyStrategy) SelectContent(account *models.Account, request *dto.GenerateContent) (*models.ContentAccount, error) {
	video, err := s.service.store.ContentAccounts.GetNextToGenerateInSequence(account.ID, request.Type, 0)
	if err != nil || video == nil {
		return nil, ErrNoContentAvailable
	}
	return video, nil
}

func (s *storyStrategy) GenerateContent(
	account *models.Account,
	request *dto.GenerateContent,
	contentAccount *models.ContentAccount,
	content *models.Content,
	textOverlay *models.TextOverlay,
) (*models.GeneratedContent, error) {
	if len(content.ContentFiles) == 0 {
		return nil, ErrNoContentFiles
	}

	sourceContentFile := content.ContentFiles[0]

	generatedContent := &models.GeneratedContent{
		ContentID:        content.ID,
		ContentAccountID: contentAccount.ID,
		TextOverlayID:    textOverlay.ID,
		AccountID:        account.ID,
		Type:             enums.ContentTypeStory,
		Files: []*models.GeneratedContentFile{{
			FileID:      sourceContentFile.File.ID,
			ThumbnailID: sourceContentFile.Thumbnail.ID,
		}},
	}

	if err := s.service.store.GeneratedContents.Create(generatedContent); err != nil {
		return nil, fmt.Errorf("failed to create generated content: %w", err)
	}

	return s.service.store.GeneratedContents.Get(generatedContent.ID)
}

type slideshowStrategy struct {
	service *GenerationServiceImpl
}

func newSlideshowStrategy(service *GenerationServiceImpl) contentStrategy {
	return &slideshowStrategy{service: service}
}

func (s *slideshowStrategy) SelectContent(account *models.Account, request *dto.GenerateContent) (*models.ContentAccount, error) {
	limit := 0
	if account.AccountRole == enums.AccountRoleMain {
		limit = 2
	}

	slideshow, err := s.service.store.ContentAccounts.GetNextToGenerateInSequence(account.ID, request.Type, limit)
	if err != nil || slideshow == nil {
		return nil, ErrNoContentAvailable
	}
	return slideshow, nil
}

func (s *slideshowStrategy) GenerateContent(
	account *models.Account,
	request *dto.GenerateContent,
	contentAccount *models.ContentAccount,
	content *models.Content,
	textOverlay *models.TextOverlay,
) (*models.GeneratedContent, error) {
	if len(content.ContentFiles) == 0 {
		return nil, ErrNoContentFiles
	}

	var files []*models.GeneratedContentFile
	for _, cf := range content.ContentFiles {
		files = append(files, &models.GeneratedContentFile{
			FileID:      cf.File.ID,
			ThumbnailID: cf.Thumbnail.ID,
		})
	}

	generatedContent := &models.GeneratedContent{
		ContentID:        content.ID,
		ContentAccountID: contentAccount.ID,
		TextOverlayID:    textOverlay.ID,
		AccountID:        account.ID,
		Type:             enums.ContentTypeSlideshow,
		Files:            files,
	}

	if err := s.service.store.GeneratedContents.Create(generatedContent); err != nil {
		return nil, fmt.Errorf("failed to create generated content: %w", err)
	}

	return s.service.store.GeneratedContents.Get(generatedContent.ID)
}

// Utility functions
func findLeastRecentlyGenerated(contents []*models.ContentAccount) *models.ContentAccount {
	for _, content := range contents {
		if content.LastGeneratedAt.IsZero() {
			return content
		}
	}

	leastRecent := contents[0]
	for _, content := range contents[1:] {
		if content.LastGeneratedAt.Before(leastRecent.LastGeneratedAt) {
			leastRecent = content
		}
	}

	return leastRecent
}

func shouldMirror(useMirror bool, timesGenerated int) bool {
	return useMirror && timesGenerated%2 == 0
}
