package interval

import "math/rand/v2"

type Interval struct {
	Min       float64
	Max       float64
	Center    float64
	Intensity float64 // Multiplier for the interval's effect
}

func New(min, max float64) Interval {
	return NewWithIntensity(min, max, 1.0)
}

func NewWithIntensity(min, max, intensity float64) Interval {
	if min > max {
		min, max = max, min
	}
	center := (min + max) / 2
	return Interval{
		Min:       min,
		Max:       max,
		Center:    center,
		Intensity: intensity,
	}
}

func (i Interval) Random() float64 {
	distance := (i.Max - i.Min) / 2
	adjustedDistance := distance * i.Intensity
	adjustedMin := i.Center - adjustedDistance
	adjustedMax := i.Center + adjustedDistance
	return adjustedMin + rand.Float64()*(adjustedMax-adjustedMin)
}

func (i Interval) RandomWithBias() float64 {
	distance := (i.Max - i.Min) / 2
	adjustedDistance := distance * i.Intensity
	adjustedMin := i.Center - adjustedDistance
	adjustedMax := i.Center + adjustedDistance

	t := rand.Float64()

	// U suave: mezcla entre uniforme y cuadrática
	// factor controla cuánto sesgo: 0.0 = uniforme, 1.0 = U fuerte
	factor := 0.3 // Ajusta entre 0.0 y 1.0

	var biasedT float64
	if t < 0.5 {
		biasedT = 2 * t * t
	} else {
		biasedT = 1 - 2*(1-t)*(1-t)
	}

	// Interpolar entre uniforme (t) y sesgado (biasedT)
	finalT := t*(1-factor) + biasedT*factor

	return adjustedMin + finalT*(adjustedMax-adjustedMin)
}
