package decoy

import (
	"math"
	"testing"

	"github.com/aaron70/decoy/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
		testutil.RunConcurrently(testGoroutines, testIterations, func(_, _ int) {
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
		testutil.RunConcurrently(testGoroutines, testIterations, func(_, _ int) {
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
		testutil.RunConcurrently(testGoroutines, testIterations, func(_, _ int) {
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
		testutil.RunConcurrently(testGoroutines, testIterations, func(_, _ int) {
			_ = d.Probablity(0.5)
		})
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
		testutil.RunConcurrently(testGoroutines, testIterations, func(_, _ int) {
			got := d.RandomChoiceList(choices)
			assert.Contains(t, choices, got)
		})
	})
}
