package decoy

import (
	"testing"

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
