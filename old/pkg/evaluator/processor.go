package evaluator

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// VideoProcessor handles video processing
type VideoProcessor struct {
	TempDir             string
	FFmpegPath          string
	SimilarityThreshold float64
	mu                  sync.RWMutex
	videoHashes         map[string]*VideoHash
}

// NewVideoProcessor creates a new video processor
func NewVideoProcessor(tempDir, ffmpegPath string) *VideoProcessor {
	return &VideoProcessor{
		TempDir:             tempDir,
		FFmpegPath:          ffmpegPath,
		SimilarityThreshold: 0.85, // 85% similarity to consider duplicate
		videoHashes:         make(map[string]*VideoHash),
	}
}

// ExtractFrames extracts frames from video using FFmpeg
func (vp *VideoProcessor) ExtractFrames(videoPath string, outputDir string, fps int) error {
	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("error creating directory: %w", err)
	}

	// FFmpeg command to extract frames
	cmd := exec.Command(vp.FFmpegPath,
		"-i", videoPath,
		"-vf", fmt.Sprintf("fps=%d,scale=64:64", fps), // Scale to 64x64 for consistency
		"-y", // Overwrite existing files
		filepath.Join(outputDir, "frame_%04d.jpg"))

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error extracting frames: %w", err)
	}

	return nil
}

// GetVideoDuration gets the video duration
func (vp *VideoProcessor) GetVideoDuration(videoPath string) (float64, error) {
	cmd := exec.Command(vp.FFmpegPath,
		"-i", videoPath,
		"-f", "null", "-")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("ffmpeg command should fail for duration extraction")
	}

	// Parse duration from FFmpeg output
	outputStr := string(output)
	lines := strings.Split(outputStr, "\n")

	for _, line := range lines {
		if strings.Contains(line, "Duration:") {
			parts := strings.Split(line, "Duration: ")
			if len(parts) > 1 {
				durationStr := strings.Split(parts[1], ",")[0]
				duration := parseDuration(durationStr)
				return duration, nil
			}
		}
	}

	return 0, fmt.Errorf("could not get video duration")
}

// parseDuration converts duration string HH:MM:SS.ms to seconds
func parseDuration(durationStr string) float64 {
	parts := strings.Split(durationStr, ":")
	if len(parts) != 3 {
		return 0
	}

	hours, _ := strconv.ParseFloat(parts[0], 64)
	minutes, _ := strconv.ParseFloat(parts[1], 64)
	seconds, _ := strconv.ParseFloat(parts[2], 64)

	return hours*3600 + minutes*60 + seconds
}

// ProcessVideo processes a complete video and generates its hash
func (vp *VideoProcessor) ProcessVideo(videoPath string) (*VideoHash, error) {
	// Create temporary directory for frames
	videoID := generateVideoID(videoPath)
	frameDir := filepath.Join(vp.TempDir, videoID)

	// Cleanup at the end
	defer os.RemoveAll(frameDir)

	// Get video duration
	duration, err := vp.GetVideoDuration(videoPath)
	if err != nil {
		return nil, fmt.Errorf("error getting duration: %w", err)
	}

	// Extract frames (1 frame per second)
	fps := 1
	if err := vp.ExtractFrames(videoPath, frameDir, fps); err != nil {
		return nil, fmt.Errorf("error extracting frames: %w", err)
	}

	// Get list of frames
	frameFiles, err := filepath.Glob(filepath.Join(frameDir, "*.jpg"))
	if err != nil {
		return nil, fmt.Errorf("error listing frames: %w", err)
	}

	if len(frameFiles) == 0 {
		return nil, fmt.Errorf("no frames were extracted from the video")
	}

	// Calculate hash for each frame
	frameHashes := make([]uint64, 0, len(frameFiles))
	for _, frameFile := range frameFiles {
		hash, err := ComputePerceptualHash(frameFile)
		if err != nil {
			log.Printf("Error calculating hash for frame %s: %v", frameFile, err)
			continue
		}
		frameHashes = append(frameHashes, hash)
	}

	if len(frameHashes) == 0 {
		return nil, fmt.Errorf("could not calculate hashes for any frame")
	}

	// Create composite video hash
	videoHash := createCompositeHash(frameHashes)

	return &VideoHash{
		FilePath:    videoPath,
		Hash:        videoHash,
		Duration:    duration,
		FrameHashes: frameHashes,
		Timestamp:   time.Now(),
	}, nil
}
