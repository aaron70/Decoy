package random

import "math/rand/v2"

func RandomInt(rand *rand.Rand, min, max int) int {
	return rand.IntN(max-min) + min
}
