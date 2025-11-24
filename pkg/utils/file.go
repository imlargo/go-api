package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func ExtractFileNameFromURL(urlStr string) (filename, extension string, err error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", "", fmt.Errorf("error parsing URL: %v", err)
	}

	// Get the path from the URL
	fullPath := parsedURL.Path

	// Extract the file name from the path
	fileName := path.Base(fullPath)
	fileName = normalizeString(fileName)

	// If there is no file in the URL, return empty
	if fileName == "/" || fileName == "." {
		return "", "", nil
	}

	// Extract the extension
	ext := path.Ext(fileName)

	// Remove the extension from the name to get only the base name
	name := strings.TrimSuffix(fileName, ext)

	// Remove the dot from the extension if it exists
	if ext != "" {
		ext = strings.TrimPrefix(ext, ".")
	}

	return name, ext, nil
}

// ExtractFileName extracts the file name and extension from a full filename.
func ExtractFileName(fullFilename string) (filename, extension string) {

	normalized := normalizeString(fullFilename)

	// Extract the extension
	ext := filepath.Ext(normalized)

	// Remove the extension from the name to get only the base name
	name := strings.TrimSuffix(normalized, ext)

	// Remove the dot from the extension if it exists
	if ext != "" {
		ext = strings.TrimPrefix(ext, ".")
	}

	return name, ext
}

// DetectContentType returns the content type based on the file extension.
func DetectContentType(ext string) string {
	contentTypes := map[string]string{
		"jpg":  "image/jpeg",
		"jpeg": "image/jpeg",
		"png":  "image/png",
		"gif":  "image/gif",
		"webp": "image/webp",
		"pdf":  "application/pdf",
		"txt":  "text/plain",
		"doc":  "application/msword",
		"docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"mp4":  "video/mp4",
		"mp3":  "audio/mpeg",
		"zip":  "application/zip",
		"json": "application/json",
		"xml":  "application/xml",
	}

	if contentType, exists := contentTypes[strings.ToLower(ext)]; exists {
		return contentType
	}

	return "application/octet-stream"
}

// Extracts the filename from a Content-Disposition header
func ExtractFilenameFromDisposition(disposition string) string {
	parts := strings.Split(disposition, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "filename=") {
			filename := strings.TrimPrefix(part, "filename=")
			filename = strings.Trim(filename, `"`)
			return filename
		}
	}
	return ""
}

// ResolveContentTypeExtension returns the file extension based on the content type
func ResolveContentTypeExtension(contentType string) string {
	if idx := strings.Index(contentType, ";"); idx != -1 {
		contentType = contentType[:idx]
	}

	contentType = strings.TrimSpace(strings.ToLower(contentType))

	switch contentType {
	case "image/jpeg":
		return "jpg"
	case "image/png":
		return "png"
	case "image/gif":
		return "gif"
	case "image/webp":
		return "webp"
	case "text/plain":
		return "txt"
	case "application/pdf":
		return "pdf"
	case "application/json":
		return "json"
	case "text/html":
		return "html"
	case "application/zip":
		return "zip"
	default:
		return ""
	}
}

func normalizeString(filename string) string {
	filename = strings.TrimSpace(filename)
	filename = strings.ToLower(filename)
	filename = strings.ReplaceAll(filename, " ", "-")
	return filename
}

// ComputeFileHash calculates the SHA-256 hash of a file
func ComputeFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to compute hash: %w", err)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// supportedVideoExtensionsList is the single source of truth for supported video extensions
var supportedVideoExtensionsList = []string{"mp4", "mov", "avi", "mkv", "webm", "flv", "wmv", "m4v", "mpg", "mpeg"}

// supportedVideoExtensionsMap is used for efficient lookup
var supportedVideoExtensionsMap = func() map[string]bool {
	m := make(map[string]bool, len(supportedVideoExtensionsList))
	for _, ext := range supportedVideoExtensionsList {
		m[ext] = true
	}
	return m
}()

// supportedImageExtensionsList is the single source of truth for supported image extensions
var supportedImageExtensionsList = []string{"jpg", "jpeg", "png", "gif", "webp", "heic", "heif"}

// supportedImageExtensionsMap is used for efficient lookup
var supportedImageExtensionsMap = func() map[string]bool {
	m := make(map[string]bool, len(supportedImageExtensionsList))
	for _, ext := range supportedImageExtensionsList {
		m[ext] = true
	}
	return m
}()

// IsVideoFile checks if a file extension corresponds to a supported video format
func IsVideoFile(filename string) bool {
	_, ext := ExtractFileName(filename)
	return IsVideoExtension(ext)
}

// IsVideoExtension checks if an extension corresponds to a supported video format
func IsVideoExtension(ext string) bool {
	ext = strings.ToLower(strings.TrimPrefix(ext, "."))
	return supportedVideoExtensionsMap[ext]
}

// GetSupportedVideoExtensions returns a list of supported video file extensions
func GetSupportedVideoExtensions() []string {
	return supportedVideoExtensionsList
}

// IsImageFile checks if a file extension corresponds to a supported image format
func IsImageFile(filename string) bool {
	_, ext := ExtractFileName(filename)
	return IsImageExtension(ext)
}

// IsImageExtension checks if an extension corresponds to a supported image format
func IsImageExtension(ext string) bool {
	ext = strings.ToLower(strings.TrimPrefix(ext, "."))
	return supportedImageExtensionsMap[ext]
}

// GetSupportedImageExtensions returns a list of supported image file extensions
func GetSupportedImageExtensions() []string {
	return supportedImageExtensionsList
}

// IsVideoOrImageFile checks if a file is either a video or image
func IsVideoOrImageFile(filename string) bool {
	return IsVideoFile(filename) || IsImageFile(filename)
}
