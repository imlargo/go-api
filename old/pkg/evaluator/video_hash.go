package evaluator

import "time"

// VideoHash representa el hash perceptual de un video
type VideoHash struct {
	FilePath    string
	Hash        []uint64
	Duration    float64
	FrameHashes []uint64
	Timestamp   time.Time
}
