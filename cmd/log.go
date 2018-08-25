package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli"
)

func logVerbose(ctx *cli.Context, format string, args ...interface{}) {
	if !ctx.GlobalBool("verbose") {
		return
	}

	if !strings.HasSuffix(format, "\n") {
		format = format + "\n"
	}

	fmt.Fprintf(os.Stderr, "-- "+format, args...)
}
