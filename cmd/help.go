package cmd

import (
	"fmt"
	"strings"

	"github.com/urfave/cli"
)

type Help struct {
	Usage       string
	ArgsUsage   string
	Description string
	Flags       map[string]string
	Complete    cli.BashCompleteFunc
}

func die(msg string) {
	//panic(msg)
}

var HelpTexts = map[string]Help{
	"init": {
		Usage:     "Initialize a new repository",
		ArgsUsage: "<username>",
		Complete:  completeArgsUsage,
		Description: `Initialize a new repository with a certain backend.

   If BRIG_PATH is set, the new repository will be created at this place.  If not,
   the current working directory is used. If the directory is not empty, brig will
   warn you about it and abort.

   The username can be specified as pretty much any string, but it is recommended
   to use the special format »user@domain.something/resource«. This is similar to
   XMPP IDs. Specifying a resource can help you use the same name for different
   computers and specifying a domain makes it possible to indicate groups.  This
   is especially important for commands like »brig net locate«.
`,
		Flags: map[string]string{
			"backend": "Choose what backend to use",
		},
	},
	"whoami": {
		Usage:    "Print the own remote identity",
		Complete: completeArgsUsage,
		Description: `This command prints your name, fingerprint and what store
   you are looking at. When you initialized your repository, you chose
   the name and a fingerprint (a long hash value) was created for you.
   The store you're looking at, is by default your own. Every user has it's
   own store. See also the documentation for the »become« command.
`,
	},
	"remote": {
		Usage:    "Manage remotes (other authenticated users)",
		Complete: completeSubcommands,
		Description: `Add, list, edit and remove remotes.

   You need to add a remote for every peer or place you want to sync with. In
   order to add a remote, you need their fingerprint, which should be exchange in
   prior. You can view your own fingerprint with the »whoami« command. Also see
   the »net locate« command for more information.
`,
	},
	"remote.add": {
		Usage:       "Add a new remote",
		ArgsUsage:   "<name> <fingerprint>",
		Complete:    completeArgsUsage,
		Description: "Add a new remote under a handy name with their fingerprint.",
	},
	"remote.remove": {
		Usage:       "Remove a remote by name",
		ArgsUsage:   "<name>",
		Complete:    completeArgsUsage,
		Description: "Remove a remote by name",
	},
	"remote.list": {
		Usage:       "List all remotes",
		Complete:    completeArgsUsage,
		Description: "Show a list of each remote's name and corresponding fingerprints.",
	},
	"remote.clear": {
		Usage:       "Clear the complete remote list",
		Complete:    completeArgsUsage,
		Description: "Clear the complete remote list. Note that you cannot undo this operation.",
	},
	"remote.edit": {
		Usage:    "Edit the current list",
		Complete: completeArgsUsage,
		Description: `Edit the current list using $EDITOR as YAML file.
   It will be updated upon saving`,
	},
	"remote.ping": {
		Usage:    "Ping a remote",
		Complete: completeArgsUsage,
		Description: `Ping a remote and check if we can reach them.

   There is a small difference to the »net list« command. »ping« will only work
   if both sides authenticated each other and can thus be used as a test for this.
   Additonally, it shows the roundtrip time, the ping request took to travel.
`,
	},
	"pin": {
		Usage:     "Commands to pin a certain file",
		ArgsUsage: "<file>",
		Complete:  completeSubcommands,
		Description: `Pinning a file to keep it in local storage.

   When you retrieve a file from a remote machine, the file will be cached (maybe
   partially) for some time on your machine.  After some time, or after hitting a
   certain space limit, the file will be cleaned up to reclaim space. By adding a
   pin, you can make sure that this file will not be deleted.  Pinning can be
   useful for keeping old versions or to download files for offline use.

   Note however that pinning a file does not cause it to be automatically downloaded.
   Until we have a proper way to do this, you can use »brig cat <file> > /dev/null«.

   This command contains the subcommand 'add', but for usability reasons,
   »brig pin add <path>« is the same as »brig pin <path>«.

   See also the »gc« command as counterpart of pinning.
`,
	},
	"pin.add": {
		Usage:     "Pin a file or directory to local storage",
		ArgsUsage: "<file>",
		Complete:  completeArgsUsage,
		Description: `A node that is pinned to local storage will not be
   deleted by the garbage collector.`,
	},
	"pin.rm": {
		Usage:     "Remove a pin",
		ArgsUsage: "<file>",
		Complete:  completeArgsUsage,
		Description: `A node that is pinned to local storage will not be
   deleted by the garbage collector.`,
	},
	"net": {
		Usage:    "Commands to go online/offline, list other users and locate them",
		Complete: completeSubcommands,
		Description: `This command offers various subcommands that allow to
   change and check your online status (»online/offline/status«), and to
   see what other remotes are online. Most importantly, you can search
   other users by their name of by parts of it.`,
	},
	"net.offline": {
		Usage:    "Disconnect from the global network",
		Complete: completeArgsUsage,
		Description: `After going offline, other peers will not be able to
   contact you any more and vice versa. The daemon keeps running in this
   time and you can do all offline operations.`,
	},
	"net.online": {
		Usage:    "Connect to the global network",
		Complete: completeArgsUsage,
		Description: `This is the opposite of going offline. It might take a
   few minutes until you are fully connected. You are online by default.`,
	},
	"net.status": {
		Usage:       "Check if you're connected to the global network",
		Complete:    completeArgsUsage,
		Description: `This will either print the string »online« or »offline«`,
	},
	"net.list": {
		Usage:    "Check which remotes are currently online",
		Complete: completeArgsUsage,
		Description: `This goes over every entry in your remote list and prins
   his name, network address, rountrip and when we was last seen this
   remote.`,
	},
	"net.locate": {
		Usage:     "Try to locate a remote by their name or by a part of it",
		ArgsUsage: "<name-or-part-of-it>",
		Complete:  completeArgsUsage,
		Description: `brig is able to find the fingerprint of other users (that
   are online) by a part of their name. See the help of »brig init« to see
   out of what components the name is built of.

   Each found item shows the name, the fingerprint and what part of their name
   matched with your query.  Sometimes other peers are offline and cannot send
   your their fingerprint. In this case the peer will still be shown, but as
   »offline«.

   IMPORTANT: Locating a remote DOES NOT replace proper authentication. It is
   relatively easy to fake a fingerprint or even to have two peers with the same
   name. Always authenticate your peer properly via a sidechannel (mail,
   telephone, in person). »locate« is supposed to be only a help of discovering
   other nodes.

   Note that this operation might take quite a few seconds. Specifying »--timeout« can help,
   but currently it still might take longer than the given timeout.`,
	},
	"status": {
		Usage: "Show what has changed in the current commit",
		Description: `This a shortcut for »brig diff HEAD CURR«.
See the »diff« command for more information.`,
	},
	"diff": {
		Usage:     "Show what changed between two commits",
		ArgsUsage: "[<REMOTE>] [<REMOTE_REV> [<OTHER_REMOTE> [<OTHER_REV>]]]]",
		Complete:  completeArgsUsage,
		Description: `View what sync would do when being called on the specified points in history.

   Diff does not show what changed inside of the files, but shows how the files themselves
   changed. To describe this, brig knows seven different change types:

   - Added (+): The file was added on the remote side.
   - Removed (-): The file was removed on the remote side.
   - Missing (_): The file is missing on the remote side (e.g. we added it)
   - Moved (→): The file was moved to a new location.
   - Ignored (*): This file was ignored because we chose to due to our settings.
   - Mergeable (⇄): Both sides have changes, but they can be merged.
   - Conflict (⚡): Both sides have changes but they conflict.

   See »brig commit« for a general explanation of commits.

EXAMPLES:

   $ brig diff                      # Show diff from our CURR to our HEAD
   $ brig diff alice                # Show diff from our CURR to alice's last state
   $ brig diff alice some_tag       # Show diff from our CURR to 'some_tag' of alice
   $ brig diff alice HEAD bob HEAD  # Show diff between alice and bob's HEAD
`,
	},
	"tag": {
		Usage:     "Tag a commit with a specific name",
		Complete:  completeArgsUsage,
		ArgsUsage: "<commit> <name>",
		Description: `Give a name to a commit, which is easier to remember than the hash.
   You can use the name you gave in all places where brig requires you to specify a commit.

   There are three special tags pre-defined for you:

   - CURR: A reference to the staging commit.
   - HEAD: The last fully completed commit.
   - INIT: The very first commit in the chain.

   Tags are case insensitive. That means that »HEAD« and »head« mean the same.

EXAMPLES:

   $ brig tag SEfXUAH6AR my-tag-name   # Name the commit SEfXUAH6AR 'my-tag-name'.
   $ brig tag -d my-tag-name           # Delete the tag name again.
`,
	},
	"log": {
		Usage:    "Show all commits in a certain range",
		Complete: completeArgsUsage,
		Description: `Show a list of commits from a start (--from) up to and end (--to).
   If omitted »--from INIT --to HEAD« will be assumed.

   The output will show one commit per line, each including the (short) hash of the commit,
   the date it was committed and the (optional) commit message.
`,
	},
	"fetch": {
		Usage:     "Fetch all metadata from another peer",
		ArgsUsage: "<remote>",
		Complete:  completeArgsUsage,
		Description: `Get all the latest metadata of a certain peer.
   This does not download any actual data, but only the metadata of it.
   You have to be authenticated to the user to get his data.

   Fetch will be done automatically by »sync« and »diff« and is usually
   only helpful when doing it together with »become«.`,
	},
	"sync": {
		Usage:     "Sync with another peer",
		ArgsUsage: "<remote>",
		Complete:  completeArgsUsage,
		Description: `Sync and merge all metadata of another peer with our metadata.
   After this operation you might see new files in your folder.
   Those files were not downloaded yet and will be only on the first access.

   It is recommended that your first check what will be synced with »brig diff«.

   TODO: write some more documentation on this.
`,
	},
	"commit": {
		Usage:    "Create a new commit",
		Complete: completeArgsUsage,
		Description: `Create a new commit.

   The message (»--message«) is optional. If you do not pass it, a message will
   be generated with contains the current time. The commit history can be
   viewed by »brig log«.

   Think of commits as snapshots that can be created explicitly by you or even
   automated in an interval. It is important to remember that »commit« will
   only create a snapshot of the metadata. It is not guaranteed that you can
   still access the actual data.

   In the current implementation all old file states are unpinned during a
   commit. That means that you can only access old data until the garbage
   collector ran (see »brig gc«) or if you pin the old file explicitly.
`,
	},
	"reset": {
		Usage:     "Reset a file or the whole commit to an old state",
		ArgsUsage: "<remote>",
		Complete:  completeArgsUsage,
		Description: `Reset a file to an old state by specifying the commit it
   should be reverted to. If you do not pass »<file>« the whole commit will be
   filled with the contents of the old commit.

   If you reset to an old commit and you have uncommitted changes, brig will warn you
   about that and refuse the »reset« unless you pass »--force«.

   Note for git users: It is not possible to go back in history and branch out from there.
   »reset« simply overwrites the staging commit (CURR) with an old state, thus keeping
   all the previous history.

   If you notice that you do not like the state you've resetted to,
   »brig reset head« will bring you back to the last known good state.
`,
	},
	"become": {
		Usage:     "View the data of another user",
		ArgsUsage: "<remote>",
		Complete:  completeArgsUsage,
		Description: `View the data of another user.

   You can temporarily explore the metadata of another user, by »becoming«
   them. Once you became a certain user (which needs to be in your remote list
   and on which you called »brig fetch« before), you can look around in the
   data like in yours. You can also modify files, but keep in mind that they
   will be reset on he next fetch.

   This is a rather esoteric command and it's likely that you won't use it
   often. It's currently mainly useful for debugging, but could be used with
   »fetch« to do some editing before the actual »sync«.
`,
	},
	"history": {
		Usage:     "Show the history of a file or directory",
		ArgsUsage: "<path>",
		Complete:  completeArgsUsage,
		Description: `Show a list of all changes that were made to this path.

   Not every change you ever made is recorded, but the change between each commit.
   Every line shows the type of change and what commits were involved. If it's a
   move, it will also show from and to where the path was moved.

   Possible types of changes are:

   - added: The file was added in this commit.
   - moved: The file was modified in this commit.
   - removed: The file was removed in this commit.
   - modified: The file was modified in this commit.

   Furthermore, the following combination are possible:

   - moved & modified: The file was moved and modified.
   - add & modified: The file was removed before and now re-added with different content.
   - moved & removed: The file was moved to another location.
`,
	},
	"stage": {
		Usage:     "Add a local file to the storage",
		ArgsUsage: "(<local-path> [<path>]|--stdin <path>)",
		Complete:  completeArgsUsage,
		Description: `Read a local file (given by »local-path«) and try to read
   it. This is the conceptual equivalent of »git add«. The stream will be encrypted
   and possibly compressed before saving it to ipfs.

   If you omit »path«, the file will be added under the root
   directory, with the basename of »local-path«. You can change this by
   specifying where to save the local file by additionally passing »path«.

   Additonally you can read the file from standard input, if you pass »--stdin«.
   In this case you pass only one path: The path where the stream is stored.

EXAMPLES:

   $ brig stage file.png                   # gets added as /file.png
   $ brig stage file.png /photos/me.png    # gets added as /photos/me.png
   $ cat file.png | brig --stdin /file.png # gets added as /file.png`,
	},
	"touch": {
		Usage:     "Create an empty file under the specified path",
		ArgsUsage: "<path>",
		Complete:  completeArgsUsage,
		Description: `Convinience command for adding empty files.

   If the file or directory already exists, the modification time is updated to
   the current timestamp (like the original touch(1) does).
`,
	},
	"cat": {
		Usage:     "Output the content of a file to standard output",
		ArgsUsage: "<path>",
		Complete:  completeArgsUsage,
		Description: `Decrypt and decompress the stream from ipfs and write it to standard output.

Outputting a directory is currently not allowed (but might be in the future by
outputting a .tar archive of the directory contents).
`,
	},
	"info": {
		Usage:     "Show metadata of a file or directory",
		ArgsUsage: "<path>",
		Complete:  completeArgsUsage,
		Description: `Show all metadata attributes known for a file or directory.

   Path:    Absolute path of the file inside of the storage.
   User:    User which modified the file last.
   Type:    »file« or »directory«.
   Size:    Exact content size in bytes.
   Hash:    Hash of the node.
   Inode:   Internal inode. Also shown as inode in FUSE.
   Pinned:  »yes« if the file is pinned, »no« else.
   ModTime: Timestamp of last modification.
   Content: Content hash of the file in ipfs.
`,
	},
}

func injectHelp(cmd *cli.Command, path string) {
	help, ok := HelpTexts[cmd.Name]
	if !ok {
		die(fmt.Sprintf("bug: no such help entry: %v", cmd.Name))
	}

	cmd.Usage = help.Usage
	cmd.ArgsUsage = help.ArgsUsage
	cmd.Description = help.Description
	cmd.BashComplete = help.Complete

	// for _, flag := range cmd.Flags {
	// 	// flagNames := strings.Split(flag.GetName(), ",")
	// 	// flagUsage, ok := help.Flags[flagNames[0]]
	// 	// if !ok {
	// 	// 	panic(fmt.Sprintf("no documentation for flag %s (of %s)", flagNames[0], path))
	// 	// }
	// }

	// Be really pedantic to catch any deprecated flags early.
	if len(cmd.Flags) != len(help.Flags) {
		die(fmt.Sprintf("flag help catalogue of %s contains too many items", path))
	}
}

func translateHelp(cmds []cli.Command, prefix []string) {
	for idx := range cmds {
		path := append([]string{cmds[idx].Name}, prefix...)
		injectHelp(&cmds[idx], strings.Join(path, "."))
		translateHelp(cmds[idx].Subcommands, path)
	}
}

func TranslateHelp(cmds []cli.Command) []cli.Command {
	translateHelp(cmds, nil)
	return cmds
}
