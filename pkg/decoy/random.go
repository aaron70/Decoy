package decoy

import (
	"fmt"
	"math"

	"github.com/aaron70/goaty/errors"
	"github.com/aaron70/goaty/utils"
)

func (d *Decoy) RandomInt(min, max int) (int, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	return d.randomInt(min, max)
}

func (d *Decoy) randomInt(min, max int) (int, error) {
	if min >= max {
		return 0, fmt.Errorf("%w: RandomInt: Invalid limits [%d, %d). max (%d) must be greater than min (%d)", errors.ErrInvalidArgument, min, max, max, min)
	}
	return int(uint64(min) + d.random.Uint64N(uint64(max)-uint64(min))), nil
}

func (d *Decoy) RandomFloat(min, max float64) (float64, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	return d.randomFloat(min, max)
}

func (d *Decoy) randomFloat(min, max float64) (float64, error) {
	if math.IsNaN(min) || math.IsNaN(max) || math.IsInf(min, 0) || math.IsInf(max, 0) {
		return 0.0, fmt.Errorf("%w: RandomFloat: Invalid limits [%v, %v). Non-finite range ", ErrRangeTooLarge, min, max)
	}
	if min >= max {
		return 0.0, fmt.Errorf("%w: RandomFloat: Invalid limits [%v, %v). max (%v) must be greater than min (%v)", errors.ErrInvalidArgument, min, max, max, min)
	}

	span := max - min
	if math.IsInf(span, 0) {
		return 0.0, fmt.Errorf("%w: RandomFloat: Invalid limits [%v, %v). Range is too large", ErrRangeTooLarge, min, max)
	}

	result := d.random.Float64()*span + min
	if result >= max {
		result = math.Nextafter(max, min)
	}
	return result, nil
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

func (d *Decoy) RandomChoice[C any](choices ...C) (C, error) {
	var zero C
	if len(choices) == 0 {
		return zero, fmt.Errorf("%w: RandomChoice: Requires at least one choice, got no choices", errors.ErrInvalidArgument)
	}

	d.mutex.Lock()
	defer d.mutex.Unlock()
	return choices[utils.Must(d.randomInt(0, len(choices)))], nil
}
