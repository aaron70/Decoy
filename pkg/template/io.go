package template

import (
	"os"
)

func (t *Template) ReadFileBytes(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (t *Template) ReadFileString(path string) (string, error) {
	bytes, err := os.ReadFile(path)
	return string(bytes), err
}
