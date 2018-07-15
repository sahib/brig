package main

import (
	"flag"
	"fmt"

	"gx/ipfs/QmPWdUPQGR7djd2UCwnAkNaRMZ8pfTiBa35Kj7Lmky1QLX/ginkgo/config"
)

func BuildVersionCommand() *Command {
	return &Command{
		Name:         "version",
		FlagSet:      flag.NewFlagSet("version", flag.ExitOnError),
		UsageCommand: "ginkgo version",
		Usage: []string{
			"Print Ginkgo's version",
		},
		Command: printVersion,
	}
}

func printVersion([]string, []string) {
	fmt.Printf("Ginkgo Version %s\n", config.VERSION)
}
