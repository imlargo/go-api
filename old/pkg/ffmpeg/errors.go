package ffmpeg

import "errors"

var (
	ErrFFmpegNotFound = errors.New("ffmpeg binary not found in PATH")
	ErrInvalidInput   = errors.New("invalid input file")
	ErrInvalidOutput  = errors.New("invalid output file")
	ErrCommandFailed  = errors.New("ffmpeg command failed")
	ErrTimeout        = errors.New("ffmpeg command timed out")
	ErrNoInput        = errors.New("no input file specified")
	ErrFileNotFound   = errors.New("file not found")
)
