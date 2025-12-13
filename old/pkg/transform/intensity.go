package transform

type IntensityLevel float64

const (
	IntensityLow  IntensityLevel = 0.5
	IntensityBase IntensityLevel = 1.0
	IntensityHigh IntensityLevel = 1.5
	IntensityMax  IntensityLevel = 2.0
)

func (i IntensityLevel) IsValid() bool {
	switch i {
	case IntensityLow, IntensityBase, IntensityHigh, IntensityMax:
		return true
	default:
		return false
	}
}

func (i IntensityLevel) Float64() float64 {
	return float64(i)
}
