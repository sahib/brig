package main

import (
	"os"

	"github.com/sahib/brig/cmd"
)

func main() {
	os.Exit(cmd.RunCmdline(os.Args))
}
