package template

import (
	"fmt"
	"reflect"

	"github.com/aaron70/goaty/errors"
)

func (t *Template) RandomChoiceList[T any](choices []T) (T, error) {
	return t.RandomChoice(choices...)
}

func (t *Template) randomChoiceList(choices any) (any, error) {
	value := reflect.ValueOf(choices)
	if !value.IsValid() || (value.Kind() != reflect.Slice && value.Kind() != reflect.Array) {
		return nil, fmt.Errorf("%w: randomChoiceList: expected a slice or array, got %T", errors.ErrInvalidArgument, choices)
	}
	if value.Len() == 0 {
		return nil, fmt.Errorf("%w: randomChoiceList: requires at least one choice", errors.ErrInvalidArgument)
	}

	index, err := t.RandomInt(0, value.Len())
	if err != nil {
		return nil, err
	}
	return value.Index(index).Interface(), nil
}
