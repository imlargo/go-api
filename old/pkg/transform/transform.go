package transform

import (
	"math/rand"
	"strings"
	"time"

	"github.com/nicolailuther/butter/pkg/interval"
	"github.com/nicolailuther/butter/pkg/transform/generators"
)

const (
	// fontSizeRatio is the font size scaling ratio: base font size (56px) / base video height (1280px) = 0.04375
	// This maintains text at 4.375% of video height across all resolutions
	fontSizeRatio = 0.04375

	// charWidthRatio is the character width ratio: base char width (35px) / base font size (56px) = 0.625
	// This maintains proper character width proportional to font size
	charWidthRatio = 0.625
)

func NewRandomTransformationConfig(params *Parameters, options TransformationOptions) *Transformation {
	if params.IntensityLevel == 0 {
		params.IntensityLevel = IntensityBase
	}

	// Normalization
	brightnessCorrection := interval.NewWithIntensity(params.Brightness.Min, params.Brightness.Max, params.IntensityLevel.Float64())
	contrastCorrection := interval.NewWithIntensity(params.Contrast.Min, params.Contrast.Max, params.IntensityLevel.Float64())
	gammaCorrection := interval.NewWithIntensity(params.Gamma.Min, params.Gamma.Max, params.IntensityLevel.Float64())

	// Visual
	rotationAngle := interval.NewWithIntensity(params.Rotate.Min, params.Rotate.Max, params.IntensityLevel.Float64()).RandomWithBias()

	// Adjust offset X according on angle rotation
	//offsetXFactor := interval.NewWithIntensity(params.OffsetX.Min, params.OffsetX.Max, params.IntensityLevel.Float64()).RandomWithBias()
	//offsetYFactor := interval.NewWithIntensity(params.OffsetY.Min, params.OffsetY.Max, params.IntensityLevel.Float64()).RandomWithBias()

	/*
		offsetX := int(offsetXFactor * float64(videoInfo.Width))
		offsetY := int(offsetYFactor * float64(videoInfo.Height))

		// if angle is positive then it will rotate clockwise so try to not make black parts
		if rotationAngle > 0 && offsetX > 0 {
			offsetX = -offsetX
		}

		if rotationAngle < 0 && offsetX < 0 {
			offsetX = -offsetX
		}
	*/

	// Lens distortion
	k1 := interval.NewWithIntensity(params.LennsDistortionK1.Min, params.LennsDistortionK1.Max, params.IntensityLevel.Float64()).RandomWithBias()
	k2 := interval.NewWithIntensity(params.LennsDistortionK2.Min, params.LennsDistortionK2.Max, params.IntensityLevel.Float64()).RandomWithBias()

	// Subtle perspective shift
	perspectiveShiftRange := interval.NewWithIntensity(params.PerspectiveShift.Min, params.PerspectiveShift.Max, params.IntensityLevel.Float64())

	metadataProfile := generators.GenerateCompleteDeviceProfile()

	// Calculate scaled font size based on video height
	scaledFontSize := float64(options.VideoInfo.Height) * fontSizeRatio

	return &Transformation{
		Normalization: NormalizationConfig{
			Enabled:      true,
			TargetWidth:  0,
			TargetHeight: 0,
			AspectRatio:  "",
			ColorCorrection: ColorCorrectionConfig{
				Enabled:    !options.IsMain && true,
				Brightness: brightnessCorrection.RandomWithBias(),
				Contrast:   contrastCorrection.RandomWithBias(),
				Gamma:      gammaCorrection.RandomWithBias(),
			},
		},
		Visual: VisualConfig{
			Scaling: ScalingConfig{
				Enabled:  !options.IsMain && true,
				Factor:   0, //scaleRange.RandomWithBias(),
				Adaptive: false,
				Platform: "",
			},
			Panning: PanningConfig{
				Enabled: !options.IsMain && true,
				OffsetX: 0, //offsetX,
				OffsetY: 0, //offsetY,
			},
			Rotation: RotationConfig{
				Enabled: !options.IsMain && true,
				Angle:   rotationAngle,
				Static:  true,
			},
			Flipping: FlippingConfig{
				Enabled:    !options.IsMain && true,
				Horizontal: options.UseMirror,
				Vertical:   false,
			},
			Cropping: CroppingConfig{},
			Geometric: GeometricConfig{
				Enabled:          !options.IsMain && true,
				LensDistortionK1: k1,
				LensDistortionK2: k2,
				PerspectiveX1:    perspectiveShiftRange.RandomWithBias(),
				PerspectiveX2:    perspectiveShiftRange.RandomWithBias(),
				PerspectiveX3:    perspectiveShiftRange.RandomWithBias(),
				PerspectiveY0:    perspectiveShiftRange.RandomWithBias(),
				PerspectiveY1:    perspectiveShiftRange.RandomWithBias(),
				PerspectiveY2:    perspectiveShiftRange.RandomWithBias(),
				PerspectiveY3:    perspectiveShiftRange.RandomWithBias(),
			},
			Histogram: HistogramConfig{},
		},
		Temporal: TemporalConfig{},
		ColorStyle: ColorStyleConfig{
			Brightness: BrightnessConfig{
				Enabled: !options.IsMain && true,
				Value:   0.5,
			},
			Contrast: ContrastConfig{
				Enabled: !options.IsMain && true,
				Value:   0.5,
			},
			Saturation: SaturationConfig{
				Enabled: !options.IsMain && true,
				Value:   0.5,
			},
			ColorMixing: ColorMixingConfig{
				Enabled:    !options.IsMain && false,
				RedShift:   0,
				GreenShift: 0,
				BlueShift:  0,
			},
			LUT:     LUTConfig{},
			Effects: EffectsConfig{},
		},
		Overlays: OverlayConfig{
			TextOverlays: TextOverlay{
				Text:            strings.TrimSpace(options.TextOverlay),
				baseFontSize:    scaledFontSize,
				baseLineHeight:  0.56,
				MaxWidth:        float64(options.VideoInfo.Width) * 0.87,
				approxCharWidth: scaledFontSize * charWidthRatio,
			},
		},
		Audio:        AudioConfig{},
		NoiseTexture: NoiseTextureConfig{},
		Structural: StructuralConfig{
			Enabled: true,
			Metadata: MetadataConfig{
				Enabled:      true,
				RemoveEXIF:   false, // Mejor selectivo que todo
				ModifyDate:   true,
				NewDate:      time.Now().AddDate(0, 0, -rand.Intn(30)), // 0-30 días atrás
				DeviceInfo:   metadataProfile.DeviceInfo,
				SoftwareInfo: metadataProfile.SoftwareInfo,
			},
			Container: ContainerConfig{
				Enabled:       true,
				ModifyHeaders: true,
				ModifyAtoms:   true,
				ReorderAtoms:  true,
			},
			Fingerprint: FingerprintConfig{
				Enabled:           true,
				SimulateCamera:    true,
				RemoveEditTraces:  true,
				AddCameraMetadata: true,
			},
		},
	}
}
