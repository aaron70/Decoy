package template

import (
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/aaron70/goaty/errors"
)

func (t *Template) ReadFileBytes(path string) ([]byte, error) {
	return t.readFile(path)
}

func (t *Template) ReadFileString(path string) (string, error) {
	data, err := t.readFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (t *Template) readFile(path string) ([]byte, error) {
	resolved, err := t.resolvePath(path)
	if err != nil {
		return nil, err
	}
	if t.maxFileSize <= 0 {
		return os.ReadFile(resolved)
	}

	file, err := os.Open(resolved)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	limit := t.maxFileSize
	if limit < math.MaxInt64 {
		limit++
	}
	data, err := io.ReadAll(io.LimitReader(file, limit))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > t.maxFileSize {
		return nil, fmt.Errorf("%w: file %q exceeds the maximum allowed size of %d bytes", errors.ErrInvalidArgument, path, t.maxFileSize)
	}
	return data, nil
}

func (t *Template) resolvePath(path string) (string, error) {
	if t.baseDir == "" {
		return path, nil
	}

	base, err := filepath.Abs(t.baseDir)
	if err != nil {
		return "", err
	}
	base, err = filepath.EvalSymlinks(base)
	if err != nil {
		return "", err
	}

	candidate := path
	if !filepath.IsAbs(candidate) {
		candidate = filepath.Join(base, candidate)
	}
	candidate, err = filepath.Abs(candidate)
	if err != nil {
		return "", err
	}
	candidate, err = filepath.EvalSymlinks(candidate)
	if err != nil {
		return "", err
	}

	relative, err := filepath.Rel(base, candidate)
	if err != nil {
		return "", err
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("%w: path %q is outside the allowed base directory", errors.ErrInvalidArgument, path)
	}
	return candidate, nil
}
