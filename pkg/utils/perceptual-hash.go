package utils

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/corona10/goimagehash"
)

type VideoFingerprint struct {
	FrameHashes   []uint64  `json:"frame_hashes"`
	Duration      float64   `json:"duration"`
	FrameCount    int       `json:"frame_count"`
	Timestamp     time.Time `json:"timestamp"`
	SampleRate    float64   `json:"sample_rate"`
	StrategicHash uint64    `json:"strategic_hash"`
}

// GenerateVideoFingerprint genera un fingerprint de video desde una URL
func GenerateVideoFingerprint(videoURL string) (*VideoFingerprint, error) {
	tmpDir, err := os.MkdirTemp("", "vfp_*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	videoPath := tmpDir + "/v.mp4"

	if err := downloadFile(videoURL, videoPath); err != nil {
		return nil, fmt.Errorf("download: %w", err)
	}

	duration, err := getVideoDuration(videoPath)
	if err != nil {
		return nil, fmt.Errorf("get duration: %w", err)
	}

	if duration < MinVideoDuration {
		return nil, fmt.Errorf("video too short: %.1f seconds", duration)
	}

	timestamps := calculateSafeTimestamps(duration)
	hashes := make([]uint64, 0, MaxFramesToExtract)

	for _, ts := range timestamps {
		hash, err := extractAndHashSingleFrame(videoPath, ts)
		if err != nil {
			continue
		}
		hashes = append(hashes, hash)
	}

	if len(hashes) == 0 {
		return nil, fmt.Errorf("no valid hashes computed")
	}

	strategicIdx := len(hashes) / 2

	return &VideoFingerprint{
		FrameHashes:   hashes,
		Duration:      duration,
		FrameCount:    len(hashes),
		Timestamp:     time.Now(),
		SampleRate:    float64(len(hashes)) / duration,
		StrategicHash: hashes[strategicIdx],
	}, nil
}

// CompareAgainstFingerprint compara un video contra un fingerprint existente
func CompareAgainstFingerprint(videoURL string, originalFingerprint *VideoFingerprint) (bool, float64, error) {
	tmpDir, err := os.MkdirTemp("", "vcmp_*")
	if err != nil {
		return false, 0, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	videoPath := tmpDir + "/v.mp4"

	if err := downloadFile(videoURL, videoPath); err != nil {
		return false, 0, fmt.Errorf("download: %w", err)
	}

	duration, err := getVideoDuration(videoPath)
	if err != nil {
		return false, 0, fmt.Errorf("get duration: %w", err)
	}

	// Early rejection por duración (tolera cortado de 1-2 segundos al final)
	durationDiff := absDiff(duration - originalFingerprint.Duration)
	if durationDiff/originalFingerprint.Duration > 0.15 {
		return false, 0, nil
	}

	if duration < MinVideoDuration {
		return false, 0, fmt.Errorf("video too short: %.1f seconds", duration)
	}

	timestamps := calculateSafeTimestamps(duration)
	hashes := make([]uint64, 0, MaxFramesToExtract)

	for _, ts := range timestamps {
		hash, err := extractAndHashSingleFrame(videoPath, ts)
		if err != nil {
			continue
		}
		hashes = append(hashes, hash)
	}

	if len(hashes) == 0 {
		return false, 0, fmt.Errorf("no valid hashes computed")
	}

	// Quick check: comparar frame del medio
	strategicIdx := len(hashes) / 2
	quickSim := CompareHashes(originalFingerprint.StrategicHash, hashes[strategicIdx])
	if quickSim < 0.78 {
		return false, quickSim, nil
	}

	// Comparación completa
	similarity := compareHashArrays(originalFingerprint.FrameHashes, hashes)
	return similarity > SimilarityThreshold, similarity, nil
}

// CompareFingerprints compara dos fingerprints entre sí
func CompareFingerprints(fp1, fp2 *VideoFingerprint) float64 {
	if fp1 == nil || fp2 == nil {
		return 0
	}

	// Quick check
	quickSim := CompareHashes(fp1.StrategicHash, fp2.StrategicHash)
	if quickSim < 0.78 {
		return 0
	}

	return compareHashArrays(fp1.FrameHashes, fp2.FrameHashes)
}

// CompareHashes compara dos hashes individuales y retorna similitud 0-1
func CompareHashes(hash1, hash2 uint64) float64 {
	if hash1 == 0 || hash2 == 0 {
		return 0
	}

	distance := hammingDistance(hash1, hash2)
	return 1.0 - (float64(distance) * 0.015625) // 1/64
}

// extractAndHashSingleFrame extrae un frame en un timestamp específico y lo hashea
func extractAndHashSingleFrame(videoPath string, timestamp float64) (uint64, error) {
	cmd := exec.Command("ffmpeg",
		"-ss", fmt.Sprintf("%.2f", timestamp),
		"-i", videoPath,
		"-vframes", "1",
		"-vf", "scale="+FFmpegScale,
		"-f", "image2pipe",
		"-vcodec", "mjpeg",
		"-q:v", "1",
		"-loglevel", "error",
		"pipe:1",
	)

	out, err := cmd.Output()
	if err != nil || len(out) == 0 {
		return 0, fmt.Errorf("ffmpeg failed: %w", err)
	}

	img, _, err := image.Decode(bytes.NewReader(out))
	if err != nil {
		return 0, fmt.Errorf("decode image: %w", err)
	}

	hash, err := goimagehash.PerceptionHash(img)
	if err != nil {
		return 0, fmt.Errorf("compute hash: %w", err)
	}

	return hash.GetHash(), nil
}

// compareHashArrays compara dos arrays de hashes y retorna similitud promedio
func compareHashArrays(hashes1, hashes2 []uint64) float64 {
	minLen := len(hashes1)
	if len(hashes2) < minLen {
		minLen = len(hashes2)
	}

	if minLen == 0 {
		return 0
	}

	totalSimilarity := 0.0
	for i := 0; i < minLen; i++ {
		distance := hammingDistance(hashes1[i], hashes2[i])
		totalSimilarity += 1.0 - (float64(distance) * 0.015625)
	}

	return totalSimilarity / float64(minLen)
}

// hammingDistance calcula la distancia de Hamming entre dos hashes
func hammingDistance(hash1, hash2 uint64) int {
	xor := hash1 ^ hash2
	count := 0
	for xor != 0 {
		count++
		xor &= xor - 1
	}
	return count
}

// calculateSafeTimestamps calcula timestamps seguros sin ir negativo ni fuera de rango
// También tolera videos cortados al final (1-2 segundos)
func calculateSafeTimestamps(duration float64) []float64 {
	minSafe := FrameSafeMargin
	maxSafe := duration - FrameSafeMargin

	// Si el video es muy corto, usar márgenes más pequeños
	if duration < 5 {
		minSafe = 0.5
		maxSafe = duration - 0.5
	}

	if maxSafe <= minSafe {
		// Video muy corto, tomar un frame en el medio
		return []float64{duration / 2}
	}

	timestamps := make([]float64, 0, MaxFramesToExtract)

	// Inicio seguro
	timestamps = append(timestamps, minSafe)

	// Mitad
	mid := duration / 2
	if mid > minSafe && mid < maxSafe {
		timestamps = append(timestamps, mid)
	} else if mid >= maxSafe {
		// Si la mitad está en la zona peligrosa, tomar un punto entre inicio y final
		timestamps = append(timestamps, (minSafe+maxSafe)/2)
	}

	// Final seguro (FrameSafeMargin segundos antes del fin)
	if len(timestamps) < MaxFramesToExtract {
		timestamps = append(timestamps, maxSafe)
	}

	return timestamps
}

// getVideoDuration obtiene la duración de un video usando ffprobe
func getVideoDuration(videoPath string) (float64, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		videoPath,
	)

	out, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe failed: %w", err)
	}

	durationStr := string(bytes.TrimSpace(out))
	if durationStr == "" {
		return 0, fmt.Errorf("empty duration response")
	}

	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, fmt.Errorf("parse duration: %w", err)
	}

	return duration, nil
}

// downloadFile descarga un archivo desde una URL
func downloadFile(url, filepath string) error {
	client := &http.Client{Timeout: 3 * time.Minute}

	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %d", resp.StatusCode)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer out.Close()

	_, err = io.CopyBuffer(out, resp.Body, make([]byte, 64*1024))
	if err != nil {
		return fmt.Errorf("copy file: %w", err)
	}

	return nil
}

// absDiff calcula el valor absoluto de la diferencia
func absDiff(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
