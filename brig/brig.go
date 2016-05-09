package main

import (
	"github.com/disorganizer/brig/cmdline"
	"os"
)

func main() {
	x := cmdline.RunCmdline()
	os.Exit(x)
}
