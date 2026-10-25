package decoy

import (
	"fmt"
	"math"
)

func (d *Decoy) RandomInt(limit1, limit2 int) int {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	return d.randomInt(limit1, limit2)
}

func (d *Decoy) randomInt(limit1, limit2 int) int {
	min, max := limit1, limit2
	if min > max {
		min, max = max, min
	}
	if min == max {
		panic(fmt.Sprintf("decoy: empty range [%d, %d)", min, max))
	}
	span := int64(max) - int64(min)
	if span <= 0 {
		panic(fmt.Sprintf("decoy: range [%d, %d) is too large", min, max))
	}
	return int(int64(min) + d.random.Int64N(span))
}

func (d *Decoy) RandomFloat(limit1, limit2 float64) float64 {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	return d.randomFloat(limit1, limit2)
}

func (d *Decoy) randomFloat(limit1, limit2 float64) float64 {
	if math.IsNaN(limit1) || math.IsNaN(limit2) || math.IsInf(limit1, 0) || math.IsInf(limit2, 0) {
		panic(fmt.Sprintf("decoy: non-finite range [%v, %v)", limit1, limit2))
	}

	min, max := limit1, limit2
	if min > max {
		min, max = max, min
	}
	if min == max {
		return min
	}
	span := max - min
	if math.IsInf(span, 0) {
		panic(fmt.Sprintf("decoy: range [%v, %v) is too large", min, max))
	}
	return d.random.Float64()*span + min
}

func (d *Decoy) RandomBoolean() bool {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	return d.randomBool()
}

func (d *Decoy) randomBool() bool {
	return d.random.IntN(2) == 0
}

func (d *Decoy) Probability(probability float64) bool {
	if probability <= 0 {
		return false
	}
	if probability >= 1 {
		return true
	}

	d.mutex.Lock()
	defer d.mutex.Unlock()
	return d.random.Float64() < probability
}

func (d *Decoy) RandomChoice[C any](choices ...C) C {
	if len(choices) == 0 {
		panic("decoy: RandomChoice requires at least one choice")
	}

	d.mutex.Lock()
	defer d.mutex.Unlock()
	return choices[d.randomInt(0, len(choices))]
}
