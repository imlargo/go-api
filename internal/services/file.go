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
	"github.com/imlargo/go-api/internal/dto"
	"github.com/imlargo/go-api/internal/models"
	"github.com/imlargo/go-api/pkg/storage"
	"github.com/imlargo/go-api/pkg/utils"
)

type FileService interface {
	UploadFromMultipart(file *multipart.FileHeader) (*models.File, error)
	UploadFromReader(file *storage.File) (*models.File, error)
	UploadFromUrl(url string) (*models.File, error)
	GetFile(id uint) (*models.File, error)
	DeleteFile(id uint) error
	GetPresignedURL(fileID uint, expiryMins int) (*dto.PresignedURL, error)
	BulkDeleteFiles(fileIDs []uint) error
	DownloadFile(fileID uint) (*models.File, *storage.FileDownload, error)
}

type fileService struct {
	*Service
	storageService     storage.FileStorage
	maxFileSize        int64
	urlDownloadTimeout time.Duration
}

func NewFileService(
	service *Service,
	storageService storage.FileStorage,
) FileService {
	return &fileService{
		Service:            service,
		storageService:     storageService,
		maxFileSize:        1 * 1024 * 1024 * 1024, // 1GB
		urlDownloadTimeout: 60 * time.Second * 15,  // Timeout for URL downloads
	}
}

func (s *fileService) UploadFromUrl(url string) (*models.File, error) {
	file, err := s.createFileFromUrl(url)
	if err != nil {
		return nil, fmt.Errorf("failed to create file from URL: %w", err)
	}

	return s.UploadFromReader(file)
}

func (s *fileService) UploadFromReader(file *storage.File) (*models.File, error) {

	if file.Size > s.maxFileSize {
		return nil, fmt.Errorf("file size exceeds limit of 1GB")
	}

	fileID := uuid.New()

	_, ext := utils.ExtractFileName(file.Filename)

	key := s.createFileKey(fileID, ext)

	contentType := file.ContentType
	if contentType == "" {
		contentType = utils.DetectContentType(ext)
	}

	uploadResult, err := s.storageService.Upload(
		key,
		file.Reader,
		contentType,
		file.Size,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to R2: %w", err)
	}

	createdFile := &models.File{
		ContentType: uploadResult.ContentType,
		Size:        uploadResult.Size,
		Etag:        uploadResult.Etag,
		Path:        uploadResult.Key,
		Url:         uploadResult.Url,
	}
	if err := s.store.Files.Create(createdFile); err != nil {
		s.storageService.Delete(key)
		return nil, fmt.Errorf("failed to save file record: %w", err)
	}

	return createdFile, nil
}

func (s *fileService) UploadFromMultipart(file *multipart.FileHeader) (*models.File, error) {
	if file.Size > s.maxFileSize {
		return nil, fmt.Errorf("file size exceeds limit of 1GB")
	}

	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	defer src.Close()

	fileID := uuid.New()

	_, ext := utils.ExtractFileName(file.Filename)

	key := s.createFileKey(fileID, ext)

	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = utils.DetectContentType(ext)
	}

	uploadResult, err := s.storageService.Upload(
		key,
		src,
		contentType,
		file.Size,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to upload file to R2: %w", err)
	}

	createdFile := &models.File{
		ContentType: uploadResult.ContentType,
		Size:        uploadResult.Size,
		Etag:        uploadResult.Etag,
		Path:        uploadResult.Key,
		Url:         uploadResult.Url,
	}
	if err := s.store.Files.Create(createdFile); err != nil {
		s.storageService.Delete(key)
		return nil, fmt.Errorf("failed to save file record: %w", err)
	}

	return createdFile, nil
}

func (s *fileService) GetFile(id uint) (*models.File, error) {
	file, err := s.store.Files.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	return file, nil
}

func (s *fileService) DeleteFile(id uint) error {
	file, err := s.store.Files.GetByID(id)
	if err != nil {
		return fmt.Errorf("file not found: %w", err)
	}

	if err := s.storageService.Delete(file.Path); err != nil {
		return fmt.Errorf("failed to delete from storage: %w", err)
	}

	// Eliminar de BD
	if err := s.store.Files.Delete(id); err != nil {
		return fmt.Errorf("failed to delete file record: %w", err)
	}

	return nil
}

func (s *fileService) GetPresignedURL(fileID uint, expiryMins int) (*dto.PresignedURL, error) {
	file, err := s.store.Files.GetByID(fileID)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	expiry := time.Duration(expiryMins) * time.Minute
	if expiry == 0 {
		expiry = 15 * time.Minute // default 15 minutes
	}

	url, err := s.storageService.GetPresignedURL(file.Path, expiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return &dto.PresignedURL{
		Url:       url,
		ExpiresAt: time.Now().Add(expiry).Format(time.RFC3339),
	}, nil
}

func (s *fileService) createFileFromUrl(urlStr string) (*storage.File, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, fmt.Errorf("only HTTP and HTTPS URLs are supported")
	}

	client := &http.Client{
		Timeout: s.urlDownloadTimeout,
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
	if resp.ContentLength > 0 && resp.ContentLength > s.maxFileSize {
		return nil, fmt.Errorf("file size %d exceeds limit of %d bytes", resp.ContentLength, s.maxFileSize)
	}

	// Use LimitReader to prevent memory exhaustion even if Content-Length is not set
	limitedReader := io.LimitReader(resp.Body, s.maxFileSize+1) // +1 to detect if limit exceeded

	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check if we hit the limit (meaning file is too large)
	if int64(len(data)) > s.maxFileSize {
		return nil, fmt.Errorf("file size exceeds limit of %d bytes", s.maxFileSize)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	filename := s.extractUrlFilename(urlStr, resp.Header.Get("Content-Disposition"), contentType)

	return &storage.File{
		Reader:      bytes.NewReader(data),
		Filename:    filename,
		Size:        int64(len(data)),
		ContentType: contentType,
	}, nil
}

func (s *fileService) BulkDeleteFiles(fileIDs []uint) error {
	if len(fileIDs) == 0 {
		return nil // No files to delete
	}

	fileKeys, err := s.store.Files.GetFilesKeys(fileIDs)
	if err != nil {
		return fmt.Errorf("failed to retrieve files: %w", err)
	}

	if err := s.storageService.BulkDelete(fileKeys); err != nil {
		s.logger.Errorln("Error deleting multiple files from storage:", err.Error())
		return err
	}

	if err := s.store.Files.DeleteFiles(fileIDs); err != nil {
		return fmt.Errorf("failed to delete file records: %w", err)
	}

	return nil
}

func (s *fileService) DownloadFile(fileID uint) (*models.File, *storage.FileDownload, error) {

	file, err := s.store.Files.GetByID(fileID)
	if err != nil {
		return nil, nil, fmt.Errorf("file not found: %w", err)
	}

	if file.Path == "" {
		return nil, nil, fmt.Errorf("file path is empty, cannot download")
	}

	downloadData, err := s.storageService.GetFileForDownload(file.Path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to download file: %w", err)
	}

	return file, downloadData, nil
}

func (s *fileService) createFileKey(fileID uuid.UUID, ext string) string {
	now := time.Now()
	datePrefix := now.Format("2006/01/02")
	return fmt.Sprintf("%s/%s.%s", datePrefix, fileID.String(), ext)
}

func (s *fileService) extractUrlFilename(urlStr, disposition, contentType string) string {
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
