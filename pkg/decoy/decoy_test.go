package decoy

import (
	"errors"
	"math/rand/v2"
	"runtime"
	"testing"

	"github.com/aaron70/decoy/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDecoy(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		d, err := NewDecoy()
		require.NoError(t, err)
		require.NotNil(t, d)

		assert.NotNil(t, d.intIncrementals, "counter map must be initialized")
		assert.NotNil(t, d.randPool.New, "random pool must have a constructor")
		assert.NotZero(t, d.randSeed)
	})

	t.Run("no options", func(t *testing.T) {
		d, err := NewDecoy()
		require.NoError(t, err)
		assert.NotNil(t, d)
	})

	t.Run("applies options in order", func(t *testing.T) {
		var order []string
		d, err := NewDecoy(
			func(*Decoy) error { order = append(order, "first"); return nil },
			func(*Decoy) error { order = append(order, "second"); return nil },
			func(*Decoy) error { order = append(order, "third"); return nil },
		)
		require.NoError(t, err)
		require.NotNil(t, d)
		assert.Equal(t, []string{"first", "second", "third"}, order)
	})

	t.Run("option mutates instance", func(t *testing.T) {
		d, err := NewDecoy(func(decoy *Decoy) error {
			decoy.randSeed = 777
			return nil
		})
		require.NoError(t, err)
		require.NotNil(t, d)
		assert.Equal(t, uint64(777), d.randSeed)
	})

	t.Run("option error aborts construction", func(t *testing.T) {
		wantErr := errors.New("option boom")
		d, err := NewDecoy(
			func(*Decoy) error { return nil },
			func(*Decoy) error { return wantErr },
			func(*Decoy) error { t.Fatal("must not run after error"); return nil },
		)
		require.ErrorIs(t, err, wantErr)
		assert.Nil(t, d)
	})
}

func TestWithSeed(t *testing.T) {
	t.Run("sets the seed", func(t *testing.T) {
		d, err := NewDecoy(WithSeed(7))
		require.NoError(t, err)
		assert.Equal(t, uint64(7), d.randSeed)
	})

	t.Run("overrides the default time seed", func(t *testing.T) {
		d, err := NewDecoy(WithSeed(0))
		require.NoError(t, err)
		assert.Zero(t, d.randSeed)
	})

	t.Run("same seed produces identical sequences", func(t *testing.T) {
		d1, err := NewDecoy(WithSeed(123))
		require.NoError(t, err)
		d2, err := NewDecoy(WithSeed(123))
		require.NoError(t, err)

		r1 := d1.randPool.New().(*rand.Rand)
		r2 := d2.randPool.New().(*rand.Rand)
		for range 1000 {
			require.Equal(t, r1.Int64(), r2.Int64())
		}
	})

	t.Run("different seeds produce different sequences", func(t *testing.T) {
		d1, err := NewDecoy(WithSeed(1))
		require.NoError(t, err)
		d2, err := NewDecoy(WithSeed(2))
		require.NoError(t, err)

		r1 := d1.randPool.New().(*rand.Rand)
		r2 := d2.randPool.New().(*rand.Rand)
		same := 0
		for range 1000 {
			if r1.Int64() == r2.Int64() {
				same++
			}
		}
		assert.Less(t, same, 1000, "different seeds must not be identical")
	})

	t.Run("resets the internal stream", func(t *testing.T) {
		d, err := NewDecoy(WithSeed(9))
		require.NoError(t, err)

		first := d.randPool.New().(*rand.Rand)
		second := d.randPool.New().(*rand.Rand)
		same := 0
		for range 1000 {
			if first.Int64() == second.Int64() {
				same++
			}
		}
		assert.Less(t, same, 1000, "each pool constructor call must use a new stream")
	})

	t.Run("first draw is reproducible through the public API", func(t *testing.T) {
		d1, err := NewDecoy(WithSeed(555))
		require.NoError(t, err)
		d2, err := NewDecoy(WithSeed(555))
		require.NoError(t, err)

		assert.Equal(t, d1.RandomInt(0, 1<<30), d2.RandomInt(0, 1<<30))
	})
}

// TestDecoyConcurrencySafety exercises a single Decoy instance from many
// goroutines at once, mimicking the intended shared-instance usage. It is meant
// to be run with the race detector enabled (go test -race).
func TestDecoyConcurrencySafety(t *testing.T) {
	d := newTestDecoy(t)

	const (
		goroutines  = 64
		iterations  = 200
		counterID   = "shared"
		counterStep = 1
	)
	total := goroutines * iterations

	baseline := runtime.NumGoroutine()

	testutil.RunConcurrently(goroutines, iterations, func(g, i int) {
		switch (g + i) % 9 {
		case 0:
			assert.GreaterOrEqual(t, d.RandomInt(-50, 50), -50)
		case 1:
			assert.Less(t, d.RandomFloat(0, 1), 1.0)
		case 2:
			_ = d.RandomBoolean()
		case 3:
			_ = d.Probablity(0.5)
		case 4:
			_ = d.NextIncrementalInt(counterID, 0, counterStep)
		case 5:
			_ = d.CurrentIncrementalInt(counterID, -1)
		case 6:
			assert.NotEmpty(t, d.RandomName())
		case 7:
			assert.NotEmpty(t, d.RandomLastName())
		case 8:
			assert.NotEmpty(t, d.RandomFullName(0.5))
		}
	})

	final := d.CurrentIncrementalInt(counterID, -1)
	assert.GreaterOrEqual(t, final, int64(0))
	assert.Less(t, final, int64(total))

	// Goroutine (leak) safety: no goroutine must be left permanently blocked.
	testutil.WaitForGoroutines(t, baseline+2)
	testutil.AssertNoGoroutineLeak(t)
}

func TestDecoyNoGoroutineLeak(t *testing.T) {
	d := newTestDecoy(t)
	baseline := runtime.NumGoroutine()

	for range 100 {
		d.RandomInt(0, 10)
		d.RandomFloat(-1, 1)
		d.RandomBoolean()
		d.Probablity(0.5)
		d.NextIncrementalInt("leak", 0, 1)
		d.CurrentIncrementalInt("leak", 0)
		d.RandomName()
		d.RandomLastName()
		d.RandomFullName(0.5)
	}

	testutil.WaitForGoroutines(t, baseline+2)
	testutil.AssertNoGoroutineLeak(t)
}
