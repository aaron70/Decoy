package decoy

import (
	"bytes"
	"errors"
	"math"
	"regexp"
	"runtime"
	"runtime/pprof"
	"sort"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testGoroutines = 32
	testIterations = 100
)

func newTestDecoy(t *testing.T) *Decoy {
	t.Helper()
	d, err := NewDecoy(WithSeed(42))
	require.NoError(t, err)
	require.NotNil(t, d)
	return d
}

// runConcurrently runs fn from goroutines*iterations goroutines that are all
// released at the same time, maximizing the chance of contention.
func runConcurrently(t *testing.T, goroutines, iterations int, fn func(g, i int)) {
	t.Helper()
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := range goroutines {
		go func(g int) {
			defer wg.Done()
			<-start
			for i := range iterations {
				fn(g, i)
			}
		}(g)
	}
	close(start)
	wg.Wait()
}

var goroutineLeakTotalRe = regexp.MustCompile(`(?m)^goroutineleak profile: total (\d+)$`)

// assertNoGoroutineLeak uses the Go 1.27 goroutine leak profiler. It triggers a
// leak-detecting GC cycle and asserts that no goroutine is permanently blocked
// on an unreachable synchronization primitive.
func assertNoGoroutineLeak(t *testing.T) {
	t.Helper()
	profile := pprof.Lookup("goroutineleak")
	if profile == nil {
		t.Skip("goroutineleak profile is not available (requires Go 1.27+)")
	}

	var buf bytes.Buffer
	require.NoError(t, profile.WriteTo(&buf, 1))

	match := goroutineLeakTotalRe.FindStringSubmatch(buf.String())
	require.NotNilf(t, match, "unexpected goroutineleak profile format:\n%s", buf.String())

	total, err := strconv.Atoi(match[1])
	require.NoError(t, err)
	assert.Zerof(t, total, "leaked goroutines detected:\n%s", buf.String())
}

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

func TestRandomInt(t *testing.T) {
	d := newTestDecoy(t)

	t.Run("within bounds", func(t *testing.T) {
		cases := []struct{ a, b int }{
			{0, 1},
			{0, 10},
			{10, 0},
			{1, 2},
			{1, 3},
			{-5, 5},
			{-10, -1},
			{-1, -10},
			{100, 1000},
		}
		for _, c := range cases {
			lo, hi := min(c.a, c.b), max(c.a, c.b)
			for range 1000 {
				got := d.RandomInt(c.a, c.b)
				require.GreaterOrEqualf(t, got, lo, "RandomInt(%d, %d)=%d", c.a, c.b, got)
				require.Lessf(t, got, hi, "RandomInt(%d, %d)=%d", c.a, c.b, got)
			}
		}
	})

	t.Run("equal limits do not panic and return the limit", func(t *testing.T) {
		assert.Equal(t, 5, d.RandomInt(5, 5))
		assert.Equal(t, -3, d.RandomInt(-3, -3))
		assert.Equal(t, 0, d.RandomInt(0, 0))
	})

	t.Run("single value range is constant", func(t *testing.T) {
		for range 100 {
			assert.Equal(t, 3, d.RandomInt(3, 4))
		}
	})

	t.Run("extreme range does not overflow", func(t *testing.T) {
		for range 1000 {
			got := d.RandomInt(math.MinInt, math.MaxInt)
			require.GreaterOrEqual(t, got, math.MinInt)
			require.Less(t, got, math.MaxInt)
		}
	})

	t.Run("produces more than one value", func(t *testing.T) {
		seen := map[int]struct{}{}
		for range 1000 {
			seen[d.RandomInt(0, 100)] = struct{}{}
		}
		assert.Greater(t, len(seen), 1)
	})

	t.Run("concurrent access", func(t *testing.T) {
		runConcurrently(t, testGoroutines, testIterations, func(_, _ int) {
			got := d.RandomInt(0, 100)
			assert.GreaterOrEqual(t, got, 0)
			assert.Less(t, got, 100)
		})
	})
}

func TestRandomFloat(t *testing.T) {
	d := newTestDecoy(t)

	t.Run("within bounds", func(t *testing.T) {
		cases := []struct{ a, b float64 }{
			{0, 1},
			{1, 2},
			{10, 20},
			{20, 10},
			{-5, 5},
			{-1, -2},
			{-2, -1},
			{0.5, 0.75},
		}
		for _, c := range cases {
			lo, hi := math.Min(c.a, c.b), math.Max(c.a, c.b)
			for range 1000 {
				got := d.RandomFloat(c.a, c.b)
				require.GreaterOrEqualf(t, got, lo, "RandomFloat(%v, %v)=%v", c.a, c.b, got)
				require.Lessf(t, got, hi, "RandomFloat(%v, %v)=%v", c.a, c.b, got)
			}
		}
	})

	t.Run("respects non-zero lower bound", func(t *testing.T) {
		for range 1000 {
			got := d.RandomFloat(100, 200)
			require.GreaterOrEqual(t, got, 100.0)
			require.Less(t, got, 200.0)
		}
	})

	t.Run("equal limits return the limit", func(t *testing.T) {
		assert.Equal(t, 3.5, d.RandomFloat(3.5, 3.5))
		assert.Equal(t, -7.25, d.RandomFloat(-7.25, -7.25))
	})

	t.Run("concurrent access", func(t *testing.T) {
		runConcurrently(t, testGoroutines, testIterations, func(_, _ int) {
			got := d.RandomFloat(0, 1)
			assert.GreaterOrEqual(t, got, 0.0)
			assert.Less(t, got, 1.0)
		})
	})
}

func TestRandomBoolean(t *testing.T) {
	d := newTestDecoy(t)

	t.Run("roughly balanced", func(t *testing.T) {
		const n = 20000
		trueCount := 0
		for range n {
			if d.RandomBoolean() {
				trueCount++
			}
		}
		ratio := float64(trueCount) / float64(n)
		assert.Greater(t, ratio, 0.45)
		assert.Less(t, ratio, 0.55)
	})

	t.Run("returns both values", func(t *testing.T) {
		seenTrue, seenFalse := false, false
		for range 1000 {
			if d.RandomBoolean() {
				seenTrue = true
			} else {
				seenFalse = true
			}
		}
		assert.True(t, seenTrue && seenFalse)
	})

	t.Run("concurrent access", func(t *testing.T) {
		runConcurrently(t, testGoroutines, testIterations, func(_, _ int) {
			_ = d.RandomBoolean()
		})
	})
}

func TestProbablity(t *testing.T) {
	d := newTestDecoy(t)

	t.Run("zero or below is always false", func(t *testing.T) {
		for _, p := range []float64{-1, -0.5, -math.SmallestNonzeroFloat64, 0, math.Inf(-1), math.NaN()} {
			assert.Falsef(t, d.Probablity(p), "Probablity(%v)", p)
		}
	})

	t.Run("one or above is always true", func(t *testing.T) {
		for _, p := range []float64{1, 1.5, math.MaxFloat64, math.Inf(1)} {
			assert.Truef(t, d.Probablity(p), "Probablity(%v)", p)
		}
	})

	t.Run("empirical probability", func(t *testing.T) {
		cases := []struct {
			p         float64
			tolerance float64
		}{
			{0.25, 0.03},
			{0.5, 0.03},
			{0.75, 0.03},
		}
		const n = 40000
		for _, c := range cases {
			trueCount := 0
			for range n {
				if d.Probablity(c.p) {
					trueCount++
				}
			}
			ratio := float64(trueCount) / float64(n)
			assert.InDeltaf(t, c.p, ratio, c.tolerance, "Probablity(%v) ratio", c.p)
		}
	})

	t.Run("concurrent access", func(t *testing.T) {
		runConcurrently(t, testGoroutines, testIterations, func(_, _ int) {
			_ = d.Probablity(0.5)
		})
	})
}

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
		runConcurrently(t, testGoroutines, testIterations, func(g, _ int) {
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
		runConcurrently(t, testGoroutines, testIterations, func(_, _ int) {
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
		runConcurrently(t, testGoroutines, testIterations, func(g, _ int) {
			d.NextIncrementalInt("id-"+strconv.Itoa(g), 0, 1)
		})
		for g := range testGoroutines {
			assert.Equal(t, int64(testIterations-1), d.CurrentIncrementalInt("id-"+strconv.Itoa(g), -1))
		}
	})
}

func TestRandomChoice(t *testing.T) {
	t.Run("empty returns zero value", func(t *testing.T) {
		d := newTestDecoy(t)
		assert.Equal(t, 0, d.RandomChoice[int]())
		assert.Equal(t, "", d.RandomChoice[string]())
		assert.Equal(t, "", d.RandomChoiceList([]string{}))
	})

	t.Run("single choice is constant", func(t *testing.T) {
		d := newTestDecoy(t)
		for range 100 {
			assert.Equal(t, "only", d.RandomChoice("only"))
		}
	})

	t.Run("returns one of the choices", func(t *testing.T) {
		d := newTestDecoy(t)
		choices := []string{"a", "b", "c"}
		seen := map[string]struct{}{}
		for range 1000 {
			got := d.RandomChoice(choices...)
			assert.Contains(t, choices, got)
			seen[got] = struct{}{}
		}
		assert.Len(t, seen, len(choices))
	})

	t.Run("list variant", func(t *testing.T) {
		d := newTestDecoy(t)
		choices := []int{1, 2, 3}
		seen := map[int]struct{}{}
		for range 1000 {
			got := d.RandomChoiceList(choices)
			assert.Contains(t, choices, got)
			seen[got] = struct{}{}
		}
		assert.Len(t, seen, len(choices))
	})

	t.Run("concurrent access", func(t *testing.T) {
		d := newTestDecoy(t)
		choices := []int{1, 2, 3, 4}
		runConcurrently(t, testGoroutines, testIterations, func(_, _ int) {
			got := d.RandomChoiceList(choices)
			assert.Contains(t, choices, got)
		})
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

	runConcurrently(t, goroutines, iterations, func(g, i int) {
		switch (g + i) % 6 {
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
		}
	})

	final := d.CurrentIncrementalInt(counterID, -1)
	assert.GreaterOrEqual(t, final, int64(0))
	assert.Less(t, final, int64(total))

	// Goroutine (leak) safety: no goroutine must be left permanently blocked.
	waitForGoroutines(t, baseline+2)
	assertNoGoroutineLeak(t)
}

func waitForGoroutines(t *testing.T, want int) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if runtime.NumGoroutine() <= want {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	assert.LessOrEqualf(t, runtime.NumGoroutine(), want,
		"goroutines did not settle back (runtime.NumGoroutine)")
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
	}

	waitForGoroutines(t, baseline+2)
	assertNoGoroutineLeak(t)
}
