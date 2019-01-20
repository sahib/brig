package parcello

import (
	"os"
	"path/filepath"
)

func match(pattern, path, name string) (bool, error) {
	matched, err := filepath.Match(pattern, path)
	if err != nil {
		return false, err
	}

	try, _ := filepath.Match(pattern, name)
	return matched || try, nil
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}
