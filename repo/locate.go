package repo

import (
	log "github.com/Sirupsen/logrus"
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

// GuessFolder tries to find the desired brig repo by heuristics.
// Current heuristics: check env var BRIG_PATH, then the working dir.
// On failure, it will return an empty string.
func GuessFolder() string {
	wd := os.Getenv("BRIG_PATH")
	if wd == "" {
		var err error
		wd, err = os.Getwd()
		if err != nil {
			log.Errorf("Unable to fetch working dir: %q", err)
			return ""
		}
	}

	actualPath := FindRepo(wd)
	if actualPath == "" {
		log.Errorf("Unable to find repo in path or any parents: %q", wd)
	}

	return actualPath
}
