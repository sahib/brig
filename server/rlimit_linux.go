// +build linux

package server

import (
	"syscall"

	log "github.com/sirupsen/logrus"
)

func increaseMaxOpenFds() error {
	rLimit := syscall.Rlimit{}
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		return err
	}

	// We're good already.
	if rLimit.Cur >= rLimit.Max {
		return nil
	}

	rLimit.Cur = rLimit.Max
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		return err
	}

	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		return err
	}

	log.Debugf("Increased max number of open fds to %d (hard: %d)", rLimit.Cur, rLimit.Max)
	return nil
}
