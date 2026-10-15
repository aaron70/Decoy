package decoy

import (
	"sort"
	"strconv"
	"sync"
	"testing"

	"github.com/aaron70/decoy/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestCurrentIncrementalInt(t *testing.T) {
	t.Run("unknown id returns default", func(t *testing.T) {
		d := newTestDecoy(t)
		assert.Equal(t, int64(42), d.CurrentIncrementalInt("missing", 42))
		assert.Equal(t, int64(0), d.CurrentIncrementalInt("missing", 0))
	})

	t.Run("reflects created counter", func(t *testing.T) {
		d := newTestDecoy(t)
		assert.Equal(t, int64(10), d.NextIncrementalInt("counter", 10, 5))
		assert.Equal(t, int64(10), d.CurrentIncrementalInt("counter", -1))
		assert.Equal(t, int64(15), d.NextIncrementalInt("counter", 10, 5))
		assert.Equal(t, int64(15), d.CurrentIncrementalInt("counter", -1))
	})

	t.Run("independent ids", func(t *testing.T) {
		d := newTestDecoy(t)
		d.NextIncrementalInt("a", 1, 1)
		d.NextIncrementalInt("b", 100, 10)
		assert.Equal(t, int64(1), d.CurrentIncrementalInt("a", 0))
		assert.Equal(t, int64(100), d.CurrentIncrementalInt("b", 0))
	})

	t.Run("concurrent reads and writes", func(t *testing.T) {
		d := newTestDecoy(t)
		testutil.RunConcurrently(testGoroutines, testIterations, func(g, _ int) {
			id := "shared-reader"
			if g%2 == 0 {
				_ = d.NextIncrementalInt(id, 0, 1)
			} else {
				_ = d.CurrentIncrementalInt(id, -1)
			}
		})
	})
}

func TestNextIncrementalInt(t *testing.T) {
	t.Run("first call returns start", func(t *testing.T) {
		d := newTestDecoy(t)
		assert.Equal(t, int64(100), d.NextIncrementalInt("seq", 100, 3))
	})

	t.Run("applies step on subsequent calls", func(t *testing.T) {
		d := newTestDecoy(t)
		assert.Equal(t, int64(0), d.NextIncrementalInt("seq", 0, 3))
		assert.Equal(t, int64(3), d.NextIncrementalInt("seq", 0, 3))
		assert.Equal(t, int64(6), d.NextIncrementalInt("seq", 0, 3))
	})

	t.Run("negative step", func(t *testing.T) {
		d := newTestDecoy(t)
		assert.Equal(t, int64(0), d.NextIncrementalInt("neg", 0, -5))
		assert.Equal(t, int64(-5), d.NextIncrementalInt("neg", 0, -5))
		assert.Equal(t, int64(-10), d.NextIncrementalInt("neg", 0, -5))
	})

	t.Run("zero step stays constant", func(t *testing.T) {
		d := newTestDecoy(t)
		for range 5 {
			assert.Equal(t, int64(7), d.NextIncrementalInt("zero", 7, 0))
		}
	})

	t.Run("independent ids", func(t *testing.T) {
		d := newTestDecoy(t)
		assert.Equal(t, int64(1), d.NextIncrementalInt("a", 1, 1))
		assert.Equal(t, int64(100), d.NextIncrementalInt("b", 100, 10))
		assert.Equal(t, int64(2), d.NextIncrementalInt("a", 1, 1))
		assert.Equal(t, int64(110), d.NextIncrementalInt("b", 100, 10))
	})

	t.Run("concurrent calls yield every value exactly once", func(t *testing.T) {
		d := newTestDecoy(t)
		total := testGoroutines * testIterations

		var mu sync.Mutex
		seen := make([]int64, 0, total)
		testutil.RunConcurrently(testGoroutines, testIterations, func(_, _ int) {
			v := d.NextIncrementalInt("race", 0, 1)
			mu.Lock()
			seen = append(seen, v)
			mu.Unlock()
		})

		sort.Slice(seen, func(i, j int) bool { return seen[i] < seen[j] })
		expected := make([]int64, total)
		for i := range total {
			expected[i] = int64(i)
		}
		assert.Equal(t, expected, seen)
		assert.Equal(t, int64(total-1), d.CurrentIncrementalInt("race", -1))
	})

	t.Run("concurrent creation across many ids", func(t *testing.T) {
		d := newTestDecoy(t)
		testutil.RunConcurrently(testGoroutines, testIterations, func(g, _ int) {
			d.NextIncrementalInt("id-"+strconv.Itoa(g), 0, 1)
		})
		for g := range testGoroutines {
			assert.Equal(t, int64(testIterations-1), d.CurrentIncrementalInt("id-"+strconv.Itoa(g), -1))
		}
	})
}
