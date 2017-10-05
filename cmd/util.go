package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/VividCortex/godaemon"
	"github.com/codegangsta/cli"
	"github.com/disorganizer/brig/brigd/client"
)

// ExitCode is an error that maps the error interface to a specific error
// message and a unix exit code
type ExitCode struct {
	Code    int
	Message string
}

func (err ExitCode) Error() string {
	return err.Message
}

// guessRepoFolder tries to find the repository path
// by using a number of sources.
func guessRepoFolder() string {
	path := os.Getenv("BRIG_PATH")
	if path == "" {
		return "."
	}

	return path
}

func readPassword() (string, error) {
	// TODO: Implement again.
	return "klaus", nil
}

func prefixSlash(s string) string {
	if !strings.HasPrefix(s, "/") {
		return "/" + s
	}

	return s
}

type cmdHandlerWithClient func(ctx *cli.Context, ctl *client.Client) error

func startDaemon(repoPath string, port int) (*client.Client, error) {
	exePath, err := godaemon.GetExecutablePath()
	if err != nil {
		return nil, err
	}

	// Start a new daemon process:
	log.Info("Starting daemon from: ", exePath)

	// TODO: Fill in correct password.
	proc := exec.Command(
		exePath, "-l", "/tmp/brig.log", "-x", "klaus", "daemon", "launch",
	)

	if err := proc.Start(); err != nil {
		log.Infof("Failed to start the daemon: %v", err)
		return nil, err
	}

	// This will likely suffice for most cases:
	time.Sleep(100 * time.Millisecond)

	for i := 0; i < 15; i++ {
		ctl, err := client.Dial(context.Background(), port)
		if err != nil {
			log.Infof("Waiting to bootup...")
			time.Sleep(500 * time.Millisecond)
			continue
		}

		return ctl, nil
	}

	return nil, fmt.Errorf("Daemon could not be started or took to long")
}

func withDaemon(handler cmdHandlerWithClient, startNew bool) func(*cli.Context) {
	// If not, make sure we start a new one:
	return withExit(func(ctx *cli.Context) error {
		port := guessPort()

		// Check if the daemon is running:
		ctl, err := client.Dial(context.Background(), port)
		if err == nil {
			return handler(ctx, ctl)
		}

		if !startNew {
			// Daemon was not running and we may not start a new one.
			return ExitCode{DaemonNotResponding, "Daemon not running"}
		}

		// Start the server & pass the password:
		ctl, err = startDaemon(guessRepoFolder(), port)
		if err != nil {
			return ExitCode{
				DaemonNotResponding,
				fmt.Sprintf("Unable to start daemon: %v", err),
			}
		}

		// Run the actual handler:
		return handler(ctx, ctl)
	})
}

type checkFunc func(ctx *cli.Context) int

func withArgCheck(checker checkFunc, handler func(*cli.Context)) func(*cli.Context) {
	return func(ctx *cli.Context) {
		if checker(ctx) != Success {
			os.Exit(BadArgs)
		}

		handler(ctx)
	}
}

func withExit(handler func(*cli.Context) error) func(*cli.Context) {
	return func(ctx *cli.Context) {
		if err := handler(ctx); err != nil {
			log.Error(err.Error())
			cerr, ok := err.(ExitCode)
			if !ok {
				os.Exit(UnknownError)
			}

			os.Exit(cerr.Code)
		}

		os.Exit(Success)
	}
}

func needAtLeast(min int) checkFunc {
	return func(ctx *cli.Context) int {
		if ctx.NArg() < min {
			if min == 1 {
				log.Warningf("Need at least %d argument.", min)
			} else {
				log.Warningf("Need at least %d arguments.", min)
			}
			cli.ShowCommandHelp(ctx, ctx.Command.Name)
			return BadArgs
		}

		return Success
	}
}

func guessPort() int {
	envPort := os.Getenv("BRIG_PORT")
	if envPort != "" {
		// Somebody tried to set BRIG_PORT.
		// Try to parse and spit errors if wrong.
		port, err := strconv.Atoi(envPort)
		if err != nil {
			log.Fatalf("Could not parse $BRIG_PORT: %v", err)
		}

		return port
	}

	// Guess the default port.
	log.Warning("BRIG_PORT not given, assuming :6666")
	return 6666
}
