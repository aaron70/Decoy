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
