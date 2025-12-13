package files

import (
	"fmt"
	"os"
	"path/filepath"
)

func CheckFile(filePath string) error {
	if filePath == "" {
		return ErrFileEmpty
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return ErrFileNotExist
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return ErrFileAccess(filePath, err)
	}

	if fileInfo.IsDir() {
		return ErrFileIsDirectory(filePath)
	}

	return nil
}

// EnsureDirectoryExists creates a directory if it doesn't exist, with options for handling existing files/directories
func EnsureDirectoryExists(dirPath string, overwrite bool) error {
	if dirPath == "" {
		return fmt.Errorf("directory path cannot be empty")
	}

	dirPath = filepath.Clean(dirPath)

	fileInfo, err := os.Stat(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Path doesn't exist, create directory
			if err := os.MkdirAll(dirPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory '%s': %w", dirPath, err)
			}

			return nil
		}
		// Other error occurred (permission, etc.)
		return fmt.Errorf("failed to access path '%s': %w", dirPath, err)
	}

	// Path exists, check if it's a directory
	if !fileInfo.IsDir() {
		if overwrite {
			// Remove the existing file and create directory
			if err := os.Remove(dirPath); err != nil {
				return fmt.Errorf("failed to remove existing file '%s': %w", dirPath, err)
			}

			if err := os.MkdirAll(dirPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory after removing file '%s': %w", dirPath, err)
			}

			return nil
		}
		return fmt.Errorf("path '%s' exists but is a file, not a directory", dirPath)
	}

	return nil
}
