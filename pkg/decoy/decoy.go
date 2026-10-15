package decoy

import (
	"math/rand/v2"
	"sync"
	"sync/atomic"
	"time"
)

type DecoyOption func(*Decoy) error

type Decoy struct {
	randSeed        uint64
	randStream      atomic.Uint64
	randPool        sync.Pool
	intIncrementals map[string]*atomic.Int64
	mutex           sync.RWMutex
}

func WithSeed(seed uint64) DecoyOption {
	return func(d *Decoy) error {
		d.randSeed = seed
		d.randStream.Store(0)
		d.randPool.New = func() any {
			return rand.New(rand.NewPCG(seed, d.randStream.Add(1)))
		}
		return nil
	}
}

func NewDecoy(options ...DecoyOption) (*Decoy, error) {
	seed := uint64(time.Now().UnixNano())
	decoy := &Decoy{
		randSeed:        seed,
		intIncrementals: make(map[string]*atomic.Int64),
	}
	decoy.randPool.New = func() any {
		return rand.New(rand.NewPCG(seed, decoy.randStream.Add(1)))
	}

	for _, option := range options {
		if err := option(decoy); err != nil {
			return nil, err
		}
	}

	return decoy, nil
}

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

func (d *Decoy) CurrentIncrementalInt(id string, defaultValue int64) int64 {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	value, exists := d.intIncrementals[id]
	if !exists {
		return defaultValue
	}
	return value.Load()
}

func (d *Decoy) NextIncrementalInt(id string, start, step int64) int64 {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	value, exists := d.intIncrementals[id]
	if !exists {
		a := &atomic.Int64{}
		a.Store(start)
		d.intIncrementals[id] = a
		return start
	}
	return value.Add(step)
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
