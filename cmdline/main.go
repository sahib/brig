package main

import "github.com/sahib/climax"

func main() {
	demo := climax.New("brig")
	demo.Brief = "brig is a decentralized file syncer based on IPFS and XMPP."
	demo.Version = "unstable"

	commands := []climax.Command{
		climax.Command{
			Name:     "init",
			Brief:    "Initialize an empty repository and open it",
			Category: "repository",
			Usage:    `<JID> [<PATH>]`,
			Help:     `Create an empty repository, open it and associate it with the JID`,
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
			Name:     "clone",
			Brief:    "Clone an repository from somebody else",
			Category: "repository",
			Usage:    `<OTHER_JID> <YOUR_JID> [<PATH>]`,
			Help:     `...`,
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
				return 0
			},
		},
		climax.Command{
			Name:     "open",
			Category: "repository",
			Brief:    "Open an encrypted port. Asks for passphrase.",
		},
		climax.Command{
			Name:     "close",
			Category: "repository",
			Brief:    "Encrypt all metadata in the port and go offline.",
		},
		climax.Command{
			Name:     "sync",
			Category: "repository",
			Brief:    "Sync with all or selected trusted peers.",
		},
		climax.Command{
			Name:     "push",
			Category: "repository",
			Brief:    "Push your content to all or selected trusted peers.",
		},
		climax.Command{
			Name:     "pull",
			Category: "repository",
			Brief:    "Pull content from all or selected trusted peers.",
		},
		climax.Command{
			Name:     "watch",
			Category: "repository",
			Brief:    "Enable or disable watch mode.",
		},
		climax.Command{
			Name:     "discover",
			Category: "xmpp helper",
			Brief:    "Try to find other brig users near you.",
		},
		climax.Command{
			Name:     "friends",
			Category: "xmpp helper",
			Brief:    "List your trusted peers.",
		},
		climax.Command{
			Name:     "beg",
			Category: "xmpp helper",
			Brief:    "Request authorisation from a buddy.",
		},
		climax.Command{
			Name:     "ban",
			Category: "xmpp helper",
			Brief:    "Discontinue friendship with a peer.",
		},
		climax.Command{
			Name:     "prio",
			Category: "xmpp helper",
			Brief:    "Change priority of a peer.",
		},
		climax.Command{
			Name:     "status",
			Category: "working dir",
			Brief:    "Give an overview of brig's current state.",
		},
		climax.Command{
			Name:     "add",
			Category: "working dir",
			Brief:    "Make file to be managed by brig.",
		},
		climax.Command{
			Name:     "find",
			Category: "working dir",
			Brief:    "Find filenames in the fleet.",
		},
		climax.Command{
			Name:     "rm",
			Category: "working dir",
			Brief:    "Remove file from brig's control.",
		},
		climax.Command{
			Name:     "log",
			Category: "working dir",
			Brief:    "Visualize changelog tree.",
		},
		climax.Command{
			Name:     "checkout",
			Category: "working dir",
			Brief:    "Attempt to checkout previous version of a file.",
		},
		climax.Command{
			Name:     "lock",
			Category: "advanced",
			Brief:    "Disallow any modification of the repository.",
		},
		climax.Command{
			Name:     "unlock",
			Category: "advanced",
			Brief:    "Remove a previous write lock.",
		},
		climax.Command{
			Name:     "fsck",
			Category: "advanced",
			Brief:    "Verify, and possibly fix, broken files.",
		},
		climax.Command{
			Name:     "passwd",
			Category: "advanced",
			Brief:    "Set your XMPP and access password.",
		},
		climax.Command{
			Name:     "yubi",
			Category: "advanced",
			Brief:    "Manage YubiKeys.",
		},
		climax.Command{
			Name:     "config",
			Category: "misc",
			Brief:    "Access, list and modify configuration values.",
		},
		climax.Command{
			Name:     "update",
			Category: "misc",
			Brief:    "Try to securely update brig.",
		},
		climax.Command{
			Name:     "help",
			Category: "misc",
			Brief:    "Print some help",
			Usage:    "Did you really need help on help?",
		},
		climax.Command{
			Name:     "version",
			Category: "misc",
			Brief:    "Print current version.",
			Usage:    "Print current version.",
		},
	}

	for _, command := range commands {
		demo.AddCommand(command)
	}

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
