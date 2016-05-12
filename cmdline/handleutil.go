package cmdline

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/disorganizer/brig/daemon"
	"github.com/disorganizer/brig/repo"
	repoconfig "github.com/disorganizer/brig/repo/config"
	pwdutil "github.com/disorganizer/brig/util/pwd"
	"github.com/olebedev/config"
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
	folder := repo.GuessFolder()
	if folder == "" {
		log.Errorf("This does not like a brig repository (missing .brig)")
		os.Exit(BadArgs)
	}

	return folder
}

func readPassword() (string, error) {
	repoFolder := guessRepoFolder()
	pwd, err := pwdutil.PromptPasswordMaxTries(4, func(pwd string) bool {
		err := repo.CheckPassword(repoFolder, pwd)
		return err == nil
	})

	return pwd, err
}

func prefixSlash(s string) string {
	if !strings.HasPrefix(s, "/") {
		return "/" + s
	}

	return s
}

type cmdHandlerWithClient func(ctx *cli.Context, client *daemon.Client) error

func withDaemon(handler cmdHandlerWithClient, startNew bool) func(*cli.Context) {
	// If not, make sure we start a new one:
	return withExit(func(ctx *cli.Context) error {
		port := guessPort()

		// Check if the daemon is running:
		client, err := daemon.Dial(port)
		if err == nil {
			return handler(ctx, client)
		}

		if !startNew {
			// Daemon was not running and we may not start a new one.
			return ExitCode{DaemonNotResponding, "Daemon not running"}
		}

		// Check if the password was supplied via a commandline flag.
		pwd := ctx.String("password")
		if pwd == "" {
			// Prompt the user:
			var cmdPwd string

			cmdPwd, err = readPassword()
			if err != nil {
				return ExitCode{
					BadPassword,
					fmt.Sprintf("Could not read password: %v", pwd),
				}
			}

			pwd = cmdPwd
		}

		// Start the dameon & pass the password:
		client, err = daemon.Reach(pwd, guessRepoFolder(), port)
		if err != nil {
			return ExitCode{
				DaemonNotResponding,
				fmt.Sprintf("Unable to start daemon: %v", err),
			}
		}

		// Run the actual handler:
		return handler(ctx, client)
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

func withConfig(handler func(*cli.Context, *config.Config) error) func(*cli.Context) error {
	return func(ctx *cli.Context) error {
		return handler(ctx, loadConfig())
	}
}

func loadConfig() *config.Config {
	// We do not use guessRepoFolder() here. It might abort
	folder := repo.GuessFolder()
	cfg, err := repoconfig.LoadConfig(filepath.Join(folder, ".brig", "config"))
	if err != nil {
		log.Warningf("Could not load config: %v", err)
		log.Warningf("Falling back on config defaults...")
		return repoconfig.CreateDefaultConfig()
	}

	return cfg
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

	// Trie the config elsewhise:
	config := loadConfig()
	port, err := config.Int("daemon.port")
	if err != nil {
		log.Fatalf("Cannot find out daemon port: %v", err)
	}

	return port
}
