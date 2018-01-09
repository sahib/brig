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

func handleHistory(ctx *cli.Context, ctl *client.Client) error {
	path := ctx.Args().First()

	history, err := ctl.History(path)
	if err != nil {
		return ExitCode{UnknownError, fmt.Sprintf("history: %v", err)}
	}

	tabW := tabwriter.NewWriter(
		os.Stdout, 0, 0, 2, ' ',
		tabwriter.StripEscape,
	)

	if len(history) != 0 {
		fmt.Fprintf(tabW, "CHANGE\tPATH\tREV\t\n")
	}

	for _, entry := range history {
		fmt.Fprintf(
			tabW,
			"%s\t%s\t%s\t\n",
			colors.Colorize(entry.Change, colors.Yellow),
			colors.Colorize(entry.Path, colors.Green),
			colors.Colorize(entry.Commit.Hash.B58String()[:10], colors.Red),
		)
	}

	return tabW.Flush()
}

func printDiff(diff *client.Diff) {
	simpleSection := func(heading string, infos []client.StatInfo) {
		if len(infos) == 0 {
			return
		}

		fmt.Println(heading)
		for _, info := range diff.Added {
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
	remoteRev := ctx.Args().Get(0)
	if remoteRev == "" {
		remoteRev = "HEAD"
	}

	localRev := ctx.String("rev")
	remoteName := ctx.String("remote")
	if remoteName == "" {
		self, err := ctl.Whoami()
		if err != nil {
			return err
		}

		remoteName = self.Owner
	}

	diff, err := ctl.MakeDiff(remoteName, localRev, remoteRev)
	if err != nil {
		return ExitCode{UnknownError, fmt.Sprintf("diff: %v", err)}
	}

	printDiff(diff)
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

	diff, err := ctl.MakeDiff(self.Owner, "HEAD", "CURR")
	if err != nil {
		return err
	}

	printDiff(diff)
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
