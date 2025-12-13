package evaluator

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"image"
	"image/jpeg"
	"io/fs"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// ComputePerceptualHash calculates the perceptual hash of an image
func ComputePerceptualHash(imgPath string) (uint64, error) {
	file, err := os.Open(imgPath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	img, err := jpeg.Decode(file)
	if err != nil {
		return 0, err
	}

	// Convert to grayscale and resize to 8x8
	grayImg := convertToGrayscale(img)
	resized := resizeImage(grayImg, 8, 8)

	// Calculate perceptual hash (pHash)
	hash := computeDCTHash(resized)
	return hash, nil
}

// convertToGrayscale converts image to grayscale
func convertToGrayscale(img image.Image) [][]float64 {
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	gray := make([][]float64, height)
	for y := 0; y < height; y++ {
		gray[y] = make([]float64, width)
		for x := 0; x < width; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			// Convert to grayscale using luminance
			gray[y][x] = 0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(b>>8)
		}
	}
	return gray
}

// resizeImage resizes the image using simple bilinear interpolation
func resizeImage(img [][]float64, newWidth, newHeight int) [][]float64 {
	oldHeight := len(img)
	oldWidth := len(img[0])

	resized := make([][]float64, newHeight)
	for y := 0; y < newHeight; y++ {
		resized[y] = make([]float64, newWidth)
		for x := 0; x < newWidth; x++ {
			// Simple interpolation - take the nearest pixel
			oldX := (x * oldWidth) / newWidth
			oldY := (y * oldHeight) / newHeight

			if oldX >= oldWidth {
				oldX = oldWidth - 1
			}
			if oldY >= oldHeight {
				oldY = oldHeight - 1
			}

			resized[y][x] = img[oldY][oldX]
		}
	}
	return resized
}

// computeDCTHash calculates hash using DCT (Discrete Cosine Transform)
func computeDCTHash(img [][]float64) uint64 {
	// Simplified DCT for 8x8
	dct := computeDCT(img)

	// Calculate average of DC components (excluding [0][0])
	var sum float64
	count := 0
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			if i == 0 && j == 0 {
				continue // Skip DC component
			}
			sum += dct[i][j]
			count++
		}
	}
	avg := sum / float64(count)

	// Generate binary hash by comparing with average
	var hash uint64
	bit := 0
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			if i == 0 && j == 0 {
				continue
			}
			if bit >= 64 {
				break
			}
			if dct[i][j] > avg {
				hash |= (1 << bit)
			}
			bit++
		}
		if bit >= 64 {
			break
		}
	}

	return hash
}

// computeDCT calculates the discrete cosine transform
func computeDCT(img [][]float64) [][]float64 {
	size := len(img)
	dct := make([][]float64, size)
	for i := range dct {
		dct[i] = make([]float64, size)
	}

	for u := 0; u < size; u++ {
		for v := 0; v < size; v++ {
			var sum float64
			for x := 0; x < size; x++ {
				for y := 0; y < size; y++ {
					sum += img[x][y] *
						math.Cos((2*float64(x)+1)*float64(u)*math.Pi/(2*float64(size))) *
						math.Cos((2*float64(y)+1)*float64(v)*math.Pi/(2*float64(size)))
				}
			}

			cu := 1.0
			if u == 0 {
				cu = 1.0 / math.Sqrt(2)
			}
			cv := 1.0
			if v == 0 {
				cv = 1.0 / math.Sqrt(2)
			}

			dct[u][v] = 0.25 * cu * cv * sum
		}
	}

	return dct
}

// generateVideoID generates a unique ID for the video
func generateVideoID(videoPath string) string {
	hash := md5.Sum([]byte(videoPath + time.Now().String()))
	return hex.EncodeToString(hash[:])
}

// createCompositeHash creates a composite hash of the video
func createCompositeHash(frameHashes []uint64) []uint64 {
	if len(frameHashes) == 0 {
		return []uint64{}
	}

	// Divide the video into segments and create hash per segment
	segmentSize := 5 // 5 frames per segment
	segments := make([]uint64, 0)

	for i := 0; i < len(frameHashes); i += segmentSize {
		end := i + segmentSize
		if end > len(frameHashes) {
			end = len(frameHashes)
		}

		// Segment hash (XOR of all frames)
		var segmentHash uint64
		for j := i; j < end; j++ {
			segmentHash ^= frameHashes[j]
		}
		segments = append(segments, segmentHash)
	}

	return segments
}

// CompareVideos compara dos videos y calcula similitud
func (vp *VideoProcessor) CompareVideos(hash1, hash2 *VideoHash) *SimilarityResult {
	if len(hash1.Hash) == 0 || len(hash2.Hash) == 0 {
		return &SimilarityResult{
			OriginalVideo: hash1.FilePath,
			ComparedVideo: hash2.FilePath,
			Similarity:    0.0,
			IsOriginal:    true,
		}
	}

	// Compare segments
	matchedSegments := 0
	totalSegments := len(hash1.Hash)

	if len(hash2.Hash) < totalSegments {
		totalSegments = len(hash2.Hash)
	}

	totalDistance := 0

	for i := 0; i < totalSegments; i++ {
		distance := hammingDistance(hash1.Hash[i], hash2.Hash[i])
		totalDistance += distance

		// Consider match if Hamming distance < 10
		if distance < 10 {
			matchedSegments++
		}
	}

	// Calculate similarity
	similarity := float64(matchedSegments) / float64(totalSegments)

	// Also compare individual frame hashes for greater precision
	frameSimilarity := compareFrameHashes(hash1.FrameHashes, hash2.FrameHashes)

	// Weighted average
	finalSimilarity := 0.7*similarity + 0.3*frameSimilarity

	return &SimilarityResult{
		OriginalVideo:   hash1.FilePath,
		ComparedVideo:   hash2.FilePath,
		Similarity:      finalSimilarity,
		IsOriginal:      finalSimilarity < vp.SimilarityThreshold,
		MatchedSegments: matchedSegments,
		TotalSegments:   totalSegments,
		HashDistance:    totalDistance,
	}
}

// hammingDistance calculates the Hamming distance between two hashes
func hammingDistance(hash1, hash2 uint64) int {
	xor := hash1 ^ hash2
	distance := 0

	for xor != 0 {
		distance++
		xor &= xor - 1 // Remove the least significant bit
	}

	return distance
}

// compareFrameHashes compara arrays de hashes de frames
func compareFrameHashes(hashes1, hashes2 []uint64) float64 {
	if len(hashes1) == 0 || len(hashes2) == 0 {
		return 0.0
	}

	minLen := len(hashes1)
	if len(hashes2) < minLen {
		minLen = len(hashes2)
	}

	matches := 0
	for i := 0; i < minLen; i++ {
		if hammingDistance(hashes1[i], hashes2[i]) < 8 {
			matches++
		}
	}

	return float64(matches) / float64(minLen)
}

// ProcessDirectory processes all videos in a directory
func (vp *VideoProcessor) ProcessDirectory(dirPath string, originalVideoPath string) ([]*SimilarityResult, error) {
	// Process original video
	fmt.Printf("Processing original video: %s\n", originalVideoPath)
	originalHash, err := vp.ProcessVideo(originalVideoPath)
	if err != nil {
		return nil, fmt.Errorf("error processing original video: %w", err)
	}

	// Store original hash
	vp.mu.Lock()
	vp.videoHashes[originalVideoPath] = originalHash
	vp.mu.Unlock()

	var results []*SimilarityResult
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Process videos in the directory
	err = filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		// Check if it's a video file
		ext := strings.ToLower(filepath.Ext(path))
		if !isVideoFile(ext) {
			return nil
		}

		// Skip the original video if it's in the directory
		if path == originalVideoPath {
			return nil
		}

		wg.Add(1)
		go func(videoPath string) {
			defer wg.Done()

			fmt.Printf("Processing: %s\n", videoPath)

			// Process video
			videoHash, err := vp.ProcessVideo(videoPath)
			if err != nil {
				log.Printf("Error processing %s: %v", videoPath, err)
				return
			}

			// Store hash
			vp.mu.Lock()
			vp.videoHashes[videoPath] = videoHash
			vp.mu.Unlock()

			// Compare with original
			result := vp.CompareVideos(originalHash, videoHash)

			mu.Lock()
			results = append(results, result)
			mu.Unlock()

			// Show result
			status := "ORIGINAL"
			if !result.IsOriginal {
				status = "DUPLICATE"
			}

			fmt.Printf("✓ %s - Similarity: %.2f%% - %s\n",
				filepath.Base(videoPath),
				result.Similarity*100,
				status)

		}(path)

		return nil
	})

	wg.Wait()

	if err != nil {
		return nil, fmt.Errorf("error processing directory: %w", err)
	}

	// Sort results by similarity (highest to lowest)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})

	return results, nil
}

// isVideoFile checks if the extension corresponds to a video file
func isVideoFile(ext string) bool {
	videoExts := []string{".mp4", ".avi", ".mov", ".mkv", ".flv", ".wmv", ".webm", ".m4v"}
	for _, videoExt := range videoExts {
		if ext == videoExt {
			return true
		}
	}
	return false
}

// GenerateReport generates a report of the results
func GenerateReport(results []*SimilarityResult, originalVideo string) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Printf("DUPLICATE CONTENT ANALYSIS REPORT\n")
	fmt.Printf("Original Video: %s\n", originalVideo)
	fmt.Printf("Total videos analyzed: %d\n", len(results))
	fmt.Println(strings.Repeat("=", 80))

	duplicates := 0
	for _, result := range results {
		if !result.IsOriginal {
			duplicates++
		}
	}

	fmt.Printf("Original videos: %d\n", len(results)-duplicates)
	fmt.Printf("Duplicate videos: %d\n", duplicates)
	fmt.Printf("Percentage of duplicates: %.1f%%\n\n", float64(duplicates)/float64(len(results))*100)

	fmt.Println("DETAILED RESULTS:")
	fmt.Println(strings.Repeat("-", 80))

	for i, result := range results {
		status := "✅ ORIGINAL"
		if !result.IsOriginal {
			status = "🚫 DUPLICATE"
		}

		fmt.Printf("%d. %s\n", i+1, filepath.Base(result.ComparedVideo))
		fmt.Printf("   Similarity: %.2f%%\n", result.Similarity*100)
		fmt.Printf("   Status: %s\n", status)
		fmt.Printf("   Matching segments: %d/%d\n", result.MatchedSegments, result.TotalSegments)
		fmt.Printf("   Hash distance: %d\n\n", result.HashDistance)
	}
}
