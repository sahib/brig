package cmd

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/sahib/brig/client"
	"github.com/sahib/brig/version"
	"github.com/toqueteos/webbrowser"
	"github.com/urfave/cli"
)

// printError simply prints a nicely formatted error to stderr.
func printError(msg string) {
	fmt.Fprintln(os.Stderr, color.RedString("*** ")+msg)
}

// cmdOutput runs a command at `path` with `args` and returns it's output.
// No real error checking is done, on errors an empty string is returned.
func cmdOutput(path string, args ...string) string {
	out, err := exec.Command(path, args...).Output()
	if err != nil {
		// No other error checking here, `brig bug` is best effort.
		printError(fmt.Sprintf("failed to run %s %s", path, strings.Join(args, " ")))
		return ""
	}

	return strings.TrimSpace(string(out))
}

// handleBugReport compiles a report of useful info when providing a bug report.
func handleBugReport(ctx *cli.Context) error {
	buf := &bytes.Buffer{}
	fmt.Fprintln(buf, `Please answer these questions before submitting your issue.
Please include anything else you think is helpful. Thanks!

### What did you do?

### What did you expect to see?

### What did you see instead?

### Do you still see this issue with a development binary?

### Did you check if a similar bug report was already opened?

### System details:`)

	fmt.Fprintf(buf, "go version:     ``%s``\n", cmdOutput("go", "version"))
	fmt.Fprintf(buf, "uname -s -v -m: ``%s``\n", cmdOutput("uname", "-s", "-v", "-m"))
	fmt.Fprintf(buf, "\n")

	fmt.Fprintf(
		buf,
		"brig client version: ``%s [build: %s]``\n",
		version.String(),
		version.BuildTime,
	)

	port := ctx.GlobalInt("port")
	ctl, err := client.Dial(context.Background(), port)
	if err == nil {
		// Try to get the server side / ipfs version.
		version, err := ctl.Version()
		if err == nil {
			fmt.Fprintf(
				buf,
				"brig server version: ``%s+%s``\n",
				version.ServerSemVer,
				version.ServerRev,
			)
			fmt.Fprintf(
				buf,
				"ipfs version:        ``%s+%s``\n",
				version.BackendSemVer,
				version.BackendRev,
			)
		}
	} else {
		printError("Cannot get server and ipfs version.")
		printError("If it is possible to start the daemon, do it now.")
		printError("This will make the bug report more helpful. Thanks.")
	}

	printToStdout := ctx.Bool("stdout")
	if !printToStdout {
		// Try to open the issue tracker for convinience:
		urlVal := url.Values{}
		urlVal.Set("body", buf.String())
		reportUrl := "https://github.com/sahib/brig/issues/new?"

		if err := webbrowser.Open(reportUrl + urlVal.Encode()); err != nil {
			printToStdout = true
		}
	}

	if printToStdout {
		// If not, ask the user to print it directly:
		if !ctx.Bool("stdout") {
			printError("I failed to open the issue tracker in your browser.")
			printError("Please paste the underlying text manually at this URL:")
			printError("https://github.com/sahib/brig/issues")
		}
		fmt.Println(buf.String())
	}

	return nil
}
