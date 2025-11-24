package services

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/nicolailuther/butter/internal/dto"
	"github.com/nicolailuther/butter/internal/models"
	"github.com/nicolailuther/butter/pkg/storage"
	"github.com/nicolailuther/butter/pkg/utils"
)

type FileService interface {
	UploadFileFromMultipart(file *multipart.FileHeader) (*models.File, error)
	UploadFileFromReader(file *storage.File) (*models.File, error)
	UploadFileFromUrl(url string) (*models.File, error)
	GetFile(id uint) (*models.File, error)
	DeleteFile(id uint) error
	GetPresignedURL(fileID uint, expiryMins int) (*dto.PresignedURLResponse, error)
	BulkDeleteFiles(fileIDs []uint) error
	DownloadFile(fileID uint) (*models.File, *storage.FileDownload, error)
}

type fileServiceImpl struct {
	*Service
	storageService     storage.FileStorage
	defaultBucket      string
	maxFileSize        int64
	urlDownloadTimeout time.Duration
}

func NewFileService(
	container *Service,
	storageService storage.FileStorage,
	defaultBucket string,
) FileService {
	return &fileServiceImpl{
		Service:            container,
		storageService:     storageService,
		defaultBucket:      defaultBucket,
		maxFileSize:        1 * 1024 * 1024 * 1024, // 1GB
		urlDownloadTimeout: 60 * time.Second * 15,  // Timeout for URL downloads
	}
}

func (u *fileServiceImpl) UploadFileFromUrl(url string) (*models.File, error) {
	file, err := u.createFileFromUrl(url)
	if err != nil {
		return nil, fmt.Errorf("failed to create file from URL: %w", err)
	}

	return u.UploadFileFromReader(file)
}

func (u *fileServiceImpl) UploadFileFromReader(file *storage.File) (*models.File, error) {

	if file.Size > u.maxFileSize {
		return nil, fmt.Errorf("file size exceeds limit of 1GB")
	}

	fileID := uuid.New()

	_, ext := utils.ExtractFileName(file.Filename)

	key := u.createFileKey(fileID, ext)

	contentType := file.ContentType
	if contentType == "" {
		contentType = utils.DetectContentType(ext)
	}

	upload, err := u.storageService.Upload(
		key,
		file.Reader,
		contentType,
		file.Size,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to R2: %w", err)
	}

	created := &models.File{
		ContentType: upload.ContentType,
		Size:        upload.Size,
		Etag:        upload.Etag,
		Path:        upload.Key,
		Url:         upload.Url,
	}
	if err := u.store.Files.Create(created); err != nil {
		u.storageService.Delete(key)
		return nil, fmt.Errorf("failed to save file record: %w", err)
	}

	return created, nil
}

func (u *fileServiceImpl) UploadFileFromMultipart(file *multipart.FileHeader) (*models.File, error) {
	if file.Size > u.maxFileSize {
		return nil, fmt.Errorf("file size exceeds limit of 1GB")
	}

	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	defer src.Close()

	fileID := uuid.New()

	_, ext := utils.ExtractFileName(file.Filename)

	key := u.createFileKey(fileID, ext)

	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = utils.DetectContentType(ext)
	}

	upload, err := u.storageService.Upload(
		key,
		src,
		contentType,
		file.Size,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to upload file to R2: %w", err)
	}

	created := &models.File{
		ContentType: upload.ContentType,
		Size:        upload.Size,
		Etag:        upload.Etag,
		Path:        upload.Key,
		Url:         upload.Url,
	}
	if err := u.store.Files.Create(created); err != nil {
		u.storageService.Delete(key)
		return nil, fmt.Errorf("failed to save file record: %w", err)
	}

	return created, nil
}

func (u *fileServiceImpl) GetFile(id uint) (*models.File, error) {
	file, err := u.store.Files.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	return file, nil
}

func (u *fileServiceImpl) DeleteFile(id uint) error {
	if id == 0 {
		return fmt.Errorf("file ID cannot be zero")
	}

	file, err := u.store.Files.GetByID(id)
	if err != nil {
		return fmt.Errorf("file not found: %w", err)
	}

	// Delete from storage first
	if file.Path != "" {
		if err := u.storageService.Delete(file.Path); err != nil {
			u.logger.Errorf("Failed to delete file from storage (path: %s): %v", file.Path, err)
			// Continue with database deletion - storage cleanup can be handled separately
		}
	}

	// Delete from database
	if err := u.store.Files.Delete(id); err != nil {
		return fmt.Errorf("failed to delete file record: %w", err)
	}

	return nil
}

func (u *fileServiceImpl) GetPresignedURL(fileID uint, expiryMins int) (*dto.PresignedURLResponse, error) {
	file, err := u.store.Files.GetByID(fileID)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	expiry := time.Duration(expiryMins) * time.Minute
	if expiry == 0 {
		expiry = 15 * time.Minute // default 15 minutes
	}

	url, err := u.storageService.GetPresignedURL(file.Path, expiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return &dto.PresignedURLResponse{
		Url:       url,
		ExpiresAt: time.Now().Add(expiry).Format(time.RFC3339),
	}, nil
}

func (u *fileServiceImpl) createFileFromUrl(urlStr string) (*storage.File, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, fmt.Errorf("only HTTP and HTTPS URLs are supported")
	}

	client := &http.Client{
		Timeout: u.urlDownloadTimeout,
	}

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("User-Agent", "Go-HTTP-Client/1.1")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}

	// Check Content-Length before reading to prevent downloading huge files
	if resp.ContentLength > 0 && resp.ContentLength > u.maxFileSize {
		return nil, fmt.Errorf("file size %d exceeds limit of %d bytes", resp.ContentLength, u.maxFileSize)
	}

	// Use LimitReader to prevent memory exhaustion even if Content-Length is not set
	limitedReader := io.LimitReader(resp.Body, u.maxFileSize+1) // +1 to detect if limit exceeded

	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check if we hit the limit (meaning file is too large)
	if int64(len(data)) > u.maxFileSize {
		return nil, fmt.Errorf("file size exceeds limit of %d bytes", u.maxFileSize)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	filename := u.extractUrlFilename(urlStr, resp.Header.Get("Content-Disposition"), contentType)

	return &storage.File{
		Reader:      bytes.NewReader(data),
		Filename:    filename,
		Size:        int64(len(data)),
		ContentType: contentType,
	}, nil
}

func (u *fileServiceImpl) BulkDeleteFiles(fileIDs []uint) error {
	if len(fileIDs) == 0 {
		return nil // No files to delete
	}

	// Filter out any zero IDs
	validFileIDs := make([]uint, 0, len(fileIDs))
	for _, id := range fileIDs {
		if id > 0 {
			validFileIDs = append(validFileIDs, id)
		}
	}

	if len(validFileIDs) == 0 {
		return nil // No valid files to delete
	}

	fileKeys, err := u.store.Files.GetFilesKeys(validFileIDs)
	if err != nil {
		return fmt.Errorf("failed to retrieve file keys: %w", err)
	}

	// Only proceed with storage deletion if we have file keys
	if len(fileKeys) > 0 {
		if err := u.storageService.BulkDelete(fileKeys); err != nil {
			u.logger.Errorf("Error deleting files from storage: %v", err)
			// Continue with database deletion even if storage deletion fails
		}
	}

	// Delete file records from database
	if err := u.store.Files.DeleteFiles(validFileIDs); err != nil {
		return fmt.Errorf("failed to delete file records: %w", err)
	}

	return nil
}

func (u *fileServiceImpl) DownloadFile(fileID uint) (*models.File, *storage.FileDownload, error) {

	file, err := u.store.Files.GetByID(fileID)
	if err != nil {
		return nil, nil, fmt.Errorf("file not found: %w", err)
	}

	if file.Path == "" {
		return nil, nil, fmt.Errorf("file path is empty, cannot download")
	}

	downloadData, err := u.storageService.GetFileForDownload(file.Path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to download file: %w", err)
	}

	return file, downloadData, nil
}

func (u *fileServiceImpl) createFileKey(fileID uuid.UUID, ext string) string {
	now := time.Now()
	datePrefix := now.Format("2006/01/02")
	return fmt.Sprintf("%s/%s.%s", datePrefix, fileID.String(), ext)
}

func (u *fileServiceImpl) extractUrlFilename(urlStr, disposition, contentType string) string {
	// Try URL first
	filename, ext, err := utils.ExtractFileNameFromURL(urlStr)
	if err == nil && filename != "" {
		if ext != "" {
			return filename + "." + ext
		}
		return filename
	}

	// Try Content-Disposition header
	if disposition != "" {
		if dispositionFilename := utils.ExtractFilenameFromDisposition(disposition); dispositionFilename != "" {
			return dispositionFilename
		}
	}

	// Generate filename from content type
	ext = utils.ResolveContentTypeExtension(contentType)
	if ext == "" {
		ext = "bin" // fallback extension
	}

	return fmt.Sprintf("file_%d.%s", time.Now().Unix(), ext)
}
