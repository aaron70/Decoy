package decoy

import (
	"maps"
	"math/rand/v2"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aaron70/goaty/errors"
	"github.com/aaron70/goaty/options"
)

type DecoyOption func(*DecoyConfig) error

type DecoyConfig struct {
	Seed          uint64
	Ngram         int
	Ngrams        map[string][]string
	MaxTextLength int
}

type Decoy struct {
	seed   uint64
	random *rand.Rand
	randMu sync.Mutex

	corpusMu      sync.RWMutex
	ngram         int
	ngrams        map[string][]string
	keys          []string
	keyCumulative []int
	words         []string
	maxTextLength int
	defaultNgrams bool

	incrementalMu   sync.RWMutex
	intIncrementals map[string]*atomic.Int64
}

func WithSeed(seed uint64) DecoyOption {
	return func(d *DecoyConfig) error {
		d.Seed = seed
		return nil
	}
}

func WithNgram(n int) DecoyOption {
	return func(d *DecoyConfig) error {
		if n <= 0 {
			return errors.NewError(errors.ErrInvalidArgument, nil, "WithNgram: The ngram size (%d) must be greater than 0", n)
		}
		d.Ngram = n
		return nil
	}
}

func WithNgrams(ngrams map[string][]string) DecoyOption {
	return func(d *DecoyConfig) error {
		if len(ngrams) == 0 {
			return errors.NewError(errors.ErrInvalidArgument, nil, "WithNgrams: The ngrams map can't be empty")
		}
		d.Ngrams = ngrams
		return nil
	}
}

func WithMaxTextLength(length int) DecoyOption {
	return func(d *DecoyConfig) error {
		if length < 0 {
			return errors.NewError(errors.ErrInvalidArgument, nil, "WithMaxTextLength: The maximum text length can't be negative")
		}
		d.MaxTextLength = length
		return nil
	}
}

func NewDecoy(opts ...DecoyOption) (*Decoy, error) {
	config := DecoyConfig{Seed: uint64(time.Now().UnixNano()), Ngram: 4}
	if err := options.Apply(&config, opts...); err != nil {
		return nil, err
	}
	if config.Ngram <= 0 {
		config.Ngram = 4
	}

	decoy := &Decoy{
		seed:            config.Seed,
		random:          rand.New(rand.NewPCG(config.Seed, 0)),
		ngram:           config.Ngram,
		ngrams:          make(map[string][]string),
		maxTextLength:   config.MaxTextLength,
		intIncrementals: make(map[string]*atomic.Int64),
	}

	for key, successors := range config.Ngrams {
		keyWords := strings.Fields(key)
		if len(keyWords) != config.Ngram {
			return nil, errors.NewError(errors.ErrInvalidArgument, nil, "WithNgrams: The key %q must contain exactly %d words", key, config.Ngram)
		}
		if len(successors) == 0 {
			return nil, errors.NewError(errors.ErrInvalidArgument, nil, "WithNgrams: The key %q must have at least one successor", key)
		}
		for _, successor := range successors {
			successorWords := strings.Fields(successor)
			if len(successorWords) != 1 || successorWords[0] != successor {
				return nil, errors.NewError(errors.ErrInvalidArgument, nil, "WithNgrams: The successor %q for key %q is empty or contains whitespace", successor, key)
			}
		}
		canonicalKey := strings.Join(keyWords, " ")
		decoy.ngrams[canonicalKey] = append(decoy.ngrams[canonicalKey], slices.Clone(successors)...)
	}
	decoy.rebuildKeysLocked()

	return decoy, nil
}

func (d *Decoy) Seed() uint64 {
	return d.seed
}

func (d *Decoy) EnableDefaultNgrams() {
	d.corpusMu.Lock()
	defer d.corpusMu.Unlock()
	d.defaultNgrams = true
}

func (d *Decoy) ensureRandomLocked() {
	if d.random == nil {
		seed := d.seed
		if seed == 0 {
			seed = uint64(time.Now().UnixNano())
		}
		d.random = rand.New(rand.NewPCG(seed, 0))
	}
}

func (d *Decoy) ensureCorpusLocked() {
	if d.ngram <= 0 {
		d.ngram = 4
	}
	if d.ngrams == nil {
		d.ngrams = make(map[string][]string)
	}
}

func (d *Decoy) ensureIncrementalsLocked() {
	if d.intIncrementals == nil {
		d.intIncrementals = make(map[string]*atomic.Int64)
	}
}

func (d *Decoy) rebuildKeysLocked() {
	d.keys = slices.Sorted(maps.Keys(d.ngrams))
	d.keyCumulative = make([]int, len(d.keys))
	total := 0
	for index, key := range d.keys {
		total += len(d.ngrams[key])
		d.keyCumulative[index] = total
	}
}
