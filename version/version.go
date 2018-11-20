package version

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
	// GitRev is the current HEAD of git of this release
	GitRev = ""
	// BuildTime is the ISO8601 timestamp of the current build
	BuildTime = ""

	// MajorInt is "Major" as parsed integer
	MajorInt = -1
	// MinorInt as "Minor" as parsed integer
	MinorInt = -1
	// PatchInt as "Patch" as parsed integer
	PatchInt = -1
)

func parseVersionNum(v, what string) int {
	if len(v) <= 0 {
		return 0
	}

	num, err := strconv.Atoi(Major)
	if err != nil {
		panic(fmt.Sprintf("Cannot parse %s version: %v", what, err))
	}

	return num
}

func init() {
	MajorInt = parseVersionNum(Major, "major")
	MinorInt = parseVersionNum(Minor, "minor")
	PatchInt = parseVersionNum(Patch, "patch")
}

// Numbers returns a tuple of (major, minor, patch)
func Numbers() (int, int, int) {
	return MajorInt, MinorInt, PatchInt
}

// String returns a Maj.Min.Patch string.
func String() string {
	base := fmt.Sprintf("v%s.%s.%s", Major, Minor, Patch)
	if ReleaseType != "" {
		base += "-" + ReleaseType
	}

	if len(GitRev) >= 7 {
		base += "+" + GitRev[:7]
	}

	return base
}
