package ffmpeg

import "time"

// Result contains the result of the execution
type Result struct {
	ExitCode int
	Stdout   string
	Stderr   string
	Duration time.Duration
}
