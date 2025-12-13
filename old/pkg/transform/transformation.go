package transform

import (
	"math"
	"strings"
	"time"
)

// Transformation contains all the params for transforming a video
type Transformation struct {
	// Normalization
	Normalization NormalizationConfig `json:"normalization"`

	// Visual Transformations
	Visual VisualConfig `json:"visual"`

	// Temporal Variations
	Temporal TemporalConfig `json:"temporal"`

	// Color and Style
	ColorStyle ColorStyleConfig `json:"color_style"`

	// Overlays
	Overlays OverlayConfig `json:"overlays"`

	// Audio
	Audio AudioConfig `json:"audio"`

	// Noise and Texture
	NoiseTexture NoiseTextureConfig `json:"noise_texture"`

	// Structural
	Structural StructuralConfig `json:"structural"`
}

// NormalizationConfig for the initial phase
type NormalizationConfig struct {
	Enabled         bool                  `json:"enabled"`
	TargetWidth     int                   `json:"target_width"`
	TargetHeight    int                   `json:"target_height"`
	AspectRatio     string                `json:"aspect_ratio"` // "16:9", "9:16", "1:1", etc.
	ColorCorrection ColorCorrectionConfig `json:"color_correction"`
}

type ColorCorrectionConfig struct {
	Enabled    bool    `json:"enabled"`
	Brightness float64 `json:"brightness"` // -1.0 to 1.0
	Contrast   float64 `json:"contrast"`   // -1.0 to 1.0
	Gamma      float64 `json:"gamma"`      // 0.1 to 3.0
}

// VisualConfig for visual transformations
type VisualConfig struct {
	Scaling   ScalingConfig   `json:"scaling"`
	Rotation  RotationConfig  `json:"rotation"`
	Flipping  FlippingConfig  `json:"flipping"`
	Cropping  CroppingConfig  `json:"cropping"`
	Geometric GeometricConfig `json:"geometric"`
	Histogram HistogramConfig `json:"histogram"`
	Panning   PanningConfig   `json:"panning"`
}

type PanningConfig struct {
	Enabled bool `json:"enabled"`
	OffsetX int  `json:"offset_x"`
	OffsetY int  `json:"offset_y"`
}

type ScalingConfig struct {
	Enabled  bool    `json:"enabled"`
	Factor   float64 `json:"factor"`   // 0.1 to 3.0
	Adaptive bool    `json:"adaptive"` // Platform-adaptive scaling
	Platform string  `json:"platform"` // "tiktok", "instagram", "youtube"
}

type RotationConfig struct {
	Enabled  bool            `json:"enabled"`
	Angle    float64         `json:"angle"`  // Degrees
	Static   bool            `json:"static"` // true = static, false = animated
	Animated AnimationConfig `json:"animated"`
}

type AnimationConfig struct {
	Duration  time.Duration `json:"duration"`
	Oscillate bool          `json:"oscillate"`
	MinAngle  float64       `json:"min_angle"`
	MaxAngle  float64       `json:"max_angle"`
}

type FlippingConfig struct {
	Enabled    bool `json:"enabled"`
	Horizontal bool `json:"horizontal"`
	Vertical   bool `json:"vertical"`
}

type CroppingConfig struct {
	Enabled bool    `json:"enabled"`
	Top     float64 `json:"top"` // Percentage 0-1
	Bottom  float64 `json:"bottom"`
	Left    float64 `json:"left"`
	Right   float64 `json:"right"`
}

type GeometricConfig struct {
	Enabled          bool    `json:"enabled"`
	LensDistortionK1 float64 `json:"lens_distortion_k1"`
	LensDistortionK2 float64 `json:"lens_distortion_k2"`
	PerspectiveX0    float64 `json:"perspective_x0"` // top-left x
	PerspectiveX1    float64 `json:"perspective_x1"` // top-left x
	PerspectiveX2    float64 `json:"perspective_x2"` // top-left x
	PerspectiveX3    float64 `json:"perspective_x3"` // top-left x
	PerspectiveY0    float64 `json:"perspective_y0"` // top-left y
	PerspectiveY1    float64 `json:"perspective_y1"` // top-left y
	PerspectiveY2    float64 `json:"perspective_y2"` // top-left y
	PerspectiveY3    float64 `json:"perspective_y3"` // top-left y
}

type HistogramConfig struct {
	Enabled    bool    `json:"enabled"`
	RedShift   float64 `json:"red_shift"`
	GreenShift float64 `json:"green_shift"`
	BlueShift  float64 `json:"blue_shift"`
}

// TemporalConfig for temporal variations
type TemporalConfig struct {
	Enabled        bool                 `json:"enabled"`
	FrameReorder   FrameReorderConfig   `json:"frame_reorder"`
	FrameSkip      FrameSkipConfig      `json:"frame_skip"`
	FrameDuplicate FrameDuplicateConfig `json:"frame_duplicate"`
	SpeedVariation SpeedVariationConfig `json:"speed_variation"`
}

type FrameReorderConfig struct {
	Enabled bool  `json:"enabled"`
	Pattern []int `json:"pattern"` // Reordering pattern
	Random  bool  `json:"random"`
}

type FrameSkipConfig struct {
	Enabled  bool `json:"enabled"`
	Interval int  `json:"interval"` // Skip every N frames
}

type FrameDuplicateConfig struct {
	Enabled bool `json:"enabled"`
	Factor  int  `json:"factor"` // Duplicate every N frame X times
}

type SpeedVariationConfig struct {
	Enabled    bool    `json:"enabled"`
	MinSpeed   float64 `json:"min_speed"` // 0.1 to 2.0
	MaxSpeed   float64 `json:"max_speed"`
	Sinusoidal bool    `json:"sinusoidal"`
	Random     bool    `json:"random"`
}

// ColorStyleConfig for color and style alterations
type ColorStyleConfig struct {
	Brightness  BrightnessConfig  `json:"brightness"`
	Contrast    ContrastConfig    `json:"contrast"`
	Saturation  SaturationConfig  `json:"saturation"`
	ColorMixing ColorMixingConfig `json:"color_mixing"`
	LUT         LUTConfig         `json:"lut"`
	Effects     EffectsConfig     `json:"effects"`
}

type BrightnessConfig struct {
	Enabled bool    `json:"enabled"`
	Value   float64 `json:"value"` // -1.0 to 1.0
}

type ContrastConfig struct {
	Enabled bool    `json:"enabled"`
	Value   float64 `json:"value"` // -1.0 to 1.0
}

type SaturationConfig struct {
	Enabled bool    `json:"enabled"`
	Value   float64 `json:"value"` // 0.0 to 2.0
}

type ColorMixingConfig struct {
	Enabled    bool    `json:"enabled"`
	RedShift   float64 `json:"red_shift"` // -1.0 to 1.0
	GreenShift float64 `json:"green_shift"`
	BlueShift  float64 `json:"blue_shift"`
}

type LUTConfig struct {
	Enabled  bool   `json:"enabled"`
	FilePath string `json:"file_path"`
	Style    string `json:"style"` // "cinematic", "vintage", "warm", etc.
}

type EffectsConfig struct {
	Vintage          bool    `json:"vintage"`
	Sepia            bool    `json:"sepia"`
	BlackWhite       bool    `json:"black_white"`
	VintageIntensity float64 `json:"vintage_intensity"`
}

// OverlayConfig for overlay elements
type OverlayConfig struct {
	Enabled      bool             `json:"enabled"`
	TextOverlays TextOverlay      `json:"text_overlays"`
	Graphics     []GraphicOverlay `json:"graphics"`
	GIFs         []GIFOverlay     `json:"gifs"`
	PiP          PictureInPicture `json:"pip"`
}

type TextOverlay struct {
	Text            string `json:"text"`
	X               string `json:"x"` // Position percentage 0-1
	Y               string `json:"y"`
	Color           string `json:"color"` // Hex color
	baseFontSize    float64
	baseLineHeight  float64
	MaxWidth        float64
	approxCharWidth float64
}

func (t TextOverlay) FontSize() float64 {

	const fontReductionFactor = 0.09

	estimatedCharsPerLine := int(math.Floor(float64(t.MaxWidth) / float64(t.approxCharWidth)))
	visualLineCount := int(math.Ceil(float64(len(t.Text)) / float64(estimatedCharsPerLine)))
	actualLineCount := len(strings.Split(t.Text, "\n"))
	lineCount := int(math.Max(float64(visualLineCount), float64(actualLineCount)))

	adjustedFontSize := int(math.Round(t.baseFontSize * (1 - fontReductionFactor*float64(lineCount-1))))
	fontSize := int(math.Max(float64(adjustedFontSize), 40))
	return float64(fontSize)
}

func (t TextOverlay) LineHeight() float64 {
	const lineReductionFactor = 0.07

	resizedCharWidth := t.approxCharWidth * (t.FontSize() / t.baseFontSize)
	estimatedCharsPerLine := int(math.Floor(float64(t.MaxWidth) / float64(resizedCharWidth)))
	visualLineCount := int(math.Ceil(float64(len(t.Text)) / float64(estimatedCharsPerLine)))
	actualLineCount := len(strings.Split(t.Text, "\n"))
	lineCount := int(math.Max(float64(visualLineCount), float64(actualLineCount)))

	adjustedLineSize := t.baseLineHeight * (1 - lineReductionFactor*float64(lineCount-1))
	lineHeight := math.Max(adjustedLineSize, 0.7)
	return lineHeight
}

// GetTextSegments splits text into lines based on \n delimiters and maximum width,
// intelligently wrapping long lines at word boundaries when possible.
// Returns text with REAL newline characters (\n) not escaped sequences.
func (t TextOverlay) GetTextSegments() string {
	// Calculate scaled character width based on font size
	resizedCharWidth := t.approxCharWidth * (t.FontSize() / t.baseFontSize)
	maxCharsPerLine := int(math.Floor(float64(t.MaxWidth) / float64(resizedCharWidth)))

	// Ensure minimum of 1 character per line
	if maxCharsPerLine < 1 {
		maxCharsPerLine = 1
	}

	var lines []string
	segments := strings.Split(t.Text, "\n")

	for _, segment := range segments {
		trimmed := strings.TrimSpace(segment)

		// Preserve empty lines to maintain intentional spacing
		if len(trimmed) == 0 {
			lines = append(lines, "")
			continue
		}

		// Wrap segment if it exceeds maximum width
		wrapped := wrapLine(trimmed, maxCharsPerLine)
		lines = append(lines, wrapped...)
	}

	// Join with ACTUAL newline character, not escaped
	return strings.Join(lines, "\n")
}

// wrapLine breaks a long line into multiple lines at word boundaries.
func wrapLine(text string, maxChars int) []string {
	// Text fits in one line
	if len(text) <= maxChars {
		return []string{text}
	}

	var result []string
	remaining := text

	for len(remaining) > maxChars {
		splitPoint := maxChars

		// Look for the last space within the character limit
		// Use maxChars+1 to include the character at maxChars position, but do not exceed string length
		searchLimit := maxChars
		if searchLimit >= len(remaining) {
			searchLimit = len(remaining) - 1
		}

		if idx := strings.LastIndex(remaining[:searchLimit+1], " "); idx > 0 {
			splitPoint = idx
		}

		// Extract and trim the line
		line := strings.TrimSpace(remaining[:splitPoint])
		if line != "" {
			result = append(result, line)
		}

		// Continue with remaining text
		remaining = strings.TrimSpace(remaining[splitPoint:])
	}

	// Add final remaining text if any
	if remaining != "" {
		result = append(result, remaining)
	}

	return result
}

type GraphicOverlay struct {
	FilePath  string        `json:"file_path"`
	X         float64       `json:"x"`
	Y         float64       `json:"y"`
	Scale     float64       `json:"scale"`
	Duration  time.Duration `json:"duration"`
	StartTime time.Duration `json:"start_time"`
}

type GIFOverlay struct {
	FilePath  string        `json:"file_path"`
	X         float64       `json:"x"`
	Y         float64       `json:"y"`
	Scale     float64       `json:"scale"`
	Loop      bool          `json:"loop"`
	Duration  time.Duration `json:"duration"`
	StartTime time.Duration `json:"start_time"`
}

type PictureInPicture struct {
	Enabled   bool          `json:"enabled"`
	VideoPath string        `json:"video_path"`
	X         float64       `json:"x"`
	Y         float64       `json:"y"`
	Width     float64       `json:"width"` // Percentage of main video
	Height    float64       `json:"height"`
	StartTime time.Duration `json:"start_time"`
	Duration  time.Duration `json:"duration"`
}

// AudioConfig for audio manipulation
type AudioConfig struct {
	Enabled     bool             `json:"enabled"`
	Replacement AudioReplacement `json:"replacement"`
	Pitch       PitchConfig      `json:"pitch"`
	Speed       AudioSpeedConfig `json:"speed"`
	Noise       AudioNoiseConfig `json:"noise"`
	Muting      MutingConfig     `json:"muting"`
}

type AudioReplacement struct {
	Enabled  bool    `json:"enabled"`
	FilePath string  `json:"file_path"`
	Volume   float64 `json:"volume"` // 0.0 to 2.0
}

type PitchConfig struct {
	Enabled   bool    `json:"enabled"`
	Semitones float64 `json:"semitones"` // -12 to 12
}

type AudioSpeedConfig struct {
	Enabled bool    `json:"enabled"`
	Factor  float64 `json:"factor"` // 0.5 to 2.0
}

type AudioNoiseConfig struct {
	Enabled    bool    `json:"enabled"`
	WhiteNoise bool    `json:"white_noise"`
	Volume     float64 `json:"volume"` // 0.0 to 1.0
}

type MutingConfig struct {
	Enabled   bool          `json:"enabled"`
	Full      bool          `json:"full"`
	StartTime time.Duration `json:"start_time"`
	Duration  time.Duration `json:"duration"`
}

// NoiseTextureConfig for noise and texture effects
type NoiseTextureConfig struct {
	Enabled          bool            `json:"enabled"`
	LuminanceNoise   NoiseConfig     `json:"luminance_noise"`
	ChrominanceNoise NoiseConfig     `json:"chrominance_noise"`
	FilmGrain        FilmGrainConfig `json:"film_grain"`
	Patterns         PatternConfig   `json:"patterns"`
}

type NoiseConfig struct {
	Enabled   bool    `json:"enabled"`
	Intensity float64 `json:"intensity"` // 0.0 to 1.0
}

type FilmGrainConfig struct {
	Enabled   bool    `json:"enabled"`
	Intensity float64 `json:"intensity"`
	Size      float64 `json:"size"` // Grain size
}

type PatternConfig struct {
	Enabled   bool    `json:"enabled"`
	Type      string  `json:"type"`      // "lines", "dots", "grid"
	Intensity float64 `json:"intensity"` // Nearly invisible
	Frequency float64 `json:"frequency"`
}

// StructuralConfig for structural file alterations
type StructuralConfig struct {
	Enabled     bool              `json:"enabled"`
	Metadata    MetadataConfig    `json:"metadata"`
	Container   ContainerConfig   `json:"container"`
	Fingerprint FingerprintConfig `json:"fingerprint"`
}

type MetadataConfig struct {
	Enabled      bool              `json:"enabled"`
	RemoveEXIF   bool              `json:"remove_exif"`
	ModifyDate   bool              `json:"modify_date"`
	NewDate      time.Time         `json:"new_date"`
	DeviceInfo   map[string]string `json:"device_info"`
	SoftwareInfo map[string]string `json:"software_info"`
}

type ContainerConfig struct {
	Enabled       bool `json:"enabled"`
	ModifyHeaders bool `json:"modify_headers"`
	ModifyAtoms   bool `json:"modify_atoms"` // For MP4
	ReorderAtoms  bool `json:"reorder_atoms"`
}

type FingerprintConfig struct {
	Enabled           bool `json:"enabled"`
	SimulateCamera    bool `json:"simulate_camera"`
	RemoveEditTraces  bool `json:"remove_edit_traces"`
	AddCameraMetadata bool `json:"add_camera_metadata"`
}
