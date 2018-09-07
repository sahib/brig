package cmd

import (
	"fmt"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/fatih/color"
	isatty "github.com/mattn/go-isatty"
	formatter "github.com/sahib/brig/util/log"
	"github.com/sahib/brig/version"
	"github.com/urfave/cli"
)

func init() {
	log.SetOutput(os.Stderr)
	log.SetLevel(log.DebugLevel)

	// Only use fancy logging if we print to a terminal:
	if isatty.IsTerminal(os.Stdout.Fd()) {
		log.SetFormatter(&formatter.FancyLogFormatter{
			UseColors: true,
		})
	}
}

func formatGroup(category string) string {
	return strings.ToUpper(category) + " COMMANDS"
}

////////////////////////////
// Commandline definition //
////////////////////////////

// RunCmdline starts a brig commandline tool.
func RunCmdline(args []string) int {
	app := cli.NewApp()
	app.Name = "brig"
	app.Usage = "Secure and dezentralized file synchronization"
	app.EnableBashCompletion = true
	app.Version = fmt.Sprintf(
		"%s [buildtime: %s] (client version)",
		version.String(),
		version.BuildTime,
	)
	app.CommandNotFound = commandNotFound
	app.Description = "brig can be used to securely store, version and synchronize files between many peers."

	// Set global options here:
	app.Before = func(ctx *cli.Context) error {
		if ctx.Bool("no-color") {
			color.NoColor = true
		}

		return nil
	}

	// Groups:
	repoGroup := formatGroup("repository")
	wdirGroup := formatGroup("working tree")
	vcscGroup := formatGroup("version control")
	netwGroup := formatGroup("network")

	// Autocomplete all commands, but not their aliases.
	app.BashComplete = func(ctx *cli.Context) {
		for _, cmd := range app.Commands {
			fmt.Println(cmd.Name)
		}
	}

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "nodaemon,n",
			Usage: "Don't start the daemon automatically",
		},
		cli.BoolFlag{
			Name:  "no-password,x",
			Usage: "Use 'no-pass' as password",
		},
		cli.StringFlag{
			Name:   "bind",
			Usage:  "To what host to bind to (default: localhost)",
			Value:  "localhost",
			EnvVar: "BRIG_BIND",
		},
		cli.IntFlag{
			Name:   "port,p",
			Usage:  "On what port the daemon listens on (default: 6666)",
			EnvVar: "BRIG_PORT",
			Value:  6666,
		},
		cli.BoolFlag{
			Name:  "no-color,",
			Usage: "Use 'no-pass' as password",
		},
		cli.StringFlag{
			Name:  "password,P",
			Usage: "Supply user password. Usage is not recommended.",
			Value: "",
		},
		cli.BoolFlag{
			Name:  "verbose,V",
			Usage: "Show certain messages during client startup (helpful for debugging)",
		},
		cli.StringFlag{
			Name:   "path",
			Usage:  "Path of the repository",
			Value:  "",
			EnvVar: "BRIG_PATH",
		},
	}

	app.Commands = TranslateHelp([]cli.Command{
		{
			Name:     "init",
			Category: repoGroup,
			Action:   withArgCheck(needAtLeast(1), withDaemonAlways(handleInit)),
		}, {
			Name:     "whoami",
			Category: netwGroup,
			Action:   withDaemon(handleWhoami, true),
		}, {
			Name:     "remote",
			Aliases:  []string{"rmt"},
			Category: netwGroup,
			Subcommands: []cli.Command{
				{
					Name:   "add",
					Action: withArgCheck(needAtLeast(2), withDaemon(handleRemoteAdd, true)),
				}, {
					Name:    "remove",
					Aliases: []string{"rm"},
					Action:  withArgCheck(needAtLeast(1), withDaemon(handleRemoteRemove, true)),
				}, {
					Name:    "list",
					Aliases: []string{"ls"},
					Action:  withDaemon(handleRemoteList, true),
				}, {
					Name:   "clear",
					Action: withDaemon(handleRemoteClear, true),
				}, {
					Name:   "edit",
					Action: withDaemon(handleRemoteEdit, true),
				}, {
					Name:   "ping",
					Action: withArgCheck(needAtLeast(1), withDaemon(handleRemotePing, true)),
				},
			},
		}, {
			Name:     "pin",
			Category: vcscGroup,
			Action:   withArgCheck(needAtLeast(1), withDaemon(handlePin, true)),
			Subcommands: []cli.Command{
				{
					Name:   "add",
					Action: withArgCheck(needAtLeast(1), withDaemon(handlePin, true)),
				}, {
					Name:    "remove",
					Aliases: []string{"rm"},
					Action:  withArgCheck(needAtLeast(1), withDaemon(handleUnpin, true)),
				}, {
					Name:   "set",
					Action: withDaemon(handlePinSet, true),
				}, {
					Name:   "clear",
					Action: withDaemon(handlePinClear, true),
				}, {
					Name:    "list",
					Aliases: []string{"ls"},
					Action:  withDaemon(handlePinList, true),
				},
			},
		}, {
			Name:     "net",
			Category: netwGroup,
			Subcommands: []cli.Command{
				{
					Name:   "offline",
					Action: withDaemon(handleOffline, true),
				}, {
					Name:   "online",
					Action: withDaemon(handleOnline, true),
				}, {
					Name:   "status",
					Action: withDaemon(handleIsOnline, true),
				}, {
					Name:   "locate",
					Action: withArgCheck(needAtLeast(1), withDaemon(handleNetLocate, true)),
				},
			},
		}, {
			Name:     "status",
			Aliases:  []string{"st"},
			Category: vcscGroup,
			Action:   withDaemon(handleStatus, true),
		}, {
			Name:     "diff",
			Category: vcscGroup,
			Action:   withDaemon(handleDiff, true),
		}, {
			Name:     "tag",
			Category: vcscGroup,
			Action:   withArgCheck(needAtLeast(1), withDaemon(handleTag, true)),
		}, {
			Name:     "log",
			Category: vcscGroup,
			Action:   withDaemon(handleLog, true),
		}, {
			Name:     "fetch",
			Category: vcscGroup,
			Action:   withArgCheck(needAtLeast(1), withDaemon(handleFetch, true)),
		}, {
			// TODO: option to auto-download (parts of?) the synced result.
			Name:     "sync",
			Category: vcscGroup,
			Action:   withArgCheck(needAtLeast(1), withDaemon(handleSync, true)),
		}, {
			// TODO: Do re-pinning of old files only after a commit (to allow safe jump backs)
			Name:     "commit",
			Aliases:  []string{"cmt"},
			Category: vcscGroup,
			Action:   withDaemon(handleCommit, true),
		}, {
			// TODO: Figure out/test exact way of pinning and write docs for it.
			Name:     "reset",
			Aliases:  []string{"re"},
			Category: vcscGroup,
			Action:   withArgCheck(needAtLeast(1), withDaemon(handleReset, true)),
		}, {
			Name:     "become",
			Category: vcscGroup,
			Action:   withDaemon(handleBecome, true),
		}, {
			Name:     "history",
			Aliases:  []string{"hst", "hist"},
			Category: vcscGroup,
			Action:   withArgCheck(needAtLeast(1), withDaemon(handleHistory, true)),
		}, {
			Name:     "stage",
			Aliases:  []string{"stg", "add", "a"},
			Category: wdirGroup,
			Action:   withArgCheck(needAtLeast(1), withDaemon(handleStage, true)),
		}, {
			Name:     "touch",
			Aliases:  []string{"t"},
			Category: wdirGroup,
			Action:   withArgCheck(needAtLeast(1), withDaemon(handleTouch, true)),
		}, {
			Name:     "cat",
			Category: wdirGroup,
			Action:   withArgCheck(needAtLeast(1), withDaemon(handleCat, true)),
		}, {
			Name:     "info",
			Category: wdirGroup,
			Action:   withArgCheck(needAtLeast(1), withDaemon(handleInfo, true)),
		}, {
			Name:     "rm",
			Aliases:  []string{"remove"},
			Category: wdirGroup,
			Action:   withArgCheck(needAtLeast(1), withDaemon(handleRm, true)),
		}, {
			Name:     "ls",
			Category: wdirGroup,
			Action:   withDaemon(handleList, true),
		}, {
			Name:     "tree",
			Category: wdirGroup,
			Action:   withDaemon(handleTree, true),
		}, {
			Name:     "mkdir",
			Category: wdirGroup,
			Action:   withArgCheck(needAtLeast(1), withDaemon(handleMkdir, true)),
		}, {
			Name:     "mv",
			Category: wdirGroup,
			Action:   withArgCheck(needAtLeast(2), withDaemon(handleMv, true)),
		}, {
			Name:     "cp",
			Category: wdirGroup,
			Action:   withArgCheck(needAtLeast(2), withDaemon(handleCp, true)),
		}, {
			Name:     "edit",
			Category: wdirGroup,
			Action:   withArgCheck(needAtLeast(1), withDaemon(handleEdit, true)),
		}, {
			Name:     "daemon",
			Category: repoGroup,
			Subcommands: []cli.Command{
				{
					Name:   "launch",
					Action: withExit(handleDaemonLaunch),
				}, {
					Name:   "quit",
					Action: withDaemon(handleDaemonQuit, false),
				}, {
					Name:   "ping",
					Action: withDaemon(handleDaemonPing, false),
				},
			},
		}, {
			Name:     "config",
			Aliases:  []string{"cfg"},
			Category: repoGroup,
			Action:   withDaemon(handleConfigList, true),
			Subcommands: []cli.Command{
				{
					Name:    "list",
					Aliases: []string{"ls"},
					Action:  withDaemon(handleConfigList, true),
				}, {
					Name:   "get",
					Action: withArgCheck(needAtLeast(1), withDaemon(handleConfigGet, true)),
				}, {
					Name:   "doc",
					Action: withArgCheck(needAtLeast(1), withDaemon(handleConfigDoc, true)),
				}, {
					Name:   "set",
					Action: withArgCheck(needAtLeast(2), withDaemon(handleConfigSet, true)),
				},
			},
		}, {
			Name:     "fstab",
			Category: repoGroup,
			Action:   withArgCheck(needAtLeast(0), withDaemon(handleFstabList, true)),
			Subcommands: []cli.Command{
				{
					Name:   "add",
					Action: withArgCheck(needAtLeast(2), withDaemon(handleFstabAdd, true)),
				}, {
					Name:    "remove",
					Aliases: []string{"rm"},
					Action:  withArgCheck(needAtLeast(1), withDaemon(handleFstabRemove, true)),
				}, {
					Name:   "apply",
					Action: withArgCheck(needAtLeast(0), withDaemon(handleFstabApply, true)),
				}, {
					Name:    "list",
					Aliases: []string{"ls"},
					Action:  withArgCheck(needAtLeast(0), withDaemon(handleFstabList, true)),
				},
			},
		}, {
			Name:     "mount",
			Category: repoGroup,
			Action:   withDaemon(handleMount, true),
		}, {
			Name:     "unmount",
			Category: repoGroup,
			Action:   withDaemon(handleUnmount, true),
		}, {
			Name:     "version",
			Category: repoGroup,
			Action:   withDaemon(handleVersion, false),
		}, {
			Name:     "gc",
			Category: repoGroup,
			Action:   withDaemon(handleGc, true),
		}, {
			Name:   "docs",
			Action: handleOpenHelp,
			Hidden: true,
		}, {
			Name:   "bug",
			Action: handleBugReport,
		},
	})

	if err := app.Run(args); err != nil {
		return 1
	}
	return 0
}
