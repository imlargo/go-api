package ffmpeg

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// FFmpeg represents an instance of FFmpeg with configuration
type FFmpeg struct {
	binaryPath string
	workingDir string
	timeout    time.Duration
	env        []string
}

// New creates a new instance of FFmpeg
func New(opts *Options) (*FFmpeg, error) {
	if opts == nil {
		opts = &Options{
			WorkingDir: "./tmp/ffmpeg",
		}
	}

	// Determine the binary path
	binaryPath := opts.BinaryPath
	if binaryPath == "" {
		path, err := exec.LookPath("ffmpeg")
		if err != nil {
			return nil, ErrFFmpegNotFound
		}
		binaryPath = path
	}

	// Check that the binary exists and is executable
	if err := checkFFmpegBinary(binaryPath); err != nil {
		return nil, err
	}

	return &FFmpeg{
		binaryPath: binaryPath,
		workingDir: opts.WorkingDir,
		timeout:    opts.Timeout,
		env:        opts.Env,
	}, nil
}

// NewCommand creates a new FFmpeg command
func (f *FFmpeg) NewCommand() *Command {
	return &Command{
		ffmpeg:       f,
		globalArgs:   []string{},
		inputArgs:    []string{},
		outputArgs:   []string{},
		videoFilters: []string{},
		audioFilters: []string{},
	}
}

// Probe retrieves information about a media file
func (f *FFmpeg) Probe(filePath string) (*Result, error) {
	probePath, err := exec.LookPath("ffprobe")
	if err != nil {
		return nil, errors.New("ffprobe binary not found in PATH")
	}

	cmd := exec.Command(probePath, "-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", filePath)

	if f.workingDir != "" {
		cmd.Dir = f.workingDir
	}

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	startTime := time.Now()
	err = cmd.Run()
	duration := time.Since(startTime)

	result := &Result{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: duration,
	}

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				result.ExitCode = status.ExitStatus()
			}
		}
		return result, fmt.Errorf("ffprobe failed: %s", stderr.String())
	}

	return result, nil
}

// Version retrieves the version of FFmpeg
func (f *FFmpeg) Version() (*Result, error) {
	cmd := exec.Command(f.binaryPath, "-version")

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	startTime := time.Now()
	err := cmd.Run()
	duration := time.Since(startTime)

	result := &Result{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: duration,
	}

	if err != nil {
		return result, fmt.Errorf("failed to get ffmpeg version: %w", err)
	}

	return result, nil
}

// GetVideoInfo obtiene información detallada del video incluyendo rotación
func (f *FFmpeg) GetVideoInfo(filePath string) (*VideoInfo, error) {
	result, err := f.Probe(filePath)
	if err != nil {
		return nil, fmt.Errorf("error al ejecutar ffprobe: %w", err)
	}

	var probeResult ProbeResult
	if err := json.Unmarshal([]byte(result.Stdout), &probeResult); err != nil {
		return nil, fmt.Errorf("error al parsear JSON de ffprobe: %w", err)
	}

	videoInfo := &VideoInfo{}

	if duration, err := strconv.ParseFloat(probeResult.Format.Duration, 64); err == nil {
		videoInfo.Duration = duration
	}

	if bitrate, err := strconv.Atoi(probeResult.Format.Bitrate); err == nil {
		videoInfo.Bitrate = bitrate
	}

	videoInfo.Format = probeResult.Format.Format

	var videoStream, audioStream *Stream
	for i := range probeResult.Streams {
		stream := &probeResult.Streams[i]
		switch stream.CodecType {
		case "video":
			if videoStream == nil {
				videoStream = stream
			}
		case "audio":
			if audioStream == nil {
				audioStream = stream
			}
		}
	}

	if videoStream != nil {
		videoInfo.Width = videoStream.Width
		videoInfo.Height = videoStream.Height
		videoInfo.Codec = videoStream.CodecName

		if fps := parseFrameRate(videoStream.RFrameRate); fps > 0 {
			videoInfo.FPS = fps
		} else if fps := parseFrameRate(videoStream.AvgFrameRate); fps > 0 {
			videoInfo.FPS = fps
		}

		if videoInfo.Bitrate == 0 {
			if bitrate, err := strconv.Atoi(videoStream.Bitrate); err == nil {
				videoInfo.Bitrate = bitrate
			}
		}
	}

	if audioStream != nil {
		videoInfo.AudioCodec = audioStream.CodecName
		videoInfo.AudioChannels = audioStream.Channels

		if bitrate, err := strconv.Atoi(audioStream.Bitrate); err == nil {
			videoInfo.AudioBitrate = bitrate
		}
	}

	videoInfo.Rotation = probeResult.GetVideoRotation()

	return videoInfo, nil
}

// ValidateVideoFile valida que el archivo sea un video válido
func (f *FFmpeg) ValidateVideoFile(filePath string) error {
	info, err := f.GetVideoInfo(filePath)
	if err != nil {
		return fmt.Errorf("file is not a valid video: %w", err)
	}

	if info.Width == 0 || info.Height == 0 {
		return errors.New("file does not contain a valid video stream")
	}

	if info.Duration <= 0 {
		return errors.New("invalid video duration")
	}

	return nil
}
