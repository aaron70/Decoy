package template

func (t *Template) RandomChoiceList[T any](choices []T) (T, error) {
	return t.RandomChoice(choices...)
}
