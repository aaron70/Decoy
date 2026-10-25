package decoy

import (
	"math"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestDecoy(t *testing.T, seed uint64) *Decoy {
	t.Helper()
	d, err := NewDecoy(WithSeed(seed))
	require.NoError(t, err, "NewDecoy(WithSeed(%d))", seed)
	require.NotNil(t, d)
	return d
}

func requirePanicContains(t *testing.T, want string, fn func()) {
	t.Helper()
	var recovered any
	func() {
		defer func() {
			recovered = recover()
		}()
		fn()
	}()
	require.NotNil(t, recovered, "expected panic containing %q, got none", want)
	msg, ok := recovered.(string)
	require.True(t, ok, "expected string panic containing %q, got %T: %v", want, recovered, recovered)
	require.Contains(t, msg, want)
}

func runWithTimeout(t *testing.T, timeout time.Duration, fn func()) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		defer close(done)
		fn()
	}()
	select {
	case <-done:
	case <-time.After(timeout):
		require.FailNow(t, "timed out waiting for operation; possible deadlock")
	}
}

func waitWithTimeout(t *testing.T, wg *sync.WaitGroup, timeout time.Duration) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(timeout):
		require.FailNow(t, "timed out waiting for goroutines; possible deadlock")
	}
}

func TestSeedDeterminism(t *testing.T) {
	seeds := []uint64{0, 1, 42, math.MaxUint64}
	for _, seed := range seeds {
		t.Run(strconv.FormatUint(seed, 10), func(t *testing.T) {
			a := newTestDecoy(t, seed)
			b := newTestDecoy(t, seed)

			for i := range 200 {
				require.Equal(t, a.RandomInt(-100, 100), b.RandomInt(-100, 100), "draw %d: RandomInt", i)
				require.Equal(t, a.RandomFloat(-100, 100), b.RandomFloat(-100, 100), "draw %d: RandomFloat", i)
				require.Equal(t, a.RandomBoolean(), b.RandomBoolean(), "draw %d: RandomBoolean", i)
				require.Equal(t, a.Probability(0.5), b.Probability(0.5), "draw %d: Probability", i)
				require.Equal(t, a.RandomChoice("a", "b", "c"), b.RandomChoice("a", "b", "c"), "draw %d: RandomChoice", i)
			}
		})
	}
}

func TestDifferentSeeds(t *testing.T) {
	a := newTestDecoy(t, 1)
	b := newTestDecoy(t, 2)

	different := false
	for range 100 {
		if a.RandomInt(0, 1_000_000) != b.RandomInt(0, 1_000_000) {
			different = true
			break
		}
	}
	require.True(t, different, "expected different seeds to produce different sequences")
}

func TestRandomIntBounds(t *testing.T) {
	tests := []struct {
		name           string
		limit1, limit2 int
		min, max       int
	}{
		{"positive", 0, 10, 0, 10},
		{"single_value_range", 0, 1, 0, 1},
		{"negative", -10, -5, -10, -5},
		{"crossing_zero", -5, 5, -5, 5},
		{"reversed_positive", 10, 0, 0, 10},
		{"reversed_negative", -5, -10, -10, -5},
		{"wide", 0, math.MaxInt, 0, math.MaxInt},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := newTestDecoy(t, 12345)
			for i := range 1000 {
				got := d.RandomInt(tt.limit1, tt.limit2)
				require.GreaterOrEqual(t, got, tt.min, "draw %d: RandomInt(%d, %d)", i, tt.limit1, tt.limit2)
				require.Less(t, got, tt.max, "draw %d: RandomInt(%d, %d)", i, tt.limit1, tt.limit2)
			}
		})
	}
}

func TestRandomIntCoverage(t *testing.T) {
	tests := []struct {
		name           string
		limit1, limit2 int
		min, max       int
	}{
		{"non_negative", 0, 10, 0, 10},
		{"negative", -10, -5, -10, -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := newTestDecoy(t, 7)
			seen := make(map[int]bool)
			for range 10000 {
				got := d.RandomInt(tt.limit1, tt.limit2)
				require.GreaterOrEqual(t, got, tt.min, "RandomInt(%d, %d)", tt.limit1, tt.limit2)
				require.Less(t, got, tt.max, "RandomInt(%d, %d)", tt.limit1, tt.limit2)
				seen[got] = true
			}
			for v := tt.min; v < tt.max; v++ {
				assert.True(t, seen[v], "value %d never generated in [%d, %d)", v, tt.min, tt.max)
			}
		})
	}
}

func TestRandomIntSingleValue(t *testing.T) {
	tests := []struct {
		name                   string
		limit1, limit2, expect int
	}{
		{"positive", 5, 6, 5},
		{"negative", -3, -2, -3},
		{"reversed", 6, 5, 5},
		{"min_int", math.MinInt, math.MinInt + 1, math.MinInt},
		{"max_int", math.MaxInt - 1, math.MaxInt, math.MaxInt - 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := newTestDecoy(t, 1)
			require.Equal(t, tt.expect, d.RandomInt(tt.limit1, tt.limit2), "RandomInt(%d, %d)", tt.limit1, tt.limit2)
		})
	}
}

func TestRandomIntEmptyRangePanics(t *testing.T) {
	for _, limit := range []int{0, 3, -7, math.MinInt, math.MaxInt} {
		t.Run(strconv.Itoa(limit), func(t *testing.T) {
			d := newTestDecoy(t, 1)
			requirePanicContains(t, "empty range", func() {
				d.RandomInt(limit, limit)
			})
		})
	}
}

func TestRandomIntOverflowPanics(t *testing.T) {
	if strconv.IntSize != 64 {
		t.Skip("requires 64-bit int")
	}

	tests := []struct {
		name           string
		limit1, limit2 int
	}{
		{"full_range", math.MinInt, math.MaxInt},
		{"min_to_zero", math.MinInt, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := newTestDecoy(t, 1)
			requirePanicContains(t, "too large", func() {
				d.RandomInt(tt.limit1, tt.limit2)
			})
		})
	}
}

func TestRandomIntReversedEquivalent(t *testing.T) {
	a := newTestDecoy(t, 123)
	b := newTestDecoy(t, 123)

	for i := range 100 {
		require.Equal(t, a.RandomInt(0, 100), b.RandomInt(100, 0), "draw %d", i)
	}
}

func TestRandomIntPanicReleasesLock(t *testing.T) {
	d := newTestDecoy(t, 1)
	requirePanicContains(t, "empty range", func() {
		d.RandomInt(3, 3)
	})
	runWithTimeout(t, 5*time.Second, func() {
		got := d.RandomInt(0, 10)
		assert.GreaterOrEqual(t, got, 0)
		assert.Less(t, got, 10)
	})
}

func TestRandomFloatBounds(t *testing.T) {
	tests := []struct {
		name           string
		limit1, limit2 float64
		min, max       float64
	}{
		{"unit", 0, 1, 0, 1},
		{"negative", -10, -5, -10, -5},
		{"crossing_zero", -5, 5, -5, 5},
		{"reversed", 10, 0, 0, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := newTestDecoy(t, 99)
			for i := range 10000 {
				got := d.RandomFloat(tt.limit1, tt.limit2)
				require.False(t, math.IsNaN(got), "draw %d: RandomFloat(%v, %v) returned NaN", i, tt.limit1, tt.limit2)
				require.False(t, math.IsInf(got, 0), "draw %d: RandomFloat(%v, %v) returned Inf", i, tt.limit1, tt.limit2)
				require.GreaterOrEqual(t, got, tt.min, "draw %d: RandomFloat(%v, %v)", i, tt.limit1, tt.limit2)
				require.Less(t, got, tt.max, "draw %d: RandomFloat(%v, %v)", i, tt.limit1, tt.limit2)
			}
		})
	}
}

func TestRandomFloatEqualBounds(t *testing.T) {
	values := []float64{-5.5, 0, 1.25, math.SmallestNonzeroFloat64, math.MaxFloat64, -math.MaxFloat64}
	for _, value := range values {
		t.Run(strconv.FormatFloat(value, 'g', -1, 64), func(t *testing.T) {
			d := newTestDecoy(t, 1)
			require.Equal(t, value, d.RandomFloat(value, value))
		})
	}
}

func TestRandomFloatNonFinitePanics(t *testing.T) {
	tests := []struct {
		name           string
		limit1, limit2 float64
	}{
		{"nan_first", math.NaN(), 1},
		{"nan_second", 1, math.NaN()},
		{"nan_both", math.NaN(), math.NaN()},
		{"positive_inf_first", math.Inf(1), 1},
		{"positive_inf_second", 1, math.Inf(1)},
		{"negative_inf_first", math.Inf(-1), 1},
		{"negative_inf_second", 1, math.Inf(-1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := newTestDecoy(t, 1)
			requirePanicContains(t, "non-finite", func() {
				d.RandomFloat(tt.limit1, tt.limit2)
			})
		})
	}
}

func TestRandomFloatOverflowingSpanPanics(t *testing.T) {
	tests := []struct {
		name           string
		limit1, limit2 float64
	}{
		{"max_symmetric", -math.MaxFloat64, math.MaxFloat64},
		{"max_symmetric_reversed", math.MaxFloat64, -math.MaxFloat64},
		{"large_symmetric", -1e308, 1e308},
		{"large_symmetric_reversed", 1e308, -1e308},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := newTestDecoy(t, 1)
			requirePanicContains(t, "too large", func() {
				d.RandomFloat(tt.limit1, tt.limit2)
			})
		})
	}
}

func TestRandomFloatWideFiniteRange(t *testing.T) {
	const limit = 1e300
	d := newTestDecoy(t, 1)

	for i := range 1000 {
		got := d.RandomFloat(-limit, limit)
		require.False(t, math.IsNaN(got), "draw %d: RandomFloat(%v, %v) returned NaN", i, -limit, limit)
		require.False(t, math.IsInf(got, 0), "draw %d: RandomFloat(%v, %v) returned Inf", i, -limit, limit)
		require.GreaterOrEqual(t, got, -limit, "draw %d", i)
		require.LessOrEqual(t, got, limit, "draw %d", i)
	}
}

func TestRandomFloatReversedEquivalent(t *testing.T) {
	a := newTestDecoy(t, 123)
	b := newTestDecoy(t, 123)

	for i := range 100 {
		require.Equal(t, a.RandomFloat(-50, 50), b.RandomFloat(50, -50), "draw %d", i)
	}
}

func TestRandomFloatPanicReleasesLock(t *testing.T) {
	d := newTestDecoy(t, 1)
	requirePanicContains(t, "non-finite", func() {
		d.RandomFloat(math.NaN(), 0)
	})
	runWithTimeout(t, 5*time.Second, func() {
		got := d.RandomFloat(0, 1)
		assert.False(t, math.IsNaN(got))
		assert.GreaterOrEqual(t, got, 0.0)
		assert.Less(t, got, 1.0)
	})
}

func TestRandomBoolean(t *testing.T) {
	const draws = 1000
	d := newTestDecoy(t, 1)
	trueCount := 0

	for range draws {
		if d.RandomBoolean() {
			trueCount++
		}
	}
	require.NotZero(t, trueCount, "RandomBoolean never returned true")
	require.NotEqual(t, draws, trueCount, "RandomBoolean never returned false")
}

func TestProbabilityEndpoints(t *testing.T) {
	d := newTestDecoy(t, 1)

	for _, p := range []float64{-1, 0, math.Inf(-1), math.NaN()} {
		assert.False(t, d.Probability(p), "Probability(%v)", p)
	}
	for _, p := range []float64{1, 2, math.Inf(1)} {
		assert.True(t, d.Probability(p), "Probability(%v)", p)
	}
}

func TestProbabilityInterior(t *testing.T) {
	const draws = 10000
	d := newTestDecoy(t, 123)
	trueCount := 0

	for range draws {
		if d.Probability(0.5) {
			trueCount++
		}
	}
	require.GreaterOrEqual(t, trueCount, 4000, "Probability(0.5) over %d draws", draws)
	require.LessOrEqual(t, trueCount, 6000, "Probability(0.5) over %d draws", draws)
}

func TestProbabilityVerySmall(t *testing.T) {
	d := newTestDecoy(t, 5)
	for i := range 1000 {
		require.False(t, d.Probability(math.SmallestNonzeroFloat64), "draw %d", i)
	}
}

func TestRandomChoiceEmptyPanics(t *testing.T) {
	d := newTestDecoy(t, 1)
	requirePanicContains(t, "at least one choice", func() {
		d.RandomChoice[int]()
	})
	requirePanicContains(t, "at least one choice", func() {
		d.RandomChoice([]int{}...)
	})
	var nothing []string
	requirePanicContains(t, "at least one choice", func() {
		d.RandomChoice(nothing...)
	})
}

func TestRandomChoiceSingle(t *testing.T) {
	type sample struct {
		ID   int
		Name string
	}

	d := newTestDecoy(t, 1)
	want := sample{ID: 7, Name: "seven"}

	for range 100 {
		require.Equal(t, 42, d.RandomChoice(42))
		require.Equal(t, "only", d.RandomChoice("only"))
		require.Equal(t, want, d.RandomChoice(want))
	}
}

func TestRandomChoiceMultiple(t *testing.T) {
	d := newTestDecoy(t, 123)
	choices := []string{"alpha", "beta", "gamma", "delta"}
	valid := make(map[string]bool, len(choices))
	seen := make(map[string]bool, len(choices))
	for _, choice := range choices {
		valid[choice] = true
	}

	for range 1000 {
		got := d.RandomChoice(choices...)
		require.True(t, valid[got], "RandomChoice returned %q, want one of %v", got, choices)
		seen[got] = true
	}
	for _, choice := range choices {
		assert.True(t, seen[choice], "choice %q never returned", choice)
	}
}

func TestConcurrentRandomCalls(t *testing.T) {
	d := newTestDecoy(t, 1234)
	const goroutines = 64
	const iterations = 500

	var wg sync.WaitGroup
	for range goroutines {
		wg.Go(func() {
			for range iterations {
				intGot := d.RandomInt(-10, 10)
				assert.GreaterOrEqual(t, intGot, -10, "RandomInt(-10, 10)")
				assert.Less(t, intGot, 10, "RandomInt(-10, 10)")

				floatGot := d.RandomFloat(-1, 1)
				assert.False(t, math.IsNaN(floatGot), "RandomFloat(-1, 1) returned NaN")
				assert.False(t, math.IsInf(floatGot, 0), "RandomFloat(-1, 1) returned Inf")
				assert.GreaterOrEqual(t, floatGot, -1.0, "RandomFloat(-1, 1)")
				assert.Less(t, floatGot, 1.0, "RandomFloat(-1, 1)")

				_ = d.RandomBoolean()
				_ = d.Probability(0.5)
				assert.Contains(t, []string{"a", "b", "c"}, d.RandomChoice("a", "b", "c"))
			}
		})
	}
	waitWithTimeout(t, &wg, 10*time.Second)
}

func TestConcurrentPanicRecovery(t *testing.T) {
	d := newTestDecoy(t, 7)
	const goroutines = 8
	const iterations = 25

	var wg sync.WaitGroup
	for range goroutines {
		wg.Go(func() {
			for range iterations {
				assert.Panics(t, func() {
					d.RandomInt(1, 1)
				}, "expected RandomInt(1, 1) to panic")

				intGot := d.RandomInt(0, 10)
				assert.GreaterOrEqual(t, intGot, 0, "RandomInt(0, 10)")
				assert.Less(t, intGot, 10, "RandomInt(0, 10)")

				assert.Panics(t, func() {
					d.RandomFloat(math.NaN(), 0)
				}, "expected RandomFloat(NaN, 0) to panic")

				floatGot := d.RandomFloat(0, 1)
				assert.False(t, math.IsNaN(floatGot), "RandomFloat(0, 1) returned NaN")
				assert.GreaterOrEqual(t, floatGot, 0.0, "RandomFloat(0, 1)")
				assert.Less(t, floatGot, 1.0, "RandomFloat(0, 1)")
			}
		})
	}
	waitWithTimeout(t, &wg, 15*time.Second)
}
