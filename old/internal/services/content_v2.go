package services

import (
	"errors"
	"fmt"

	"github.com/nicolailuther/butter/internal/domain"
	"github.com/nicolailuther/butter/internal/dto"
	"github.com/nicolailuther/butter/internal/enums"
	"github.com/nicolailuther/butter/internal/models"
	"github.com/nicolailuther/butter/pkg/content"
	"github.com/nicolailuther/butter/pkg/taskqueue"
	"github.com/nicolailuther/butter/pkg/utils"
	"gorm.io/gorm"
)

type ContentService interface {
	GetContents(filters *dto.GetFolderFilters) (*models.ContentFolder, error)
	CreateFolder(request *dto.CreateFolder) (*models.ContentFolder, error)
	UpdateFolder(folderID uint, request map[string]interface{}) (*models.ContentFolder, error)
	DeleteFolder(folderID uint) error
	CreateContent(request *dto.CreateContent) (*models.Content, error)
	UpdateContent(contentID uint, updateData map[string]interface{}) (*models.Content, error)
	DeleteContent(contentID uint) error

	AssignAccountToContent(contentID, accountID uint) (*models.Content, error)
	UnassignAccountFromContent(contentID, accountID uint) (*models.Content, error)
	GetContentsByAccount(accountID uint) ([]*models.ContentAccount, error)
	UpdateContentAccount(contentID, accountID uint, updateData map[string]interface{}) (*models.ContentAccount, error)

	// Generation methods
	GenerateContent(request *dto.GenerateContent) ([]*models.GeneratedContent, error)
	GetGeneratedContent(accountID uint) ([]*models.GeneratedContent, error)
	GetGeneratedContentByID(generatedContentID uint) (*models.GeneratedContent, error)
	UpdateGeneratedContent(generatedContentID uint, request *dto.UpdateGeneratedContent) (*models.GeneratedContent, error)
	DeleteGeneratedContent(generatedContentID uint) error

	// Thumbnail generation v2
	GenerateThumbnailV2(request *dto.GenerateThumbnail) (*dto.ThumbnailResult, error)
}

type contentServiceImpl struct {
	*Service
	generationServiceV2  ContentGenerationService
	concurrentGeneration ConcurrentContentGenerationService
	taskManager          taskqueue.TaskManager
	mediaService         content.MediaService
	repurposeService     content.MediaService
	fileService          FileService
}

// NewContentService creates a content service with concurrent generation support
func NewContentService(
	container *Service,
	generationServiceV2 ContentGenerationService,
	concurrentGeneration ConcurrentContentGenerationService,
	taskManager taskqueue.TaskManager,
	mediaService content.MediaService,
	repurposeService content.MediaService,
	fileService FileService,
) ContentService {
	return &contentServiceImpl{
		Service:              container,
		generationServiceV2:  generationServiceV2,
		concurrentGeneration: concurrentGeneration,
		taskManager:          taskManager,
		mediaService:         mediaService,
		repurposeService:     repurposeService,
		fileService:          fileService,
	}
}

func (s *contentServiceImpl) CreateFolder(request *dto.CreateFolder) (*models.ContentFolder, error) {
	if err := domain.ValidateCreateFolder(request); err != nil {
		return nil, err
	}

	// TODO: Check user permissions

	// Check client exists
	client, err := s.store.Clients.GetByID(request.ClientID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if client == nil {
		return nil, errors.New("client not found")
	}

	// Check parent folder exists if parent_id is provided
	if request.ParentID != 0 {
		parentFolder, err := s.store.ContentFolders.Get(request.ParentID)
		if err != nil {
			return nil, err
		}

		// Check parent folder belongs to the same client
		if parentFolder.ClientID != request.ClientID {
			return nil, errors.New("parent folder does not belong to the same client")
		}
	}

	// Create folder in DB
	folder := &models.ContentFolder{
		Name:     request.Name,
		ClientID: request.ClientID,
		ParentID: request.ParentID,
	}

	if err := s.store.ContentFolders.Create(folder); err != nil {
		return nil, err
	}

	return folder, nil
}

func (s *contentServiceImpl) UpdateFolder(folderID uint, request map[string]interface{}) (*models.ContentFolder, error) {
	if folderID == 0 {
		return nil, errors.New("folder ID is required")
	}

	var payload dto.UpdateFolder
	if err := utils.MapToStruct(request, &payload); err != nil {
		return nil, err
	}

	if err := domain.ValidateUpdateFolder(&payload); err != nil {
		return nil, err
	}

	// Get existing folder
	folder, err := s.store.ContentFolders.Get(folderID)
	if err != nil {
		return nil, err
	}

	if folder == nil {
		return nil, errors.New("folder not found")
	}

	// TODO: Check user permissions to ensure they can update this folder

	// Validate parent folder if provided
	if payload.ParentID != nil && *payload.ParentID != 0 {
		// Check that parent folder exists
		parentFolder, err := s.store.ContentFolders.Get(*payload.ParentID)
		if err != nil {
			return nil, err
		}

		// Check parent folder belongs to the same client
		if parentFolder.ClientID != folder.ClientID {
			return nil, errors.New("parent folder does not belong to the same client")
		}

		// Prevent circular reference (folder cannot be its own parent or descendant)
		if *payload.ParentID == folderID {
			return nil, errors.New("folder cannot be its own parent")
		}

		// TODO: Add more sophisticated cycle detection for deeper hierarchies if needed
	}

	// Update in database
	if err := s.store.ContentFolders.Patch(folderID, request); err != nil {
		return nil, err
	}

	updatedFolder, err := s.store.ContentFolders.Get(folderID)
	if err != nil {
		return nil, err
	}

	return updatedFolder, nil
}

func (s *contentServiceImpl) DeleteFolder(folderID uint) error {
	if folderID == 0 {
		return errors.New("folder ID is required")
	}

	// Get existing folder to validate existence and permissions
	_, err := s.store.ContentFolders.Get(folderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("folder not found")
		}
		return err
	}

	// TODO: Check user permissions to ensure they can delete this folder

	// Check if folder has content (subfolders or content items)
	hasContent, err := s.store.ContentFolders.HasContent(folderID)
	if err != nil {
		return err
	}

	if hasContent {
		return errors.New("cannot delete folder that contains content or subfolders")
	}

	// Delete the folder
	if err := s.store.ContentFolders.Delete(folderID); err != nil {
		return err
	}

	return nil
}

func (s *contentServiceImpl) GetContents(filters *dto.GetFolderFilters) (*models.ContentFolder, error) {

	if err := domain.ValidateGetFolderFilters(filters); err != nil {
		return nil, err
	}

	// Get root folder
	if filters.FolderID == 0 {
		rootFolders, err := s.store.ContentFolders.GetRootFolders(filters.ClientID)
		if err != nil {
			return nil, err
		}

		rootContents, err := s.store.Contents.GetRootContents(filters.ClientID)
		if err != nil {
			return nil, err
		}

		folder := &models.ContentFolder{
			Name:     "Root",
			ClientID: filters.ClientID,
			Children: rootFolders,
			Contents: rootContents,
		}

		return folder, nil
	}

	// Get specific folder by ID
	folder, err := s.store.ContentFolders.GetPopulated(filters.FolderID)
	if err != nil {
		return nil, err
	}

	// Check folder belongs to the client
	if folder.ClientID != filters.ClientID {
		return nil, errors.New("folder does not belong to the client")
	}

	return folder, nil
}

func (s *contentServiceImpl) CreateContent(request *dto.CreateContent) (*models.Content, error) {

	if err := domain.ValidateCreateContent(request); err != nil {
		return nil, err
	}

	// Check client exists
	client, err := s.store.Clients.GetByID(request.ClientID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if client == nil {
		return nil, errors.New("client not found")
	}

	// Check folder exists if folder_id is provided
	if request.FolderID != 0 {
		folder, err := s.store.ContentFolders.Get(request.FolderID)
		if err != nil {
			return nil, err
		}

		// Check folder belongs to the same client
		if folder.ClientID != request.ClientID {
			return nil, errors.New("folder does not belong to the same client")
		}
	}

	cleanupFiles := func(files []*models.File) {
		if len(files) == 0 {
			return
		}
		ids := make([]uint, len(files))
		for i, f := range files {
			ids[i] = f.ID
		}
		s.fileService.BulkDeleteFiles(ids)
	}

	cleanupContentFiles := func(contentFiles []*models.ContentFile) {
		if len(contentFiles) == 0 {
			return
		}
		var ids []uint
		for _, cf := range contentFiles {
			ids = append(ids, cf.FileID, cf.ThumbnailID)
		}
		s.fileService.BulkDeleteFiles(ids)
	}

	// Process files
	var contentFiles []*models.ContentFile
	var files []*models.File
	for i, fileHeader := range request.ContentFiles {

		// Validate file type based on content type before upload
		if err := s.validateFileTypeForContent(request.Type, fileHeader.Filename); err != nil {
			return nil, err
		}

		// Upload file
		createdFile, err := s.fileService.UploadFileFromMultipart(fileHeader)
		if err != nil {
			cleanupFiles(files)
			return nil, err
		}
		files = append(files, createdFile)

		exists, err := s.store.Contents.FileExistsForClient(createdFile.Etag, request.ClientID)
		if err != nil {
			s.logger.Errorln("Failed to check for existing file:", err.Error())
			cleanupFiles(files)
			return nil, err
		}

		if exists {
			cleanupFiles(files)
			return nil, errors.New("a file with the name '" + fileHeader.Filename + "' already exists for this client")
		}

		// Create thumbnail
		thumbnail, err := s.createThumbnail(createdFile.ID)
		if err != nil {
			cleanupFiles(files)
			return nil, err
		}
		files = append(files, thumbnail)

		// Add to content files
		contentFiles = append(contentFiles, &models.ContentFile{
			FileID:      createdFile.ID,
			ThumbnailID: thumbnail.ID,
			Order:       i,
		})
	}

	content := &models.Content{
		Name:         request.ContentFiles[0].Filename,
		Type:         request.Type,
		FolderID:     request.FolderID,
		ClientID:     request.ClientID,
		ContentFiles: contentFiles,
		Enabled:      true,
	}

	if err := s.store.Contents.Create(content); err != nil {
		cleanupContentFiles(contentFiles)
		return nil, err
	}

	created, err := s.store.Contents.GetPopulated(content.ID)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *contentServiceImpl) UpdateContent(contentID uint, updateData map[string]interface{}) (*models.Content, error) {

	if contentID == 0 {
		return nil, errors.New("content id is required")
	}

	// Get existing content first to validate permissions
	existingContent, err := s.store.Contents.Get(contentID)
	if err != nil {
		return nil, err
	}

	if existingContent == nil {
		return nil, errors.New("content not found")
	}

	var request dto.UpdateContent
	if err := utils.MapToStruct(updateData, &request); err != nil {
		return nil, err
	}

	if err := domain.ValidateUpdateContent(&request); err != nil {
		return nil, err
	}

	// Validate folder movement if folder_id is being updated
	if request.FolderID != nil {
		if err := s.validateFolderMovement(existingContent, *request.FolderID); err != nil {
			return nil, err
		}
	}

	if err := s.store.Contents.Patch(contentID, updateData); err != nil {
		return nil, err
	}

	content, err := s.store.Contents.GetPopulated(contentID)
	if err != nil {
		return nil, err
	}

	return content, nil
}

func (s *contentServiceImpl) DeleteContent(contentID uint) error {
	if contentID == 0 {
		return errors.New("content ID cannot be zero")
	}

	// Get content first to validate existence
	content, err := s.store.Contents.Get(contentID)
	if err != nil {
		return err
	}

	if content == nil {
		return errors.New("content not found")
	}

	var filesToDelete []uint

	// Get content files
	contentFiles, err := s.store.ContentFiles.GetByContent(contentID)
	if err != nil {
		return err
	}

	for _, cf := range contentFiles {
		filesToDelete = append(filesToDelete, cf.FileID, cf.ThumbnailID)
	}

	// Get generated contents and their files
	generatedContentsV2, err := s.store.GeneratedContents.GetByContent(contentID)
	if err != nil {
		return err
	}

	// Only collect file IDs from video-type generated content
	// Stories and slideshows reference original files and should not be deleted
	filesToDelete = append(filesToDelete, collectVideoFileIDs(generatedContentsV2)...)

	var generatedContentIDs []uint
	for _, gc := range generatedContentsV2 {
		generatedContentIDs = append(generatedContentIDs, gc.ID)
	}

	// Delete content and associated content files using transaction
	if err := s.store.Contents.DeleteWithFiles(contentID); err != nil {
		return err
	}

	// Delete all generated contents
	if err := s.store.GeneratedContents.DeleteMany(generatedContentIDs); err != nil {
		return err
	}

	// After successful database deletion, delete files from storage
	// If file deletion fails, we log it but don't fail the entire operation
	if len(filesToDelete) > 0 {
		go func() {
			if err := s.fileService.BulkDeleteFiles(filesToDelete); err != nil {
				s.logger.Errorf("Failed to delete files from storage for content %d: %v", contentID, err)
				// Continue execution - database cleanup was successful
			}
		}()
	}

	return nil
}

func (s *contentServiceImpl) createThumbnail(fileID uint) (*models.File, error) {
	if fileID == 0 {
		return nil, errors.New("file ID cannot be empty")
	}

	// Get the file to check its type
	file, err := s.fileService.GetFile(fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %v", err)
	}

	// Validate that the file is either a video or image format
	if !utils.IsVideoFile(file.Url) && !utils.IsImageFile(file.Url) {
		return nil, fmt.Errorf("unsupported file type for thumbnail generation. Only video and image files are supported")
	}

	// Use repurpose engine v2 to generate thumbnail (works for both images and videos)
	result, err := s.repurposeService.GenerateThumbnailV2(int(fileID))
	if err != nil {
		return nil, err
	}

	if result.Result == nil {
		return nil, errors.New("thumbnail generation completed but no result returned")
	}

	thumbnailFileID := uint(result.Result.ThumbnailID)

	// Get the thumbnail file from the database
	thumbnailFile, err := s.fileService.GetFile(thumbnailFileID)
	if err != nil {
		return nil, err
	}

	return thumbnailFile, nil
}

// validateFolderMovement validates that content can be moved to the target folder
func (s *contentServiceImpl) validateFolderMovement(existingContent *models.Content, targetFolderID uint) error {
	// If moving to root (folder_id = 0), no additional validation needed
	if targetFolderID == 0 {
		return nil
	}

	// Check if target folder exists
	targetFolder, err := s.store.ContentFolders.Get(targetFolderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("target folder not found")
		}
		return err
	}

	// Check if target folder belongs to the same client as the content
	if targetFolder.ClientID != existingContent.ClientID {
		return errors.New("cannot move content to a folder belonging to a different client")
	}

	return nil
}

func (s *contentServiceImpl) AssignAccountToContent(contentID, accountID uint) (*models.Content, error) {
	if contentID == 0 {
		return nil, errors.New("content id is required")
	}
	if accountID == 0 {
		return nil, errors.New("account id is required")
	}

	// Check if content exists
	content, err := s.store.Contents.Get(contentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("content not found")
		}
		return nil, err
	}

	// Check if account exists
	account, err := s.store.Accounts.GetByID(accountID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("account not found")
		}
		return nil, err
	}

	// TODO: Add permission checks here if needed
	// Ensure account belongs to the same client as content
	if account.ClientID != content.ClientID {
		return nil, errors.New("account does not belong to the same client as content")
	}

	err = s.store.ContentAccounts.Create(&models.ContentAccount{
		ContentID: contentID,
		AccountID: accountID,
		Enabled:   true,
	})
	if err != nil {
		return nil, err
	}

	updatedContent, err := s.store.Contents.GetPopulated(contentID)
	if err != nil {
		return nil, err
	}

	return updatedContent, nil
}

func (s *contentServiceImpl) UnassignAccountFromContent(contentID, accountID uint) (*models.Content, error) {

	if contentID == 0 {
		return nil, errors.New("content id is required")
	}

	if accountID == 0 {
		return nil, errors.New("account id is required")
	}

	// Check if content exists
	_, err := s.store.Contents.Get(contentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("content not found")
		}
		return nil, err
	}

	// Check if account exists
	_, err = s.store.Accounts.GetByID(accountID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("account not found")
		}
		return nil, err
	}

	// Check if assignment exists
	assignment, err := s.store.ContentAccounts.GetByContentAndAccountRaw(contentID, accountID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("content is not assigned to the account")
		}
		return nil, err
	}

	if assignment == nil {
		return nil, errors.New("content is not assigned to the account")
	}

	// Get generated contents for this specific content-account assignment
	generatedContentsV2, err := s.store.GeneratedContents.GetByContentAccount(assignment.ID)
	if err != nil {
		return nil, err
	}

	// Only collect file IDs from video-type generated content
	// Stories and slideshows reference original files and should not be deleted
	filesToDelete := collectVideoFileIDs(generatedContentsV2)

	var generatedContentIDs []uint
	for _, gc := range generatedContentsV2 {
		generatedContentIDs = append(generatedContentIDs, gc.ID)
	}

	// Delete assignment
	err = s.store.ContentAccounts.Delete(contentID, accountID)
	if err != nil {
		return nil, err
	}

	// Delete all generated contents
	if err := s.store.GeneratedContents.DeleteMany(generatedContentIDs); err != nil {
		return nil, err
	}

	// After successful database deletion, delete files from storage
	// If file deletion fails, we log it but don't fail the entire operation
	if len(filesToDelete) > 0 {
		go func() {
			if err := s.fileService.BulkDeleteFiles(filesToDelete); err != nil {
				s.logger.Errorf("Failed to delete files from storage for content %d: %v", contentID, err)
				// Continue execution - database cleanup was successful
			}
		}()
	}

	updatedContent, err := s.store.Contents.GetPopulated(contentID)
	if err != nil {
		return nil, err
	}

	return updatedContent, nil
}

func (s *contentServiceImpl) GetContentsByAccount(accountID uint) ([]*models.ContentAccount, error) {
	if accountID == 0 {
		return nil, errors.New("account id is required")
	}

	return s.store.ContentAccounts.GetByAccount(accountID)
}

func (s *contentServiceImpl) UpdateContentAccount(contentID, accountID uint, updateData map[string]interface{}) (*models.ContentAccount, error) {
	if contentID == 0 {
		return nil, errors.New("content id is required")
	}
	if accountID == 0 {
		return nil, errors.New("account id is required")
	}

	// Check if the ContentAccount exists
	existingContentAccount, err := s.store.ContentAccounts.GetByContentAndAccount(contentID, accountID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("content account assignment not found")
		}
		return nil, err
	}

	// TODO: Add permission checks here if needed
	// Ensure the content and account belong to the same client
	if existingContentAccount.Content != nil && existingContentAccount.Account != nil {
		if existingContentAccount.Content.ClientID != existingContentAccount.Account.ClientID {
			return nil, errors.New("content and account do not belong to the same client")
		}
	}

	// Validate and convert the updateData to DTO for validation
	var request dto.UpdateContentAccount
	if err := utils.MapToStruct(updateData, &request); err != nil {
		return nil, errors.New("invalid data: " + err.Error())
	}

	if err := domain.ValidateUpdateContentAccount(&request); err != nil {
		return nil, err
	}

	// Update the ContentAccount in the database
	if err := s.store.ContentAccounts.UpdateByContentAndAccount(contentID, accountID, updateData); err != nil {
		return nil, err
	}

	// Return the updated ContentAccount
	updatedContentAccount, err := s.store.ContentAccounts.GetByContentAndAccount(contentID, accountID)
	if err != nil {
		return nil, err
	}

	return updatedContentAccount, nil
}

func (s *contentServiceImpl) GenerateContent(request *dto.GenerateContent) ([]*models.GeneratedContent, error) {
	if err := domain.ValidateGenerateContent(request); err != nil {
		return nil, err
	}

	// Use concurrent generation with Redis task queue (always available)
	return s.concurrentGeneration.GenerateContentConcurrent(request, s.taskManager)
}

func (s *contentServiceImpl) GetGeneratedContent(accountID uint) ([]*models.GeneratedContent, error) {
	if accountID == 0 {
		return nil, errors.New("account ID cannot be zero")
	}

	return s.store.GeneratedContents.GetByAccount(accountID)
}

func (s *contentServiceImpl) DeleteGeneratedContent(generatedContentID uint) error {
	if generatedContentID == 0 {
		return errors.New("generated content ID cannot be zero")
	}

	// Check generated content
	existing, err := s.store.GeneratedContents.Get(generatedContentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("generated content not found")
		}
		return err
	}

	contentFiles, err := s.store.GeneratedContentFiles.GetByGeneratedContent(generatedContentID)
	if err != nil {
		return err
	}

	// Collect associated file IDs before deletion
	// Only delete files for video content type, as stories and slideshows reference original files
	var fileIDs []uint
	if existing.Type == enums.ContentTypeVideo {
		for _, cf := range contentFiles {
			fileIDs = append(fileIDs, cf.FileID, cf.ThumbnailID)
		}
	}

	// Delete the generated content record and its associated files
	if err := s.store.GeneratedContents.DeleteTX(generatedContentID); err != nil {
		return err
	}

	if err := s.fileService.BulkDeleteFiles(fileIDs); err != nil {
		s.logger.Errorf("Failed to delete files from storage for generated content %d: %v", generatedContentID, err)
	}

	return nil
}

// collectVideoFileIDs collects file IDs only from video-type generated content.
// Stories and slideshows reference original files and should not have their files deleted.
// Note: Images cannot be generated (no strategy exists), so they are not handled here.
func collectVideoFileIDs(generatedContents []*models.GeneratedContent) []uint {
	var fileIDs []uint
	for _, gc := range generatedContents {
		if gc.Type == enums.ContentTypeVideo {
			for _, cf := range gc.Files {
				fileIDs = append(fileIDs, cf.FileID, cf.ThumbnailID)
			}
		}
	}
	return fileIDs
}

func (s *contentServiceImpl) GetGeneratedContentByID(generatedContentID uint) (*models.GeneratedContent, error) {
	if generatedContentID == 0 {
		return nil, errors.New("generated content ID cannot be zero")
	}

	generatedContent, err := s.store.GeneratedContents.Get(generatedContentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("generated content not found")
		}
		return nil, err
	}

	return generatedContent, nil
}

func (s *contentServiceImpl) UpdateGeneratedContent(generatedContentID uint, request *dto.UpdateGeneratedContent) (*models.GeneratedContent, error) {
	if generatedContentID == 0 {
		return nil, errors.New("generated content ID cannot be zero")
	}

	if err := domain.ValidateUpdateGeneratedContent(request); err != nil {
		return nil, err
	}

	// Check if generated content exists
	_, err := s.store.GeneratedContents.Get(generatedContentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("generated content not found")
		}
		return nil, err
	}

	// Build update data from allowed fields only
	updateData := make(map[string]interface{})
	if request.IsPosted != nil {
		updateData["is_posted"] = *request.IsPosted
	}
	if request.MaybePosted != nil {
		updateData["maybe_posted"] = *request.MaybePosted
	}

	// Only update if there are fields to update
	if len(updateData) > 0 {
		if err := s.store.GeneratedContents.Patch(generatedContentID, updateData); err != nil {
			return nil, err
		}
	}

	// Return the updated generated content
	return s.store.GeneratedContents.Get(generatedContentID)
}

func (s *contentServiceImpl) GenerateThumbnailV2(request *dto.GenerateThumbnail) (*dto.ThumbnailResult, error) {
	if err := domain.ValidateGenerateThumbnail(request); err != nil {
		return nil, err
	}

	// Verify file exists
	file, err := s.fileService.GetFile(request.FileID)
	if err != nil {
		return nil, errors.New("file not found")
	}

	// Generate thumbnail using repurpose engine v2
	result, err := s.repurposeService.GenerateThumbnailV2(int(file.ID))
	if err != nil {
		return nil, err
	}

	if result.Result == nil {
		return nil, errors.New("thumbnail generation completed but no result returned")
	}

	// Return the result
	return &dto.ThumbnailResult{
		ProcessingTime:  int64(result.Result.ProcessingTimeMs),
		ThumbnailFileID: uint(result.Result.ThumbnailID),
		ThumbnailHash:   result.Result.ThumbnailHash,
	}, nil
}

// validateFileTypeForContent validates that a file type is appropriate for the given content type
func (s *contentServiceImpl) validateFileTypeForContent(contentType enums.ContentType, filename string) error {
	switch contentType {
	case enums.ContentTypeVideo:
		// Video content only accepts video files
		if !utils.IsVideoFile(filename) {
			supportedFormats := utils.GetSupportedVideoExtensions()
			return fmt.Errorf("unsupported file type for '%s'. Video content only accepts video files. Supported formats: %v", filename, supportedFormats)
		}
	case enums.ContentTypeStory, enums.ContentTypeSlideshow, enums.ContentTypeImage:
		// Story, slideshow, and image content accept both images and videos
		if !utils.IsVideoOrImageFile(filename) {
			videoFormats := utils.GetSupportedVideoExtensions()
			imageFormats := utils.GetSupportedImageExtensions()
			return fmt.Errorf("unsupported file type for '%s'. This content type accepts images and videos. Supported image formats: %v, supported video formats: %v", filename, imageFormats, videoFormats)
		}
	default:
		return fmt.Errorf("unknown content type: %s", contentType)
	}
	return nil
}

// DeleteClientContents deletes all content, content files, and content folders for a client in a transaction.
// This method should be called within an existing transaction (tx) to ensure atomicity with other operations.
