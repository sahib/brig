package main

import (
	"github.com/disorganizer/brig/cmd"
	"os"
)

func main() {
	os.Exit(cmdline.RunCmdline())
}
