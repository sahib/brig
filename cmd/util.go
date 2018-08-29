package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/fatih/color"
	"github.com/sahib/brig/client"
	"github.com/sahib/brig/cmd/pwd"
	"github.com/sahib/brig/defaults"
	"github.com/sahib/brig/repo"
	"github.com/sahib/brig/util/pwutil"
	"github.com/urfave/cli"
)

var (
	// backend delivers overly descriptive error messages including
	// the string below. Simply filter this info:
	rpcErrPattern = regexp.MustCompile(`\s*server/capnp/api.capnp.*rpc exception:\s*`)
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

func getRepoEntryFromRegistry() (*repo.RegistryEntry, error) {
	registry, err := repo.OpenRegistry()
	if err != nil {
		return nil, err
	}

	entries, err := registry.List()
	if err != nil {
		return nil, err
	}

	// Shortcut: If there's only one repo, always connect to that.
	if len(entries) == 1 {
		return entries[0], nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// Scan twice.
	// First: Check if we're in some repo:
	for _, entry := range entries {
		if entry.Path == cwd {
			return entry, nil
		}
	}

	// Second: Check for defaults:
	for _, entry := range entries {
		if entry.IsDefault {
			return entry, nil
		}
	}

	return nil, fmt.Errorf("no suitable registry entry found")
}

func mustAbsPath(path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		fmt.Printf("Failed to get absolute repo path: %v", err)
		os.Exit(1)
	}

	return absPath
}

func yesify(val bool) string {
	if val {
		return color.GreenString("yes")
	}

	return color.RedString("no")
}

func checkmarkify(val bool) string {
	if val {
		return color.GreenString("✔")
	}

	return ""
}

// guessRepoFolder tries to find the repository path
// by using a number of sources.
// This helper may call exit when it fails to get the path.
func guessRepoFolder(lookupGlobal bool) string {
	envPath := os.Getenv("BRIG_PATH")
	if envPath != "" {
		return mustAbsPath(envPath)
	}

	if lookupGlobal {
		entry, err := getRepoEntryFromRegistry()
		if err == nil {
			return mustAbsPath(entry.Path)
		}

		fmt.Printf("Failed to get path from registry: %v\n", err)
	}

	cwdPath, err := os.Getwd()
	if err != nil {
		fmt.Printf("Failed to get current working dir: %v; aborting.", err)
		os.Exit(1)
	}

	return mustAbsPath(cwdPath)
}

func guessNextFreePort(ctx *cli.Context) (int, error) {
	if ctx.GlobalIsSet("port") {
		return ctx.GlobalInt("port"), nil
	}

	registry, err := repo.OpenRegistry()
	if err != nil {
		return 0, err
	}

	entries, err := registry.List()
	if err != nil {
		return 0, err
	}

	maxPort := 0
	if len(entries) == 0 {
		// Use the default value.
		maxPort = ctx.GlobalInt("port")
	} else {
		for _, entry := range entries {
			if int(entry.Port) > maxPort {
				maxPort = int(entry.Port)
			}
		}

		// Always start checking with at least
		// the next higher port.
		maxPort += 1
	}

	maxAttempts := 1000
	for off := 0; off <= maxAttempts; off++ {
		conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", maxPort+off))
		if err != nil {
			return maxPort + off, nil
		}

		conn.Close()
	}

	return 0, fmt.Errorf(
		"failed to find next free port after %d attempts",
		maxAttempts,
	)
}

func guessPort(ctx *cli.Context) int {
	if ctx.GlobalIsSet("port") {
		return ctx.GlobalInt("port")
	}

	entry, err := getRepoEntryFromRegistry()
	if err == nil && entry.Port > 0 {
		logVerbose(ctx, "found port from global registry: %d", entry.Port)
		return int(entry.Port)
	}

	port := ctx.GlobalInt("port")
	logVerbose(ctx, "using port from global --port option: %d", port)
	return port
}

func readPasswordFromArgs(basePath string, ctx *cli.Context) string {
	if pwHelper := ctx.String("pw-helper"); pwHelper != "" {
		password, err := pwutil.ReadPasswordFromHelper(basePath, pwHelper)

		if err == nil {
			return password
		}

		logVerbose(ctx, "failed to read password from '%s': %v\n", pwHelper, err)
	}

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
		logVerbose(ctx, "repository is not initialized, skipping password entry")
		return "", nil
	}

	// Try to read the password from -x or fallback to the default
	// password if requested by the --no-pass switch.
	if password := readPasswordFromArgs(repoPath, ctx); password != "" {
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

func getExecutablePath() (string, error) {
	// NOTE: This might not work on other platforms.
	//       In this case we fall back to LookPath().
	exePath, err := os.Readlink("/proc/self/exe")
	if err != nil {
		return exec.LookPath("brig")
	}

	return filepath.Clean(exePath), nil
}

func startDaemon(ctx *cli.Context, repoPath string, port int) (*client.Client, error) {
	exePath, err := getExecutablePath()
	if err != nil {
		return nil, err
	}

	logVerbose(ctx, "using executable path: %s", exePath)

	// If a password helper is configured, we should not ask the password right here.
	askPassword := true
	cfg, err := defaults.OpenMigratedConfig(filepath.Join(repoPath, "config.yml"))
	if err != nil {
		logVerbose(ctx, "failed to open config for guessing password method: %v", err)
	} else {
		if cfg.String("repo.password_command") != "" {
			askPassword = false
		}
	}

	bindHost := ctx.GlobalString("bind")

	log.Infof(
		"No Daemon running at %s:%d. Starting daemon from binary: %s",
		bindHost,
		port,
		exePath,
	)

	daemonArgs := []string{}
	if askPassword {
		logVerbose(ctx, "asking password since no password command was given")
		pwd, err := readPassword(ctx, repoPath)
		if err != nil {
			return nil, err
		}

		// TODO: Do not pass the password via a cmdline argument.
		// Use a environment variable.
		if len(pwd) != 0 {
			daemonArgs = append(daemonArgs, "--password", pwd)
		}
	}

	daemonArgs = append(daemonArgs, []string{
		"--port", strconv.FormatInt(int64(port), 10),
		"--bind", bindHost,
		"daemon", "launch",
	}...)

	argString := fmt.Sprintf("'%s'", strings.Join(daemonArgs, "' '"))
	logVerbose(ctx, "Starting daemon as: %s %s", exePath, argString)

	proc := exec.Command(exePath, daemonArgs...)
	if err := proc.Start(); err != nil {
		log.Infof("Failed to start the daemon: %v", err)
		return nil, err
	}

	// This will likely suffice for most cases:
	time.Sleep(300 * time.Millisecond)

	warningPrinted := false
	for i := 0; i < 25; i++ {
		ctl, err := client.Dial(context.Background(), port)
		if err != nil {
			// Only print this warning once...
			if !warningPrinted && i >= 10 {
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

func guessNextRepoFolder(ctx *cli.Context) string {
	absify := func(path string) string {
		absPath, err := filepath.Abs(path)
		if err != nil {
			// Better stop, before we do something stupid:
			log.Errorf("failed to get absolute path: %s", path)
			os.Exit(1)
		}

		return absPath
	}

	if folder := ctx.Args().First(); len(folder) > 0 {
		return absify(folder)
	}

	if folder := ctx.GlobalString("path"); len(folder) > 0 {
		return absify(folder)
	}

	folder, err := os.Getwd()
	if err != nil {
		log.Warningf("failed to get current working dir: %s", folder)
		return "."
	}

	return folder
}

func withDaemonAlways(handler cmdHandlerWithClient) cli.ActionFunc {
	return withExit(func(ctx *cli.Context) error {
		port, err := guessNextFreePort(ctx)
		if err != nil {
			return err
		}

		logVerbose(ctx, "using port %d for new daemon.", port)

		folder := guessNextRepoFolder(ctx)
		logVerbose(ctx, "using path %s for new daemon.", folder)

		// Start the server & pass the password:
		ctl, err := startDaemon(ctx, folder, port)
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

func withDaemon(handler cmdHandlerWithClient, startNew bool) cli.ActionFunc {
	// If not, make sure we start a new one:
	return withExit(func(ctx *cli.Context) error {
		port := guessPort(ctx)

		if startNew {
			logVerbose(ctx, "using port %d to check for running daemon.", port)
		} else {
			logVerbose(ctx, "using port %d to connect to old daemon.", port)
		}

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
		folder := guessRepoFolder(true)
		logVerbose(ctx, "starting new daemon in background on folder %s", folder)

		ctl, err = startDaemon(ctx, folder, port)
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
	return rpcErrPattern.ReplaceAllString(err.Error(), " ")
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

func repoIsInitialized(dir string) (bool, error) {
	fd, err := os.Open(dir)
	if err != nil && os.IsNotExist(err) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	names, err := fd.Readdirnames(-1)
	if err != nil {
		return true, err
	}

	return len(names) >= 1, nil
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

	if _, err := fd.Seek(0, io.SeekStart); err != nil {
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

// parseDuration tries to convert the string `s` to
// a duration in seconds (+ fractions).
// It uses time.ParseDuration() internally, but allows
// whole numbers which are counted as seconds.
func parseDuration(s string) (float64, error) {
	sec, err := strconv.ParseFloat(s, 64)
	if err == nil {
		return sec, nil
	}

	dur, err := time.ParseDuration(s)
	if err != nil {
		return 0.0, err
	}

	return float64(dur) / float64(time.Second), nil
}
