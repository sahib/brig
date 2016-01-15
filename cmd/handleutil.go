package cmdline

import (
	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/daemon"
	"github.com/tucnak/climax"
)

type CmdHandlerWithClient func(ctx climax.Context, client *daemon.Client) int

func withDaemon(handler CmdHandlerWithClient) climax.CmdHandler {
	// Check if the daemon is running:
	client, err := daemon.Dial(6666)
	if err == nil {
		return func(ctx climax.Context) int {
			defer client.Close()
			return handler(ctx, client)
		}
	}

	// If not, make sure we start a new one:
	return func(ctx climax.Context) int {
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
		client, err := daemon.Reach(pwd, guessRepoFolder(), 6666)
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
