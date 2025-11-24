package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestComputeFileHash(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	testContent := []byte("Hello, World!")
	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Compute hash
	hash, err := ComputeFileHash(testFile)
	if err != nil {
		t.Fatalf("ComputeFileHash failed: %v", err)
	}

	// Expected SHA-256 hash of "Hello, World!"
	expectedHash := "dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f"

	if hash != expectedHash {
		t.Errorf("Hash mismatch. Got %s, expected %s", hash, expectedHash)
	}

	// Verify hash is 64 characters (SHA-256 hex)
	if len(hash) != 64 {
		t.Errorf("Hash length is incorrect. Got %d, expected 64", len(hash))
	}
}

func TestComputeFileHash_NonExistentFile(t *testing.T) {
	_, err := ComputeFileHash("/nonexistent/file.txt")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestComputeFileHash_EmptyFile(t *testing.T) {
	// Create a temporary empty file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "empty.txt")

	if err := os.WriteFile(testFile, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	hash, err := ComputeFileHash(testFile)
	if err != nil {
		t.Fatalf("ComputeFileHash failed: %v", err)
	}

	// Expected SHA-256 hash of empty content
	expectedHash := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

	if hash != expectedHash {
		t.Errorf("Hash mismatch for empty file. Got %s, expected %s", hash, expectedHash)
	}
}

func TestComputeFileHash_LargeFile(t *testing.T) {
	// Create a temporary large file (1MB)
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "large.bin")

	// Create 1MB of data
	data := make([]byte, 1024*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	if err := os.WriteFile(testFile, data, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	hash, err := ComputeFileHash(testFile)
	if err != nil {
		t.Fatalf("ComputeFileHash failed: %v", err)
	}

	// Verify hash is valid format
	if len(hash) != 64 {
		t.Errorf("Hash length is incorrect. Got %d, expected 64", len(hash))
	}

	// Compute hash again to verify consistency
	hash2, err := ComputeFileHash(testFile)
	if err != nil {
		t.Fatalf("ComputeFileHash failed on second attempt: %v", err)
	}

	if hash != hash2 {
		t.Errorf("Hash is not consistent. Got %s and %s", hash, hash2)
	}
}
