package cmdline

import (
	"fmt"
	"os"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig"
	"github.com/disorganizer/brig/daemon"
	"github.com/disorganizer/brig/repo"
	"github.com/disorganizer/brig/util/colors"
	colorlog "github.com/disorganizer/brig/util/log"
	"github.com/tsuibin/goxmpp2/xmpp"
	"github.com/tucnak/climax"
)

func init() {
	log.SetOutput(os.Stderr)

	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)

	// Log pretty text
	log.SetFormatter(&colorlog.ColorfulLogFormatter{})
}

///////////////////////
// Utility functions //
///////////////////////

func formatGroup(category string) string {
	return strings.ToUpper(category) + " COMMANDS:"
}

// guessRepoFolder tries to find the repository path
// by using a number of sources.
func guessRepoFolder() string {
	wd := os.Getenv("BRIG_PATH")
	if wd != "" {
		return wd
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Error(err)
	}
	return wd
}

///////////////////////
// Handler functions //
///////////////////////

func handleVersion(ctx climax.Context) int {
	fmt.Println(brig.VersionString())
	return 0
}

func handleOpen(ctx climax.Context) int {
	repo.PromptPasswordMaxTries(4, func(pwd string) bool {
		return pwd == "bob"
	})

	repository, err := repo.LoadFsRepository(guessRepoFolder())
	if err != nil {
		log.Error("Could not create repository", err)
	}
	fmt.Println(repository)
	return 0
}

func handleDaemonPing() int {
	client, err := daemon.Dial(6666)
	if err != nil {
		log.Warning("Unable to dial to daemon: ", err)
		return 1
	}
	defer client.Close()

	for i := 0; ; i++ {
		before := time.Now()
		symbol := colors.Colorize("✔", colors.Green)
		if !client.Ping() {
			symbol = colors.Colorize("✘", colors.Red)
		}

		delay := time.Since(before)

		fmt.Printf("#%02d %s ➔ %s: %s (%v)\n",
			i+1,
			client.LocalAddr().String(),
			client.RemoteAddr().String(),
			symbol, delay)
		time.Sleep(1 * time.Second)
	}

	return 0
}

func handleDaemonQuit() int {
	client, err := daemon.Dial(6666)
	if err != nil {
		log.Warning("Unable to dial to daemon: ", err)
		return 1
	}
	defer client.Close()

	client.Exorcise()
	return 0
}

func handleDaemon(ctx climax.Context) int {
	if ctx.Is("ping") {
		return handleDaemonPing()
	} else if ctx.Is("quit") {
		return handleDaemonQuit()
	} else {
		baal, err := daemon.Summon(guessRepoFolder(), 6666)
		if err != nil {
			log.Warning("Unable to start daemon: ", err)
			return 1
		}

		baal.Serve()
	}

	return 0
}

func handleInit(ctx climax.Context) int {
	if len(ctx.Args) < 1 {
		log.Error("Need your Jabber ID.")
		return 1
	}

	jid := xmpp.JID(ctx.Args[0])
	if jid.Domain() == "" {
		log.Error("Your JabberID needs a domain.")
		return 2
	}

	// Extract the folder from the resource name by default:
	folder := jid.Resource()
	if folder == "" {
		log.Error("Need a resource in your JID.")
		return 3
	}

	if envFolder := os.Getenv("BRIG_PATH"); envFolder != "" {
		folder = envFolder
	}

	if ctx.Is("folder") {
		folder, _ = ctx.Get("folder")
	}

	pwd := ""
	// pwd, err := repo.PromptNewPassword(40.0)
	// if err != nil {
	// 	log.Error(err)
	// 	return 4
	// }

	if _, err := repo.NewFsRepository(string(jid), string(pwd), folder); err != nil {
		log.Error(err)
		return 5
	}

	return 0
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
					Name:     "--folder",
					Short:    "o",
					Usage:    `--depth="N"`,
					Help:     `Only clone up to this depth of pinned files`,
					Variable: true,
				},
			},
			Examples: []climax.Example{
				{
					Usecase:     `alice@jabber.de/laptop`,
					Description: `Create a folder laptop/ with hidden directories`,
				},
			},
			Handle: func(ctx climax.Context) int {
				return handleInit(ctx)
			},
		},
		climax.Command{
			Name:  "clone",
			Brief: "Clone an repository from somebody else",
			Group: repoGroup,
			Usage: `<OTHER_JID> <YOUR_JID> [<PATH>]`,
			Help:  `...`,
			Flags: []climax.Flag{
				{
					Name:     "--depth",
					Short:    "d",
					Usage:    `--depth="N"`,
					Help:     `Only clone up to this depth of pinned files`,
					Variable: true,
				},
			},
			Examples: []climax.Example{
				{
					Usecase:     `alice@jabber.de/laptop bob@jabber.de/desktop`,
					Description: `Clone Alice' contents`,
				},
			},
			Handle: func(ctx climax.Context) int {
				// TODO: Utils to convert string to int.
				// TODO: Utils to get default value.
				depth, ok := ctx.Get("--depth")
				if !ok {
					depth = "-1"
				}

				fmt.Println(depth)
				return 0
			},
		},
		climax.Command{
			Name:  "open",
			Group: repoGroup,
			Brief: "Open an encrypted port. Asks for passphrase.",
			Handle: func(ctx climax.Context) int {
				return handleOpen(ctx)
			},
		},
		climax.Command{
			Name:  "close",
			Group: repoGroup,
			Brief: "Encrypt all metadata in the port and go offline.",
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
			Name:  "pull",
			Group: repoGroup,
			Brief: "Pull content from all or selected trusted peers.",
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
			Name:  "status",
			Group: wdirGroup,
			Brief: "Give an overview of brig's current state.",
		},
		climax.Command{
			Name:  "add",
			Group: wdirGroup,
			Brief: "Make file to be managed by brig.",
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
			Name:  "lock",
			Group: advnGroup,
			Brief: "Disallow any modification of the repository.",
		},
		climax.Command{
			Name:  "unlock",
			Group: advnGroup,
			Brief: "Remove a previous write lock.",
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
					Name:  "ping",
					Short: "p",
					Usage: `--ping`,
					Help:  `Ping the dameon to check if it's running.`,
				},
				{
					Name:  "quit",
					Short: "q",
					Usage: `--quit`,
					Help:  `Kill a running daemon.`,
				},
			},
			Handle: handleDaemon,
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
			Name:  "config",
			Group: miscGroup,
			Brief: "Access, list and modify configuration values.",
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
		Text:  "TODO: write.",
	})
	demo.AddTopic(climax.Topic{
		Name:  "tutorial",
		Brief: "A slightly longer introduction.",
		Text:  "TODO: write.",
	})
	demo.AddTopic(climax.Topic{
		Name:  "terms",
		Brief: "Cheat sheet for often used terms.",
		Text:  "TODO: write.",
	})

	return demo.Run()
}
