package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/dustin/go-humanize"
	tw "github.com/olekukonko/tablewriter"
	"github.com/sahib/brig/client"
	"github.com/sahib/brig/util/colors"
	"github.com/urfave/cli"
)

func handleStage(ctx *cli.Context, ctl *client.Client) error {
	localPath := ctx.Args().Get(0)

	repoPath := filepath.Base(localPath)
	if len(ctx.Args()) > 1 {
		repoPath = ctx.Args().Get(1)
	}

	if err := ctl.Stage(localPath, repoPath); err != nil {
		return ExitCode{
			UnknownError,
			fmt.Sprintf("stage: %v", err),
		}
	}
	return nil
}

func handleCat(ctx *cli.Context, ctl *client.Client) error {
	stream, err := ctl.Cat(ctx.Args().First())
	if err != nil {
		// TODO: Make those exit codes a wrapper function.
		return ExitCode{
			UnknownError,
			fmt.Sprintf("cat: %v", err),
		}
	}

	defer stream.Close()

	if _, err := io.Copy(os.Stdout, stream); err != nil {
		return ExitCode{
			UnknownError,
			fmt.Sprintf("cat: %v", err),
		}
	}

	return nil
}

func handleRm(ctx *cli.Context, ctl *client.Client) error {
	path := ctx.Args().First()

	if err := ctl.Remove(path); err != nil {
		return ExitCode{
			UnknownError,
			fmt.Sprintf("rm: %v", err),
		}
	}

	return nil
}

func handleMv(ctx *cli.Context, ctl *client.Client) error {
	srcPath := ctx.Args().Get(0)
	dstPath := ctx.Args().Get(1)
	return ctl.Move(srcPath, dstPath)
}

func handleList(ctx *cli.Context, ctl *client.Client) error {
	maxDepth := ctx.Int("depth")
	if ctx.Bool("recursive") {
		maxDepth = -1
	}

	root := "/"
	if ctx.Args().Present() {
		root = ctx.Args().First()
	}

	entries, err := ctl.List(root, maxDepth)
	if err != nil {
		return err
	}

	tabW := tabwriter.NewWriter(
		os.Stdout, 0, 0, 2, ' ',
		tabwriter.StripEscape,
	)

	// TODO: golangs' tabwriter falls short when using colors in the middle.
	//       Better use https://github.com/olekukonko/tablewriter on the
	//       next occassion.
	if len(entries) != 0 {
		fmt.Fprintln(tabW, "SIZE\tMODTIME\tPATH\tPIN\t")
	}

	for _, entry := range entries {
		pinState := ""
		if entry.IsPinned {
			pinState += " " + colors.Colorize("🖈", colors.Cyan)
		}

		fmt.Fprintf(
			tabW,
			"%s\t%s\t%s\t%s\t\n",
			humanize.Bytes(entry.Size),
			entry.ModTime.Format(time.Stamp),
			entry.Path,
			pinState,
		)
	}

	return tabW.Flush()
}

func handleTree(ctx *cli.Context, ctl *client.Client) error {
	entries, err := ctl.List("/", -1)
	if err != nil {
		return ExitCode{
			UnknownError,
			fmt.Sprintf("ls: %v", err),
		}
	}

	return showTree(entries, -1)
}

func handleMkdir(ctx *cli.Context, ctl *client.Client) error {
	path := ctx.Args().First()
	createParents := ctx.Bool("parents")

	if err := ctl.Mkdir(path, createParents); err != nil {
		return ExitCode{UnknownError, fmt.Sprintf("mkdir: %v", err)}
	}

	return nil
}

func handleInfo(ctx *cli.Context, ctl *client.Client) error {
	path := ctx.Args().First()
	info, err := ctl.Stat(path)
	if err != nil {
		return err
	}

	pinState := colors.Colorize("yes", colors.Green)
	if !info.IsPinned {
		pinState += " " + colors.Colorize("no", colors.Red)
	}

	nodeType := "file"
	if info.IsDir {
		nodeType = "directory"
	}

	w := tw.NewWriter(os.Stdout)
	w.SetBorder(false)
	w.SetColumnSeparator("")
	w.SetColumnAlignment([]int{
		tw.ALIGN_RIGHT,
		tw.ALIGN_LEFT,
	})

	// TODO: This still shows an empty header line.
	w.SetHeader([]string{"", ""})
	w.SetHeaderLine(false)

	w.Append([]string{"Path", info.Path})
	w.Append([]string{"Type", nodeType})
	w.Append([]string{"Size", humanize.Bytes(info.Size)})
	w.Append([]string{"Hash", info.Hash.B58String()})
	w.Append([]string{"Inode", strconv.FormatUint(info.Inode, 10)})
	w.Append([]string{"Pinned", pinState})
	w.Append([]string{"ModTime", info.ModTime.Format(time.RFC3339)})
	w.Append([]string{"Content", info.Content.B58String()})

	w.SetColumnColor(
		tw.Colors{tw.FgWhiteColor, tw.Bold},
		tw.Colors{},
	)

	w.Render()
	return nil
}
