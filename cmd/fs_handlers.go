package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"

	"github.com/sahib/brig/cmd/tabwriter"
	"github.com/sahib/brig/util"

	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/sahib/brig/client"
	"github.com/urfave/cli"
)

func handleStage(ctx *cli.Context, ctl *client.Client) error {
	localPath, err := filepath.Abs(ctx.Args().Get(0))
	if err != nil {
		return fmt.Errorf("Failed to retrieve absolute path: %v", err)
	}

	repoPath := filepath.Base(localPath)
	if len(ctx.Args()) > 1 {
		repoPath = ctx.Args().Get(1)
	}

	info, err := os.Stat(localPath)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return handleStageDirectory(ctx, ctl, localPath, repoPath)
	}

	return ctl.Stage(localPath, repoPath)
}

type stagePair struct {
	local, repo string
}

func handleStageDirectory(ctx *cli.Context, ctl *client.Client, root, repoRoot string) error {
	// First create all directories:
	// (tbh: I'm not exactly sure what "lexical" order means in the docs of Walk,
	//  i.e. breadth-first or depth first, so better be safe)
	toBeStaged := []stagePair{}

	root = filepath.Clean(root)
	repoRoot = filepath.Clean(repoRoot)

	err := filepath.Walk(root, func(childPath string, info os.FileInfo, err error) error {
		repoPath := filepath.Join(repoRoot, childPath[len(root):])

		if info.IsDir() {
			if err := ctl.Mkdir(repoPath, true); err != nil {
				return err
			}
		} else {
			toBeStaged = append(toBeStaged, stagePair{childPath, repoPath})
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("Failed to create sub directories: %v", err)
	}

	for _, child := range toBeStaged {
		if err := ctl.Stage(child.local, child.repo); err != nil {
			return err
		}
	}

	return nil
}

func handleCat(ctx *cli.Context, ctl *client.Client) error {
	stream, err := ctl.Cat(ctx.Args().First())
	if err != nil {
		return ExitCode{
			UnknownError,
			fmt.Sprintf("cat: %v", err),
		}
	}

	defer util.Closer(stream)

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

func handleCp(ctx *cli.Context, ctl *client.Client) error {
	srcPath := ctx.Args().Get(0)
	dstPath := ctx.Args().Get(1)
	return ctl.Copy(srcPath, dstPath)
}

func colorForSize(size uint64) func(f string, a ...interface{}) string {
	switch {
	case size >= 1024 && size < 1024<<10:
		return color.CyanString
	case size >= 1024<<10 && size < 1024<<20:
		return color.YellowString
	case size >= 1024<<20 && size < 1024<<30:
		return color.RedString
	case size >= 1024<<30:
		return color.MagentaString
	default:
		return func(f string, a ...interface{}) string {
			return f
		}
	}
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

	if len(entries) != 0 {
		fmt.Fprintln(tabW, "SIZE\tMODTIME\tPATH\tPIN\t")
	}

	for _, entry := range entries {
		pinState := ""
		if entry.IsPinned {
			pinState += " " + color.CyanString("🖈")
		}

		coloredPath := ""
		if entry.IsDir {
			coloredPath = color.GreenString(entry.Path)
		} else {
			coloredPath = color.WhiteString(entry.Path)
		}

		fmt.Fprintf(
			tabW,
			"%s\t%s\t%s\t%s\t\n",
			colorForSize(entry.Size)(humanize.Bytes(entry.Size)),
			entry.ModTime.Format(time.Stamp),
			coloredPath,
			pinState,
		)
	}

	return tabW.Flush()
}

func handleTree(ctx *cli.Context, ctl *client.Client) error {
	root := "/"
	if ctx.NArg() > 0 {
		root = ctx.Args().First()
	}

	entries, err := ctl.List(root, -1)
	if err != nil {
		return ExitCode{
			UnknownError,
			fmt.Sprintf("tree: %v", err),
		}
	}

	showTree(entries, &treeCfg{
		showPin: true,
	})
	return nil
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

	pinState := color.GreenString("yes")
	if !info.IsPinned {
		pinState = color.RedString("no")
	}

	nodeType := "file"
	if info.IsDir {
		nodeType = "directory"
	}

	tabW := tabwriter.NewWriter(
		os.Stdout, 0, 0, 2, ' ',
		tabwriter.StripEscape,
	)

	fmt.Fprintln(tabW, "ATTR\tVALUE\t")

	printPair := func(name string, val interface{}) {
		fmt.Fprintf(
			tabW,
			"%s\t%v\t\n",
			color.WhiteString(name),
			val,
		)
	}

	printPair("Path", info.Path)
	printPair("Type", nodeType)
	printPair("Size", humanize.Bytes(info.Size))
	printPair("Hash", info.Hash.B58String())
	printPair("Inode", strconv.FormatUint(info.Inode, 10))
	printPair("Pinned", pinState)
	printPair("ModTime", info.ModTime.Format(time.RFC3339))
	printPair("Content", info.Content.B58String())

	return tabW.Flush()
}

func handleEdit(ctx *cli.Context, ctl *client.Client) error {
	repoPath := ctx.Args().First()

	exists, err := ctl.Exists(repoPath)
	if err != nil {
		return err
	}

	data := []byte{}
	if exists {
		r, err := ctl.Cat(repoPath)
		if err != nil {
			return err
		}

		defer util.Closer(r)

		data, err = ioutil.ReadAll(r)
		if err != nil {
			return err
		}
	}

	tempPath, err := editToPath(data, path.Ext(repoPath))
	if err != nil {
		return err
	}

	defer func() {
		if err := os.Remove(tempPath); err != nil {
			fmt.Printf("Failed to remove temp file: %v\n", err)
		}
	}()

	return ctl.Stage(tempPath, repoPath)
}

func handleTouch(ctx *cli.Context, ctl *client.Client) error {
	repoPath := ctx.Args().First()
	return ctl.Touch(repoPath)
}
