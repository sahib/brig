package cmdline

import (
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	colorlog "github.com/disorganizer/brig/util/log"
	"github.com/tucnak/climax"
)

func init() {
	log.SetOutput(os.Stderr)

	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)

	// Log pretty text
	log.SetFormatter(&colorlog.ColorfulLogFormatter{})
}

func formatGroup(category string) string {
	return strings.ToUpper(category) + " COMMANDS:"
}

////////////////////////////
// Commandline definition //
////////////////////////////

// RunCmdline starts a brig commandline tool.
func RunCmdline() int {
	demo := climax.New("brig")
	demo.Brief = "brig is a decentralized file syncer based on IPFS and XMPP."
	demo.Version = "unstable"

	repoGroup := demo.AddGroup(formatGroup("repository"))
	xmppGroup := demo.AddGroup(formatGroup("xmpp helper"))
	wdirGroup := demo.AddGroup(formatGroup("working"))
	advnGroup := demo.AddGroup(formatGroup("advanced"))
	miscGroup := demo.AddGroup(formatGroup("misc"))

	commands := []climax.Command{
		climax.Command{
			Name:  "init",
			Brief: "Initialize an empty repository and open it",
			Group: repoGroup,
			Usage: `<JID> [<PATH>]`,
			Help:  `Create an empty repository, open it and associate it with the JID`,
			Flags: []climax.Flag{
				{
					Name:  "nodaemon",
					Short: "n",
					Help:  `Do not start the daemon.`,
				}, {
					Name:     "password",
					Short:    "x",
					Usage:    `--password PWD`,
					Help:     `Supply password.`,
					Variable: true,
				},
			},
			Examples: []climax.Example{
				{
					Usecase:     `alice@jabber.de/laptop`,
					Description: `Create a folder laptop/ with hidden directories`,
				},
			},
			Handle: withArgCheck(needAtLeast(1), handleInit),
		},
		climax.Command{
			Name:   "clone",
			Brief:  "Clone an repository from somebody else",
			Group:  repoGroup,
			Usage:  `<YOUR_JID> <REMOTE_JID> [<PATH>]`,
			Help:   `...`,
			Handle: withArgCheck(needAtLeast(2), handleClone),
			Flags: []climax.Flag{
				{
					Name:  "nodaemon",
					Short: "n",
					Help:  `Do not start the daemon.`,
				}, {
					Name:     "password",
					Short:    "x",
					Usage:    `--password PWD`,
					Help:     `Supply password.`,
					Variable: true,
				},
			},
			Examples: []climax.Example{
				{
					Usecase:     `alice@jabber.de/laptop bob@jabber.de/desktop`,
					Description: `Clone Alice' contents`,
				},
			},
		},
		climax.Command{
			Name:   "open",
			Group:  repoGroup,
			Brief:  "Open an encrypted repository. Asks for passphrase.",
			Handle: withDaemon(handleOpen, true),
		},
		climax.Command{
			Name:   "close",
			Group:  repoGroup,
			Brief:  "Encrypt all metadata in the repository and go offline.",
			Handle: withDaemon(handleClose, false),
		},
		climax.Command{
			Name:   "history",
			Group:  repoGroup,
			Brief:  "Show the history of a file.",
			Handle: withArgCheck(needAtLeast(1), withDaemon(handleHistory, true)),
		},
		climax.Command{
			Name:   "offline",
			Group:  repoGroup,
			Brief:  "Disconnect from the outside world. The daemon will continue running.",
			Handle: withDaemon(handleOffline, true),
		},
		climax.Command{
			Name:   "online",
			Group:  repoGroup,
			Brief:  "Connect the daemon to the outside world.",
			Handle: withDaemon(handleOnline, true),
		},
		climax.Command{
			Name:   "is-online",
			Group:  repoGroup,
			Brief:  "Check if the daemon is online.",
			Handle: withDaemon(handleIsOnline, true),
		},
		climax.Command{
			Name:  "sync",
			Group: repoGroup,
			Brief: "Sync with all or selected trusted peers.",
		},
		climax.Command{
			Name:  "push",
			Group: repoGroup,
			Brief: "Push your content to all or selected trusted peers.",
		},
		climax.Command{
			Name:   "pull",
			Group:  repoGroup,
			Brief:  "Pull content from all or selected trusted peers.",
			Handle: withArgCheck(needAtLeast(1), withDaemon(handlePull, true)),
		},
		climax.Command{
			Name:  "watch",
			Group: repoGroup,
			Brief: "Enable or disable watch mode.",
		},
		climax.Command{
			Name:  "discover",
			Group: xmppGroup,
			Brief: "Try to find other brig users near you.",
		},
		climax.Command{
			Name:  "friends",
			Group: xmppGroup,
			Brief: "List your trusted peers.",
		},
		climax.Command{
			Name:  "beg",
			Group: xmppGroup,
			Brief: "Request authorisation from a buddy.",
		},
		climax.Command{
			Name:  "ban",
			Group: xmppGroup,
			Brief: "Discontinue friendship with a peer.",
		},
		climax.Command{
			Name:  "prio",
			Group: xmppGroup,
			Brief: "Change priority of a peer.",
		},
		climax.Command{
			Name:   "tree",
			Brief:  "List files in a tree",
			Group:  wdirGroup,
			Usage:  `[<PATH>]`,
			Help:   `...`,
			Handle: withDaemon(handleTree, true),
			Flags: []climax.Flag{
				{
					Name:     "depth",
					Short:    "d",
					Help:     `List directories only into certain depth.`,
					Variable: true,
				},
			},
		},
		climax.Command{
			Name:   "ls",
			Brief:  "List files",
			Group:  wdirGroup,
			Usage:  `[<PATH>]`,
			Help:   `...`,
			Handle: withDaemon(handleList, true),
			Flags: []climax.Flag{
				{
					Name:     "depth",
					Short:    "d",
					Help:     `List directories only into certain depth.`,
					Variable: true,
				}, {
					Name:  "recursive",
					Short: "r",
					Help:  `List directories recursively.`,
				},
			},
		},
		climax.Command{
			Name:  "status",
			Group: wdirGroup,
			Brief: "Give an overview of brig's current state.",
		},
		climax.Command{
			Name:   "add",
			Group:  wdirGroup,
			Brief:  "Transer file into brig's control.",
			Usage:  `FILE_OR_FOLDER [PATH_INSIDE_BRIG]`,
			Help:   `Add a file or directory to brig. The second path is where it will appear in the mount.`,
			Handle: withArgCheck(needAtLeast(1), withDaemon(handleAdd, true)),
		},
		climax.Command{
			Name:   "rm",
			Group:  wdirGroup,
			Brief:  "Remove the file and optionally old versions of it.",
			Usage:  `FILE_OR_FOLDER PATH_INSIDE_BRIG`,
			Handle: withArgCheck(needAtLeast(1), withDaemon(handleRm, true)),
		},
		climax.Command{
			Name:   "mv",
			Group:  wdirGroup,
			Brief:  "Move a file from SOURCE to DEST.",
			Usage:  `SOURCE_FILE_OR_FOLDER DEST_FILE_OR_FOLDER`,
			Handle: withArgCheck(needAtLeast(2), withDaemon(handleMv, true)),
		},
		climax.Command{
			Name:   "cat",
			Group:  wdirGroup,
			Brief:  "Write ",
			Usage:  `FILE_OR_FOLDER DEST_PATH`,
			Handle: withArgCheck(needAtLeast(1), withDaemon(handleCat, true)),
		},
		climax.Command{
			Name:  "find",
			Group: wdirGroup,
			Brief: "Find filenames in the fleet.",
		},
		climax.Command{
			Name:  "rm",
			Group: wdirGroup,
			Brief: "Remove file from brig's control.",
		},
		climax.Command{
			Name:  "log",
			Group: wdirGroup,
			Brief: "Visualize changelog tree.",
		},
		climax.Command{
			Name:  "checkout",
			Group: wdirGroup,
			Brief: "Attempt to checkout previous version of a file.",
		},
		climax.Command{
			Name:  "fsck",
			Group: advnGroup,
			Brief: "Verify, and possibly fix, broken files.",
		},
		climax.Command{
			Name:  "daemon",
			Group: advnGroup,
			Brief: "Manually run the daemon process.",
			Flags: []climax.Flag{
				{
					Name:     "password",
					Short:    "x",
					Usage:    `--password PWD`,
					Help:     `Supply password.`,
					Variable: true,
				},
			},
			Handle: handleDaemon,
		},
		climax.Command{
			Name:   "daemon-quit",
			Group:  advnGroup,
			Brief:  "Manually kill the daemon process.",
			Handle: withDaemon(handleDaemonQuit, false),
		},
		climax.Command{
			Name:   "daemon-ping",
			Group:  advnGroup,
			Brief:  "See if the daemon responds in a timely fashion.",
			Handle: withDaemon(handleDaemonPing, false),
		},
		climax.Command{
			Name:   "daemon-wait",
			Group:  advnGroup,
			Brief:  "Block until the daemon is available.",
			Handle: handleDaemonWait,
		},
		climax.Command{
			Name:  "passwd",
			Group: advnGroup,
			Brief: "Set your XMPP and access password.",
		},
		climax.Command{
			Name:  "yubi",
			Group: advnGroup,
			Brief: "Manage YubiKeys.",
		},
		climax.Command{
			Name:   "config",
			Group:  miscGroup,
			Brief:  "Access, list and modify configuration values.",
			Handle: handleConfig,
		},
		climax.Command{
			Name:  "mount",
			Group: miscGroup,
			Brief: "Handle FUSE mountpoints.",
			Flags: []climax.Flag{
				{
					Name:  "unmount",
					Short: "u",
					Usage: `--unmount`,
					Help:  `Unmount the filesystem.`,
				},
			},
			Handle: withArgCheck(needAtLeast(1), withDaemon(handleMount, true)),
		},
		climax.Command{
			Name:  "update",
			Group: miscGroup,
			Brief: "Try to securely update brig.",
		},
		climax.Command{
			Name:  "help",
			Group: miscGroup,
			Brief: "Print some help",
			Usage: "Did you really need help on help?",
		},
		climax.Command{
			Name:   "version",
			Group:  miscGroup,
			Brief:  "Print current version.",
			Usage:  "Print current version.",
			Handle: handleVersion,
		},
	}

	for _, command := range commands {
		demo.AddCommand(command)
	}

	// Help topics:
	demo.AddTopic(climax.Topic{
		Name:  "quickstart",
		Brief: "A very short introduction to brig",
		Text:  "Needs to be written.",
	})
	demo.AddTopic(climax.Topic{
		Name:  "tutorial",
		Brief: "A slightly longer introduction.",
		Text:  "Needs to be written.",
	})
	demo.AddTopic(climax.Topic{
		Name:  "terms",
		Brief: "Cheat sheet for often used terms.",
		Text:  "Needs to be written.",
	})

	return demo.Run()
}
