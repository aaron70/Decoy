package decoy

import (
	"math"
	"strings"
	"testing"

	"github.com/aaron70/decoy/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRandomName(t *testing.T) {
	d := newTestDecoy(t)

	t.Run("returns a known first name", func(t *testing.T) {
		for range 1000 {
			got := d.RandomName()
			assert.NotEmpty(t, got)
			assert.Contains(t, firstNames, got)
		}
	})

	t.Run("produces variety", func(t *testing.T) {
		seen := map[string]struct{}{}
		for range 1000 {
			seen[d.RandomName()] = struct{}{}
		}
		assert.Greater(t, len(seen), 1)
	})

	t.Run("empty corpus returns empty string", func(t *testing.T) {
		saved := firstNames
		firstNames = nil
		defer func() { firstNames = saved }()

		assert.Equal(t, "", d.RandomName())
	})

	t.Run("concurrent access", func(t *testing.T) {
		allowed := make(map[string]struct{}, len(firstNames))
		for _, n := range firstNames {
			allowed[n] = struct{}{}
		}
		testutil.RunConcurrently(testGoroutines, testIterations, func(_, _ int) {
			_, ok := allowed[d.RandomName()]
			assert.True(t, ok)
		})
	})
}

func TestRandomLastName(t *testing.T) {
	d := newTestDecoy(t)

	t.Run("returns a known last name", func(t *testing.T) {
		for range 1000 {
			got := d.RandomLastName()
			assert.NotEmpty(t, got)
			assert.Contains(t, lastNames, got)
		}
	})

	t.Run("produces variety", func(t *testing.T) {
		seen := map[string]struct{}{}
		for range 1000 {
			seen[d.RandomLastName()] = struct{}{}
		}
		assert.Greater(t, len(seen), 1)
	})

	t.Run("empty corpus returns empty string", func(t *testing.T) {
		saved := lastNames
		lastNames = nil
		defer func() { lastNames = saved }()

		assert.Equal(t, "", d.RandomLastName())
	})

	t.Run("concurrent access", func(t *testing.T) {
		allowed := make(map[string]struct{}, len(lastNames))
		for _, n := range lastNames {
			allowed[n] = struct{}{}
		}
		testutil.RunConcurrently(testGoroutines, testIterations, func(_, _ int) {
			_, ok := allowed[d.RandomLastName()]
			assert.True(t, ok)
		})
	})
}

// mockNames temporarily replaces the name corpora for the duration of the test.
// Tests using it must not run in parallel with other tests that read the corpora.
func mockNames(t *testing.T, first, last []string) {
	t.Helper()
	savedFirst, savedLast := firstNames, lastNames
	firstNames, lastNames = first, last
	t.Cleanup(func() {
		firstNames, lastNames = savedFirst, savedLast
	})
}

func toSet(names []string) map[string]struct{} {
	set := make(map[string]struct{}, len(names))
	for _, n := range names {
		set[n] = struct{}{}
	}
	return set
}

func TestRandomFullName(t *testing.T) {
	t.Run("no middle name when probability is not positive", func(t *testing.T) {
		mockNames(t, []string{"A", "B"}, []string{"C", "D"})
		d := newTestDecoy(t)
		for _, p := range []float64{-1, 0, math.NaN()} {
			parts := strings.Fields(d.RandomFullName(p))
			require.Lenf(t, parts, 2, "RandomFullName(%v)", p)
			assert.Contains(t, []string{"A", "B"}, parts[0])
			assert.Contains(t, []string{"C", "D"}, parts[1])
		}
	})

	t.Run("always middle name when probability is at least one", func(t *testing.T) {
		mockNames(t, []string{"A", "B"}, []string{"C", "D"})
		d := newTestDecoy(t)
		for _, p := range []float64{1, 2} {
			parts := strings.Fields(d.RandomFullName(p))
			require.Lenf(t, parts, 3, "RandomFullName(%v)", p)
			assert.Contains(t, []string{"A", "B"}, parts[0])
			assert.Contains(t, []string{"C", "D"}, parts[1])
			assert.Contains(t, []string{"C", "D"}, parts[2])
		}
	})

	t.Run("middle name probability", func(t *testing.T) {
		mockNames(t, []string{"A"}, []string{"B"})
		d := newTestDecoy(t)
		const (
			n         = 40000
			p         = 0.5
			tolerance = 0.03
		)
		withMiddle := 0
		for range n {
			if len(strings.Fields(d.RandomFullName(p))) == 3 {
				withMiddle++
			}
		}
		ratio := float64(withMiddle) / float64(n)
		assert.InDelta(t, p, ratio, tolerance)
	})

	t.Run("deterministic with single-entry corpora", func(t *testing.T) {
		mockNames(t, []string{"A"}, []string{"B"})
		d := newTestDecoy(t)
		assert.Equal(t, "A B", d.RandomFullName(0))
		assert.Equal(t, "A B B", d.RandomFullName(1))
	})

	t.Run("uses real first and last names", func(t *testing.T) {
		d := newTestDecoy(t)
		first, last := toSet(firstNames), toSet(lastNames)
		for range 1000 {
			name := d.RandomFullName(0.5)

			hasFirst := false
			for f := range first {
				if strings.HasPrefix(name, f+" ") {
					hasFirst = true
					break
				}
			}
			hasLast := false
			for l := range last {
				if strings.HasSuffix(name, " "+l) {
					hasLast = true
					break
				}
			}
			require.Truef(t, hasFirst, "name %q does not start with a known first name", name)
			require.Truef(t, hasLast, "name %q does not end with a known last name", name)
		}
	})

	t.Run("concurrent access", func(t *testing.T) {
		d := newTestDecoy(t)
		testutil.RunConcurrently(testGoroutines, testIterations, func(_, _ int) {
			assert.NotEmpty(t, d.RandomFullName(0.5))
		})
	})
}
