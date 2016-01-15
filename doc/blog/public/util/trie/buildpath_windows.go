package trie

import (
	"path/filepath"
)

func buildPath(s []string) string {
	return filepath.Join(s...)
}
