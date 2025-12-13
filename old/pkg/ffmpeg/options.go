package ffmpeg

import "time"

// Options to configure FFmpeg
type Options struct {
	BinaryPath string        // Path to the ffmpeg binary (optional, searches in PATH if empty)
	WorkingDir string        // Working directory
	Timeout    time.Duration // Timeout for commands (0 = no timeout)
	Env        []string      // Additional environment variables
}
