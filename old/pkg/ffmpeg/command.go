package ffmpeg

import (
	"context"
	"fmt"
	"math"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

// Command represents a constructed FFmpeg command
type Command struct {
	ffmpeg       *FFmpeg
	globalArgs   []string // Global arguments (go before -i)
	inputArgs    []string // Input specific arguments
	input        string
	outputArgs   []string // Output specific arguments (go after -i but before output)
	output       string
	videoFilters []string
	audioFilters []string
}

// Input sets the input file
func (c *Command) Input(path string) *Command {
	if path != "" {
		c.input = path
	}
	return c
}

// Output sets the output file
func (c *Command) Output(path string) *Command {
	if path != "" {
		c.output = path
	}
	return c
}

// Args adds custom arguments to output args (this is the most common case)
func (c *Command) Args(args ...string) *Command {
	c.outputArgs = append(c.outputArgs, args...)
	return c
}

// InputArgs adds arguments that go after global args but before -i
// Examples: -noautorotate, -thread_queue_size, etc.
func (c *Command) InputArgs(args ...string) *Command {
	c.inputArgs = append(c.inputArgs, args...)
	return c
}

// GlobalArgs adds arguments that go before -i (like -y, -hide_banner, etc.)
func (c *Command) GlobalArgs(args ...string) *Command {
	c.globalArgs = append(c.globalArgs, args...)
	return c
}

// VideoCodec sets the video codec
func (c *Command) VideoCodec(codec string) *Command {
	return c.Args("-c:v", codec)
}

// AudioCodec sets the audio codec
func (c *Command) AudioCodec(codec string) *Command {
	return c.Args("-c:a", codec)
}

// VideoBitrate sets the video bitrate
func (c *Command) VideoBitrate(bitrate string) *Command {
	return c.Args("-b:v", bitrate)
}

// AudioBitrate sets the audio bitrate
func (c *Command) AudioBitrate(bitrate string) *Command {
	return c.Args("-b:a", bitrate)
}

// Scale resizes the video
func (c *Command) Scale(width, height int) *Command {
	return c.VideoFilter(fmt.Sprintf("scale=%d:%d", width, height))
}

// ScaleByFactors scales the video with different factors for width and height
func (c *Command) ScaleByFactors(widthFactor, heightFactor float64) *Command {
	return c.VideoFilter(fmt.Sprintf("scale=iw*%.3f:ih*%.3f", widthFactor, heightFactor))
}

func (c *Command) ZoomAndPan(zoomFactor float64, offsetX, offsetY int) *Command {
	var offsetXStr, offsetYStr string
	if offsetX >= 0 {
		offsetXStr = fmt.Sprintf("+%d", offsetX)
	} else {
		offsetXStr = fmt.Sprintf("%d", offsetX)
	}
	if offsetY >= 0 {
		offsetYStr = fmt.Sprintf("+%d", offsetY)
	} else {
		offsetYStr = fmt.Sprintf("%d", offsetY)
	}

	return c.VideoFilter(fmt.Sprintf("scale=iw*%.3f:ih*%.3f,crop=iw/%.3f:ih/%.3f:(ow-iw)/2%s:(oh-ih)/2%s",
		zoomFactor, zoomFactor, zoomFactor, zoomFactor, offsetXStr, offsetYStr))
}

// RotateByDegrees
func (c *Command) RotateByDegrees(degrees float64) *Command {
	radians := degrees * math.Pi / 180.0
	return c.VideoFilter(fmt.Sprintf("rotate=%.6f", radians))
}

// VideoFilter applies video filters
func (c *Command) VideoFilter(filter string) *Command {
	c.videoFilters = append(c.videoFilters, filter)
	return c
}

// AddVideoFilter is an alias for VideoFilter for consistency
func (c *Command) AddVideoFilter(filter string) *Command {
	return c.VideoFilter(filter)
}

// AudioFilter applies audio filters
func (c *Command) AudioFilter(filter string) *Command {
	c.audioFilters = append(c.audioFilters, filter)
	return c
}

// ClearVideoFilters clears all video filters
func (c *Command) ClearVideoFilters() *Command {
	c.videoFilters = nil
	return c
}

// ClearAudioFilters clears all audio filters
func (c *Command) ClearAudioFilters() *Command {
	c.audioFilters = nil
	return c
}

// Seek sets the start time
func (c *Command) Seek(position string) *Command {
	return c.Args("-ss", position)
}

// Duration sets the duration
func (c *Command) Duration(duration string) *Command {
	return c.Args("-t", duration)
}

// Overwrite allows overwriting output files
func (c *Command) Overwrite() *Command {
	return c.GlobalArgs("-y") // -y must go before -i
}

// Quality sets the CRF quality for x264/x265
func (c *Command) Quality(crf int) *Command {
	return c.Args("-crf", fmt.Sprintf("%d", crf))
}

// Preset sets the encoding preset
func (c *Command) Preset(preset string) *Command {
	return c.Args("-preset", preset)
}

// Format sets the output format
func (c *Command) Format(format string) *Command {
	return c.Args("-f", format)
}

// FlipHorizontal flips the video horizontally (mirror effect)
func (c *Command) FlipHorizontal() *Command {
	return c.VideoFilter("hflip")
}

// FlipVertical flips the video vertically
func (c *Command) FlipVertical() *Command {
	return c.VideoFilter("vflip")
}

// GetArgs returns all command arguments in the correct order
func (c *Command) GetArgs() []string {
	var args []string

	// 1. Global arguments (like -y, -hide_banner, etc.)
	args = append(args, c.globalArgs...)

	// 2. Input arguments (if any)
	args = append(args, c.inputArgs...)

	// 3. Input file
	if c.input != "" {
		args = append(args, "-i", c.input)
	}

	// 4. Output arguments (codecs, bitrates, etc.)
	args = append(args, c.outputArgs...)

	// 5. Video filters
	if len(c.videoFilters) > 0 {
		args = append(args, "-vf", strings.Join(c.videoFilters, ","))
	}

	// 6. Audio filters
	if len(c.audioFilters) > 0 {
		args = append(args, "-af", strings.Join(c.audioFilters, ","))
	}

	// 7. Output file (must be last)
	if c.output != "" {
		args = append(args, c.output)
	}

	return args
}

// GetFullCommand returns the full command as a string
func (c *Command) GetFullCommand() string {
	return fmt.Sprintf("%s %s", c.ffmpeg.binaryPath, strings.Join(c.GetArgs(), " "))
}

// Run executes the FFmpeg command
func (c *Command) Run() (*Result, error) {
	return c.RunWithContext(context.Background())
}

func (c *Command) String() string {
	return c.GetFullCommand()
}

// RunWithContext executes the FFmpeg command with context
func (c *Command) RunWithContext(ctx context.Context) (*Result, error) {
	if c.input == "" {
		return nil, ErrNoInput
	}

	if !fileExists(c.input) {
		return nil, ErrInvalidInput
	}

	// Get final arguments including filters
	finalArgs := c.GetArgs()

	// Create context with timeout if configured
	if c.ffmpeg.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.ffmpeg.timeout)
		defer cancel()
	}

	startTime := time.Now()
	cmd := exec.CommandContext(ctx, c.ffmpeg.binaryPath, finalArgs...)

	// Set working directory
	if c.ffmpeg.workingDir != "" {
		cmd.Dir = c.ffmpeg.workingDir
	}

	// Set environment variables
	if len(c.ffmpeg.env) > 0 {
		cmd.Env = append(os.Environ(), c.ffmpeg.env...)
	}

	// Capture stdout and stderr
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	duration := time.Since(startTime)

	result := &Result{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: duration,
	}

	// Get exit code
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return result, ErrTimeout
		}

		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				result.ExitCode = status.ExitStatus()
			} else {
				result.ExitCode = 1
			}
		} else {
			result.ExitCode = 1
		}

		return result, fmt.Errorf("%w: %s", ErrCommandFailed, stderr.String())
	}

	result.ExitCode = 0
	return result, nil
}

// Repurposer
// Phase 1: Initial Normalization Extensions
func (c *Command) SetResolution(width, height int) *Command {
	return c.VideoFilter(fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2", width, height, width, height))
}

func (c *Command) ForceAspectRatio(ratio string) *Command {
	return c.VideoFilter(fmt.Sprintf("setsar=%s", ratio))
}

func (c *Command) BasicColorCorrection(brightness float64, contrast float64, gamma float64) *Command {
	return c.VideoFilter(fmt.Sprintf("eq=brightness=%.3f:contrast=%.3f:gamma=%.3f", brightness, contrast, gamma))
}

// Phase 2: Visual Transformations Extensions
func (c *Command) CropPartial(top, bottom, left, right float64) *Command {
	return c.VideoFilter(fmt.Sprintf("crop=iw-iw*%.3f-iw*%.3f:ih-ih*%.3f-ih*%.3f:iw*%.3f:ih*%.3f", left, right, top, bottom, left, top))
}

func (c *Command) LensDistortion(k1, k2 float64) *Command {
	return c.VideoFilter(fmt.Sprintf("lenscorrection=k1=%.6f:k2=%.6f", k1, k2))
}

func (c *Command) PerspectiveShift(x0, y0, x1, y1, x2, y2, x3, y3 float64) *Command {
	return c.VideoFilter(fmt.Sprintf(
		"perspective="+
			"x0=(0+%f*W):y0=(0+%f*H):"+
			"x1=(W+%f*W):y1=(0+%f*H):"+
			"x2=(0+%f*W):y2=(H+%f*H):"+
			"x3=(W+%f*W):y3=(H+%f*H):"+
			"sense=destination:interpolation=linear",
		x0, y0, x1, y1, x2, y2, x3, y3,
	))
}

// Phase 3: Color and Style Alterations
func (c *Command) AdjustBrightness(value float64) *Command {
	return c.VideoFilter(fmt.Sprintf("eq=brightness=%.3f", value))
}

func (c *Command) AdjustContrast(value float64) *Command {
	return c.VideoFilter(fmt.Sprintf("eq=contrast=%.3f", value))
}

func (c *Command) AdjustGamma(value float64) *Command {
	return c.VideoFilter(fmt.Sprintf("eq=gamma=%.3f", value))
}

func (c *Command) AdjustSaturation(value float64) *Command {
	return c.VideoFilter(fmt.Sprintf("eq=saturation=%.3f", value))
}

func (c *Command) ColorChannelMix(rr, rg, rb, gr, gg, gb, br, bg, bb float64) *Command {
	return c.VideoFilter(fmt.Sprintf("colorchannelmixer=rr=%.3f:rg=%.3f:rb=%.3f:gr=%.3f:gg=%.3f:gb=%.3f:br=%.3f:bg=%.3f:bb=%.3f", rr, rg, rb, gr, gg, gb, br, bg, bb))
}

func (c *Command) RGBShift(rh, rv, gh, gv, bh, bv float64) *Command {
	return c.VideoFilter(fmt.Sprintf("rgbashift=rh=%.1f:rv=%.1f:gh=%.1f:gv=%.1f:bh=%.1f:bv=%.1f", rh, rv, gh, gv, bh, bv))
}

func (c *Command) ApplyLUT(lutFile string) *Command {
	return c.VideoFilter(fmt.Sprintf("lut3d=%s", lutFile))
}

func (c *Command) VintageEffect() *Command {
	return c.VideoFilter("curves=vintage")
}

func (c *Command) SepiaEffect() *Command {
	return c.VideoFilter("colorchannelmixer=.393:.769:.189:0:.349:.686:.168:0:.272:.534:.131")
}

func (c *Command) BlackWhiteEffect() *Command {
	return c.VideoFilter("hue=s=0")
}

func (c *Command) WarmColorGrading(intensity float64) *Command {
	return c.VideoFilter(fmt.Sprintf("colorbalance=rs=%.3f:gs=-%.3f:bs=-%.3f", intensity*0.1, intensity*0.05, intensity*0.1))
}

func (c *Command) ColdColorGrading(intensity float64) *Command {
	return c.VideoFilter(fmt.Sprintf("colorbalance=rs=-%.3f:gs=-%.3f:bs=%.3f", intensity*0.1, intensity*0.05, intensity*0.1))
}

// Phase 4: Noise and Texture Effects
func (c *Command) AddLuminanceNoise(strength float64) *Command {
	return c.VideoFilter(fmt.Sprintf("noise=alls=%d:allf=t+u", int(strength*100)))
}

func (c *Command) AddChrominanceNoise(strength float64) *Command {
	return c.VideoFilter(fmt.Sprintf("noise=c0s=%d:c1s=%d:c2s=%d:c3s=%d:allf=t+u", int(strength*50), int(strength*50), int(strength*50), int(strength*50)))
}

func (c *Command) FilmGrain(strength, size float64) *Command {
	return c.VideoFilter(fmt.Sprintf("noise=alls=%d:allf=t+u", int(strength*75)))
}

func (c *Command) AddSubtleLines(opacity float64, spacing int) *Command {
	// Create a subtle lines pattern using geq
	return c.VideoFilter(fmt.Sprintf("geq=r='if(mod(Y,%d),r(X,Y),r(X,Y)*%.3f)':g='if(mod(Y,%d),g(X,Y),g(X,Y)*%.3f)':b='if(mod(Y,%d),b(X,Y),b(X,Y)*%.3f)'", spacing, 1-opacity, spacing, 1-opacity, spacing, 1-opacity))
}

func (c *Command) AddVignette(intensity float64) *Command {
	return c.VideoFilter(fmt.Sprintf("vignette=PI/4*%.3f", intensity))
}

// Histogram adjustments
func (c *Command) AdjustHistogram(shadows, midtones, highlights float64) *Command {
	return c.VideoFilter(fmt.Sprintf("curves=shadows=%.3f:midtones=%.3f:highlights=%.3f", shadows, midtones, highlights))
}

// Combined effects for efficiency
func (c *Command) ColorGradeBasic(brightness, contrast, saturation, gamma float64) *Command {
	return c.VideoFilter(fmt.Sprintf("eq=brightness=%.3f:contrast=%.3f:saturation=%.3f:gamma=%.3f", brightness, contrast, saturation, gamma))
}

// scaleToFillNoBorders returns the minimum scale factor required
// so that, after rotating by angleDeg, no black borders/triangles appear.
// w,h: original dimensions of the video (desired output)
func scaleToFillNoBorders(w, h int, angleDeg float64) float64 {
	if angleDeg == 0 {
		return 1.0
	}

	theta := angleDeg * math.Pi / 180.0
	absCos := math.Abs(math.Cos(theta))
	absSin := math.Abs(math.Sin(theta))

	// After rotation, the bounding box of the video becomes larger.
	// To fill the original frame without black borders, we need to scale up
	// so that the rotated content covers the entire original area.
	//
	// For a rectangle (w x h) rotated by θ:
	// New bounding box width:  w' = w*|cos(θ)| + h*|sin(θ)|
	// New bounding box height: h' = w*|sin(θ)| + h*|cos(θ)|
	//
	// To ensure the original frame is filled, we need:
	// scale * w >= w' and scale * h >= h'
	// Therefore: scale >= max(w'/w, h'/h)

	wf := float64(w)
	hf := float64(h)

	// Calculate how much the bounding box grows in each dimension
	scaleForWidth := (wf*absCos + hf*absSin) / wf
	scaleForHeight := (wf*absSin + hf*absCos) / hf

	// Use the larger scale to ensure both dimensions are covered
	return math.Max(scaleForWidth, scaleForHeight)
}

// RotateZoomToFill rotates and applies a compensatory zoom to fill,
// and finally crops to the original size (w x h) centered.
// Uses rotation with bilinear smoothing.
func (c *Command) RotateZoomToFill(angleDeg float64, videoInfo *VideoInfo, applyHFlip bool) *Command {
	originalW := videoInfo.Width
	originalH := videoInfo.Height

	// Dimensiones después de aplicar transpose (si aplica)
	currentW, currentH := originalW, originalH
	needsTranspose := videoInfo.NeedsTranspose()

	// 1) CRÍTICO: Aplicar rotación de metadata primero
	if needsTranspose {
		transposeFilter := videoInfo.GetTransposeFilter()
		if transposeFilter != "" {
			c.VideoFilter(transposeFilter)

			// Si es rotación de 90° o 270°, intercambiar dimensiones
			if videoInfo.Rotation == 90 || videoInfo.Rotation == -90 ||
				videoInfo.Rotation == 270 || videoInfo.Rotation == -270 {
				currentW, currentH = currentH, currentW
			}
		}
	}

	// 2) Aplicar rotación personalizada
	if angleDeg != 0 {
		rad := angleDeg * math.Pi / 180.0
		c.VideoFilter(fmt.Sprintf("rotate=%.6f:bilinear=1", rad))

		// 3) Zoom compensatorio
		zoom := scaleToFillNoBorders(currentW, currentH, angleDeg)
		println("Zoom: ", zoom)
		c.VideoFilter(fmt.Sprintf("scale=iw*%.6f:ih*%.6f", zoom, zoom))
	}

	// 4) Crop
	c.VideoFilter(fmt.Sprintf("crop=%d:%d:(iw-%d)/2:(ih-%d)/2",
		currentW, currentH, currentW, currentH))

	// 5) IMPORTANTE: Aplicar hflip ANTES del transpose inverso
	if applyHFlip {
		c.VideoFilter("hflip")
	}

	// 6) Volver a aplicar la rotación inversa
	if needsTranspose {
		inverseTranspose := getInverseTranspose(videoInfo.Rotation)
		if inverseTranspose != "" {
			c.VideoFilter(inverseTranspose)
		}
	}

	return c
}

// getInverseTranspose retorna el transpose inverso para volver a la orientación original
func getInverseTranspose(rotation int) string {
	switch rotation {
	case 90, -270:
		return "transpose=2" // Inverso de transpose=1
	case -90, 270:
		return "transpose=1" // Inverso de transpose=2
	case 180, -180:
		return "transpose=2,transpose=2" // Inverso de 180° es otro 180°
	default:
		return ""
	}
}

// METADATA OPERATIONS
func (c *Command) ApplyMaximalContainerModification() *Command {
	return c.Args("-movflags", "+faststart+frag_keyframe+separate_moof+omit_tfhd_offset")
}

// RemoveAllEXIF removes all EXIF metadata
func (c *Command) RemoveAllEXIF() *Command {
	return c.Args("-map_metadata", "-1")
}

// RemoveSelectiveMetadata removes specific metadata
func (c *Command) RemoveSelectiveMetadata(keys []string) *Command {
	for _, key := range keys {
		c.Args("-metadata", fmt.Sprintf("%s=", key))
	}
	return c
}

// RemoveSelectiveMetadata removes specific metadata
func (c *Command) RemoveMetadata(key string) *Command {
	c.Args("-metadata", fmt.Sprintf("%s=", key))
	return c
}

func (c *Command) SetSelectiveMetadata(data map[string]string) *Command {
	for key, value := range data {
		c.Args("-metadata", fmt.Sprintf("%s=%s", key, value))
	}
	return c
}

// SetCustomDate sets a custom date
func (c *Command) SetCustomDate(newDate time.Time) *Command {
	return c.Args("-metadata", fmt.Sprintf("creation_time=%s", newDate.Format("2006-01-02T15:04:05.000000Z")))
}

// TEXT OVERLAY OPERATIONS

// TextOverlayOptions configures options for text overlays
type TextOverlayOptions struct {
	Text        string
	TextPath    string // Path to a text file with multiple lines
	FontFile    string
	Font        string
	FontSize    int
	LineHeight  float64
	MaxWidth    int
	FontColor   string
	X           string  // X position (number, left, center, right, or expression)
	Y           string  // Y position (number, top, center, bottom, or expression)
	Alpha       float64 // Transparency (0.0 to 1.0)
	BoxEnable   bool    // Enable background box
	BoxColor    string  // Box color
	BoxOpacity  float64 // Box opacity
	BorderW     int     // Border width
	BorderColor string  // Border color
	ShadowX     int     // Shadow X offset
	ShadowY     int     // Shadow Y offset
	ShadowColor string  // Shadow color
	StartTime   string  // Start time (format: HH:MM:SS)
	EndTime     string  // End time (format: HH:MM:SS)
	FadeIn      float64 // Fade in duration in seconds
	FadeOut     float64 // Fade out duration in seconds
}

// SimpleTextOverlay adds simple text at specific position
func (c *Command) SimpleTextOverlay(text string, x, y int) *Command {
	filter := fmt.Sprintf("drawtext=text='%s':x=%d:y=%d", text, x, y)
	return c.VideoFilter(filter)
}

// TextOverlayWithFont adds text with custom font
func (c *Command) TextOverlayWithFont(text, fontFile string, fontSize int, x, y int) *Command {
	filter := fmt.Sprintf("drawtext=text='%s':fontfile='%s':fontsize=%d:x=%d:y=%d",
		text, fontFile, fontSize, x, y)
	return c.VideoFilter(filter)
}

// TextOverlayWithColor adds text with custom color
func (c *Command) TextOverlayWithColor(text, color string, fontSize, x, y int) *Command {
	filter := fmt.Sprintf("drawtext=text='%s':fontcolor=%s:fontsize=%d:x=%d:y=%d",
		text, color, fontSize, x, y)
	return c.VideoFilter(filter)
}

// AdvancedTextOverlay adds text with advanced options
func (c *Command) AdvancedTextOverlay(opts TextOverlayOptions) *Command {
	var parts []string

	// CRITICAL: Use textfile for multi-line support
	if opts.TextPath != "" {
		// Escape single quotes in path for FFmpeg
		escapedPath := strings.ReplaceAll(opts.TextPath, "'", "'\\''")
		parts = append(parts, fmt.Sprintf("textfile='%s'", escapedPath))

		// CRITICAL: Set text_shaping=1 to enable proper text rendering
		// This helps with complex scripts and proper glyph positioning
		parts = append(parts, "text_shaping=1")
	} else if opts.Text != "" {
		// Fallback: inline text (not recommended for multi-line)
		escapedText := opts.Text
		escapedText = strings.ReplaceAll(escapedText, "'", "'\\''")
		escapedText = strings.ReplaceAll(escapedText, ":", "\\:")
		escapedText = strings.ReplaceAll(escapedText, "\n", "\\n")
		parts = append(parts, fmt.Sprintf("text='%s'", escapedText))
	}

	// Font configuration
	if opts.FontFile != "" {
		escapedFont := strings.ReplaceAll(opts.FontFile, "'", "'\\''")
		parts = append(parts, fmt.Sprintf("fontfile='%s'", escapedFont))
	} else if opts.Font != "" {
		parts = append(parts, fmt.Sprintf("font='%s'", opts.Font))
	}

	// Text alignment
	parts = append(parts, "text_align=center")

	// Font size
	if opts.FontSize > 0 {
		parts = append(parts, fmt.Sprintf("fontsize=%d", opts.FontSize))
	}

	// Line spacing - CRITICAL for multi-line text readability
	if opts.LineHeight > 0 {
		parts = append(parts, fmt.Sprintf("line_spacing=%.2f", opts.LineHeight))
	} else {
		// Default line spacing if not specified
		parts = append(parts, "line_spacing=0")
	}

	// Font color
	if opts.FontColor != "" {
		parts = append(parts, fmt.Sprintf("fontcolor=%s", opts.FontColor))
	}

	// Maximum text width
	if opts.MaxWidth > 0 {
		parts = append(parts, fmt.Sprintf("textw=%d", opts.MaxWidth))
	}

	// Position
	if opts.X != "" {
		parts = append(parts, fmt.Sprintf("x=%s", opts.X))
	}
	if opts.Y != "" {
		parts = append(parts, fmt.Sprintf("y=%s", opts.Y))
	}

	// Transparency
	if opts.Alpha > 0 && opts.Alpha <= 1 {
		parts = append(parts, fmt.Sprintf("alpha=%.2f", opts.Alpha))
	}

	// Background box
	if opts.BoxEnable {
		parts = append(parts, "box=1")
		if opts.BoxColor != "" {
			parts = append(parts, fmt.Sprintf("boxcolor=%s", opts.BoxColor))
		}
		if opts.BoxOpacity > 0 && opts.BoxOpacity <= 1 {
			parts = append(parts, fmt.Sprintf("boxborderw=%d", int(opts.BoxOpacity*255)))
		}
	}

	// Border (outline)
	if opts.BorderW > 0 {
		parts = append(parts, fmt.Sprintf("borderw=%d", opts.BorderW))
		if opts.BorderColor != "" {
			parts = append(parts, fmt.Sprintf("bordercolor=%s", opts.BorderColor))
		}
	}

	// Shadow
	if opts.ShadowX != 0 || opts.ShadowY != 0 {
		parts = append(parts, fmt.Sprintf("shadowx=%d", opts.ShadowX))
		parts = append(parts, fmt.Sprintf("shadowy=%d", opts.ShadowY))
		if opts.ShadowColor != "" {
			parts = append(parts, fmt.Sprintf("shadowcolor=%s", opts.ShadowColor))
		}
	}

	// Timing
	if opts.StartTime != "" && opts.EndTime != "" {
		parts = append(parts, fmt.Sprintf("enable='between(t,%s,%s)'",
			convertTimeToSeconds(opts.StartTime), convertTimeToSeconds(opts.EndTime)))
	}

	// Fade effects
	if opts.FadeIn > 0 || opts.FadeOut > 0 {
		var fadeExpr string
		if opts.FadeIn > 0 && opts.FadeOut > 0 {
			fadeExpr = fmt.Sprintf("fade(t,0,%.2f)*fade(t,%.2f,%.2f)",
				opts.FadeIn,
				convertTimeToSecondsFloat(opts.EndTime)-opts.FadeOut,
				opts.FadeOut)
		} else if opts.FadeIn > 0 {
			fadeExpr = fmt.Sprintf("fade(t,0,%.2f)", opts.FadeIn)
		} else if opts.FadeOut > 0 {
			fadeExpr = fmt.Sprintf("fade(t,%.2f,%.2f)",
				convertTimeToSecondsFloat(opts.EndTime)-opts.FadeOut,
				opts.FadeOut)
		}
		if fadeExpr != "" {
			parts = append(parts, fmt.Sprintf("alpha='%s'", fadeExpr))
		}
	}

	// Explicitly set expansion=normal to avoid expression expansion in text content
	parts = append(parts, "expansion=normal")
	filter := fmt.Sprintf("drawtext=%s", strings.Join(parts, ":"))

	// Debug: Print generated filter
	fmt.Printf("DEBUG: FFmpeg drawtext filter: %s\n", filter)

	return c.VideoFilter(filter)
}

// Convenience methods for common positions
func (c *Command) TextOverlayTopLeft(text string, fontSize int) *Command {
	return c.SimpleTextOverlay(text, 10, 10)
}

func (c *Command) TextOverlayTopRight(text string, fontSize int) *Command {
	filter := fmt.Sprintf("drawtext=text='%s':fontsize=%d:x=w-tw-10:y=10", text, fontSize)
	return c.VideoFilter(filter)
}

func (c *Command) TextOverlayBottomLeft(text string, fontSize int) *Command {
	filter := fmt.Sprintf("drawtext=text='%s':fontsize=%d:x=10:y=h-th-10", text, fontSize)
	return c.VideoFilter(filter)
}

func (c *Command) TextOverlayBottomRight(text string, fontSize int) *Command {
	filter := fmt.Sprintf("drawtext=text='%s':fontsize=%d:x=w-tw-10:y=h-th-10", text, fontSize)
	return c.VideoFilter(filter)
}

func (c *Command) TextOverlayCenter(text string, fontSize int) *Command {
	filter := fmt.Sprintf("drawtext=text='%s':fontsize=%d:x=(w-tw)/2:y=(h-th)/2", text, fontSize)
	return c.VideoFilter(filter)
}

func (c *Command) TextOverlayTopCenter(text string, fontSize int) *Command {
	filter := fmt.Sprintf("drawtext=text='%s':fontsize=%d:x=(w-tw)/2:y=10", text, fontSize)
	return c.VideoFilter(filter)
}

func (c *Command) TextOverlayBottomCenter(text string, fontSize int) *Command {
	filter := fmt.Sprintf("drawtext=text='%s':fontsize=%d:x=(w-tw)/2:y=h-th-10", text, fontSize)
	return c.VideoFilter(filter)
}

// Special text effects
func (c *Command) TextOverlayWithOutline(text, fontColor, outlineColor string, fontSize, outlineWidth, x, y int) *Command {
	filter := fmt.Sprintf("drawtext=text='%s':fontsize=%d:fontcolor=%s:borderw=%d:bordercolor=%s:x=%d:y=%d",
		text, fontSize, fontColor, outlineWidth, outlineColor, x, y)
	return c.VideoFilter(filter)
}

func (c *Command) TextOverlayWithShadow(text string, fontSize, shadowX, shadowY, x, y int) *Command {
	filter := fmt.Sprintf("drawtext=text='%s':fontsize=%d:shadowx=%d:shadowy=%d:shadowcolor=black:x=%d:y=%d",
		text, fontSize, shadowX, shadowY, x, y)
	return c.VideoFilter(filter)
}

func (c *Command) TextOverlayWithBackground(text, textColor, bgColor string, fontSize, x, y int) *Command {
	filter := fmt.Sprintf("drawtext=text='%s':fontsize=%d:fontcolor=%s:box=1:boxcolor=%s:boxborderw=5:x=%d:y=%d",
		text, fontSize, textColor, bgColor, x, y)
	return c.VideoFilter(filter)
}

// Texto animado
func (c *Command) ScrollingTextHorizontal(text string, fontSize int, speed float64) *Command {
	filter := fmt.Sprintf("drawtext=text='%s':fontsize=%d:x=w-%.2f*t:y=h/2", text, fontSize, speed)
	return c.VideoFilter(filter)
}

func (c *Command) ScrollingTextVertical(text string, fontSize int, speed float64) *Command {
	filter := fmt.Sprintf("drawtext=text='%s':fontsize=%d:x=w/2:y=h-%.2f*t", text, fontSize, speed)
	return c.VideoFilter(filter)
}

func (c *Command) TypewriterEffect(text string, fontSize int, x, y int, charsPerSecond float64) *Command {
	filter := fmt.Sprintf("drawtext=text='%s':fontsize=%d:x=%d:y=%d:text='%s':textfile=:start_number=0:fontcolor=white:enable='gte(t*%.2f,n)'",
		text, fontSize, x, y, text, charsPerSecond)
	return c.VideoFilter(filter)
}

// Multiple texts
func (c *Command) MultipleTextOverlays(textOverlays []TextOverlayOptions) *Command {
	for _, overlay := range textOverlays {
		c.AdvancedTextOverlay(overlay)
	}
	return c
}

// Texto con timestamp dinámico
func (c *Command) TimestampOverlay(format, position string, fontSize int) *Command {
	var x, y string
	switch position {
	case "top-left":
		x, y = "10", "10"
	case "top-right":
		x, y = "w-tw-10", "10"
	case "bottom-left":
		x, y = "10", "h-th-10"
	case "bottom-right":
		x, y = "w-tw-10", "h-th-10"
	default:
		x, y = "10", "10"
	}

	filter := fmt.Sprintf("drawtext=text='%s':fontsize=%d:x=%s:y=%s:fontcolor=white:box=1:boxcolor=black@0.5",
		format, fontSize, x, y)
	return c.VideoFilter(filter)
}

// Auxiliary functions
func convertTimeToSeconds(timeStr string) string {
	// Convert HH:MM:SS to seconds
	parts := strings.Split(timeStr, ":")
	if len(parts) != 3 {
		return "0"
	}

	hours := parseInt(parts[0])
	minutes := parseInt(parts[1])
	seconds := parseInt(parts[2])

	totalSeconds := hours*3600 + minutes*60 + seconds
	return fmt.Sprintf("%d", totalSeconds)
}

func convertTimeToSecondsFloat(timeStr string) float64 {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 3 {
		return 0.0
	}

	hours := parseFloat(parts[0])
	minutes := parseFloat(parts[1])
	seconds := parseFloat(parts[2])

	return hours*3600 + minutes*60 + seconds
}

func parseInt(s string) int {
	// Simple implementation of string to int conversion
	result := 0
	for _, char := range s {
		if char >= '0' && char <= '9' {
			result = result*10 + int(char-'0')
		}
	}
	return result
}

func parseFloat(s string) float64 {
	// Simple implementation of string to float conversion
	return float64(parseInt(s))
}

// Predefined text styles
func (c *Command) MovieTitleStyle(title string) *Command {
	return c.AdvancedTextOverlay(TextOverlayOptions{
		Text:        title,
		FontSize:    72,
		FontColor:   "white",
		X:           "(w-tw)/2",
		Y:           "(h-th)/2",
		BorderW:     3,
		BorderColor: "black",
		ShadowX:     5,
		ShadowY:     5,
		ShadowColor: "black@0.5",
		BoxEnable:   true,
		BoxColor:    "black@0.3",
	})
}

func (c *Command) SubtitleStyle(text string) *Command {
	return c.AdvancedTextOverlay(TextOverlayOptions{
		Text:        text,
		FontSize:    36,
		FontColor:   "white",
		X:           "(w-tw)/2",
		Y:           "h-th-50",
		BorderW:     2,
		BorderColor: "black",
		BoxEnable:   true,
		BoxColor:    "black@0.7",
	})
}

func (c *Command) WatermarkStyle(text string) *Command {
	return c.AdvancedTextOverlay(TextOverlayOptions{
		Text:      text,
		FontSize:  24,
		FontColor: "white@0.7",
		X:         "w-tw-10",
		Y:         "h-th-10",
		Alpha:     0.7,
	})
}

// USAGE EXAMPLES:

/*
// Basic usage:
cmd.SimpleTextOverlay("Hello World", 100, 50)

// With custom font:
cmd.TextOverlayWithFont("Custom Text", "/path/to/font.ttf", 48, 200, 100)

// Advanced text:
cmd.AdvancedTextOverlay(TextOverlayOptions{
    Text:       "Advanced Text",
    FontSize:   60,
    FontColor:  "yellow",
    X:          "center",
    Y:          "center",
    Alpha:      0.8,
    BoxEnable:  true,
    BoxColor:   "black@0.5",
    BorderW:    3,
    BorderColor: "red",
    StartTime:  "00:00:10",
    EndTime:    "00:00:20",
    FadeIn:     1.0,
    FadeOut:    1.0,
})

// Predefined positions:
cmd.TextOverlayTopLeft("Top Left Text", 32)
cmd.TextOverlayCenter("Centered Text", 48)
cmd.TextOverlayBottomRight("Bottom Right", 24)

// Predefined styles:
cmd.MovieTitleStyle("EPIC MOVIE")
cmd.SubtitleStyle("This is a subtitle")
cmd.WatermarkStyle("© 2024")

// Animated text:
cmd.ScrollingTextHorizontal("Scrolling text...", 36, 50.0)
cmd.TimestampOverlay("%%{pts\\:hms}", "top-right", 20)

// Multiple texts:
overlays := []TextOverlayOptions{
    {Text: "Title", FontSize: 60, X: "center", Y: "100"},
    {Text: "Subtitle", FontSize: 36, X: "center", Y: "200"},
}
cmd.MultipleTextOverlays(overlays)
*/
