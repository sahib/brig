package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/VividCortex/godaemon"
	"github.com/sahib/brig/client"
	"github.com/sahib/brig/cmd/pwd"
	"github.com/urfave/cli"
)

var (
	// backend delivers overly descriptive error messages including
	// the stirng below. Simply filter this info:
	rpcErrPattern = regexp.MustCompile(" server/capnp/api.capnp.* rpc exception:")
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

func readPasswordFromArgs(ctx *cli.Context) string {
	for curr := ctx; curr != nil; {
		if curr.Bool("no-password") {
			return "no-pass"
		}

		if password := curr.String("password"); password != "" {
			return password
		}

		curr = curr.Parent()
	}

	return ""
}

func readPassword(ctx *cli.Context, repoPath string) (string, error) {
	isInitialized, err := repoIsInitialized(repoPath)
	if err != nil {
		return "", err
	}

	if !isInitialized {
		return "", nil
	}

	// Try to read the password from -x or fallback to the default
	// password if requested by the --no-pass switch.
	if password := readPasswordFromArgs(ctx); password != "" {
		return password, nil
	}

	// Read the password from stdin:
	password, err := pwd.PromptPassword()
	if err != nil {
		return "", err
	}

	return password, nil
}

func prefixSlash(s string) string {
	if !strings.HasPrefix(s, "/") {
		return "/" + s
	}

	return s
}

type cmdHandlerWithClient func(ctx *cli.Context, ctl *client.Client) error

func startDaemon(ctx *cli.Context, repoPath string, port int) (*client.Client, error) {
	exePath, err := godaemon.GetExecutablePath()
	if err != nil {
		return nil, err
	}

	pwd, err := readPassword(ctx, repoPath)
	if err != nil {
		return nil, err
	}

	// Start a new daemon process:
	log.Info("No Daemon running. Starting daemon from binary: ", exePath)
	proc := exec.Command(
		exePath, "-p", pwd, "daemon", "launch",
	)

	if err := proc.Start(); err != nil {
		log.Infof("Failed to start the daemon: %v", err)
		return nil, err
	}

	// This will likely suffice for most cases:
	time.Sleep(200 * time.Millisecond)

	warningPrinted := false
	for i := 0; i < 15; i++ {
		ctl, err := client.Dial(context.Background(), port)
		if err != nil {
			// Only print this warning once...
			if !warningPrinted {
				log.Warnf("Waiting for daemon to bootup... :/")
				warningPrinted = true
			}
			time.Sleep(50 * time.Millisecond)
			continue
		}

		return ctl, nil
	}

	return nil, fmt.Errorf("Daemon could not be started or took to long. Wrong password maybe?")
}

func withDaemon(handler cmdHandlerWithClient, startNew bool) cli.ActionFunc {
	// If not, make sure we start a new one:
	// TODO: Make use of cli's error returning signatures.
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
		ctl, err = startDaemon(ctx, guessRepoFolder(), port)
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

func withArgCheck(checker checkFunc, handler cli.ActionFunc) cli.ActionFunc {
	return func(ctx *cli.Context) error {
		if checker(ctx) != Success {
			os.Exit(BadArgs)
		}

		return handler(ctx)
	}
}

func prettyPrintError(err error) string {
	return rpcErrPattern.ReplaceAllString(err.Error(), "")
}

func withExit(handler cli.ActionFunc) cli.ActionFunc {
	return func(ctx *cli.Context) error {
		if err := handler(ctx); err != nil {
			log.Error(prettyPrintError(err))
			cerr, ok := err.(ExitCode)
			if !ok {
				os.Exit(UnknownError)
			}

			os.Exit(cerr.Code)
		}

		os.Exit(Success)
		return nil
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

			if err := cli.ShowCommandHelp(ctx, ctx.Command.Name); err != nil {
				log.Warningf("Failed to display --help: %v", err)
			}

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
	return 6666
}

func repoIsInitialized(dir string) (bool, error) {
	fd, err := os.Open(dir)
	if err != nil {
		return true, err
	}

	names, err := fd.Readdirnames(-1)
	if err != nil {
		return true, err
	}

	for _, name := range names {
		switch name {
		case "meta.yml":
			fmt.Println("Meta exi")
			return true, nil
		case "logs":
			// That's okay.
		default:
			// Anything else we do not know:
			return true, nil
		}
	}

	// base case for empty dir:
	return false, nil
}

// tempFileWithSuffix works the same as ioutil.TempFile(),
// but allows for the addition of a suffix to the filepath.
// This has the nice side effect that some editors can recognize
// the filetype based on the ending and provide you syntax highlighting.
// (this is used in edit() below)
func tempFileWithSuffix(dir, prefix, suffix string) (f *os.File, err error) {
	if dir == "" {
		dir = os.TempDir()
	}

	for i := 0; i < 10000; i++ {
		mid := strconv.Itoa(rand.Int())
		name := filepath.Join(dir, prefix+mid+suffix)
		f, err = os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
		if os.IsExist(err) {
			continue
		}
		break
	}
	return
}

// editToPath opens up $EDITOR with `data` and saves the edited data
// to a temporary path that is then returned.
func editToPath(data []byte, suffix string) (string, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		// It makes my heart bleed, but assume that vi is too hard
		// for the majority I've met & that might use brig.
		editor = "nano"
	}

	fd, err := tempFileWithSuffix("", "brig-cmd-buffer-", suffix)
	if err != nil {
		return "", err
	}

	doDelete := false

	// Make sure it gets cleaned up.
	defer func() {
		if doDelete {
			if err := os.Remove(fd.Name()); err != nil {
				fmt.Printf("Failed to remove temp file: %v\n", err)
			}
		}

		if err := fd.Close(); err != nil {
			fmt.Printf("Failed to close file: %v\n", err)
		}
	}()

	if _, err := fd.Write(data); err != nil {
		return "", err
	}

	// Launch editor and hook it up with all necessary fds:
	cmd := exec.Command(editor, fd.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		doDelete = true
		return "", fmt.Errorf("Running $EDITOR (%s) failed: %v", editor, err)
	}

	if _, err := fd.Seek(0, os.SEEK_SET); err != nil {
		doDelete = true
		return "", err
	}

	return fd.Name(), nil
}

// edit opens up $EDITOR with `data` and returns the edited data.
func edit(data []byte, suffix string) ([]byte, error) {
	tempPath, err := editToPath(data, suffix)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := os.Remove(tempPath); err != nil {
			fmt.Printf("Failed to remove temp file: %v\n", err)
		}
	}()

	newData, err := ioutil.ReadFile(tempPath)
	if err != nil {
		return nil, err
	}

	// Some editors might add a trailing newline:
	return bytes.TrimRight(newData, "\n"), nil
}
