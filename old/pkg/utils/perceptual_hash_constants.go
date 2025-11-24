package utils

// Video fingerprint configuration constants
const (
	// MaxFramesToExtract is the maximum number of frames to extract from a video for fingerprinting
	MaxFramesToExtract = 3

	// SimilarityThreshold is the minimum similarity score (0-1) for two videos to be considered the same
	SimilarityThreshold = 0.82

	// FrameWidth is the width in pixels to which frames are resized for processing
	FrameWidth = 128

	// FrameHeight is the height in pixels to which frames are resized for processing
	FrameHeight = 96

	// FFmpegScale is the scale parameter passed to ffmpeg for frame extraction
	FFmpegScale = "128:96"

	// MinVideoDuration is the minimum duration in seconds for a video to be processed
	MinVideoDuration = 3.0

	// FrameSafeMargin is the safety margin in seconds from the start and end of a video
	// to avoid extracting frames from potentially corrupted or black regions
	FrameSafeMargin = 2.0
)
