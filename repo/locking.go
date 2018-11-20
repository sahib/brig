package repo

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/sahib/brig/catfs/mio/encrypt"
	"github.com/sahib/brig/util"
)

const (
	// LockPathSuffix is the suffix appended to each file in its locked state.
	LockPathSuffix = ".locked"
	// LockDirSuffix is the suffix appended to each directory in its locked state.
	LockDirSuffix = ".tgz" + LockPathSuffix
)

func lockFile(path string, key []byte) error {
	lockedPath := path + LockPathSuffix
	dstFd, err := os.OpenFile(lockedPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	defer util.Closer(dstFd)

	srcFd, err := os.Open(path)
	if err != nil {
		return err
	}

	defer util.Closer(srcFd)

	encW, err := encrypt.NewWriter(dstFd, key)
	if err != nil {
		return err
	}

	if _, err = io.Copy(encW, srcFd); err != nil {
		return err
	}

	return encW.Close()
}

func lockDirectory(path string, key []byte) error {
	lockedPath := path + LockDirSuffix
	fd, err := os.OpenFile(lockedPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	defer util.Closer(fd)

	encW, err := encrypt.NewWriter(fd, key)
	if err != nil {
		return err
	}

	archiveName := fmt.Sprintf("encrypted content of %s", lockedPath)
	if err := util.Tar(path, archiveName, encW); err != nil {
		return err
	}

	return encW.Close()
}

func isExcluded(path string, excludePatterns []string) bool {
	for _, pattern := range excludePatterns {
		matched, err := filepath.Match(pattern, filepath.Base(path))

		// Should only happen for mal-formend patterns.
		if err != nil {
			log.Warningf("BUG: Failed to compile exclude pattern: %v: %v", pattern, err)
			continue
		}

		// Ignore the file if it matched:
		if matched {
			return true
		}
	}

	return false
}

func keyFromPassword(owner, password string) []byte {
	return util.DeriveKey([]byte(password), []byte(owner), 32)
}

// LockRepo encrypts all files (except those in `lockExcludes`) in `root`,
// depending on `user` and `password`. `unlockExcludes` is only used to
// prevent warnings about not locked files.
func LockRepo(root, user, password string, lockExcludes, unlockExcludes []string) error {
	files, err := ioutil.ReadDir(root)
	if err != nil {
		return err
	}

	// user is not the perfect salt, but pretty much the only available one here.
	key := keyFromPassword(user, password)

	for _, info := range files {
		path := filepath.Join(root, info.Name())
		if strings.HasSuffix(path, LockPathSuffix) {
			if !isExcluded(path, unlockExcludes) {
				log.Warningf(
					"%s already contains a locked file: %s; Ignoring",
					root,
					path,
				)
			}
			continue
		}

		if isExcluded(path, lockExcludes) {
			continue
		}

		switch {
		case info.Mode().IsDir():
			if err := lockDirectory(path, key); err != nil {
				return err
			}
		case info.Mode().IsRegular():
			if err := lockFile(path, key); err != nil {
				return err
			}
		default:
			log.Warningf("Ignoring non-file `%s`", path)
			continue
		}

		// File was successfully locked, remove the source.
		if err := os.RemoveAll(path); err != nil {
			return err
		}
	}

	return nil
}

func checkUnlockability(path string, key []byte) error {
	srcFd, err := os.Open(path)
	if err != nil {
		return err
	}

	defer util.Closer(srcFd)

	encR, err := encrypt.NewReader(srcFd, key)
	if err != nil {
		return err
	}

	_, err = io.Copy(ioutil.Discard, encR)
	return err
}

func unlockFile(path string, key []byte) error {
	srcFd, err := os.Open(path)
	if err != nil {
		return err
	}

	defer util.Closer(srcFd)

	encR, err := encrypt.NewReader(srcFd, key)
	if err != nil {
		return err
	}

	unlockedPath := path[:len(path)-len(LockPathSuffix)]
	dstFd, err := os.OpenFile(unlockedPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	defer util.Closer(dstFd)

	_, err = io.Copy(dstFd, encR)
	if err != nil {
		// Do not leave a half-finished file behind if copy failed.
		os.Remove(unlockedPath)
		return err
	}

	return nil
}

func unlockDirectory(path string, key []byte) error {
	unlockedPath := path[:len(path)-len(LockDirSuffix)]
	fd, err := os.Open(path)
	if err != nil {
		return err
	}

	defer util.Closer(fd)

	encR, err := encrypt.NewReader(fd, key)
	if err != nil {
		return err
	}

	return util.Untar(encR, unlockedPath)
}

// UnlockRepo is the exact opposite of LockRepo.
func UnlockRepo(root, user, password string, lockExcludes, unlockExcludes []string) error {
	files, err := ioutil.ReadDir(root)
	if err != nil {
		return err
	}

	warnAboutUnlockedFiles := true
	initTagPath := filepath.Join(root, "INIT_TAG")
	if _, err := os.Stat(initTagPath); err == nil {
		warnAboutUnlockedFiles = false
		if err := os.Remove(initTagPath); err != nil {
			return err
		}
	}

	key := keyFromPassword(user, password)

	for _, info := range files {
		path := filepath.Join(root, info.Name())

		if isExcluded(path, unlockExcludes) {
			continue
		}

		switch {
		case strings.HasSuffix(path, LockDirSuffix):
			if err := unlockDirectory(path, key); err != nil {
				return err
			}
		case strings.HasSuffix(path, LockPathSuffix):
			if err := unlockFile(path, key); err != nil {
				return err
			}
		default:
			if warnAboutUnlockedFiles && !isExcluded(path, lockExcludes) {
				log.Warningf("%s was not locked. Ignoring.", path)
			}
			continue
		}

		if err := os.RemoveAll(path); err != nil {
			return err
		}
	}

	return nil
}
