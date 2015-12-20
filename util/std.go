// Package util implements small helper function that
// should be included in the stdlib in our opinion.
package util

import (
	log "github.com/Sirupsen/logrus"
	"io"
)

// Min returns the minimum of a and b.
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Max returns the maximum of a and b.
func Max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

// Clamp limits x to the range [lo, hi]
func Clamp(x, lo, hi int) int {
	return Max(lo, Min(x, hi))
}

// UMin returns the unsigned minimum of a and b
func UMin(a, b uint) uint {
	if a < b {
		return a
	}
	return b
}

// UMax returns the unsigned minimum of a and b
func UMax(a, b uint) uint {
	if a < b {
		return b
	}
	return a
}

// UClamp limits x to the range [lo, hi]
func UClamp(x, lo, hi uint) uint {
	return UMax(lo, UMin(x, hi))
}

// Closer closes c. If that fails, it will log the error.
// The intended usage is for convinient defer calls only!
func Closer(c io.Closer) {
	if err := c.Close(); err != nil {
		log.Errorf("Could not close: %v", err)
	}
}
