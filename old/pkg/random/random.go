package random

import "math/rand"

func RandBool() bool {
	return rand.Intn(2) == 0
}
