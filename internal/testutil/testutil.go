package testutil

import (
	"bytes"
	"regexp"
	"runtime"
	"runtime/pprof"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RunConcurrently runs fn from goroutines*iterations goroutines that are all
// released at the same time, maximizing the chance of contention.
func RunConcurrently(goroutines, iterations int, fn func(g, i int)) {
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

// AssertNoGoroutineLeak uses the Go 1.27 goroutine leak profiler. It triggers a
// leak-detecting GC cycle and asserts that no goroutine is permanently blocked
// on an unreachable synchronization primitive.
func AssertNoGoroutineLeak(t *testing.T) {
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

// WaitForGoroutines waits up to two seconds for the running goroutine count to
// settle back to at most want, then asserts it did.
func WaitForGoroutines(t *testing.T, want int) {
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
