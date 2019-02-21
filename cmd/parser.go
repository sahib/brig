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

	var useColor bool
	switch envVar := os.Getenv("BRIG_COLOR"); envVar {
	case "", "auto":
		useColor = isatty.IsTerminal(os.Stdout.Fd())
	case "never":
		useColor = false
	case "always":
		useColor = true
	default:
		log.Warningf("Bad value for $BRIG_COLOR: %s, disabling color", envVar)
		useColor = false
	}

	// Only use fancy logging if we print to a terminal:
	log.SetFormatter(&formatter.FancyLogFormatter{
		UseColors: useColor,
	})
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
	app.Usage = "Secure and decentralized file synchronization"
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
		cli.IntFlag{
			Name:   "port,p",
			Usage:  "Port of the daemon to connect to.",
			EnvVar: "BRIG_PORT",
			Value:  6666,
		},
		cli.StringFlag{
			Name:   "repo",
			Usage:  "Path to the repository. Only has effect for new daemons.",
			Value:  "",
			EnvVar: "BRIG_PATH",
		},
		cli.BoolFlag{
			Name:  "verbose,V",
			Usage: "Show certain messages during client startup (helpful for debugging)",
		},
		cli.StringFlag{
			Name:   "bind",
			Usage:  "To what host to bind to. Do not expose to the outside. Seriously.",
			Value:  "localhost",
			EnvVar: "BRIG_BIND",
		},
		cli.StringFlag{
			Name:   "password,P",
			Usage:  "Supply user password. Usage is not recommended.",
			EnvVar: "BRIG_PASSWORD",
			Value:  "",
		},
		cli.BoolFlag{
			Name:  "nodaemon,n",
			Usage: "Don't start the daemon automatically.",
		},
		cli.BoolFlag{
			Name:  "no-color",
			Usage: "Forbid the usage of colors.",
		},
	}

	app.Commands = TranslateHelp([]cli.Command{
		{
			Name:     "init",
			Category: repoGroup,
			Action:   withArgCheck(needAtLeast(1), withDaemon(handleInit, true, false)),
		}, {
			Name:     "whoami",
			Category: netwGroup,
			Action:   withDaemon(handleWhoami, true, true),
		}, {
			Name:     "remote",
			Aliases:  []string{"rmt"},
			Category: netwGroup,
			Subcommands: []cli.Command{
				{
					Name:    "add",
					Aliases: []string{"a"},
					Action:  withArgCheck(needAtLeast(2), withDaemon(handleRemoteAdd, true, true)),
				}, {
					Name:    "remove",
					Aliases: []string{"rm"},
					Action:  withArgCheck(needAtLeast(1), withDaemon(handleRemoteRemove, true, true)),
				}, {
					Name:    "list",
					Aliases: []string{"ls"},
					Action:  withDaemon(handleRemoteList, true, true),
				}, {
					Name:   "clear",
					Action: withDaemon(handleRemoteClear, true, true),
				}, {
					Name:   "edit",
					Action: withDaemon(handleRemoteEdit, true, true),
				}, {
					Name:   "ping",
					Action: withArgCheck(needAtLeast(1), withDaemon(handleRemotePing, true, true)),
				}, {
					Name:    "auto-update",
					Aliases: []string{"au"},
					Action:  withArgCheck(needAtLeast(2), withDaemon(handleRemoteAutoUpdate, true, true)),
				}, {
					Name:    "folder",
					Aliases: []string{"fld", "f"},
					Action:  withDaemon(handleRemoteFolderListAll, true, true),
					Subcommands: []cli.Command{
						{
							Name:   "add",
							Action: withArgCheck(needAtLeast(2), withDaemon(handleRemoteFolderAdd, true, true)),
						}, {
							Name:    "remove",
							Aliases: []string{"rm"},
							Action:  withArgCheck(needAtLeast(2), withDaemon(handleRemoteFolderRemove, true, true)),
						}, {
							Name:   "clear",
							Action: withArgCheck(needAtLeast(1), withDaemon(handleRemoteFolderClear, true, true)),
						}, {
							Name:    "list",
							Aliases: []string{"ls"},
							Action:  withArgCheck(needAtLeast(1), withDaemon(handleRemoteFolderList, true, true)),
						},
					},
				},
			},
		}, {
			Name:     "pin",
			Category: vcscGroup,
			Action:   withArgCheck(needAtLeast(1), withDaemon(handlePin, true, true)),
			Subcommands: []cli.Command{
				{
					Name:   "add",
					Action: withArgCheck(needAtLeast(1), withDaemon(handlePin, true, true)),
				}, {
					Name:   "repin",
					Action: withDaemon(handleRepin, true, true),
				}, {
					Name:    "remove",
					Aliases: []string{"rm"},
					Action:  withArgCheck(needAtLeast(1), withDaemon(handleUnpin, true, true)),
				},
			},
		}, {
			Name:     "net",
			Category: netwGroup,
			Subcommands: []cli.Command{
				{
					Name:   "offline",
					Action: withDaemon(handleOffline, true, true),
				}, {
					Name:   "online",
					Action: withDaemon(handleOnline, true, true),
				}, {
					Name:   "status",
					Action: withDaemon(handleIsOnline, true, true),
				}, {
					Name:   "locate",
					Action: withArgCheck(needAtLeast(1), withDaemon(handleNetLocate, true, true)),
				},
			},
		}, {
			Name:     "status",
			Aliases:  []string{"st"},
			Category: vcscGroup,
			Action:   withDaemon(handleStatus, true, true),
		}, {
			Name:     "diff",
			Category: vcscGroup,
			Action:   withDaemon(handleDiff, true, true),
		}, {
			Name:     "tag",
			Category: vcscGroup,
			Action:   withArgCheck(needAtLeast(1), withDaemon(handleTag, true, true)),
		}, {
			Name:     "log",
			Category: vcscGroup,
			Action:   withDaemon(handleLog, true, true),
		}, {
			Name:     "fetch",
			Category: vcscGroup,
			Action:   withArgCheck(needAtLeast(1), withDaemon(handleFetch, true, true)),
		}, {
			Name:     "sync",
			Category: vcscGroup,
			Action:   withDaemon(handleSync, true, true),
		}, {
			Name:     "commit",
			Aliases:  []string{"cmt"},
			Category: vcscGroup,
			Action:   withDaemon(handleCommit, true, true),
		}, {
			Name:     "reset",
			Aliases:  []string{"re"},
			Category: vcscGroup,
			Action:   withArgCheck(needAtLeast(1), withDaemon(handleReset, true, true)),
		}, {
			Name:     "become",
			Aliases:  []string{"be"},
			Category: vcscGroup,
			Action:   withDaemon(handleBecome, true, true),
		}, {
			Name:     "history",
			Aliases:  []string{"hst", "hist"},
			Category: vcscGroup,
			Action:   withArgCheck(needAtLeast(1), withDaemon(handleHistory, true, true)),
		}, {
			Name:     "stage",
			Aliases:  []string{"stg", "add", "a"},
			Category: wdirGroup,
			Action:   withArgCheck(needAtLeast(1), withDaemon(handleStage, true, true)),
		}, {
			Name:     "touch",
			Aliases:  []string{"t"},
			Category: wdirGroup,
			Action:   withArgCheck(needAtLeast(1), withDaemon(handleTouch, true, true)),
		}, {
			Name:     "cat",
			Category: wdirGroup,
			Action:   withDaemon(handleCat, true, true),
		}, {
			Name:     "show",
			Aliases:  []string{"s", "info"},
			Category: wdirGroup,
			Action:   withArgCheck(needAtLeast(1), withDaemon(handleShow, true, true)),
		}, {
			Name:     "rm",
			Aliases:  []string{"remove"},
			Category: wdirGroup,
			Action:   withArgCheck(needAtLeast(1), withDaemon(handleRm, true, true)),
		}, {
			Name:     "ls",
			Category: wdirGroup,
			Action:   withDaemon(handleList, true, true),
		}, {
			Name:     "tree",
			Category: wdirGroup,
			Action:   withDaemon(handleTree, true, true),
		}, {
			Name:     "mkdir",
			Category: wdirGroup,
			Action:   withArgCheck(needAtLeast(1), withDaemon(handleMkdir, true, true)),
		}, {
			Name:     "mv",
			Category: wdirGroup,
			Action:   withArgCheck(needAtLeast(2), withDaemon(handleMv, true, true)),
		}, {
			Name:     "cp",
			Category: wdirGroup,
			Action:   withArgCheck(needAtLeast(2), withDaemon(handleCp, true, true)),
		}, {
			Name:     "edit",
			Category: wdirGroup,
			Action:   withArgCheck(needAtLeast(1), withDaemon(handleEdit, true, true)),
		}, {
			Name:     "daemon",
			Category: repoGroup,
			Subcommands: []cli.Command{
				{
					Name:   "launch",
					Action: withExit(handleDaemonLaunch),
				}, {
					Name:   "quit",
					Action: withDaemon(handleDaemonQuit, false, true),
				}, {
					Name:   "ping",
					Action: withDaemon(handleDaemonPing, false, true),
				},
			},
		}, {
			Name:     "config",
			Aliases:  []string{"cfg"},
			Category: repoGroup,
			Action:   withDaemon(handleConfigList, true, true),
			Subcommands: []cli.Command{
				{
					Name:    "list",
					Aliases: []string{"ls"},
					Action:  withDaemon(handleConfigList, true, true),
				}, {
					Name:   "get",
					Action: withArgCheck(needAtLeast(1), withDaemon(handleConfigGet, true, true)),
				}, {
					Name:   "doc",
					Action: withArgCheck(needAtLeast(1), withDaemon(handleConfigDoc, true, true)),
				}, {
					Name:   "set",
					Action: withArgCheck(needAtLeast(2), withDaemon(handleConfigSet, true, true)),
				},
			},
		}, {
			Name:     "fstab",
			Category: repoGroup,
			Action:   withArgCheck(needAtLeast(0), withDaemon(handleFstabList, true, true)),
			Subcommands: []cli.Command{
				{
					Name:   "add",
					Action: withArgCheck(needAtLeast(2), withDaemon(handleFstabAdd, true, true)),
				}, {
					Name:    "remove",
					Aliases: []string{"rm"},
					Action:  withArgCheck(needAtLeast(1), withDaemon(handleFstabRemove, true, true)),
				}, {
					Name:   "apply",
					Action: withDaemon(handleFstabApply, true, true),
				}, {
					Name:    "list",
					Aliases: []string{"ls"},
					Action:  withDaemon(handleFstabList, true, true),
				},
			},
		}, {
			Name:     "trash",
			Aliases:  []string{"tr"},
			Category: repoGroup,
			Action:   handleTrashList,
			Subcommands: []cli.Command{
				{
					Name:    "list",
					Aliases: []string{"ls"},
					Action:  withDaemon(handleTrashList, true, true),
				},
				{
					Name:    "remove",
					Aliases: []string{"rm"},
					Action:  withArgCheck(needAtLeast(1), withDaemon(handleTrashRemove, true, true)),
				},
			},
		}, {
			Name:     "gateway",
			Aliases:  []string{"gw"},
			Category: repoGroup,
			Subcommands: []cli.Command{
				{
					Name:   "start",
					Action: withDaemon(handleGatewayStart, true, true),
				},
				{
					Name:   "stop",
					Action: withDaemon(handleGatewayStop, true, true),
				},
				{
					Name:   "status",
					Action: withDaemon(handleGatewayStatus, true, true),
				},
				{
					Name:   "cert",
					Action: handleGatewayCert,
				},
				{
					Name:   "url",
					Action: withArgCheck(needAtLeast(1), withDaemon(handleGatewayURL, true, true)),
				},
				{
					Name:    "user",
					Aliases: []string{"u"},
					Subcommands: []cli.Command{
						{
							Name:    "add",
							Aliases: []string{"a"},
							Action:  withArgCheck(needAtLeast(1), withDaemon(handleGatewayUserAdd, true, true)),
						},
						{
							Name:    "remove",
							Aliases: []string{"rm"},
							Action:  withArgCheck(needAtLeast(1), withDaemon(handleGatewayUserRemove, true, true)),
						},
						{
							Name:    "list",
							Aliases: []string{"ls"},
							Action:  withDaemon(handleGatewayUserList, true, true),
						},
					},
				},
			},
		}, {
			Name:     "mount",
			Category: repoGroup,
			Action:   withDaemon(handleMount, true, true),
		}, {
			Name:     "unmount",
			Category: repoGroup,
			Action:   withDaemon(handleUnmount, true, true),
		}, {
			Name:     "version",
			Category: repoGroup,
			Action:   withDaemon(handleVersion, false, true),
		}, {
			Name:     "gc",
			Category: repoGroup,
			Action:   withDaemon(handleGc, true, true),
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
		fmt.Println(err)
		return 1
	}
	return 0
}
