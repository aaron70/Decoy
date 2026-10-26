package decoy

import (
	"errors"

	gErrors "github.com/aaron70/goaty/errors"
)


var (
	ErrRangeTooLarge = errors.Join(gErrors.ErrInvalidArgument, gErrors.New("RangeTooLarge"))
)
