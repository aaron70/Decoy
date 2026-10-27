package decoy

import (
	"math/rand/v2"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aaron70/goaty/options"
)

type DecoyOption func(*DecoyConfig) error

type DecoyConfig struct {
	Seed   uint64
	Ngram  int
	Ngrams map[string][]string
}

type Decoy struct {
	seed   uint64
	random *rand.Rand
	mutex  sync.RWMutex

	// Random Text Generation
	ngram  int
	ngrams map[string][]string
	keys   []string


	// Incrementals
	intIncrementals map[string]*atomic.Int64
}

func WithSeed(seed uint64) DecoyOption {
	return func(d *DecoyConfig) error {
		d.Seed = seed
		return nil
	}
}

func NewDecoy(opts ...DecoyOption) (*Decoy, error) {
	config := DecoyConfig{Seed: uint64(time.Now().UnixNano())}
	if err := options.Apply(&config, opts...); err != nil {
		return nil, err
	}

	return &Decoy{
		seed:   config.Seed,
		random: rand.New(rand.NewPCG(config.Seed, 0)),
		intIncrementals: make(map[string]*atomic.Int64),
		ngrams: make(map[string][]string),
		ngram:  4,
	}, nil
}
