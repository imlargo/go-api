package services

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/imlargo/go-api-template/internal/dto"
	"github.com/imlargo/go-api-template/internal/models"
	"github.com/imlargo/go-api-template/internal/store"
	"github.com/imlargo/go-api-template/pkg/storage"
	"github.com/imlargo/go-api-template/pkg/utils"
)

type FileService interface {
	UploadFileFromMultipart(file *multipart.FileHeader) (*models.File, error)
	UploadFileFromReader(file *storage.File) (*models.File, error)
	UploadFileFromUrl(url string) (*models.File, error)
	GetFile(id uint) (*models.File, error)
	DeleteFile(id uint) error
	GetPresignedURL(fileID uint, expiryMins int) (*dto.PresignedURL, error)
	BulkDeleteFiles(fileIDs []uint) error
	DownloadFile(fileID uint) (*models.File, *storage.FileDownload, error)
}

type fileServiceImpl struct {
	store              *store.Store
	storageService     storage.FileStorage
	defaultBucket      string
	maxFileSize        int64
	urlDownloadTimeout time.Duration
}

func NewFileService(
	store *store.Store,
	storageService storage.FileStorage,
	defaultBucket string,
) FileService {
	return &fileServiceImpl{
		store:              store,
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

	uploadResult, err := u.storageService.Upload(
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
	if err := u.store.Files.Create(createdFile); err != nil {
		u.storageService.Delete(key)
		return nil, fmt.Errorf("failed to save file record: %w", err)
	}

	return createdFile, nil
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

	uploadResult, err := u.storageService.Upload(
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
	if err := u.store.Files.Create(createdFile); err != nil {
		u.storageService.Delete(key)
		return nil, fmt.Errorf("failed to save file record: %w", err)
	}

	return createdFile, nil
}

func (u *fileServiceImpl) GetFile(id uint) (*models.File, error) {
	file, err := u.store.Files.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	return file, nil
}

func (u *fileServiceImpl) DeleteFile(id uint) error {
	file, err := u.store.Files.GetByID(id)
	if err != nil {
		return fmt.Errorf("file not found: %w", err)
	}

	if err := u.storageService.Delete(file.Path); err != nil {
		return fmt.Errorf("failed to delete from storage: %w", err)
	}

	// Eliminar de BD
	if err := u.store.Files.Delete(id); err != nil {
		return fmt.Errorf("failed to delete file record: %w", err)
	}

	return nil
}

func (u *fileServiceImpl) GetPresignedURL(fileID uint, expiryMins int) (*dto.PresignedURL, error) {
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

	return &dto.PresignedURL{
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

	fileKeys, err := u.store.Files.GetFilesKeys(fileIDs)
	if err != nil {
		return fmt.Errorf("failed to retrieve files: %w", err)
	}

	if err := u.storageService.BulkDelete(fileKeys); err != nil {
		log.Println("Error deleting multiple files from storage:", err.Error())
		return err
	}

	if err := u.store.Files.DeleteFiles(fileIDs); err != nil {
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
