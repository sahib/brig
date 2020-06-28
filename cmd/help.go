package cmd

import (
	"fmt"
	"strings"

	"github.com/toqueteos/webbrowser"
	"github.com/urfave/cli"
)

type helpEntry struct {
	Usage       string
	ArgsUsage   string
	Description string
	Complete    cli.BashCompleteFunc
	Flags       []cli.Flag
}

func die(msg string) {
	// be really pedantic when help is missing.
	// it is a developer mistake after all and should be catched early.
	panic(msg)
}

var helpTexts = map[string]helpEntry{
	"init": {
		Usage:     "Initialize a new repository.",
		ArgsUsage: "<username>",
		Complete:  completeArgsUsage,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "backend,b",
				Value: "httpipfs",
				Usage: "What data backend to use for the new repo. One of  `mock`, `httpipfs`. This cannot be changed later!",
			},
			cli.StringFlag{
				Name:  "w,pw-helper",
				Value: "",
				Usage: "Password helper command. The stdout of this command is used as password.",
			},
			cli.BoolFlag{
				Name:  "no-password,x",
				Usage: "Use a static password. Not recommended besides testing.",
			},
			cli.BoolFlag{
				Name:  "empty,e",
				Usage: "Do not create an initial README and no initial commit.",
			},
			cli.BoolFlag{
				Name:  "no-logo,n",
				Usage: "Do not display the super pretty logo on init.",
			},
			cli.StringFlag{
				Name:  "ipfs-path,P",
				Usage: "Specify an explicit path to an IPFS repository. Useful if you have more than one.",
				Value: "",
			},
			cli.BoolFlag{
				Name:  "no-ipfs-setup",
				Usage: "Do not try to install and setup IPFS.",
			},
			cli.BoolFlag{
				Name:  "no-ipfs-config",
				Usage: "Do no changes in the IPFS config that are necessary for brig. Use only when you know what you're doing.",
			},
			cli.BoolFlag{
				Name:  "no-ipfs-optimization,o",
				Usage: "Do no changes in the IPFS config that will improve the performance of brig, but are not necessary to work.",
			},
		},
		Description: `Initialize a new repository with a certain backend.

   If BRIG_PATH or --repo is set, the new repository will be created at this
   place. If nothing is specified, the repo is created at "~/.brig".  If the
   directory is not empty, brig will warn you about it and abort.

   The user name can be specified as pretty much any string, but it is recommended
   to use the special format »user@domain.something/resource«. This is similar to
   XMPP IDs. Specifying a resource can help you use the same name for different
   computers and specifying a domain makes it possible to indicate groups.  This
   is especially important for commands like »brig net locate« but is not used
   extensively by anything else yet.

   You will be asked to enter a password if you did not specify -w. This
   password will be used to encrypt the repository while brig is not running.
   For ease of use we recommend to specify a password helper with the -w
   option. This allows you to specify a password command that will print the
   desired password as output. This output is read by brig and used as
   password. For testing you could use »-w "echo mypass"«, while for serious
   use, you should use something like »pass brig/desktop/password«.

EXAMPLES:

	# Easiest way to create a repository at ~/.brig
	$ brig init ali@wonderland.org/rabbithole

`,
	},
	"whoami": {
		Usage:    "Print the own remote identity including IPFS id, fingerprint and user name.",
		Complete: completeArgsUsage,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "fingerprint,f",
				Usage: "Only print the own fingerprint",
			},
			cli.BoolFlag{
				Name:  "name,n",
				Usage: "Only print the own name",
			},
			cli.BoolFlag{
				Name:  "addr,a",
				Usage: "Only print the IPFS id portion of the fingerprint",
			},
			cli.BoolFlag{
				Name:  "key,k",
				Usage: "Only print the key portion of the fingerprint",
			},
		},
		Description: `This command prints your name, fingerprint and what store
   you are looking at. When you initialized your repository, you chose
   the name and a fingerprint (two longer hash values) was created for you.

EXAMPLES:

   # Show the fingerprint only:
   $ brig whoami -f
   QmUYz9dbqnYPyHCLUi7ghtiwFbdU93MQKFH4qg8iXHWcPV:W1q4vzbvLPUVwDUUXxjQfnuYJxq2CYqbeqXPSv7pUr5NcP
`,
	},
	"remote": {
		Usage:    "Add, list, remove and edit remotes.",
		Complete: completeSubcommands,
		Description: `
   A remote is the data needed to contact other instances of brig in the web.
   In order to add a remote, you need their fingerprint (as shown by »brig
   whoami«). This fingerprint should be exchanged in prior over a secure side
   channel (a secure instant messenger for example). Once both sides added each
   other as remotes they are said to be »authenticated«.

   Each remote can be configured further by specifying folders they may access
   or special settings like auto-updating. See the individual commands for more
   information.

   Also see the »net locate« command for details about finding other remotes.

EXAMPLES:

   # Show a diff for each remote:
   $ brig remote list --format '{{ .Name }}' | xargs -n 1 brig diff
`,
	},
	"remote.add": {
		Usage:       "Add/Update a remote under a handy name with their fingerprint.",
		ArgsUsage:   "<name> <fingerprint>",
		Complete:    completeArgsUsage,
		Description: "",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "auto-update,a",
				Usage: "Take automatic updates from this node.",
			},
			cli.BoolFlag{
				Name:  "accept-push,p",
				Usage: "Allow this remote to push to our state.",
			},
			cli.StringSliceFlag{
				Name:  "folder,f",
				Usage: "Configure the folders this remote may see. Can be given more than once. If the first letter of the folder is »-« it is added as read-only.",
			},
			cli.StringFlag{
				Name:  "conflict-strategy,c",
				Usage: "Which conflict strategy to apply (either »marker«, »ignore« or »embrace«)",
				Value: "",
			},
		},
	},
	"remote.remove": {
		Usage:       "Remove a remote by name.",
		ArgsUsage:   "<name>",
		Complete:    completeArgsUsage,
		Description: "Remove a remote by name.",
	},
	"remote.list": {
		Usage:    "List all remotes and their online status",
		Complete: completeArgsUsage,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "offline,o",
				Usage: "Do not query the online status",
			},
			cli.StringFlag{
				Name:  "format,f",
				Usage: "Format the output according to a template",
			},
		},
		Description: `
   This goes over every entry in your remote list and prints by default
   the remote name, fingerprint, rountrip, last seen timestamp and settings.

   You can format the output by using »--format« with one the following attributes:

	   * .Name
	   * .Fingerprint
	   * .Folders
	   * .AutoUpdate

   The syntax of the template is borrowed from Go. You can read about the details here:
   https://golang.org/pkg/text/template

   Note that this command will try to peek the fingerprint of each node, even
   if we did not authenticate him yet. If you do not want this, you should use
   »--offline«.

EXAMPLES:

   $ brig rmt ls -f '{{ .Name }}'  # Show each remote name, line by line.
`,
	},
	"remote.clear": {
		Usage:       "Clear the complete remote list.",
		Complete:    completeArgsUsage,
		Description: "Note that you cannot undo this operation!",
	},
	"remote.ping": {
		Usage:    "Ping a remote.",
		Complete: completeArgsUsage,
		Description: `Ping a remote and check if we can reach them.

   There is a small difference to the »remote list« command. »ping« will only
   work if both sides authenticated each other and can thus be used as a test
   for this.  Additionally, it shows the roundtrip time (i.e. the time the ping
   request took to travel).

EXAMPLES:

   $ brig rmt ping
`,
	},
	"remote.edit": {
		Usage:    "Edit the current list.",
		Complete: completeArgsUsage,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "yml,y",
				Value: "",
				Usage: "Directly overwrite remote list with yml file",
			},
		},
		Description: `
   Edit the current list using $EDITOR as YAML file.
   It will be updated once you exit your editor.`,
	},
	"remote.auto-update": {
		Usage:    "Enable auto-updating for this remote",
		Complete: completeArgsUsage,
		Description: `When enabled you will get updates shortly after this remote made it.

EXAMPLES:

	# Enable auto-updating both for bob and charlie.
	$ brig remote auto-update enable bob charlie

	# or shorter to prevent you from RSI:
	brig rmt au e bob charlie
`,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "no-initial-sync,n",
				Usage: "Do not sync initially when upon enabling.",
			},
		},
	},
	"remote.accept-push": {
		Usage:    "Allow receiving push requests from this remote.",
		Complete: completeArgsUsage,
		Description: `When enabled, other remotes can do »brig push <name>« to us.
   When we receive a push request we will sync with this remote.

EXAMPLES:

   # Allow bob and charlie to push to us:
   $ brig remote accept-push enable bob charlie

   # or shorter to prevent you from RSI:
   brig rmt ap e bob charlie
`,
	},
	"remote.conflict-strategy": {
		Usage:    "Change what conflict resolution strategy is used on conflicts.",
		Complete: completeArgsUsage,
		Description: `The conflict strategy defines how to act on sync conflicts.
   There are three different types:

   - marker: Create a conflict file with the remote's version. (default)
   - ignore: Ignore the remote version completely and keep our version.
   - embrace: Take the remote version and replace ours with it.

   See also »brig config doc fs.sync.conflict_strategy«.
   In case of an empty string, the config value above is used.

EXAMPLES:

   # Allow bob and charlie to push to us:
   $ brig remote conflict-strategy embrace bob charlie

   # or shorter to prevent you from RSI:
   brig rmt cs embrace bob charlie
`,
	},
	"remote.folder": {
		Usage:    "Configure what folders a remote is allowed to see.",
		Complete: completeArgsUsage,
		Description: `
   By default every remote is allowed to see all of your folders.
   You might want to share only specific folders with certain remotes.
   By adding folders to this list, you're limiting the nodes other remotes can see.

   If you do not specify any subcommand, this is a shortcut for »brig rmt f ls«`,
	},
	"remote.folder.add": {
		Usage:    "Add a remote folder for a specific remote.",
		Complete: completeArgsUsage,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "read-only,r",
				Usage: "Add the folder as read-only.",
			},
			cli.StringFlag{
				Name:  "conflict-strategy,c",
				Usage: "What conflict strategy to use for this specific folder. Overwrites per-remote conflict strategy.",
				Value: "",
			},
		},
		Description: `If a folder is added as read-only, we do not accept changes when syncing from remotes.

EXAMPLES:

   $ brig remote folder add bob /public --read-only
`,
	},
	"remote.folder.set": {
		Usage:    "Update the settings of a remote folder.",
		Complete: completeArgsUsage,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "read-only,r",
				Usage: "Add the folder as read-only.",
			},
			cli.BoolFlag{
				Name:  "read-write,w",
				Usage: "Add the folder as read and writeable.",
			},
			cli.StringFlag{
				Name:  "conflict-strategy,c",
				Usage: "What conflict strategy to use for this specific folder. Overwrites per-remote conflict strategy.",
				Value: "",
			},
		},
		Description: `This works exactly like »add« but overwrites an existing folder.

EXAMPLES:

   $ brig remote folder set bob /public --read-only
`,
	},
	"remote.folder.remove": {
		Usage:       "Remove a folder from a specific remote. ",
		Complete:    completeArgsUsage,
		Description: ``,
	},
	"remote.folder.clear": {
		Usage:       "Clear all folders from a specific remote.",
		Complete:    completeArgsUsage,
		Description: ``,
	},
	"remote.folder.list": {
		Usage:       "List all allowed folders for a specific remote.",
		Complete:    completeArgsUsage,
		Description: ``,
	},
	"pin": {
		Usage:     "Commands to handle the pin state.",
		ArgsUsage: "<file>",
		Complete:  completeBrigPath(true, true),
		Description: `Pinning a file to keep it in local storage.

   When you retrieve a file from a remote machine, the file will be cached (or
   maybe only blocks of it) for some time on your machine. If the file is not pinned,
   it might be collected by the garbage collector on the next run. The garbage collector
   is currently not invoked automatically, but can be activated via »brig gc«.

   Note that you can also pin files that you do not have cached locally. The
   pin does not download a file automatically currently. Until we have a
   proper way to do this, you can use »brig cat <file> > /dev/null«.

   This command contains the subcommand 'add', but for usability reasons, »brig
   pin add <path>« is the same as »brig pin <path>«.

   See also the »gc« command as counterpart of pinning.
`,
	},
	"pin.add": {
		Usage:     "Pin a file or directory to local storage",
		ArgsUsage: "<file>",
		Complete:  completeBrigPath(true, true),
		Description: `A node that is pinned to local storage will not be
   deleted by the garbage collector.`,
	},
	"pin.remove": {
		Usage:     "Remove a pin",
		ArgsUsage: "<file>",
		Complete:  completeBrigPath(true, true),
		Description: `A node that is pinned to local storage will not be
   deleted by the garbage collector.`,
	},
	"pin.repin": {
		Usage:     "Recaculate pinning based on fs.repin.{quota,min_depth,max_depth}",
		ArgsUsage: "[<root>]",
		Complete:  completeBrigPath(true, true),
		Description: `Trigger a repin calculation.

   This uses the following configuration variables:

   - fs.repin.quota: Max. amount of data to store in a repository.
   - fs.repin.min_depth: Keep this many versions definitely pinned. Trumps quota.
   - fs.repin.max_depth: Unpin versions beyond this depth definitely. Trumps quota.

   If repin detects files that need to be unpinned, then it will first unpin all files
   that are beyond the max depth setting. If this is not sufficient to stay under the quota,
   it will delete old versions, layer by layer starting with the biggest version first.

   If the optional root path was specified, the repin is only run in this part
   of the filesystem. This can be used to give the repin algorithm a hint where
   the space should be reclaimed.
   `,
	},
	"net": {
		Usage:       "Commands that change or query the network status.",
		Complete:    completeSubcommands,
		Description: `Most of these subcommands are somewhat low-level and are not often used.`,
	},
	"net.offline": {
		Usage:    "Prevent any online usage.",
		Complete: completeArgsUsage,
		Description: `

   The daemon will be running after going offline.
   After going offline, other peers will not be able to
   contact you any more and vice versa. The daemon keeps running in this
   time and you can do all offline operations.

   BUGS: This currently does not prevent other nodes to contact us.
   Shutdown the IPFS daemon to be sure for now.`,
	},
	"net.online": {
		Usage:    "Allow online usage.",
		Complete: completeArgsUsage,
		Description: `

   Opposite of »brig net offline«. This is the default state whenever the daemon starts.`,
	},
	"net.status": {
		Usage:       "Check if you're connected to the global network.",
		Complete:    completeArgsUsage,
		Description: `This will either print the string »online« or »offline«.`,
	},
	"net.locate": {
		Usage:     "Try to locate a remote by their name or by a part of it.",
		ArgsUsage: "<name-or-part-of-it>",
		Complete:  completeArgsUsage,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "t,timeout",
				Value: "10s",
				Usage: "Wait at most <n> seconds before bailing out",
			},
			cli.StringFlag{
				Name:  "m,mask",
				Value: "exact,domain,user,email",
				Usage: "Indicate what part of the id you want to query for",
			},
		},
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
		Usage: "Show what has changed in the current commit.",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "tree,t",
				Usage: "View the status as a tree listing.",
			},
		},
		Description: `This a shortcut for »brig diff HEAD CURR«.
See the »diff« command for more information.`,
	},
	"diff": {
		Usage:     "Show what changed between two commits.",
		ArgsUsage: "[<REMOTE>] [<OTHER_REMOTE> [<REMOTE_REV> [<OTHER_REMOTE_REV>]]]]",
		Complete:  completeArgsUsage,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "list,l",
				Usage: "Output the diff as simple list (like status does by default)",
			},
			cli.BoolFlag{
				Name:  "offline,o",
				Usage: "Do no fetch operation before computing the diff.",
			},
			cli.BoolFlag{
				Name:  "self,s",
				Usage: "Assume self as owner of both sides and compare only commits.",
			},
			cli.BoolFlag{
				Name:  "missing,m",
				Usage: "Show missing files in diff output.",
			},
		},
		Description: `View what sync would do when being called on the specified points in history.

   Diff does not show what changed inside of the files, but shows how the files
   themselves changed compared to the remote. To describe this, brig knows
   seven different change types:

   - Added (+): The file was added on the remote side.
   - Removed (-): The file was removed on the remote side.
   - Missing (_): The file is missing on the remote side (e.g. we added it)
   - Moved (→): The file was moved to a new location.
   - Ignored (*): This file was ignored because we chose to due to our settings.
   - Mergeable (⇄): Both sides have changes, but they can be merged.
   - Conflict (⚡): Both sides have changes but they conflict.

   Before computing the diff, it will try to fetch the metadata from the peer,
   if necessary. If you do not want this behaviour, use the »--offline« flag.

   See »brig commit« for a general explanation of commits.

EXAMPLES:

   $ brig diff                       # Show diff from our CURR to our HEAD
   $ brig diff alice                 # Show diff from our CURR to alice's last state
   $ brig diff alice some_tag        # Show diff from our CURR to 'some_tag' of alice
   $ brig diff alice bob HEAD HEAD   # Show diff between alice and bob's HEAD
   $ brig diff -s HEAD CURR          # Show diff between HEAD and CURR of alice
`,
	},
	"tag": {
		Usage:     "Tag a commit with a specific name",
		Complete:  completeArgsUsage,
		ArgsUsage: "<commit> <name>",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "delete,d",
				Usage: "Delete the tag instead of creating it",
			},
		},
		Description: `Give a name to a commit, which is easier to remember than the hash.
   You can use the name you gave in all places where brig requires you to specify a commit.

   There are three special tags pre-defined for you:

   - CURR: A reference to the staging commit.
   - HEAD: The last fully completed commit.
   - INIT: The very first commit in the chain.

   Tags are case insensitive. That means that »HEAD« and »head« mean the same.
   If you want to specify a commit by its index, you can use the special syntax
   »commit[$idx]« where »$idx« can be a zero-indexed number. The first commit
   has the index of zero.

   If you want to access the previous commit, you can also use the special
   syntax »$rev^« where »$rev« is any revision (either a commit hash, a tag
   name or anything else).  The circumflex can be used more than once to go
   back further.

EXAMPLES:

   $ brig tag SEfXUAH6AR my-tag-name   # Name the commit SEfXUAH6AR 'my-tag-name'.
   $ brig tag -d my-tag-name           # Delete the tag name again.
   $ brig tag HEAD^ previous-head      # Tag the commit before the current HEAD with "previous-head".
   $ brig tag 'commit[1]' second       # Tag the commit directly after init with "second".
`,
	},
	"log": {
		Usage:    "Show all commits in a certain range",
		Complete: completeArgsUsage,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "format,f",
				Usage: "Format the output according to a template",
			},
		},
		Description: `Show a list of commits from a start (--from) up to and end (--to).
   If omitted »--from INIT --to CURR« will be assumed.

   The output will show one commit per line, each including the (short) hash of the commit,
   the date it was committed and the (optional) commit message.
`,
	},
	"fetch": {
		Usage:     "Fetch all metadata from another peer.",
		ArgsUsage: "<remote>",
		Complete:  completeArgsUsage,
		Description: `This is a plumbing commands and most likely is only needed for debugging.

   Get all the latest metadata of a certain peer.
   This does not download any actual data, but only the metadata of it.
   You have to be authenticated to the user to get his data.

   Fetch will be done automatically by »sync« and »diff« and is usually
   only helpful when doing it together with »become«.`,
	},
	"sync": {
		Usage:     "Sync with another peer",
		ArgsUsage: "<remote>",
		Complete:  completeArgsUsage,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "no-fetch,n",
				Usage: "Do not do a fetch before syncing.",
			},
			cli.BoolFlag{
				Name:  "quiet,q",
				Usage: "Do not print what changed.",
			},
		},
		Description: `Sync and merge all metadata of another peer with our metadata.
   After this operation you might see new files in your folder.
   Those files were not downloaded yet and will be only on the first access.

   It is recommended that your first check what will be synced with »brig diff«.

   When passing no arguments, 'sync' will synchronize with all online remotes.
   When passing a single argument, it will be used as the remote name to sync with.

   The symbols in the output prefixing every path have the following meaning:

	+	The file is only present on the remote side.
	-	The file was removed on the remote side.
	→	The file was moved to a new location.
	*	This file was ignored because we chose to, due to our settings.
	⇄	Both sides have changes, but they are compatible and can be merged.
	⚡   Both sides have changes, but they are incompatible and result in conflicts.
	_	The file is missing on the remote side.

	See also »brig help diff« for some more details.
	Files from other remotes are not pinned automatically.
`,
	},
	"push": {
		Usage:    "Ask a remote to sync with us.",
		Complete: completeArgsUsage,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "dry-run,d",
				Usage: "Do not the actual push, but check if we may push.",
			},
		},
		Description: ``,
	},
	"commit": {
		Usage:    "Create a new commit",
		Complete: completeArgsUsage,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "message,m",
				Value: "",
				Usage: "Provide a meaningful commit message.",
			},
		},
		Description: `Create a new commit.

   The message (»--message«) is optional. If you do not pass it, a message will
   be generated which contains the current time. The commit history can be
   viewed by »brig log«.

   Think of commits as snapshots that can be created explicitly by you or even
   automated in an interval. It is important to remember that »commit« will
   only create a snapshot of the metadata. It is not guaranteed that you can
   still access the actual data of very old versions (See »brig help )

   You normally do not need to issue this command manually, since there is a
   loop inside of brig that will auto-commit every 5 minute (default; see the
   "fs.autocommit.interval" config key). Sync operations will also create
   commits implicitly and every change from the gateway side will also result
   in a commit.
`,
	},
	"reset": {
		Usage:     "Reset a file or the whole commit to an old state.",
		ArgsUsage: "<commit> [<file>]",
		Complete:  completeArgsUsage,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "force,f",
				Usage: "Reset even when there are changes in the staging area",
			},
		},
		Description: `Reset a file to an old state by specifying the commit it
   should be reverted to. If you do not pass »<file>« the whole commit will be
   filled with the contents of the old commit.

   If you reset to an old commit and you have uncommitted changes, brig will warn you
   about that and refuse the »reset« unless you pass »--force«.

   Note for git users: It is not possible to go back in history and branch out
   from there.  »reset« simply overwrites the staging commit (CURR) with an old
   state, thus keeping all the previous history. You can always jump back to
   the previous state. In other words: the reset operation of brig is not
   destructive. If you notice that you do not like the state you've reseted to,
   »brig reset head« will bring you back to the last known good state.
`,
	},
	"become": {
		Usage:     "View the data of another user",
		ArgsUsage: "<remote>",
		Complete:  completeArgsUsage,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "self,s",
				Usage: "Become self (i.e. the owner of the repository)",
			},
		},
		Description: `View the data of another user.

   This is a plumbing command and meant for debugging.

   You can temporarily explore the metadata of another user, by »becoming«
   them. Once you became a certain user (which needs to be in your remote list
   and on which you called »brig fetch« before), you can look around in the
   data like in yours. You can also modify files, but keep in mind that they
   will be reset on he next fetch.
`,
	},
	"history": {
		Usage:     "Show the history of a file or directory",
		ArgsUsage: "<path>",
		Complete:  completeBrigPath(true, true),
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "empty,e",
				Usage: "Also show commits where nothing happens",
			},
		},
		Description: `Show a list of all changes that were made to this path.

   Not every change you ever made is recorded, but the change between each commit.
   In other words: If you modify a file, delete it and re-add all in one commit, then
   brig will see it only as one modification.

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
		Usage:     "Add a local file to the storage.",
		ArgsUsage: "(<local-path> [<path>]|--stdin <path>)",
		Complete:  completeLocalPath,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "stdin,i",
				Usage: "Read data from stdin.",
			},
		},
		Description: `Read a local file (given by »local-path«) and try to read
   it. This is the conceptual equivalent of »git add«. The stream will be encrypted
   and possibly compressed before saving it to IPFS.

   If you omit »path«, the file will be added under the root
   directory, with the basename of »local-path«. You can change this by
   specifying where to save the local file by additionally passing »path«.

   Additionally you can read the file from standard input if you pass »--stdin«.
   In this case you pass only one path: The path where the stream is stored.

EXAMPLES:

   $ brig stage file.png                   # gets added as /file.png
   $ brig stage file.png /photos/me.png    # gets added as /photos/me.png
   $ cat file.png | brig --stdin /file.png # gets added as /file.png`,
	},
	"touch": {
		Usage:     "Create an empty file under the specified path",
		ArgsUsage: "<path>",
		Complete:  completeBrigPath(true, false),
		Description: `Convenience command for adding empty files.

   If the file or directory already exists, the modification time is updated to
   the current timestamp (like the original touch(1) does).
`,
	},
	"cat": {
		Usage:     "Output the content of a file to standard output",
		ArgsUsage: "[<path>]",
		Complete:  completeBrigPath(true, false),
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "offline,o",
				Usage: "Only output the file if it is cached locally.",
			},
		},
		Description: `Decrypt and decompress the stream from IPFS and write it to standard output.

   When specifying a directory instead of a file, the directory content will be
   output as tar archive. This is useful when saving a whole directory tree to
   disk (see also EXAMPLES).

   When no path is specified, »/« is assumed and all contents are outputted as tar.

EXAMPLES:

   # Output a single file:
   $ brig cat photo.png
   # Create a tar from root and unpack it to the current directory.
   $ brig cat | tar xfv -
   # Create .tar.gz out of of the /photos directory.
   $ brig cat photos | gzip -f > photos.tar.gz
`,
	},
	"show": {
		Usage:     "Show metadata of a file or directory or commit",
		ArgsUsage: "<path>",
		Complete:  completeBrigPath(true, true),
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "format,f",
				Usage: "Format the output according to a template",
			},
		},
		Description: `Show all metadata attributes known for a file or directory.

   Path: Absolute path of the file inside of the storage.
   User: User which modified the file last.
   Type: »file« or »directory«.
   Size: Exact content size in bytes.
   Hash: Hash of the node.
   Inode: Internal inode. Also shown as inode in FUSE.
   IsPinned: »yes« if the file is pinned, »no« else.
   IsExplicit: »yes« if the file is pinned explicitly, »no« elsewise.
   ModTime: Timestamp of last modification.
   ContentHash: Content hash of the file before encryption.
   BackendHash: Hash of the node in ipfs (ipfs cat <this hash>)
   TreeHash: Hash that is unique to this node.
`,
	},
	"rm": {
		Usage:     "Remove a file or directory",
		ArgsUsage: "<path>",
		Complete:  completeBrigPath(true, true),
		Description: `Remove a file or directory.

   In contrast to the usual rm(1) there is no --recursive switch.
   Directories are deleted recursively by default.

   Even after deleting files, you will be able to access its history by using
   the »brig history« command and bring them back via »brig reset«. If you want
   to restore a deleted entry you are able to with the »brig reset« command.
`,
	},
	"ls": {
		Usage:     "List files and directories.",
		ArgsUsage: "<path>",
		Complete:  completeBrigPath(false, true),
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "depth,d",
				Usage: "Max depth to traverse",
				Value: 1,
			},
			cli.BoolFlag{
				Name:  "recursive,R",
				Usage: "Allow recursive traverse",
			},
			cli.StringFlag{
				Name:  "format,f",
				Usage: "Format the output according to a template",
			},
		},
		Description: `List files an directories starting with »path«.
   If no »<path>« is given, the root directory is assumed. Every line of »ls«
   shows a human readable size of each entry, the last modified time stamp, the
   user that last modified the entry (if there's more than one) and if the
   entry if pinned.
`,
	},
	"tree": {
		Usage:     "List files and directories in a tree",
		ArgsUsage: "<path>",
		Complete:  completeBrigPath(false, true),
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "depth, d",
				Usage: "Max depth to traverse",
				Value: -1,
			},
		},
		Description: `Show entries in a tree(1)-like fashion.
`,
	},
	"mkdir": {
		Usage:     "Create an empty directory",
		ArgsUsage: "<path>",
		Complete:  completeBrigPath(false, true),
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "parents, p",
				Usage: "Create parent directories as needed",
			},
		},
		Description: `Create an empty directory at the specified »path«.
   By default, parent directories are not created. You can use »--parents« to
   enable this behaviour.
`,
	},
	"mv": {
		Usage:     "Move a file or directory from »src« to »dst«",
		ArgsUsage: "<src> <dst>",
		Complete:  completeBrigPath(true, true),
		Description: `Move a file or directory from »src« to »dst.«

   If »dst« already exists and is a file, it gets overwritten with »src«.
   If »dst« already exists and is a directory, »basename(src)« is created inside,
   (if the file inside does not exist yet)

   It's not allowed to move a directory into itself.
   This includes moving the root directory.
`,
	},
	"cp": {
		Usage:     "Copy a file or directory from »src« to »dst«",
		ArgsUsage: "<src> <dst>",
		Complete:  completeBrigPath(true, true),
		Description: `Copy a file or directory from »src« to »dst«.

   The semantics are the same as for »brig mv«, except that »cp« does not remove »src«.
`,
	},
	"edit": {
		Usage:     "Edit a file in place with $EDITOR",
		ArgsUsage: "<path>",
		Complete:  completeBrigPath(true, false),
		Description: `Convenience command to read the file at »path« and display it in $EDITOR.

   Once $EDITOR quits, the file is saved back.

   If $EDITOR is not set, nano is assumed (I cried a little).
   If nano is not installed this command will fail and you neet to set $EDITOR>

`,
	},
	"daemon": {
		Usage:    "Daemon management commands.",
		Complete: completeSubcommands,
		Description: `Commands to manually start or stop the daemon.

   The daemon process is normally started whenever you issue the first command
   (like »brig init« or later on a »brig ls«). Once you entered your password
   (if you did not specify a password helper), it will be started for you in
   the background. Therefore it is seldom useful to use any of those commands -
   unless you know what you're doing.
`,
	},
	"daemon.launch": {
		Usage:    "Start the daemon process in the foreground",
		Complete: completeArgsUsage,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "trace,t",
				Usage: "Create tracing output suitable for `go tool trace`",
			},
			cli.BoolFlag{
				Name:  "s,log-to-stdout",
				Usage: "Log all messages to stdout instead of syslog",
			},
		},
		Description: `Start the dameon process in the foreground.


EXAMPLES:

   $ brig daemon quit        # Shut down any previous daemon.
   $ brig daemon launch -s   # Start in foreground and log to stdout.
`,
	},
	"daemon.quit": {
		Usage:    "Quit a running daemon process",
		Complete: completeArgsUsage,
		Description: `Quit a running daemon process.

   If no daemon process is running, it will tell you.
`,
	},
	"daemon.ping": {
		Usage:    "Check if the daemon is running and reachable",
		Complete: completeArgsUsage,
		Description: `Send up to 100 ping packages to the daemon
   and also print the roundtrip time for each.
`,
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "c,count",
				Usage: "How many times to ping the daemon",
				Value: 10,
			},
		},
	},
	"config": {
		Usage:    "View and modify config options.",
		Complete: completeSubcommands,
		Description: `Commands for getting, setting and listing configuration values.

   Each config key is a dotted path (»a.b.c«), associated with one key.  These
   configuration values can help you to fine tune the behaviour of brig. In contrast
   to many other programs the config is applied immediately after setting it (where possible).
   Furthermore, each config key will describe itself and tell you if it needs a restart.

   For more details on each config value, type 'brig config ls'.

   Without further arguments »brig cfg« is a shortcut for »brig cfg ls«.
`,
	},
	"config.get": {
		Usage:       "Get a specific config key",
		Complete:    completeArgsUsage,
		ArgsUsage:   "<key>",
		Description: `Show the current value of a key`,
	},
	"config.doc": {
		Usage:     "Show the docs for this config key",
		Complete:  completeArgsUsage,
		ArgsUsage: "<key>",
		Description: `For each config key a few metadata entries are assigned.

This includes a string describing the usage, the default value and an indicator
if the service needs a restart when setting the value.

`,
	},
	"config.set": {
		Usage:     "Set a specific config key to a new value.",
		Complete:  completeArgsUsage,
		ArgsUsage: "<key> <value>",
		Description: `Set the value at »key« to »value«.

   Some config values have associated validators that will tell you if a value is not allowed.
   Also you will be warned if the config key requires a restart.
`,
	},
	"config.list": {
		Usage:       "List all existing config keys",
		Complete:    completeArgsUsage,
		Description: `List all existing config keys.`,
	},
	"fstab": {
		Usage:       "Manage mounts that will be mounted on startup of the daemon.",
		Description: "This is the conceptual equivalent of the normal fstab(5).",
	},
	"fstab.add": {
		Usage: "Add a new mount entry to fstab",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "r,readonly",
				Usage: "Create the filesystem as readonly.",
			},
			cli.BoolFlag{
				Name:  "offline,o",
				Usage: "Error out on files that are only remotely available.",
			},
			cli.StringFlag{
				Name:  "x,root",
				Usage: "Specify a root directory other than »/«.",
			},
		},
	},
	"fstab.remove": {
		Usage: "Remove a mount from fstab.",
	},
	"fstab.list": {
		Usage: "List all items in the filesystem table.",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "format,f",
				Usage: "Format the output according to a template.",
			},
		},
	},
	"fstab.apply": {
		Usage:       "Sync the reality with the mounts in fstab.",
		Description: "Mounts and unmounts directories as necessary.",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "u,unmount",
				Usage: "Unmount all mounts in the filesystem table.",
			},
		},
	},
	"mount": {
		Usage:     "Mount the contents of brig as FUSE filesystem to »mount_path«.",
		ArgsUsage: "<mount_path>",
		Complete:  completeArgsUsage,
		Description: `Show the all files and directories inside a normal
   directory. This directory is powered by a userspace filesystem which
   allows you to read and edit data like you are use to from from normal
   files. It is compatible to existing tools and allows brig to interact
   with filebrowsers, video players and other desktop tools.

   It is possible to have more than one mount. They will show the same content.

CAVEATS

   Editing large files will currently eat big amounts of memory, proportional
   to the size of the file.  We advise you to use normal commands like »brig
   cat« and »brig stage« until this is fixed.

   At this time, the filesystem also not very robust to files that timeout or
   error out otherwise. Consider this feature to be experimental while this has
   not been worked upon.
   `,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "no-mkdir",
				Usage: "Do not create the mount directory if it does not exist",
			},
			cli.BoolFlag{
				Name:  "r,readonly",
				Usage: "Create the filesystem as readonly",
			},
			cli.BoolFlag{
				Name:  "offline,o",
				Usage: "Error out on files that are only remotely available.",
			},
			cli.StringFlag{
				Name:  "x,root",
				Usage: "Create the filesystem as readonly",
			},
		},
	},
	"unmount": {
		Usage:     "Unmount a previously mounted directory",
		ArgsUsage: "<mount_path>",
		Complete:  completeArgsUsage,
		Description: `Unmount a previously mounted directory.

   All mounts get automatically unmounted once the daemon shuts down.
   In case the daemon crashed or failed to unmount, you can manually
   use this command to reclaim the mount point:

   $ fusermount -u -z /path/to/mount
`,
	},
	"version": {
		Usage:    "Show the version of brig and IPFS",
		Complete: completeArgsUsage,
		Description: `Show the version of brig and IPFS.

   This includes the client and server version of brig.
   These two values should be ideally exactly the same to avoid problems.

   Apart from that, the version of IPFS is shown here.

   If available, also the git rev is included. This is useful to get the exact
   state of the software in case of problems.

   Additionally the build time of the binary is shown.
   Please include this information when reporting a bug.
`,
	},
	"gc": {
		Usage:    "Trigger the garbage collector",
		Complete: completeArgsUsage,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "aggressive,a",
				Usage: "Also run the garbage collector on all file systems immediately",
			},
		},
		Description: `Manually trigger the garbage collector.

   Strictly speaking there are two garbage collectors in the system.  The
   garbage collector of IPFS cleans up all unpinned files from local storage.
   This still means that the objects referenced there can be retrieved from
   other network nodes, but not locally anymore. This might save alot of space.

   The other garbage collector is not very important to the user and cleans up
   unused references inside of the metadata store. It is only run if you pass
   »--aggressive«.
`,
	},
	"docs": {
		Usage: "Open the online documentation in your default web browser.",
	},
	"trash": {
		Usage: "Control the trash bin contents.",
		Description: `

   The trash bin is a convenience interface to list and restore deleted files.
   It will list all files that were deleted and were not overwritten by other files.
		`,
	},
	"trash.list": {
		Usage: "List all items in the trash bin.",
	},
	"trash.undelete": {
		Usage: "Restore a path from the trashbin.",
	},
	"gateway": {
		Usage: "Control the HTTP/S gateway service.",
		Description: `The gateway serves a UI and download endpoints over a browser.
   This enables users that do not use brig directly to still browse, edit and download files
   For having access to the gateway, users need to be created. By default no users are created.
   Create an »admin« user (password is also »admin«) with this command:

     $ brig gw user add admin admin --role-admin

   Most of the gateway is configured exclusively via config variables.  Please
   refer to the individual config keys for more information (they all start
   with »gateway.«). The »brig gw status« command will also give you a nice, readable
   overview of what the current state is and how you can improve it.
`,
	},
	"gateway.start": {
		Usage: "Start the gateway.",
		Description: `
   It is recommended to check the state with a »brig gw status« afterwards.
   This will give you important hints if something went wrong or needs attention.
`,
	},
	"gateway.stop": {
		Usage: "Stop the gateway.",
	},
	"gateway.status": {
		Usage: "Print a diagnostic report on the status of the gateway.",
	},
	"gateway.cert": {
		Usage: "Helper to get a LetsEncrypt certificate. Needs root.",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "cache-dir,d",
				Usage: "In what directory to save the certificates in. Defaults to $HOME/.cache/brig",
			},
		},
		Description: `

   This will start an HTTP Server on port 80 (thus requiring root access) and
   negotiates a LetsEncrypt over it.  For this to work you need to set
   »gateway.cert.domain« to a valid domain name.

EXAMPLES:

   $ brig cfg set gateway.cert.domain your.domain.org
   $ sudo brig gw cert
`,
	},
	"gateway.url": {
		Usage: "Helper to print the URL to a named file or directory.",
	},
	"gateway.user": {
		Usage: "Control the user account that can access the HTTP gateway.",
	},
	"gateway.user.add": {
		Usage: "Add a new gateway user.",
		ArgsUsage: "<user> [<password> <permitted folders list>]",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "role-admin,a",
				Usage: "Add this user as admin (short for »-r 'fs.view,fs.edit,fs.download,remotes.view,remotes.edit'«)",
			},
			cli.BoolFlag{
				Name:  "role-editor,b",
				Usage: "Add this user as collaborator (short for »-r 'fs.view,fs.edit,fs.download,remotes.view'«)",
			},
			cli.BoolFlag{
				Name:  "role-collaborator,c",
				Usage: "Add this user as collaborator (short for »-r 'fs.view,fs.edit,fs.download'«)",
			},
			cli.BoolFlag{
				Name:  "role-viewer,d",
				Usage: "Add this user as viewer (short for »-r 'fs.view,fs.download'«)",
			},
			cli.BoolFlag{
				Name:  "role-link-only,e",
				Usage: "Add this user as linker (short for »-r 'fs.download'«)",
			},
			cli.StringFlag{
				Name:  "rights,r",
				Usage: "Comma separated list of rights of this user.",
			},
		},
		Description: `
   The rights are as follows:

   fs.view: View and list all files.
   fs.edit: Edit and create new files.
   fs.download: Download file content.
   remotes.view: View the remotes tab.
   remotes.edit: Edit the remotes tab.

   If the folder list is empty, this user can access all files.
   If it is non-empty, the user can only access the files including and below all folders.
`,
	},
	"gateway.user.remove": {
		Usage: "Remove a gateway user by its name.",
	},
	"gateway.user.list": {
		Usage: "List all gateway users.",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "format,f",
				Usage: "Format the output by a template.",
			},
		},
		Description: `
   List all gateway users.

   The keys accepted by »--format« are:

   - Name: Name of the user.
   - PasswordHash: Hashed password.
   - Salt: Salt of the password.
   - Folders: A list of folders this users may access (might be empty).
   - Rights: A list of rights this users has (might be empty).
`,
	},
	"debug": {
		Usage: "Various debbugging utilities. Use with care.",
	},
	"debug.pprof-port": {
		Usage: "Print the pprof port of the daemon.",
		Description: `
   This is useful if there is a performance issue (high cpu consumption in idle e.g.).
   See here for some examples of what you can do: https://golang.org/pkg/net/http/pprof

EXAMPLES:

   # Show a graph with a cpu profile of the last 30s:
   go tool pprof -web "http://localhost:$(brig d p)/debug/pprof/profile?seconds=30"
`,
	},
	"bug": {
		Usage: "Print a template for bug reports.",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "stdout,s",
				Usage: "Always print the report to stdout; do not open a browser",
			},
		},
	},
}

func injectHelp(cmd *cli.Command, path string) {
	help, ok := helpTexts[path]
	if !ok {
		die(fmt.Sprintf("bug: no such help entry: %v", path))
	}

	cmd.Usage = help.Usage
	cmd.ArgsUsage = help.ArgsUsage
	cmd.Description = help.Description
	cmd.BashComplete = help.Complete
	cmd.Flags = help.Flags
}

func translateHelp(cmds []cli.Command, prefix []string) {
	for idx := range cmds {
		path := append(append([]string{}, prefix...), cmds[idx].Name)
		injectHelp(&cmds[idx], strings.Join(path, "."))
		translateHelp(cmds[idx].Subcommands, path)
	}
}

// TranslateHelp fills in the usage and description for each command.
// This is separated from the command definition to make things more readable,
// and separate logic from the (lengthy) documentation.
func TranslateHelp(cmds []cli.Command) []cli.Command {
	translateHelp(cmds, nil)
	return cmds
}

// handleOpenHelp opens the online documentation a webbrowser.
func handleOpenHelp(ctx *cli.Context) error {
	url := "https://brig.readthedocs.org"
	if err := webbrowser.Open(url); err != nil {
		fmt.Printf("could not open browser for you: %v\n", err)
		fmt.Printf("Please open this link yourself:\n\n\t%s\n", url)
	} else {
		fmt.Printf("A new tab was opened in your browser.\n")
	}

	return nil
}
