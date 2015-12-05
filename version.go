package brig

import "fmt"

const (
	// Major will be incremented on big releases.
	Major = 0
	// Minor will be incremented on small releases.
	Minor = 0
	// Patch should be incremented on every released change.
	Patch = 0
	// PreRelease is an empty string for final releases, {alpha,beta} for pre-releases.
	PreRelease = "beta"
)

// Version returns a tuple of (major, minor, patch)
func Version() (int, int, int) {
	return Major, Minor, Patch
}

// VersionString returns a Maj.Min.Patch string.
func VersionString() string {
	return fmt.Sprintf("v%d.%d.%d-%s", Major, Minor, Patch, PreRelease)
}
