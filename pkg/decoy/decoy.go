package decoy

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/aaron70/goaty/options"
)

type DecoyOption func(*Decoy) error

type Decoy struct {
	Seed      uint64
	randSeed2 atomic.Uint64
	mutex     sync.RWMutex
}

func WithSeed(seed uint64) DecoyOption {
	return func(d *Decoy) error {
		d.Seed = seed
		d.randSeed2.Store(0)
		return nil
	}
}

func NewDecoy(opts ...DecoyOption) (*Decoy, error) {
	seed := uint64(time.Now().UnixNano())
	decoy := &Decoy{
		Seed: seed,
	}

	if err := options.Apply(decoy, opts...); err != nil {
		return nil, err
	}

	return decoy, nil
}
