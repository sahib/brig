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
	panic(msg)
}

var explicitPinFlags = []cli.Flag{
	cli.StringFlag{
		Name:  "from,f",
		Value: "HEAD",
		Usage: "Specify the commit to start from",
	},
	cli.StringFlag{
		Name:  "to,t",
		Value: "INIT",
		Usage: "Specify the maximum commit to iterate to",
	},
}

var helpTexts = map[string]helpEntry{
	"init": {
		Usage:     "Initialize a new repository",
		ArgsUsage: "<username>",
		Complete:  completeArgsUsage,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "backend,b",
				Value: "httpipfs",
				Usage: "What data backend to use for the new repo",
			},
			cli.StringFlag{
				Name:  "w,pw-helper",
				Value: "",
				Usage: "A commandline to specify a password helper",
			},
			cli.BoolFlag{
				Name:  "no-password,x",
				Usage: "Use a static password. Not recommended besides testing.",
			},
			cli.BoolFlag{
				Name:  "empty,e",
				Usage: "Do not create an initial README.",
			},
			cli.BoolFlag{
				Name:  "no-logo,n",
				Usage: "Do not display the super pretty logo on init.",
			},
			cli.IntFlag{
				Name:  "ipfs-port,P",
				Usage: "Port of the HTTP API to connect to",
				Value: 5001,
			},
		},
		Description: `Initialize a new repository with a certain backend.

   If BRIG_PATH is set, the new repository will be created at this place.  If not,
   the current working directory is used. If the directory is not empty, brig will
   warn you about it and abort.

   The username can be specified as pretty much any string, but it is recommended
   to use the special format »user@domain.something/resource«. This is similar to
   XMPP IDs. Specifying a resource can help you use the same name for different
   computers and specifying a domain makes it possible to indicate groups.  This
   is especially important for commands like »brig net locate«.

   You will be asked to enter a password if you did not specify -w. This
   password will be used to encrypt the repository while brig is not running.
   For ease of use we recommend to set the password with the -w option. This
   allows you to specify a password command that will print the desired
   password as output. This output is read by brig and used as password. For
   testing you could use »-w "echo mypass"«, while for serious use, you should
   use something like »pass brig/desktop/password«.
`,
	},
	"whoami": {
		Usage:    "Print the own remote identity",
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
				Usage: "Only print the addr portion of the fingerprint",
			},
			cli.BoolFlag{
				Name:  "key,k",
				Usage: "Only print the key portion of the fingerprint",
			},
		},
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

EXAMPLES:

   # Show a diff for each remote:
   $ brig remote list --format '{{ .Name }}' | xargs -n 1 brig diff
`,
	},
	"remote.add": {
		Usage:       "Add a new remote",
		ArgsUsage:   "<name> <fingerprint>",
		Complete:    completeArgsUsage,
		Description: "Add a new remote under a handy name with their fingerprint.",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "auto-update,a",
				Usage: "Take automatic updates from this node",
			},
			cli.StringSliceFlag{
				Name:  "folder,f",
				Usage: "Configure the folders this remote may see",
			},
		},
	},
	"remote.remove": {
		Usage:       "Remove a remote by name",
		ArgsUsage:   "<name>",
		Complete:    completeArgsUsage,
		Description: "Remove a remote by name",
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
		Description: `This goes over every entry in your remote list and prins
   his name, network address, rountrip and when we was last seen this
   remote.`,
	},
	"remote.clear": {
		Usage:       "Clear the complete remote list",
		Complete:    completeArgsUsage,
		Description: "Clear the complete remote list. Note that you cannot undo this operation.",
	},
	"remote.ping": {
		Usage:    "Ping a remote",
		Complete: completeArgsUsage,
		Description: `Ping a remote and check if we can reach them.

   There is a small difference to the »remote list« command. »ping« will only work
   if both sides authenticated each other and can thus be used as a test for this.
   Additionally, it shows the roundtrip time, the ping request took to travel.
`,
	},
	"remote.edit": {
		Usage:    "Edit the current list",
		Complete: completeArgsUsage,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "yml,y",
				Value: "",
				Usage: "Directly overwrite remote list with yml file",
			},
		},
		Description: `Edit the current list using $EDITOR as YAML file.
   It will be updated upon saving`,
	},
	"remote.auto-update": {
		Usage:       "",
		Complete:    completeArgsUsage,
		Description: ``,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "no-initial-sync,n",
				Usage: "Do not sync initially when enabling the update",
			},
		},
	},
	"remote.folder": {
		Usage:       "List the folders of all remotes.",
		Complete:    completeArgsUsage,
		Description: ``,
	},
	"remote.folder.add": {
		Usage:       "Add a remote folder for a specific remote.",
		Complete:    completeArgsUsage,
		Description: ``,
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
		Usage:     "Commands to pin a certain file",
		ArgsUsage: "<file>",
		Complete:  completeBrigPath(true, true),
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
		Usage:       "Recaculate pinning based on fs.repin.{quota,min_depth,max_depth}",
		ArgsUsage:   "[<root>]",
		Complete:    completeBrigPath(true, true),
		Description: `This can be used to easily get rid of too old versions`,
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
	"net.locate": {
		Usage:     "Try to locate a remote by their name or by a part of it",
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
		Usage: "Show what has changed in the current commit",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "tree,t",
				Usage: "View the status as a tree listing",
			},
		},
		Description: `This a shortcut for »brig diff HEAD CURR«.
See the »diff« command for more information.`,
	},
	"diff": {
		Usage:     "Show what changed between two commits",
		ArgsUsage: "[<REMOTE>] [<OTHER_REMOTE> [<REMOTE_REV> [<OTHER_REMOTE_REV>]]]]",
		Complete:  completeArgsUsage,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "list,l",
				Usage: "Output the diff as simple list (like status)",
			},
			cli.BoolFlag{
				Name:  "offline,o",
				Usage: "Do no fetch before computing the diff",
			},
			cli.BoolFlag{
				Name:  "self,s",
				Usage: "Assume self as owner of both sides and compare only commits",
			},
			cli.BoolFlag{
				Name:  "missing,m",
				Usage: "Show missing files in diff output",
			},
		},
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

EXAMPLES:

   $ brig tag SEfXUAH6AR my-tag-name   # Name the commit SEfXUAH6AR 'my-tag-name'.
   $ brig tag -d my-tag-name           # Delete the tag name again.
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
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "no-fetch,n",
				Usage: "Do not do a fetch before syncing",
			},
			cli.BoolFlag{
				Name:  "quiet,q",
				Usage: "Do not print what changed",
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
`,
	},
	"commit": {
		Usage:    "Create a new commit",
		Complete: completeArgsUsage,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "message,m",
				Value: "",
				Usage: "Provide a meaningful commit message",
			},
		},
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
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "self,s",
				Usage: "Become self (i.e. the owner of the repository)",
			},
		},
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
		Complete:  completeBrigPath(true, true),
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "empty,e",
				Usage: "Also show commits where nothing happens",
			},
		},
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
		Complete:  completeLocalPath,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "stdin,i",
				Usage: "Read data from stdin",
			},
		},
		Description: `Read a local file (given by »local-path«) and try to read
   it. This is the conceptual equivalent of »git add«. The stream will be encrypted
   and possibly compressed before saving it to ipfs.

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
		Description: `Convinience command for adding empty files.

   If the file or directory already exists, the modification time is updated to
   the current timestamp (like the original touch(1) does).
`,
	},
	"cat": {
		Usage:     "Output the content of a file to standard output",
		ArgsUsage: "[<path>]",
		Complete:  completeBrigPath(true, false),
		Description: `Decrypt and decompress the stream from ipfs and write it to standard output.

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
    Pinned: »yes« if the file is pinned, »no« else.
   ModTime: Timestamp of last modification.
   Content: Content hash of the file in ipfs.
`,
	},
	"rm": {
		Usage:     "Remove a file or directory",
		ArgsUsage: "<path>",
		Complete:  completeBrigPath(true, true),
		Description: `Remove a file or directory.
   In contrast to the usual rm(1) there is no --recursive switch.
   Directories are deleted recursively by default.

   Even after deleting files, you will be able to access its history
   by using the »brig history« command. If you want to restore a deleted
   entry you are able to with the »brig reset« command.
`,
	},
	"ls": {
		Usage:     "List files and directories",
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
   shows a human readable size of each entry, the last modified timestamp, the
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
   This command is identical to »brig ls« otherwise.
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
		Usage:     "Edit a file inplae with $EDITOR",
		ArgsUsage: "<path>",
		Complete:  completeBrigPath(true, false),
		Description: `Convinience command to read the file at »path« and display it in $EDITOR.

   Once $EDITOR quits, the file is saved back.

   If $EDITOR is not set, nano is assumed (I cried a little).
   If nano is not installed this command will fail and you neet to set $EDITOR>

`,
	},
	"daemon": {
		Usage:    "Daemon management commands",
		Complete: completeSubcommands,
		Description: `Commands to manually start or stop the daemon.

   The daemon process is normally started whenever you issue the first command
   (like »brig init« or later on a »brig ls«). Once you entered your password,
   it will be started for you in the background. Therefore is seldomely useful
   to use any of those commands - unless you know what you're doing.
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

   $ brig daemon launch
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
		Usage:    "View and modify config options",
		Complete: completeSubcommands,
		Description: `Commands for getting, setting and listing configuration values.

   Each config key is a dotted path, associated with one key.
   These configuration values can help you to finetune the behaviour of brig.

   For more details on each config value, type 'brig config ls'.

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
		Usage:     "Set a specific config key to a new value",
		Complete:  completeArgsUsage,
		ArgsUsage: "<key> <value>",
		Description: `Set the value at »key« to »value«.

   Keep in mind that it might be required to restart the daemon so the new
   values take effect.
`,
	},
	"config.list": {
		Usage:       "List all existing config keys",
		Complete:    completeArgsUsage,
		Description: `List all existing config keys. The output is valid YAML.`,
	},
	"fstab": {
		Usage:       "Manage mounts that will be mounted always",
		Description: "The list of mounts is stored in the config",
	},
	"fstab.add": {
		Usage: "Add a new mount entry to fstab",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "r,readonly",
				Usage: "Create the filesystem as readonly",
			},
			cli.StringFlag{
				Name:  "x,root",
				Usage: "Specify a root directory other than »/«",
			},
		},
	},
	"fstab.remove": {
		Usage: "Remove a mount from fstab",
	},
	"fstab.list": {
		Usage: "List all items in the filesystem table",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "format,f",
				Usage: "Format the output according to a template",
			},
		},
	},
	"fstab.apply": {
		Usage:       "Sync the reality with the mounts in fstab",
		Description: "Mounts and unmounts directories as necessary",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "u,unmount",
				Usage: "Unmount all mounts in the filesystem table",
			},
		},
	},
	"mount": {
		Usage:     "Mount the contents of brig as FUSE filesystem",
		ArgsUsage: "<mount_path>",
		Complete:  completeArgsUsage,
		Description: `Show the all files and directories inside a normal
   directory. This directory is powered by a userspace filesystem which
   allows you to read and edit data like you are use to from from normal
   files. It is compatible to existing tools and allows brig to interact
   with filebrowsers, video players and other desktop tools.

   It is possible to have more than one mount. They will show the same content.

CAVEATS

   Editing large files will currently eat huge amounts of memory.
   We advise you to use normal commands like »brig cat« and »brig stage«
   until this is fixed.`,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "no-mkdir",
				Usage: "Do not create the mount directory if it does not exist",
			},
			cli.BoolFlag{
				Name:  "r,readonly",
				Usage: "Create the filesystem as readonly",
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
   use this command to reclaim the mountpoint:

   $ fusermount -u -z /path/to/mount
`,
	},
	"version": {
		Usage:    "Show the version of brig and ipfs",
		Complete: completeArgsUsage,
		Description: `Show the version of brig and ipfs.

   This includes the client and server version of brig.
   These two values should be ideally exactly the same to avoid problems.

   Apart from that, the version of ipfs is shown here.

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
				Usage: "Also run the garbage collector on all filesystems immediately",
			},
		},
		Description: `Manually trigger the garbage gollector.

   Strictly speaking there are two garbage collectors in the system.  The
   garbage collector of ipfs cleans up all unpinned files from local storage.
   This still means that the objects referenced there can be retrieved from
   other network nodes, but not locally anymore. This might save alot of space.

   The other garbage collector is not very important to the user and cleans up
   unused references inside of the metadata store. It is only run if you pass
   »--aggressive«.

`,
	},
	"docs": {
		Usage: "Open the online documentation in webbrowser.",
	},
	"trash": {
		Usage: "Control the trash bin contents.",
	},
	"trash.list": {
		Usage: "List all items in the trash bin.",
	},
	"trash.remove": {
		Usage: "Remove an item from the trash bin.",
	},
	"gateway": {
		Usage: "Control the HTTP/S gateway service.",
	},
	"gateway.start": {
		Usage: "Start the gateway.",
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
	},
	"gateway.url": {
		Usage: "Print the URL to a named file or directory.",
	},
	"gateway.user": {
		Usage: "Control the user account that can access the HTTP gateway.",
	},
	"gateway.user.add": {
		Usage: "Add a new gateway user.",
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
	},
	"gateway.user.remove": {
		Usage: "Remove a gateway user.",
	},
	"gateway.user.list": {
		Usage: "List all gateway users.",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "format,f",
				Usage: "TODO",
			},
		},
	},
	"debug": {
		Usage: "Various debbugging utilities. Use with care.",
	},
	"debug.pprof-port": {
		Usage: "Print the pprof port of the daemon.",
	},
	"bug": {
		Usage: "Print a template for bug reports",
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
