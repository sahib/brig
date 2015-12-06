// +build !windows

package trie

import (
	"os"
	"path/filepath"
)

func buildPath(s []string) string {
	return string(os.PathSeparator) + filepath.Join(s...)
}
