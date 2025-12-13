package transform

type ValueRange struct {
	Min float64
	Max float64
}

type Parameters struct {
	IntensityLevel    IntensityLevel
	Scale             ValueRange
	Rotate            ValueRange
	OffsetX           ValueRange
	OffsetY           ValueRange
	Brightness        ValueRange
	Contrast          ValueRange
	Gamma             ValueRange
	LennsDistortionK1 ValueRange
	LennsDistortionK2 ValueRange
	PerspectiveShift  ValueRange
}

func DefaultParameters() Parameters {
	return Parameters{
		Scale:             ValueRange{Min: 1.1, Max: 1.25},     // Scale range for video resizing
		Rotate:            ValueRange{Min: -5, Max: 5},         // Rotation range in degrees
		OffsetX:           ValueRange{Min: -0.15, Max: 0.15},   // Horizontal offset percentage of video width
		OffsetY:           ValueRange{Min: -0.5, Max: 0.25},    // Vertical offset percentage of video height
		Brightness:        ValueRange{Min: -0.07, Max: 0.07},   // Brightness adjustment range
		Contrast:          ValueRange{Min: 0.92, Max: 1.08},    // Contrast adjustment range
		Gamma:             ValueRange{Min: 0.8, Max: 1.3},      // Gamma adjustment range
		LennsDistortionK1: ValueRange{Min: -0.07, Max: 0.07},   // Lens distortion coefficient K1
		LennsDistortionK2: ValueRange{Min: -0.018, Max: 0.018}, // Lens distortion coefficient K2
		PerspectiveShift:  ValueRange{Min: -0.035, Max: 0.035}, // Perspective shift range
	}
}
