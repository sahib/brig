package cmd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/sahib/brig/cmd/tabwriter"

	"github.com/fatih/color"
	"github.com/sahib/brig/client"
	"github.com/urfave/cli"
)

func handleReset(ctx *cli.Context, ctl *client.Client) error {
	force := ctx.Bool("force")
	rev := ctx.Args().First()
	path := ""

	if len(ctx.Args()) > 1 {
		path = ctx.Args().Get(1)
	}

	if err := ctl.Reset(path, rev, force); err != nil {
		return ExitCode{UnknownError, fmt.Sprintf("reset: %v", err)}
	}

	return nil
}

func commitName(cmt *client.Commit) string {
	if cmt == nil {
		return ""
	}

	if len(cmt.Tags) > 0 {
		return strings.ToUpper(cmt.Tags[0])
	}

	return cmt.Hash.ShortB58()
}

func handleHistory(ctx *cli.Context, ctl *client.Client) error {
	path := ctx.Args().First()

	history, err := ctl.History(path)
	if err != nil {
		return ExitCode{UnknownError, fmt.Sprintf("history: %v", err)}
	}

	if _, err := ctl.Stat(path); err != nil {
		fmt.Printf("%s %s",
			color.YellowString("WARNING:"),
			`This file is not part of this commit, but there's still history for it.
         Most likely this file was moved or removed in the past.

`)
	}

	tabW := tabwriter.NewWriter(
		os.Stdout, 0, 0, 2, ' ',
		tabwriter.StripEscape,
	)

	containsMoves := false
	for _, entry := range history {
		for _, detail := range entry.Mask {
			if detail == "moved" {
				containsMoves = true
				break
			}
		}

		if containsMoves {
			break
		}
	}

	if len(history) != 0 {
		if containsMoves {
			fmt.Fprintf(tabW, "CHANGE\tFROM\tTO\tHOW\tWHEN\tPIN\t\n")
		} else {
			fmt.Fprintf(tabW, "CHANGE\tFROM\tTO\t\tWHEN\tPIN\t\n")
		}
	}

	for _, entry := range history {
		what := ""
		printLine := true

		for _, detail := range entry.Mask {
			// If it was moved, let's display what moved.
			if detail == "moved" {
				src := entry.WasPreviouslyAt
				dst := entry.Path

				if entry.MovedTo != "" {
					dst = entry.MovedTo
				}

				what = fmt.Sprintf(
					"%s → %s", color.RedString(src), color.RedString(dst),
				)
			}

			// Only display empty changes if nothing happened.
			if detail == "none" && !ctx.Bool("empty") {
				printLine = false
			}
		}

		if !printLine {
			continue
		}

		changeDesc := color.YellowString(strings.Join(entry.Mask, ", "))
		when := color.MagentaString(entry.Head.Date.Format(time.UnixDate))

		fmt.Fprintf(
			tabW,
			"%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			changeDesc,
			color.CyanString(commitName(entry.Next)),
			color.GreenString(commitName(entry.Head)),
			what,
			when,
			pinStateToSymbol(entry.IsPinned, entry.IsExplicit),
			entry.Head.Msg,
		)
	}

	return tabW.Flush()
}

// makePathAbbrev tries to abbreviate the `dst` path if
// both are in the same directory.
func makePathAbbrev(srcNd, dstNd client.StatInfo) string {
	src, dst := srcNd.Path, dstNd.Path
	if path.Dir(src) == path.Dir(dst) {
		dstBase := path.Base(dst)
		if dstNd.IsDir {
			return dstBase + "/"

		}

		return dstBase
	}

	relPath, err := filepath.Rel(path.Dir(src), dst)
	if err != nil {
		fmt.Println("Failed to get relative path: ", err)
		if dstNd.IsDir {
			return dst + "/"
		}

		return dst
	}

	// We could also possibly check here if relPath is longer than dst
	// and only display the relative version then. But being consistent
	// is more valuable here I think.
	if dstNd.IsDir {
		return relPath + "/"
	}

	return relPath
}

func suffixIfDir(nd *treeNode) string {
	if nd.entry.IsDir {
		return nd.name + "/"
	}

	return nd.name
}

const (
	diffTypeNone = iota
	diffTypeAdded
	diffTypeRemoved
	diffTypeMissing
	diffTypeMoved
	diffTypeIgnored
	diffTypeConflict
	diffTypeMerged
)

type diffEntry struct {
	typ  int
	pair client.DiffPair
}

// Called to format each name in the resulting tree:
func printDiffTreeLineFormatter(types map[string]diffEntry, n *treeNode) string {
	if n.name == "/" {
		return color.MagentaString("•")
	}

	if diffEntry, ok := types[n.entry.Path]; ok {
		switch diffEntry.typ {
		case diffTypeAdded:
			return color.GreenString(" + " + "▩ ← " + suffixIfDir(n))
		case diffTypeRemoved:
			return color.RedString(" - " + suffixIfDir(n) + " ← ▩")
		case diffTypeMissing:
			return color.MagentaString(" _ " + suffixIfDir(n) + " → ▩")
		case diffTypeIgnored:
			return color.YellowString(" * " + suffixIfDir(n))
		case diffTypeMoved:
			srcPath := makePathAbbrev(diffEntry.pair.Dst, diffEntry.pair.Src)
			dstBase := path.Base(diffEntry.pair.Dst.Path)
			if diffEntry.pair.Src.IsDir {
				dstBase += "/"
			}

			return color.CyanString(fmt.Sprintf(" %s ↔ %s", dstBase, srcPath))
		case diffTypeMerged:
			dstPath := makePathAbbrev(diffEntry.pair.Dst, diffEntry.pair.Src)
			srcBase := path.Base(diffEntry.pair.Src.Path)
			if diffEntry.pair.Src.IsDir {
				srcBase += "/"
			}
			// Attempt to figure out which way merge should go
			// based on modification times.
			// This information was available at resolver time in the PairDiff
			// but server returns simplified PairDiff without modification masks.
			srcModTime := diffEntry.pair.Src.ModTime
			dstModTime := diffEntry.pair.Dst.ModTime
			var mergeSymbol string = color.MagentaString("→")
			if srcModTime.After(dstModTime) {
				mergeSymbol = color.GreenString("←")
			}

			return color.WhiteString(fmt.Sprintf(" %s %s %s ", dstPath, mergeSymbol, srcBase))
		case diffTypeConflict:
			dstPath := makePathAbbrev(diffEntry.pair.Dst, diffEntry.pair.Src)
			srcBase := path.Base(diffEntry.pair.Src.Path)
			if diffEntry.pair.Src.IsDir {
				srcBase += "/"
			}

			return color.MagentaString(fmt.Sprintf(" %s ⚡%s", dstPath, srcBase))
		}
	}

	return n.name
}

func printDiffTree(diff *client.Diff, printMissing bool) {
	entries := []client.StatInfo{}
	types := make(map[string]diffEntry)

	// Singular types:
	for _, info := range diff.Added {
		types[info.Path] = diffEntry{typ: diffTypeAdded}
		entries = append(entries, info)
	}
	for _, info := range diff.Removed {
		types[info.Path] = diffEntry{typ: diffTypeRemoved}
		entries = append(entries, info)
	}

	if printMissing {
		for _, info := range diff.Missing {
			types[info.Path] = diffEntry{typ: diffTypeMissing}
			entries = append(entries, info)
		}
	}

	for _, info := range diff.Ignored {
		types[info.Path] = diffEntry{typ: diffTypeIgnored}
		entries = append(entries, info)
	}

	// Pair types:
	for _, pair := range diff.Moved {
		types[pair.Dst.Path] = diffEntry{
			typ:  diffTypeMoved,
			pair: pair,
		}
		entries = append(entries, pair.Dst)
	}
	for _, pair := range diff.Conflict {
		types[pair.Dst.Path] = diffEntry{
			typ:  diffTypeConflict,
			pair: pair,
		}
		entries = append(entries, pair.Dst)
	}
	for _, pair := range diff.Merged {
		types[pair.Dst.Path] = diffEntry{
			typ:  diffTypeMerged,
			pair: pair,
		}
		entries = append(entries, pair.Dst)
	}

	if len(entries) == 0 {
		// Nothing to show:
		return
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Path < entries[j].Path
	})

	// Render the tree:
	showTree(entries, &treeCfg{
		format: func(n *treeNode) string {
			return printDiffTreeLineFormatter(types, n)
		},
		showPin: false,
	})
}

func isEmptyDiff(diff *client.Diff) bool {
	return 0 == 0+
		len(diff.Added)+
		len(diff.Conflict)+
		len(diff.Ignored)+
		len(diff.Merged)+
		len(diff.Missing)+
		len(diff.Moved)+
		len(diff.Removed)
}

func printDiff(diff *client.Diff, printMissing bool) {
	simpleSection := func(heading string, infos []client.StatInfo) {
		if len(infos) == 0 {
			return
		}

		fmt.Println(heading)
		for _, info := range infos {
			path := info.Path
			if info.IsDir {
				path += "/"
			}

			fmt.Printf("  %s\n", path)
		}

		fmt.Println()
	}

	pairSection := func(heading, symbol string, infos []client.DiffPair) {
		if len(infos) == 0 {
			return
		}

		fmt.Println(heading)
		for _, pair := range infos {
			srcPath := pair.Src.Path
			if pair.Src.IsDir {
				srcPath += "/"
			}

			dstPath := pair.Dst.Path
			if pair.Dst.IsDir {
				dstPath += "/"
			}

			if pair.Src.Path != pair.Dst.Path {
				fmt.Printf("  %s %s %s\n", dstPath, symbol, srcPath)
			} else {
				fmt.Printf("  %s %s\n", symbol, srcPath)
			}
		}

		fmt.Println()
	}

	var addedAtRemote []client.DiffPair
	for _, src := range diff.Added {
		var pair client.DiffPair
		pair.Dst.Path = "▩"
		pair.Src = src
		addedAtRemote = append(addedAtRemote, pair)
	}
	pairSection(color.GreenString("Added:"), "←", addedAtRemote)

	simpleSection(color.YellowString("Ignored:"), diff.Ignored)

	var removedAtRemote []client.DiffPair
	for _, dst := range diff.Removed {
		var pair client.DiffPair
		pair.Dst = dst
		pair.Src.Path = "▩"
		removedAtRemote = append(removedAtRemote, pair)
	}
	pairSection(color.RedString("Removed:"), "←", removedAtRemote)

	// split diff.Merged to changedLocally and changedRemotely arrays
	var changedLocally, changedRemotely []client.DiffPair
	for _, pair := range diff.Merged {
		srcModTime := pair.Src.ModTime
		dstModTime := pair.Dst.ModTime
		if srcModTime.After(dstModTime) {
			changedRemotely = append(changedRemotely, pair)
		} else {
			changedLocally = append(changedLocally, pair)
		}
	}

	if printMissing {
		var missedAtRemote []client.DiffPair
		for _, dst := range diff.Missing {
			var pair client.DiffPair
			pair.Dst = dst
			pair.Src.Path = "▩"
			missedAtRemote = append(missedAtRemote, pair)
		}
		pairSection(color.RedString("Missing:"), "→", missedAtRemote)
	}

	pairSection(color.CyanString("Moved:"), "↔", diff.Moved)
	pairSection(color.WhiteString("Changed Locally:"), "→", changedLocally)
	pairSection(color.WhiteString("Changed Remotely:"), "←", changedRemotely)
	pairSection(color.MagentaString("Conflicts:"), "⚡", diff.Conflict)
}

func handleDiff(ctx *cli.Context, ctl *client.Client) error {
	if ctx.NArg() > 4 {
		fmt.Println("More than four arguments can't be handled.")
	}

	self, err := ctl.Whoami()
	if err != nil {
		return err
	}

	localName := self.CurrentUser
	remoteName := self.CurrentUser

	remoteRev := "CURR"
	localRev := "CURR"

	nArgs := ctx.NArg()
	if nArgs == 0 {
		// Special case: When typing brig diff we want to show
		// the diff from our CURR to HEAD only.
		localRev = "HEAD"
	}

	if ctx.Bool("self") {
		switch {
		case nArgs >= 2:
			localRev = ctx.Args().Get(1)
			fallthrough
		case nArgs >= 1:
			remoteRev = ctx.Args().Get(0)
		}
	} else {
		switch {
		case nArgs >= 4:
			localRev = ctx.Args().Get(3)
			fallthrough
		case nArgs >= 3:
			remoteRev = ctx.Args().Get(2)
			fallthrough
		case nArgs >= 2:
			localName = ctx.Args().Get(1)
			fallthrough
		case nArgs >= 1:
			remoteName = ctx.Args().Get(0)
		}
	}

	needFetch := !ctx.Bool("offline")
	diff, err := ctl.MakeDiff(localName, remoteName, localRev, remoteRev, needFetch)
	if err != nil {
		return ExitCode{UnknownError, fmt.Sprintf("diff: %v", err)}
	}

	printMissing := ctx.Bool("missing")
	if ctx.Bool("list") {
		printDiff(diff, printMissing)
	} else {
		printDiffTree(diff, printMissing)
	}

	return nil
}

func handleFetch(ctx *cli.Context, ctl *client.Client) error {
	who := ctx.Args().First()
	return ctl.Fetch(who)
}

func handleSync(ctx *cli.Context, ctl *client.Client) error {
	if len(ctx.Args()) > 0 {
		return handleSyncSingle(ctx, ctl, ctx.Args().First())
	}

	remotes, err := ctl.RemoteLs()
	if err != nil {
		return err
	}

	for _, rmt := range remotes {
		_, err := ctl.RemotePing(rmt.Name)
		if err != nil {
			fmt.Printf("Cannot reach %s..\n", rmt.Name)
			continue
		}

		fmt.Printf("Syncing with `%s`...\n", rmt.Name)
		if err := handleSyncSingle(ctx, ctl, rmt.Name); err != nil {
			return err
		}
	}

	return nil
}

func handleSyncSingle(ctx *cli.Context, ctl *client.Client, remoteName string) error {
	needFetch := true
	if ctx.Bool("no-fetch") {
		needFetch = false
	}

	if ctx.Bool("quiet") {
		return nil
	}

	diff, err := ctl.Sync(remoteName, needFetch)
	if err != nil {
		return err
	}

	if isEmptyDiff(diff) {
		fmt.Println("Nothing changed.")
		return nil
	}

	printDiff(diff, false)
	return nil
}

func handleStatus(ctx *cli.Context, ctl *client.Client) error {
	self, err := ctl.Whoami()
	if err != nil {
		return err
	}

	curr := self.CurrentUser
	diff, err := ctl.MakeDiff(curr, curr, "HEAD", "CURR", false)
	if err != nil {
		return err
	}

	if ctx.Bool("tree") {
		printDiffTree(diff, false)
	} else {
		printDiff(diff, false)
	}

	return nil
}

func handleBecome(ctx *cli.Context, ctl *client.Client) error {
	becomeSelf := ctx.Bool("self")
	if !becomeSelf && ctx.NArg() < 1 {
		return fmt.Errorf("become needs at least one argument without -s")
	}

	whoami, err := ctl.Whoami()
	if err != nil {
		return err
	}

	who := ctx.Args().First()
	if becomeSelf {
		who = whoami.Owner
	}

	if whoami.CurrentUser == who {
		fmt.Printf("You are already %s.\n", color.GreenString(who))
		return nil
	}

	if err := ctl.Become(who); err != nil {
		return err
	}

	suffix := "Everything is read only."
	if who == whoami.Owner {
		suffix = "Welcome back!"
	}

	fmt.Printf(
		"You are viewing %s's data now. %s\n",
		color.GreenString(who), suffix,
	)
	return nil
}

func handleCommit(ctx *cli.Context, ctl *client.Client) error {
	var msg string

	// Build the message:
	if ctx.IsSet("message") {
		msg = ctx.String("message")
	} else if len(ctx.Args()) >= 1 {
		msg = strings.Join(ctx.Args(), " ")
	} else {
		msg = "manual commit"
	}

	// Send the commit:
	if err := ctl.MakeCommit(msg); err != nil {
		return ExitCode{UnknownError, fmt.Sprintf("commit: %v", err)}
	}

	return nil
}

func handleTag(ctx *cli.Context, ctl *client.Client) error {
	if ctx.Bool("delete") {
		name := ctx.Args().Get(0)

		if err := ctl.Untag(name); err != nil {
			return ExitCode{
				UnknownError,
				fmt.Sprintf("untag: %v", err),
			}
		}
	} else {
		if len(ctx.Args()) < 2 {
			return ExitCode{BadArgs, "tag needs at least two arguments"}
		}

		rev := ctx.Args().Get(0)
		name := ctx.Args().Get(1)

		if err := ctl.Tag(rev, name); err != nil {
			return ExitCode{
				UnknownError,
				fmt.Sprintf("tag: %v", err),
			}
		}
	}

	return nil
}

func handleLog(ctx *cli.Context, ctl *client.Client) error {
	entries, err := ctl.Log()
	if err != nil {
		return ExitCode{UnknownError, fmt.Sprintf("commit: %v", err)}
	}

	tmpl, err := readFormatTemplate(ctx)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if tmpl != nil {
			if err := tmpl.Execute(os.Stdout, entry); err != nil {
				return err
			}

			continue
		}

		tags := ""
		isCurr := false
		if len(entry.Tags) > 0 {
			tags = fmt.Sprintf(" (%s)", strings.Join(entry.Tags, ", "))

			for _, tag := range entry.Tags {
				if tag == "curr" {
					isCurr = true
					break
				}
			}
		}

		msg := entry.Msg
		if msg == "" {
			msg = color.RedString("•")
		}

		entry.Hash.ShortB58()

		commitHash := entry.Hash.ShortB58()
		if isCurr {
			commitHash = "      -     "
		}

		fmt.Printf(
			"%s %s %s%s\n",
			color.GreenString(commitHash),
			color.YellowString(entry.Date.Format(time.UnixDate)),
			msg,
			color.CyanString(tags),
		)
	}

	return nil
}
