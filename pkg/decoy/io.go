package decoy

import (
	"os"
)

func (d *Decoy) ReadFileBytes(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (d *Decoy) ReadFileString(path string) (string, error) {
	bytes, err := os.ReadFile(path)
	return string(bytes), err
}
