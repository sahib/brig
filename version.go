package brig

import (
	"fmt"
	"strconv"
)

var (
	// Major will be incremented on big releases.
	Major = ""
	// Minor will be incremented on small releases.
	Minor = ""
	// Patch should be incremented on every released change.
	Patch = ""
	// ReleaseType is "beta", "alpha" or "" for final releases
	ReleaseType = ""
	// Gitrev is the current HEAD of git of this release
	GitRev = ""
	// Buildtime is the ISO8601 timestamp of the current build
	BuildTime = ""

	// MajorInt is "Major" as parsed integer
	MajorInt = -1
	// MinorInt as "Minor" as parsed integer
	MinorInt = -1
	// PatchInt as "Patch" as parsed integer
	PatchInt = -1
)

func init() {
	var err error

	if len(Major) > 0 {
		if MajorInt, err = strconv.Atoi(Major); err != nil {
			panic(fmt.Sprintf("Cannot parse major version: %v", err))
		}
	}

	if len(Minor) > 0 {
		if MinorInt, err = strconv.Atoi(Minor); err != nil {
			panic(fmt.Sprintf("Cannot parse minor version: %v", err))
		}
	}

	if len(Patch) > 0 {
		if PatchInt, err = strconv.Atoi(Patch); err != nil {
			panic(fmt.Sprintf("Cannot parse patch version: %v", err))
		}
	}
}

// Version returns a tuple of (major, minor, patch)
func Version() (int, int, int) {
	return MajorInt, MinorInt, PatchInt
}

// VersionString returns a Maj.Min.Patch string.
func VersionString() string {
	base := fmt.Sprintf("v%s.%s.%s", Major, Minor, Patch)
	if ReleaseType != "" {
		base += "-" + ReleaseType
	}

	if len(GitRev) >= 7 {
		base += "+" + GitRev[:7]
	}

	return base
}
