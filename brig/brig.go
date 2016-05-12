package main

import (
	"github.com/disorganizer/brig/cmdline"
	"os"
)

var (
	Major     = "unknown"
	Minor     = "unknown"
	Patch     = "unknown"
	Gitrev    = "unknown"
	Buildtime = "unknown"
)

func main() {
	os.Exit(
		cmdline.RunCmdline(Major, Minor, Patch, Gitrev, Buildtime),
	)
}
