package decoy

import "math/rand/v2"

func (d *Decoy) RandomInt(limit1, limit2 int) int {
	min, max := limit1, limit2
	if min > max {
		min, max = limit2, limit1
	}
	if min == max {
		return min
	}
	r := d.randPool.Get().(*rand.Rand)
	defer d.randPool.Put(r)
	span := uint64(max) - uint64(min)
	return int(uint64(min) + r.Uint64N(span))
}

func (d *Decoy) RandomFloat(limit1, limit2 float64) float64 {
	min, max := limit1, limit2
	if min > max {
		min, max = limit2, limit1
	}
	if min == max {
		return min
	}
	r := d.randPool.Get().(*rand.Rand)
	defer d.randPool.Put(r)
	return r.Float64()*(max-min) + min
}

func (d *Decoy) RandomBoolean() bool {
	r := d.randPool.Get().(*rand.Rand)
	defer d.randPool.Put(r)
	return r.IntN(2) == 0
}

func (d *Decoy) Probablity(probability float64) bool {
	if probability <= 0 {
		return false
	}
	if probability >= 1 {
		return true
	}
	return d.RandomFloat(0, 1) < probability
}

func (d *Decoy) RandomChoice[C any](choices ...C) C {
	var zero C
	if len(choices) == 0 {
		return zero
	}
	return choices[d.RandomInt(0, len(choices))]
}

func (d *Decoy) RandomChoiceList[C any](choices []C) C {
	return d.RandomChoice(choices...)
}
