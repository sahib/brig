package brig

import "fmt"

const (
	// MajorVersion will be incremented on big releases.
	MajorVersion = 0
	// MinorVersion will be incremented on small releases.
	MinorVersion = 0
	// PatchVersion should be incremented on every released change.
	PatchVersion = 0
)

// Version returns a tuple of (major, minor, patch)
func Version() (int, int, int) {
	return MajorVersion, MinorVersion, PatchVersion
}

// VersionString returns a Maj.Min.Patch string.
func VersionString() string {
	return fmt.Sprintf("%d.%d.%d", MajorVersion, MinorVersion, PatchVersion)
}
