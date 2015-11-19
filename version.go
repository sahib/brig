package brig

import "fmt"

const (
	MajorVersion = 0
	MinorVersion = 0
	PatchVersion = 0
)

func Version() (int, int, int) {
	return MajorVersion, MinorVersion, PatchVersion
}

func VersionString() string {
	return fmt.Sprintf("%d.%d.%d", MajorVersion, MinorVersion, PatchVersion)
}
