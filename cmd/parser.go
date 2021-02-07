package cmd

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"strings"

	"github.com/fatih/color"
	isatty "github.com/mattn/go-isatty"
	"github.com/sahib/brig/defaults"
	formatter "github.com/sahib/brig/util/log"
	"github.com/sahib/brig/version"
	log "github.com/sirupsen/logrus"
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
	return "\n" + strings.ToUpper(category) + " COMMANDS"
}

func memProfile() {
	memPath := os.Getenv("BRIG_MEM_PROFILE")
	if memPath == "" {
		return
	}

	fd, err := os.Create(memPath)
	if err != nil {
		log.Fatal("could not create memory profile: ", err)
	}

	defer fd.Close()

	runtime.GC()
	if err := pprof.WriteHeapProfile(fd); err != nil {
		log.Fatal("could not write memory profile: ", err)
	}
}

func startCPUProfile() *os.File {
	cpuPath := os.Getenv("BRIG_CPU_PROFILE")
	if cpuPath == "" {
		return nil
	}

	fd, err := os.Create(cpuPath)
	if err != nil {
		log.Fatal("could not create memory profile: ", err)
	}

	runtime.GC()
	if err := pprof.StartCPUProfile(fd); err != nil {
		log.Fatal("could not write memory profile: ", err)
	}

	return fd
}

func stopCPUProfile(fd *os.File) {
	if os.Getenv("BRIG_CPU_PROFILE") == "" {
		return
	}

	defer fd.Close()
	pprof.StopCPUProfile()
}

////////////////////////////
// Commandline definition //
////////////////////////////

// RunCmdline starts a brig commandline tool.
func RunCmdline(args []string) int {
	profFd := startCPUProfile()
	defer stopCPUProfile(profFd)
	defer memProfile()

	debug.SetTraceback("all")

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
		cli.StringFlag{
			Name:   "url,u",
			Usage:  "URL on where to reach the brig daemon. Leave empty to allow guessing.",
			EnvVar: "BRIG_URL",
			Value:  defaults.DaemonDefaultURL(),
		},
		cli.StringFlag{
			Name:   "repo",
			Usage:  "Path to the repository. Only has effect for new daemons.",
			Value:  ".",
			EnvVar: "BRIG_PATH",
		},
		cli.BoolFlag{
			Name:  "verbose,V",
			Usage: "Show certain messages during client startup (helpful for debugging)",
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
			Action:   handleInit,
		}, {
			Name:     "whoami",
			Aliases:  []string{"id"},
			Category: netwGroup,
			Action:   withDaemon(handleWhoami, true),
		}, {
			Name:     "remote",
			Aliases:  []string{"rmt", "r"},
			Category: netwGroup,
			Subcommands: []cli.Command{
				{
					Name:    "add",
					Aliases: []string{"a", "set"},
					Action:  withArgCheck(needAtLeast(2), withDaemon(handleRemoteAdd, true)),
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
				}, {
					Name:    "auto-update",
					Aliases: []string{"au"},
					Action:  withArgCheck(needAtLeast(2), withDaemon(handleRemoteAutoUpdate, true)),
				}, {
					Name:    "accept-push",
					Aliases: []string{"ap"},
					Action:  withArgCheck(needAtLeast(2), withDaemon(handleRemoteAcceptPush, true)),
				}, {
					Name:    "conflict-strategy",
					Aliases: []string{"cs"},
					Action:  withArgCheck(needAtLeast(2), withDaemon(handleRemoteConflictStrategy, true)),
				}, {
					Name:    "folder",
					Aliases: []string{"fld", "f"},
					Action:  withDaemon(handleRemoteFolderListAll, true),
					Subcommands: []cli.Command{
						{
							Name:    "add",
							Aliases: []string{"a"},
							Action:  withArgCheck(needAtLeast(2), withDaemon(handleRemoteFolderAdd, true)),
						}, {
							Name:    "set",
							Aliases: []string{"s"},
							Action:  withArgCheck(needAtLeast(2), withDaemon(handleRemoteFolderSet, true)),
						}, {
							Name:    "remove",
							Aliases: []string{"rm"},
							Action:  withArgCheck(needAtLeast(2), withDaemon(handleRemoteFolderRemove, true)),
						}, {
							Name:   "clear",
							Action: withArgCheck(needAtLeast(1), withDaemon(handleRemoteFolderClear, true)),
						}, {
							Name:    "list",
							Aliases: []string{"ls"},
							Action:  withArgCheck(needAtLeast(1), withDaemon(handleRemoteFolderList, true)),
						},
					},
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
					Name:   "repin",
					Action: withDaemon(handleRepin, true),
				}, {
					Name:    "remove",
					Aliases: []string{"rm"},
					Action:  withArgCheck(needAtLeast(1), withDaemon(handleUnpin, true)),
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
			Name:     "sync",
			Category: vcscGroup,
			Action:   withDaemon(handleSync, true),
		}, {
			Name:     "push",
			Category: vcscGroup,
			Action:   withArgCheck(needAtLeast(1), withDaemon(handlePush, true)),
		}, {
			Name:     "commit",
			Aliases:  []string{"cmt"},
			Category: vcscGroup,
			Action:   withDaemon(handleCommit, true),
		}, {
			Name:     "reset",
			Aliases:  []string{"re"},
			Category: vcscGroup,
			Action:   withArgCheck(needAtLeast(1), withDaemon(handleReset, true)),
		}, {
			Name:     "become",
			Aliases:  []string{"be"},
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
			Action:   withDaemon(handleCat, true),
		}, {
			Name:     "show",
			Aliases:  []string{"s", "info"},
			Category: wdirGroup,
			Action:   withArgCheck(needAtLeast(1), withDaemon(handleShow, true)),
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
					Action: handleDaemonLaunch,
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
					Action: withDaemon(handleFstabApply, true),
				}, {
					Name:    "list",
					Aliases: []string{"ls"},
					Action:  withDaemon(handleFstabList, true),
				},
			},
		}, {
			Name:     "trash",
			Aliases:  []string{"tr"},
			Category: repoGroup,
			Action:   withDaemon(handleTrashList, true),
			Subcommands: []cli.Command{
				{
					Name:    "list",
					Aliases: []string{"ls"},
					Action:  withDaemon(handleTrashList, true),
				},
				{
					Name:    "undelete",
					Aliases: []string{"rm"},
					Action:  withArgCheck(needAtLeast(1), withDaemon(handleTrashRemove, true)),
				},
			},
		}, {
			Name:     "hints",
			Aliases:  []string{"hi"},
			Category: repoGroup,
			Action:   withDaemon(handleRepoHintsList, true),
			Subcommands: []cli.Command{
				{
					Name:    "list",
					Aliases: []string{"ls"},
					Action:  withDaemon(handleRepoHintsList, true),
				}, {
					Name:    "set",
					Aliases: []string{"s"},
					Action:  withArgCheck(needAtLeast(1), withDaemon(handleRepoHintsSet, true)),
				}, {
					Name:    "remove",
					Aliases: []string{"rm"},
					Action:  withArgCheck(needAtLeast(1), withDaemon(handleRepoHintsRemove, true)),
				}, {
					Name:    "recode",
					Aliases: []string{"r"},
					Action:  withDaemon(handleRepoHintsRecode, true),
				},
			},
		}, {
			Name:     "gateway",
			Aliases:  []string{"gw"},
			Category: repoGroup,
			Subcommands: []cli.Command{
				{
					Name:   "start",
					Action: withDaemon(handleGatewayStart, true),
				},
				{
					Name:   "stop",
					Action: withDaemon(handleGatewayStop, true),
				},
				{
					Name:   "status",
					Action: withDaemon(handleGatewayStatus, true),
				},
				{
					Name:   "cert",
					Action: handleGatewayCert,
				},
				{
					Name:   "url",
					Action: withArgCheck(needAtLeast(1), withDaemon(handleGatewayURL, true)),
				},
				{
					Name:    "user",
					Aliases: []string{"u"},
					Subcommands: []cli.Command{
						{
							Name:    "add",
							Aliases: []string{"a"},
							Action:  withArgCheck(needAtLeast(1), withDaemon(handleGatewayUserAdd, true)),
						},
						{
							Name:    "remove",
							Aliases: []string{"rm"},
							Action:  withArgCheck(needAtLeast(1), withDaemon(handleGatewayUserRemove, true)),
						},
						{
							Name:    "list",
							Aliases: []string{"ls"},
							Action:  withDaemon(handleGatewayUserList, true),
						},
					},
				},
			},
		}, {
			Name:     "debug",
			Aliases:  []string{"d"},
			Category: repoGroup,
			Subcommands: []cli.Command{
				{
					Name:    "pprof-port",
					Aliases: []string{"p"},
					Action:  withDaemon(handleDebugPprofPort, true),
				}, {
					Name:    "decode-stream",
					Aliases: []string{"ds"},
					Action:  handleDebugDecodeStream,
				}, {
					Name:    "encode-stream",
					Aliases: []string{"es"},
					Action:  handleDebugEncodeStream,
				}, {
					Name:    "ten-source",
					Aliases: []string{"tso"},
					Action:  handleDebugTenSource,
				}, {
					Name:    "ten-sink",
					Aliases: []string{"tsi"},
					Action:  handleDebugTenSink,
				}, {
					Name:   "iobench",
					Action: handleIOBench,
				}, {
					Name:   "fusemock",
					Action: handleDebugFuseMock,
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
			Name:     "pack-repo",
			Category: repoGroup,
			Action:   handleRepoPack,
			Aliases:  []string{"pr"},
		}, {
			Name:     "unpack-repo",
			Category: repoGroup,
			Action:   withArgCheck(needAtLeast(1), handleRepoUnpack),
			Aliases:  []string{"ur"},
		}, {
			Name:   "docs",
			Action: handleOpenHelp,
			Hidden: true,
		}, {
			Name:   "bug",
			Action: handleBugReport,
		},
	})

	exitCode := Success
	if err := app.Run(args); err != nil {
		log.Error(prettyPrintError(err))
		cerr, ok := err.(ExitCode)
		if !ok {
			exitCode = UnknownError
		}

		exitCode = cerr.Code
	}

	return exitCode
}
