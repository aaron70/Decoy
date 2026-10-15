package decoy

import (
	"fmt"

	"github.com/aaron70/decoy/internal/random"
	"github.com/aaron70/goaty/utils"
)

var Default = utils.Must(NewDecoyWithSeed(0))

func RandomChoice[T any](decoy *Decoy, choices ...T) (T, error) {
	var zero T
	if len(choices) == 0 {
		return zero, fmt.Errorf("requires at least one argument")
	}
	return random.RandomChoice(decoy.Rand, choices...), nil
}
