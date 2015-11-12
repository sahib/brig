package cmdline

import (
	"fmt"
	"github.com/disorganizer/brig"
	"github.com/tucnak/climax"
	"strings"
)

///////////////////////
// Utility functions //
///////////////////////

func formatGroup(category string) string {
	return strings.ToUpper(category) + " COMMANDS:"
}

///////////////////////
// Handler functions //
///////////////////////

func handleVersion(ctx climax.Context) int {
	fmt.Println(brig.VersingString())
	return 1
}

////////////////////////////
// Commandline definition //
////////////////////////////

func RunCmdline() {
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
			Examples: []climax.Example{
				{
					Usecase:     `alice@jabber.de/laptop`,
					Description: `Create a folder laptop/ with hidden directories`,
				},
			},
			Handle: func(ctx climax.Context) int {
				return 0
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
		Name:  "quick-start",
		Brief: "A very short introduction to brig",
		Text:  "TODO: write.",
	})
	demo.AddTopic(climax.Topic{
		Name:  "tutorial",
		Brief: "A slightly longer introduction.",
		Text:  "TODO: write.",
	})
	demo.AddTopic(climax.Topic{
		Name:  "terminology",
		Brief: "Cheat sheet for often used terms.",
		Text:  "TODO: write.",
	})
	demo.Run()
}
