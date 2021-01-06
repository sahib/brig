package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/fatih/color"
	"github.com/mitchellh/go-homedir"
	"github.com/sahib/brig/client"
	"github.com/sahib/brig/cmd/pwd"
	"github.com/sahib/brig/defaults"
	"github.com/sahib/brig/util"
	"github.com/sahib/brig/util/hashlib"
	"github.com/sahib/brig/util/pwutil"
	"github.com/sahib/config"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	// backend delivers overly descriptive error messages including
	// the string below. Simply filter this info:
	rpcErrPattern = regexp.MustCompile(`\s*server/capnp/local_api.capnp.*rpc exception:\s*`)
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

// guessRepoFolder tries to find the repository path by using a number of
// sources. This helper may call exit when it fails to get the path.
func guessRepoFolder(ctx *cli.Context) (string, error) {
	if ctx.GlobalIsSet("repo") {
		// No guessing needed, follow user wish.
		return ctx.GlobalString("repo"), nil
	}

	guessLocations := []string{
		".",
	}

	home, err := homedir.Dir()
	if err == nil {
		guessLocations = append(guessLocations, []string{
			// TODO: figure out good default locations.
			filepath.Join(home, ".brig"),
			filepath.Join(home, ".cache/brig"),
		}...)
	}

	var lastError error
	for _, guessLocation := range guessLocations {
		repoFolder := mustAbsPath(guessLocation)
		if _, err := os.Stat(filepath.Join(repoFolder, "OWNER")); err != nil {
			lastError = err
			continue
		}

		return repoFolder, nil
	}

	return "", lastError
}

func openConfig(folder string) (*config.Config, error) {
	configPath := filepath.Join(folder, "config.yml")
	cfg, err := defaults.OpenMigratedConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("could not find config: %v", err)
	}

	return cfg, nil
}

func guessDaemonURL(ctx *cli.Context) (string, error) {
	if ctx.GlobalIsSet("url") {
		// No guessing needed, follow user wish.
		return ctx.GlobalString("url"), nil
	}

	folder, err := guessRepoFolder(ctx)
	if err != nil {
		log.Warnf("note: I don't know where the repository is or cannot read it.")
		log.Warnf("      I will continue with default values, cross fingers.")
		log.Warnf("      We recommend to set BRIG_PATH or pass --repo always.")
		log.Warnf("      Alternatively you can cd to your repository.")
		return ctx.GlobalString("url"), err
	}

	cfg, err := openConfig(folder)
	if err != nil {
		// Assume default:
		return ctx.GlobalString("url"), nil
	}

	return cfg.String("daemon.url"), nil
}

func guessFreeDaemonURL(ctx *cli.Context, owner string) (string, error) {
	if ctx.GlobalIsSet("url") {
		// No guessing needed, follow user wish.
		return ctx.GlobalString("url"), nil
	}

	defaultURL := defaults.DaemonDefaultURL()
	u, err := url.Parse(defaultURL)
	if err != nil {
		// this is a programming error
		panic("invalid hardcoded default daemon url")
	}

	switch u.Scheme {
	case "unix":
		// TODO: Use the owner in clear text.
		// TODO: Remove socket if it's still lying around.
		return fmt.Sprintf(
			"%s.%s",
			defaultURL,
			hashlib.Sum([]byte(owner)).B58String(),
		), nil
	case "tcp":
		// Do a best effort by searching for a free port
		// and use that for the brig repository.
		// This might be racy, but at least try it.
		port := util.FindFreePort()
		return fmt.Sprintf("tcp://127.0.0.1:%d", port), nil
	default:
		return "", fmt.Errorf("default url has unknown ")
	}
}

func readPasswordFromArgs(basePath string, ctx *cli.Context) string {
	if ctx.Bool("no-password") {
		return "no-password"
	}

	if pwHelper := ctx.String("pw-helper"); pwHelper != "" {
		password, err := pwutil.ReadPasswordFromHelper(basePath, pwHelper)

		if err == nil {
			return password
		}

		logVerbose(ctx, "failed to read password from '%s': %v\n", pwHelper, err)
	}

	// Note: the "--no-password" switch of init is handled by
	// setting a password command that echoes a static password.
	for curr := ctx; curr != nil; {
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

func startDaemon(ctx *cli.Context, repoPath, daemonURL string) (*client.Client, error) {
	stat, err := os.Stat(repoPath)
	if err != nil {
		return nil, err
	}

	if !stat.IsDir() {
		return nil, fmt.Errorf("»%s« is not a directory", repoPath)
	}

	exePath, err := getExecutablePath()
	if err != nil {
		return nil, err
	}

	logVerbose(ctx, "using executable path: %s", exePath)

	// If a password helper is configured, we should not ask the password right here.
	askPassword := true
	cfg, err := defaults.OpenMigratedConfig(
		filepath.Join(repoPath, "config.yml"),
	)

	if err != nil {
		logVerbose(ctx, "failed to open config for guessing password method: %v", err)
	} else {
		if cfg.String("repo.password_command") != "" {
			askPassword = false
		}
	}

	logVerbose(
		ctx,
		"No Daemon running at %s. Starting daemon from binary: %s",
		daemonURL,
		exePath,
	)

	daemonArgs := []string{
		"--repo", repoPath,
		"--url", daemonURL,
		"daemon", "launch",
	}

	argString := fmt.Sprintf("'%s'", strings.Join(daemonArgs, "' '"))
	logVerbose(ctx, "Starting daemon as: %s %s", exePath, argString)

	proc := exec.Command(exePath, daemonArgs...) // #nosec

	if askPassword {
		logVerbose(ctx, "asking password since no password command was given")
		pwd, err := readPassword(ctx, repoPath)
		if err != nil {
			return nil, err
		}

		if len(pwd) != 0 {
			proc.Env = append(proc.Env, fmt.Sprintf("BRIG_PASSWORD=%s", pwd))
		}
	}

	proc.Env = append(proc.Env, fmt.Sprintf("PATH=%s", os.Getenv("PATH")))

	if err := proc.Start(); err != nil {
		log.Infof("Failed to start the daemon: %v", err)
		return nil, err
	}

	// This will likely suffice for most cases:
	time.Sleep(500 * time.Millisecond)

	warningPrinted := false
	for i := 0; i < 500; i++ {
		ctl, err := client.Dial(context.Background(), daemonURL)
		if err != nil {
			// Only print this warning once...
			if !warningPrinted && i >= 100 {
				log.Warnf("waiting a bit long for daemon to bootup...")
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
	return func(ctx *cli.Context) error {
		daemonURL, _ := guessDaemonURL(ctx)

		if startNew {
			logVerbose(ctx, "using url %s to check for running daemon.", daemonURL)
		} else {
			logVerbose(ctx, "using url %s to connect to existing daemon.", daemonURL)
		}

		// Check if the daemon is running already:
		ctl, err := client.Dial(context.Background(), daemonURL)
		if err == nil {
			defer ctl.Close()
			return handler(ctx, ctl)
		}

		if !startNew {
			// Daemon was not running and we may not start a new one.
			return ExitCode{DaemonNotResponding, "Daemon not running"}
		}

		// Start the server & pass the password:
		folder, err := guessRepoFolder(ctx)
		if err != nil {
			return ExitCode{
				BadArgs,
				fmt.Sprintf("could not guess folder: %v", err),
			}
		}

		logVerbose(ctx, "starting new daemon in background, on folder '%s'", folder)

		ctl, err = startDaemon(ctx, folder, daemonURL)
		if err != nil {
			return ExitCode{
				DaemonNotResponding,
				fmt.Sprintf("Unable to start daemon: %v", err),
			}
		}

		// Run the actual handler:
		defer ctl.Close()
		return handler(ctx, ctl)
	}
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
	fd, err := os.Open(dir) // #nosec
	if err != nil && os.IsNotExist(err) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	names, err := fd.Readdirnames(-1)
	if err != nil {
		return false, err
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
		mid := strconv.Itoa(rand.Int()) // #nosec
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
	cmd := exec.Command(editor, fd.Name()) // #nosec
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

	newData, err := ioutil.ReadFile(tempPath) // #nosec
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

func readFormatTemplate(ctx *cli.Context) (*template.Template, error) {
	if ctx.IsSet("format") {
		source := ctx.String("format") + "\n"
		tmpl, err := template.New("format").Parse(source)

		if err != nil {
			return nil, err
		}

		return tmpl, nil
	}

	return nil, nil
}

func pinStateToSymbol(isPinned, isExplicit bool) string {
	if isPinned {
		colorFn := color.CyanString
		if isExplicit {
			colorFn = color.MagentaString
		}

		return colorFn("✔")
	}

	return ""
}

func yesOrNo(v bool) string {
	if v {
		return color.GreenString("yes")
	}

	return color.RedString("no")
}

type logWriter struct{ prefix string }

func (lw *logWriter) Write(buf []byte) (int, error) {
	log.Infof("%s: %s", lw.prefix, string(bytes.TrimSpace(buf)))
	return len(buf), nil
}
