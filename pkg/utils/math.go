package utils

func CalculatePercentage(value float64, total float64) float64 {
	return DivideOrZero(value, total) * 100
}

func DivideOrZero(a float64, b float64) float64 {
	if b == 0 {
		return 0
	}
	return a / b
}
