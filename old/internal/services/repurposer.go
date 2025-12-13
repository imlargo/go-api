package services

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/nicolailuther/butter/internal/dto"
	"github.com/nicolailuther/butter/internal/enums"
	"github.com/nicolailuther/butter/internal/models"
	"github.com/nicolailuther/butter/pkg/ffmpeg"
	"github.com/nicolailuther/butter/pkg/files"
	repurposer "github.com/nicolailuther/butter/pkg/repurposer"
	"github.com/nicolailuther/butter/pkg/storage"
	"github.com/nicolailuther/butter/pkg/transform"
	"github.com/nicolailuther/butter/pkg/utils"
	"go.uber.org/zap"
)

const (
	DownloadDir = "./tmp/downloads"
)

// RepurposerService is an enhanced repurposer service that works with the task queue
type RepurposerService interface {
	// Process a video repurpose task - this is called by the task queue worker
	ProcessRepurposeTask(request *dto.ReporpuseVideo) (*dto.ReporpuseContentResult, error)

	// GenerateThumbnail generates a thumbnail synchronously (same as before)
	GenerateThumbnail(request dto.GenerateThumbnail) (*dto.ThumbnailResult, error)
}

type repurposerService struct {
	*Service
	fileService  FileService
	ffmpegClient *ffmpeg.FFmpeg
}

// NewRepurposerService creates a new task-queue-enabled repurposer service
func NewRepurposerService(service *Service, fileService FileService, ffmpegClient *ffmpeg.FFmpeg) RepurposerService {
	return &repurposerService{
		Service:      service,
		fileService:  fileService,
		ffmpegClient: ffmpegClient,
	}
}

// ProcessRepurposeTask processes a video repurpose task
func (s *repurposerService) ProcessRepurposeTask(request *dto.ReporpuseVideo) (*dto.ReporpuseContentResult, error) {
	if request.FileID == 0 {
		return nil, fmt.Errorf("file ID cannot be 0")
	}

	// Initialize repurposer engine
	engine, err := repurposer.New(s.ffmpegClient, transform.DefaultParameters(), repurposer.DefaultOptions())
	if err != nil {
		return nil, fmt.Errorf("failed to create repurposer engine: %v", err)
	}

	// Download the video
	downloadedFilePath, err := s.downloadVideo(request.FileID)
	if err != nil {
		return nil, fmt.Errorf("failed to download video: %v", err)
	}

	defer func() {
		go cleanup(
			downloadedFilePath,
			engine.GetOutputPath(downloadedFilePath, files.ExtMp4),
			engine.GetThumbnailPath(engine.GetOutputPath(downloadedFilePath, files.ExtMp4)),
		)
	}()

	// Render video
	log.Printf("[INFO] Repurpose options: UseMirror=%v, UseOverlays=%v, TextOverlay=%q", request.UseMirror, request.UseOverlays, request.TextOverlay)
	result, err := engine.RepurposeVideo(context.Background(), repurposer.RenderOptions{
		InputPath:   downloadedFilePath,
		Template:    nil, // Use default template
		UseMirror:   request.UseMirror,
		UseOverlays: request.UseOverlays,
		TextOverlay: request.TextOverlay,
		IsMain:      request.IsMain,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to render video: %v", err)
	}

	// Generate thumbnail
	thumbnailResult, err := engine.RenderThumbnail(repurposer.RenderOptions{
		InputPath: result.OutputPath,
		Template:  nil, // Use default template
	})
	if err != nil {
		return nil, fmt.Errorf("failed to render video: %v", err)
	}

	// Upload files to storage with hash computation
	repurposed, videoHash, err := s.uploadLocalFileWithHash(result.OutputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to upload repurposed video: %v", err)
	}

	thumbnail, thumbnailHash, err := s.uploadLocalFileWithHash(thumbnailResult.OutputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to upload thumbnail: %v", err)
	}

	log.Printf("[Debug]: Rendered video URL: %s, Processing time (s): %.2f", repurposed.Url, result.Duration.Seconds())
	log.Printf("[Debug]: Video hash (SHA-256): %s", videoHash)
	log.Printf("[Debug]: Thumbnail hash (SHA-256): %s", thumbnailHash)

	return &dto.ReporpuseContentResult{
		ProcessingTime:  result.Duration.Milliseconds(),
		RenderedFileID:  repurposed.ID,
		ThumbnailFileID: thumbnail.ID,
		VideoHash:       videoHash,
		ThumbnailHash:   thumbnailHash,
	}, nil
}

// GenerateThumbnail generates a thumbnail synchronously
func (s *repurposerService) GenerateThumbnail(request dto.GenerateThumbnail) (*dto.ThumbnailResult, error) {
	if request.FileID == 0 {
		return nil, fmt.Errorf("file ID cannot be 0")
	}

	// Get file to check its type before processing
	file, err := s.fileService.GetFile(request.FileID)
	if err != nil {
		return nil, fmt.Errorf("file not found: %v", err)
	}

	// Validate that the file is either a video or image format
	if !utils.IsVideoFile(file.Url) && !utils.IsImageFile(file.Url) {
		videoFormats := utils.GetSupportedVideoExtensions()
		imageFormats := utils.GetSupportedImageExtensions()
		return nil, fmt.Errorf("unsupported file type. Supported video formats: %v, supported image formats: %v", videoFormats, imageFormats)
	}

	// Initialize repurposer engine
	engine, err := repurposer.New(s.ffmpegClient, transform.DefaultParameters(), repurposer.DefaultOptions())
	if err != nil {
		return nil, fmt.Errorf("failed to create repurposer engine: %v", err)
	}

	// Download the video
	downloadedFilePath, err := s.downloadVideo(request.FileID)
	if err != nil {
		return nil, fmt.Errorf("failed to download video: %v", err)
	}

	defer func() {
		go cleanup(
			downloadedFilePath,
			"",
			engine.GetThumbnailPath(downloadedFilePath),
		)
	}()

	// Generate thumbnail directly from the downloaded video
	log.Printf("[INFO] Generating thumbnail for file ID: %d", request.FileID)
	thumbnailResult, err := engine.RenderThumbnail(repurposer.RenderOptions{
		InputPath: downloadedFilePath,
		Template:  nil, // Use default template
	})
	if err != nil {
		return nil, fmt.Errorf("failed to render thumbnail: %v", err)
	}

	// Upload thumbnail to storage with hash computation
	thumbnail, thumbnailHash, err := s.uploadLocalFileWithHash(thumbnailResult.OutputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to upload thumbnail: %v", err)
	}

	log.Printf("[Debug]: Thumbnail URL: %s, Processing time (s): %.2f", thumbnail.Url, thumbnailResult.Duration.Seconds())
	log.Printf("[Debug]: Thumbnail hash (SHA-256): %s", thumbnailHash)

	return &dto.ThumbnailResult{
		ProcessingTime:  thumbnailResult.Duration.Milliseconds(),
		ThumbnailFileID: thumbnail.ID,
		ThumbnailHash:   thumbnailHash,
	}, nil
}

// CreateRepurposerTaskHandler creates a task handler that processes video repurpose tasks
// and updates the generation status after completion
func CreateRepurposerTaskHandler(
	repurposerService RepurposerService,
	serviceContainer *Service,
	generationStatusService GenerationStatusService,
	logger *zap.SugaredLogger,
) func(ctx context.Context, request *dto.ReporpuseVideo) (*dto.ReporpuseContentResult, error) {
	return func(ctx context.Context, request *dto.ReporpuseVideo) (*dto.ReporpuseContentResult, error) {
		// Increment processing counter when task starts
		// Note: This will only succeed if total_queued > 0, which prevents
		// double-counting on retries since retries don't increment total_queued
		if request.AccountID > 0 && request.ContentType != "" {
			contentType := enums.ContentType(request.ContentType)
			if err := generationStatusService.IncrementProcessingByAccount(request.AccountID, contentType); err != nil {
				// This is expected to fail on retries when total_queued = 0
				// We only log as debug, not warning
				logger.Debugw("Skipped increment processing counter (likely a retry)",
					"account_id", request.AccountID,
					"content_type", contentType,
					"error", err,
				)
			}
		}

		// Process the video repurpose task
		result, err := repurposerService.ProcessRepurposeTask(request)
		if err != nil {
			logger.Errorw("Failed to process repurpose task",
				"file_id", request.FileID,
				"account_id", request.AccountID,
				"error", err,
			)

			// Update AccountGenerationStatus on failure
			if request.AccountID > 0 && request.ContentType != "" {
				contentType := enums.ContentType(request.ContentType)
				if updateErr := incrementGenerationStatusCounter(serviceContainer, generationStatusService, request.AccountID, contentType, false); updateErr != nil {
					logger.Warnw("Failed to update generation status after task failure",
						"account_id", request.AccountID,
						"error", updateErr,
					)
				}
			}

			return nil, err
		}

		// Task succeeded - create GeneratedContent record
		if request.ContentAccountID > 0 && request.AccountID > 0 {
			// Determine content type from request or default to video
			contentType := enums.ContentTypeVideo
			if request.ContentType != "" {
				contentType = enums.ContentType(request.ContentType)
			}

			generatedContent := &models.GeneratedContent{
				Type:             contentType,
				AccountID:        request.AccountID,
				ContentID:        request.ContentID,
				ContentAccountID: request.ContentAccountID,
				IsPosted:         false,
				UsedMirror:       request.UseMirror,
				UsedOverlay:      request.UseOverlays,
			}

			// Add generated content file
			generatedContent.Files = []*models.GeneratedContentFile{
				{
					FileID:        result.RenderedFileID,
					ThumbnailID:   result.ThumbnailFileID,
					FileHash:      result.VideoHash,
					ThumbnailHash: result.ThumbnailHash,
				},
			}

			// Save to database
			if err := serviceContainer.store.GeneratedContents.Create(generatedContent); err != nil {
				logger.Errorw("Failed to create generated content record",
					"account_id", request.AccountID,
					"content_account_id", request.ContentAccountID,
					"error", err,
				)
				// Don't fail the task - the video was generated successfully
			} else {
				logger.Infow("Generated content record created",
					"account_id", request.AccountID,
					"content_account_id", request.ContentAccountID,
					"generated_content_id", generatedContent.ID,
				)
			}
		}

		// Update AccountGenerationStatus counters
		if request.AccountID > 0 && request.ContentType != "" {
			contentType := enums.ContentType(request.ContentType)
			if err := incrementGenerationStatusCounter(serviceContainer, generationStatusService, request.AccountID, contentType, true); err != nil {
				logger.Warnw("Failed to update generation status after task success",
					"account_id", request.AccountID,
					"error", err,
				)
			}
		}

		logger.Infow("Task completed successfully",
			"file_id", request.FileID,
			"account_id", request.AccountID,
			"processing_time_ms", result.ProcessingTime,
		)

		return result, nil
	}
}

// incrementGenerationStatusCounter updates the generation status counters
func incrementGenerationStatusCounter(service *Service, statusService GenerationStatusService, accountID uint, contentType enums.ContentType, isSuccess bool) error {
	// Find the active generation status for this account and content type
	status, err := service.store.GenerationStatus.GetActiveByAccount(accountID, contentType)
	if err != nil {
		// No active status found - this might be a legacy task or the status was already completed
		return nil
	}

	if isSuccess {
		// Try to increment completed normally (works if total_processing > 0)
		if err := statusService.IncrementCompleted(status.ID); err == nil {
			return nil
		}

		// If that failed, check if this is a retry after a previous failure
		// In this case, we need to convert a failed task to completed
		status, err = service.store.GenerationStatus.GetByID(status.ID)
		if err != nil {
			return err
		}

		if status.TotalFailed > 0 && status.TotalProcessing == 0 {
			// This is a retry that succeeded after a previous failure
			// Decrement failed and increment completed
			status.TotalCompleted++
			status.TotalFailed--
			status.UpdateProgress()
			if err := service.store.GenerationStatus.Update(status); err != nil {
				return err
			}

			// Check if generation is complete
			if status.IsComplete() {
				return statusService.CompleteGeneration(status.ID)
			}
		}

		return nil
	} else {
		return statusService.IncrementFailed(status.ID)
	}
}

func (s *repurposerService) downloadVideo(fileID uint) (string, error) {
	file, downloadData, err := s.fileService.DownloadFile(fileID)
	if err != nil {
		return "", err
	}
	defer downloadData.Content.Close()

	if err := files.EnsureDirectoryExists(DownloadDir, true); err != nil {
		return "", fmt.Errorf("failed to ensure directory exists: %w", err)
	}

	filename, ext := files.ExtractFileNameFromURL(file.Url)
	downloadPath := fmt.Sprintf("%s/%s_%s.%s", DownloadDir, filename, uuid.New().String(), ext)
	localFile, err := os.Create(downloadPath)
	if err != nil {
		return "", fmt.Errorf("failed to create local file %s: %w", downloadPath, err)
	}
	defer localFile.Close()

	bytesWritten, err := io.Copy(localFile, downloadData.Content)
	if err != nil {
		return "", fmt.Errorf("failed to write file to %s: %w", downloadPath, err)
	}

	if downloadData.Size > 0 && bytesWritten != downloadData.Size {
		return "", fmt.Errorf("incomplete download: expected %d bytes, wrote %d bytes", downloadData.Size, bytesWritten)
	}

	return downloadPath, nil
}

func deleteFile(filePath string) {
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		log.Println("failed to stat file:", err)
		return
	}
	if !info.Mode().IsRegular() {
		// Not a regular file (could be a directory or special file), ignore
		return
	}
	if err := os.Remove(filePath); err != nil {
		log.Println("failed to remove file:", err)
	}
}

func cleanup(downloadPath string, videoPath string, thumbnailPath string) {
	deleteFile(downloadPath)
	deleteFile(videoPath)
	deleteFile(thumbnailPath)
}

func (s *repurposerService) uploadLocalFile(filePath string) (*models.File, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("error getting file info: %w", err)
	}

	filename := uuid.New().String() + filepath.Ext(filePath)
	contentType := mime.TypeByExtension(filepath.Ext(filename))

	storageFile := &storage.File{
		Reader:      file,
		Filename:    filename,
		Size:        fileInfo.Size(),
		ContentType: contentType,
	}

	return s.fileService.UploadFileFromReader(storageFile)
}

// uploadLocalFileWithHash uploads a file and computes its SHA-256 hash
func (s *repurposerService) uploadLocalFileWithHash(filePath string) (*models.File, string, error) {
	// Compute the hash first
	hash, err := utils.ComputeFileHash(filePath)
	if err != nil {
		return nil, "", fmt.Errorf("error computing file hash: %w", err)
	}

	// Upload the file
	file, err := s.uploadLocalFile(filePath)
	if err != nil {
		return nil, "", err
	}

	return file, hash, nil
}
