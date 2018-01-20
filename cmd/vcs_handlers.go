package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sahib/brig/cmd/tabwriter"

	"github.com/sahib/brig/client"
	"github.com/sahib/brig/util/colors"
	"github.com/urfave/cli"
)

func handleReset(ctx *cli.Context, ctl *client.Client) error {
	force := ctx.Bool("force")
	path := ctx.Args().First()
	rev := "HEAD"

	if len(ctx.Args()) > 1 {
		rev = ctx.Args().Get(1)
	}

	if err := ctl.Reset(path, rev, force); err != nil {
		return ExitCode{UnknownError, fmt.Sprintf("unpin: %v", err)}
	}

	return nil
}

func commitName(cmt *client.Commit) string {
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
		fmt.Printf("%s %s", colors.Colorize("WARNING:", colors.Yellow), `This file is not part of this commit, but there's still history for it.
         Most likely this file was moved or removed in the past.

`)
	}

	tabW := tabwriter.NewWriter(
		os.Stdout, 0, 0, 2, ' ',
		tabwriter.StripEscape,
	)

	if len(history) != 0 {
		fmt.Fprintf(tabW, "CHANGE\tFROM\tTO\tHOW\tWHEN\t\n")
	}

	for idx, entry := range history {
		what := ""
		printLine := true

		for _, detail := range entry.Mask {
			// If it was moved, let's display what moved.
			if detail == "moved" && idx+1 < len(history) {
				src := history[idx+1].Path
				dst := entry.Path

				if entry.ReferTo != "" {
					dst = entry.ReferTo
				}

				what = fmt.Sprintf(
					"%s → %s",
					colors.Colorize(src, colors.Red),
					colors.Colorize(dst, colors.Red),
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

		changeDesc := colors.Colorize(
			strings.Join(entry.Mask, ", "),
			colors.Yellow,
		)

		when := colors.Colorize(
			entry.Head.Date.Format(time.Stamp),
			colors.Magenta,
		)

		fmt.Fprintf(
			tabW,
			"%s\t%s\t%s\t%s\t%s\t\n",
			changeDesc,
			colors.Colorize(commitName(entry.Next), colors.Cyan),
			colors.Colorize(commitName(entry.Head), colors.Green),
			what,
			when,
		)
	}

	return tabW.Flush()
}

func printDiffTree(diff *client.Diff) {
	const (
		diffTypeNone = iota
		diffTypeAdded
		diffTypeRemoved
		diffTypeIgnored
		diffTypeConflict
		diffTypeMerged
	)

	type diffEntry struct {
		typ  int
		pair client.DiffPair
	}

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
	for _, info := range diff.Ignored {
		types[info.Path] = diffEntry{typ: diffTypeIgnored}
		entries = append(entries, info)
	}

	// Pair types:
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

	// Called to format each name in the resulting tree:
	formatter := func(n *treeNode) string {
		if n.name == "/" {
			return colors.Colorize("•", colors.Magenta)
		}

		if diffEntry, ok := types[n.entry.Path]; ok {
			switch diffEntry.typ {
			case diffTypeAdded:
				return colors.Colorize(" + "+n.name, colors.Green)
			case diffTypeRemoved:
				return colors.Colorize(" - "+n.name, colors.Red)
			case diffTypeIgnored:
				return colors.Colorize(" * "+n.name, colors.Yellow)
			case diffTypeMerged:
				name := fmt.Sprintf(
					" %s ⇄ %s",
					diffEntry.pair.Src.Path,
					diffEntry.pair.Dst.Path,
				)
				return colors.Colorize(name, colors.Cyan)
			case diffTypeConflict:
				name := fmt.Sprintf(
					" %s ⚡ %s",
					diffEntry.pair.Src.Path,
					diffEntry.pair.Dst.Path,
				)
				return colors.Colorize(name, colors.Magenta)
			}
		}

		return n.name
	}

	// Render the tree:
	showTree(entries, &treeCfg{
		format:  formatter,
		showPin: false,
	})
}

func printDiff(diff *client.Diff) {
	simpleSection := func(heading string, infos []client.StatInfo) {
		if len(infos) == 0 {
			return
		}

		fmt.Println(heading)
		for _, info := range infos {
			fmt.Printf("  %s\n", info.Path)
		}

		fmt.Println()
	}

	pairSection := func(heading string, infos []client.DiffPair) {
		if len(infos) == 0 {
			return
		}

		for _, pair := range diff.Merged {
			fmt.Printf("  %s <-> %s\n", pair.Src.Path, pair.Dst.Path)
		}

		fmt.Println()
	}

	simpleSection(colors.Colorize("Added:", colors.Green), diff.Added)
	simpleSection(colors.Colorize("Ignored:", colors.Yellow), diff.Ignored)
	simpleSection(colors.Colorize("Removed:", colors.Red), diff.Removed)

	pairSection(
		colors.Colorize("Resolveable Conflicts:", colors.Cyan),
		diff.Merged,
	)
	pairSection(
		colors.Colorize("Conflicts:", colors.Magenta),
		diff.Conflict,
	)
}

func handleDiff(ctx *cli.Context, ctl *client.Client) error {
	if ctx.NArg() > 4 {
		fmt.Println("More than four arguments can't be handled.")
	}

	self, err := ctl.Whoami()
	if err != nil {
		return err
	}

	remoteName := self.CurrentUser
	localName := self.CurrentUser

	remoteRev := "HEAD"
	localRev := "HEAD"

	switch n := ctx.NArg(); {
	case n >= 1:
		remoteName = ctx.Args().Get(0)
		fallthrough
	case n >= 2:
		localName = ctx.Args().Get(1)
		fallthrough
	case n >= 3:
		remoteRev = ctx.Args().Get(2)
		fallthrough
	case n >= 4:
		localRev = ctx.Args().Get(3)
	}

	diff, err := ctl.MakeDiff(localName, remoteName, localRev, remoteRev)
	if err != nil {
		return ExitCode{UnknownError, fmt.Sprintf("diff: %v", err)}
	}

	if ctx.Bool("list") {
		printDiff(diff)
	} else {
		printDiffTree(diff)
	}

	return nil
}

func handleFetch(ctx *cli.Context, ctl *client.Client) error {
	who := ctx.Args().First()
	return ctl.Fetch(who)
}

func handleSync(ctx *cli.Context, ctl *client.Client) error {
	who := ctx.Args().First()

	needFetch := true
	if ctx.Bool("no-fetch") {
		needFetch = false
	}

	return ctl.Sync(who, needFetch)
}

func handleStatus(ctx *cli.Context, ctl *client.Client) error {
	self, err := ctl.Whoami()
	if err != nil {
		return err
	}

	curr := self.CurrentUser
	diff, err := ctl.MakeDiff(curr, curr, "CURR", "HEAD")
	if err != nil {
		return err
	}

	if ctx.Bool("tree") {
		printDiffTree(diff)
	} else {
		printDiff(diff)
	}

	return nil
}

func handleBecome(ctx *cli.Context, ctl *client.Client) error {
	who := ctx.Args().First()
	if err := ctl.Become(who); err != nil {
		return err
	}

	fmt.Printf(
		"You are viewing %s's data now. Changes will be local only.\n",
		colors.Colorize(who, colors.Green),
	)
	return nil
}

func handleCommit(ctx *cli.Context, ctl *client.Client) error {
	msg := ""
	if msg = ctx.String("message"); msg == "" {
		msg = fmt.Sprintf("manual commit")
	}

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

	for _, entry := range entries {
		tags := ""
		if len(entry.Tags) > 0 {
			tags = fmt.Sprintf(" (%s)", strings.Join(entry.Tags, ", "))
		}

		msg := entry.Msg
		if msg == "" {
			msg = colors.Colorize("•", colors.Red)
		}

		entry.Hash.ShortB58()

		fmt.Printf(
			"%s %s %s%s\n",
			colors.Colorize(entry.Hash.ShortB58(), colors.Green),
			colors.Colorize(entry.Date.Format(time.Stamp), colors.Yellow),
			msg,
			colors.Colorize(tags, colors.Cyan),
		)
	}

	return nil
}
