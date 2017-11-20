package main

import (
	"os"

	"github.com/disorganizer/brig/cmd"
)

func main() {

	os.Exit(cmd.RunCmdline(os.Args))
}
