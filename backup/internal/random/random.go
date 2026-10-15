package random

import "math/rand/v2"

func RandomChoice[T any](rand *rand.Rand, choices ...T) T {
	return choices[RandomInt(rand, 0, len(choices))]
}
