package repo

import (
	"os"
	"path/filepath"
)

// IsRepo checks if `folder` contains a brig repository.
// Currently, this is implemented by checkin for the hidden .brig folder,
// but this behaviour might change in the future.
func IsRepo(folder string) bool {
	file, err := os.Stat(filepath.Join(folder, ".brig"))
	if err != nil {
		return false
	}

	return file.IsDir()
}

// FindRepo checks if `folder` or any of it's parents contains a brig
// repository.  It uses IsRepo() to check if the folder is a repository.
// The path works on both relative and absolute paths.
func FindRepo(folder string) string {
	curr, err := filepath.Abs(folder)
	if err != nil {
		return ""
	}

	for curr != "" {
		if IsRepo(curr) {
			return curr
		}

		// Try in the parent directory:
		dirname := filepath.Dir(curr)
		if dirname == curr {
			break
		}

		curr = dirname
	}
	return ""
}
