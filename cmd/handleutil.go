package cmdline

import (
	"os"
	"path/filepath"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/daemon"
	"github.com/disorganizer/brig/repo"
	repoconfig "github.com/disorganizer/brig/repo/config"
	"github.com/olebedev/config"
	"github.com/tucnak/climax"
)

type CmdHandlerWithClient func(ctx climax.Context, client *daemon.Client) int

func withDaemon(handler CmdHandlerWithClient, startNew bool) climax.CmdHandler {
	// If not, make sure we start a new one:
	return func(ctx climax.Context) int {
		port := guessPort()

		// Check if the daemon is running:
		client, err := daemon.Dial(port)
		if err == nil {
			defer client.Close()
			return handler(ctx, client)
		}

		if !startNew {
			// Daemon was not running and we may not start a new one.
			log.Warning("Daemon not running.")
			return DaemonNotResponding
		}

		// Check if the password was supplied via a commandline flag.
		pwd, ok := ctx.Get("password")
		if !ok {
			// Prompt the user:
			cmdPwd, err := readPassword()
			if err != nil {
				log.Errorf("Could not read password: %v", pwd)
				return BadPassword
			}

			pwd = cmdPwd
		}

		// Start the dameon & pass the password:
		client, err = daemon.Reach(pwd, guessRepoFolder(), port)
		if err != nil {
			log.Errorf("Unable to start daemon: %v", err)
			return DaemonNotResponding
		}

		// Run the actual handler:
		return handler(ctx, client)
	}
}

type CheckFunc func(ctx climax.Context) int

func withArgCheck(checker CheckFunc, handler climax.CmdHandler) climax.CmdHandler {
	return func(ctx climax.Context) int {
		if checker(ctx) != Success {
			return BadArgs
		}

		return handler(ctx)
	}
}

func needAtLeast(min int) CheckFunc {
	return func(ctx climax.Context) int {
		if len(ctx.Args) < min {
			log.Warningf("Need at least %d arguments.", min)
			return BadArgs
		}

		return Success
	}
}

func loadConfig() *config.Config {
	// We do not use guessRepoFolder() here. It might abort
	folder := repo.GuessFolder()
	cfg, err := repoconfig.LoadConfig(filepath.Join(folder, "config"))
	if err != nil {
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
