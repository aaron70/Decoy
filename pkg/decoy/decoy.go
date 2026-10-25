package decoy

import (
	"math/rand/v2"
	"sync"
	"time"

	"github.com/aaron70/goaty/options"
)

type DecoyOption func(*Decoy) error

type Decoy struct {
	seed   uint64
	random *rand.Rand
	mutex  sync.RWMutex
}

func WithSeed(seed uint64) DecoyOption {
	return func(d *Decoy) error {
		d.seed = seed
		d.random = rand.New(rand.NewPCG(seed, 0))
		return nil
	}
}

func NewDecoy(opts ...DecoyOption) (*Decoy, error) {
	seed := uint64(time.Now().UnixNano())
	decoy := &Decoy{
		seed:   seed,
		random: rand.New(rand.NewPCG(seed, 0)),
	}

	if err := options.Apply(decoy, opts...); err != nil {
		return nil, err
	}

	return decoy, nil
}
