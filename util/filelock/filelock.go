// filelock implements helper function for using a `lock` file
// for synchronising file system resources.
package filelock

import (
	"time"

	"github.com/nightlyone/lockfile"
)

// Acquire tries to lock the lock file at `lockPath`.
// If it is already locked it will re-try after a short timeout.
func Acquire(lockPath string) error {
	lock, err := lockfile.New(lockPath)
	if err != nil {
		return err
	}

	for {
		if err := lock.TryLock(); err != nil {
			if err == lockfile.ErrBusy {
				time.Sleep(250 * time.Millisecond)
				continue
			} else {
				return err
			}
		}

		break
	}

	return nil
}

// TryAcquire tries to acquire the lock at `lockPath`.
// It will not retry if it fails.
func TryAcquire(lockPath string) error {
	lock, err := lockfile.New(lockPath)
	if err != nil {
		return err
	}

	return lock.TryLock()
}

// Release will remove the lockfile.
func Release(lockPath string) error {
	lock, err := lockfile.New(lockPath)
	if err != nil {
		return err
	}

	return lock.Unlock()
}
