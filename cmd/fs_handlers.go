package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sahib/brig/cmd/tabwriter"
	"github.com/sahib/brig/util"

	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/sahib/brig/client"
	"github.com/urfave/cli"

	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
	terminal "github.com/wayneashleyberry/terminal-dimensions"
)

func handleStage(ctx *cli.Context, ctl *client.Client) error {
	localPath := ctx.Args().Get(0)
	readFromStdin := ctx.Bool("stdin")
	repoPath := filepath.Base(localPath)

	if len(ctx.Args()) > 1 {
		repoPath = ctx.Args().Get(1)
		if localPath == "-" {
			readFromStdin = true
		}
	}

	if readFromStdin {
		repoPath = ctx.Args().Get(0)
		return ctl.StageFromReader(repoPath, os.Stdin)
	}

	absLocalPath, err := filepath.Abs(localPath)
	if err != nil {
		return fmt.Errorf("Failed to retrieve absolute path: %v", err)
	}

	_, err = os.Stat(absLocalPath)
	if err != nil {
		return err
	}

	return handleStageDirectory(ctx, ctl, absLocalPath, repoPath)
}

// holds brig repoPaths twins (by content) to a OS file system localPath
type twins struct {
	localPath string
	repoPaths []string
}

type walkOptions struct {
	dereference     bool
	continueOnError bool
}

func walk(root, repoRoot string, depth int, opt walkOptions) (map[string]twins, error) {
	// toBeStaged map: key is local path, value is array of repoPaths using the local path
	toBeStaged := make(map[string]twins)
	depth++
	if depth > 255 {
		return toBeStaged, fmt.Errorf("Exceeded allowed dereferencing depth for %v", root)
	}
	root = filepath.Clean(root)
	repoRoot = filepath.Clean(repoRoot)
	err := filepath.Walk(root, func(childPath string, info os.FileInfo, err error) error {
		repoPath := filepath.Join("/", repoRoot, childPath[len(root):])

		if opt.dereference && info.Mode()&os.ModeSymlink != 0 {
			// NOTE: `brig` does not have concept of symlink
			//       The lack of native symlinks in `brig` has the following potential issues
			//       * Ignoring cycles limits valid use cases.
			//       * Not ignoring cycles opens room for malicious input.
			//
			//       Here we implement a dereference of symlinks as a temporary measure
			//       which works in the most likely user scenarios (we assume that user
			//       is not malicious to herself and does not create infinite symlinked loops).
			//       If the level of recursion or cycles
			//       (where link points to itself directly or indirectly) is exceeded,
			//       we just fail on such link.
			//       Currently, we have a depth limit of 255 (see couple line above).
			resolvedPath, err := filepath.EvalSymlinks(childPath)
			if err != nil {
				msg := fmt.Sprintf("Failed to resolve: %v: %v", childPath, err)
				if opt.continueOnError {
					fmt.Fprintf(os.Stderr, "WARNING: %s\n", msg)
					return nil
				}
				return fmt.Errorf(msg)
			}
			info, err = os.Stat(resolvedPath)
			if err != nil {
				msg := fmt.Sprintf("Failed to do os.Stat(%v): %v", resolvedPath, err)
				if opt.continueOnError {
					fmt.Fprintf(os.Stderr, "WARNING: %s\n", msg)
					return nil
				}
				return fmt.Errorf(msg)
			}
			childPath = resolvedPath
			if info.Mode().IsDir() {
				extra, err := walk(childPath, repoPath, depth, opt)
				if err != nil {
					if opt.continueOnError {
						fmt.Fprintf(os.Stderr, "WARNING: %s\n", err.Error())
						return nil
					}
					return err
				}
				for k, v := range extra {
					t, ok := toBeStaged[k]
					if !ok {
						t = twins{
							v.localPath,
							[]string{},
						}
					}
					t.repoPaths = append(t.repoPaths, v.repoPaths...)
					toBeStaged[k] = t
				}
				return nil
			}
		}

		if info.Mode().IsRegular() {
			k, _ := inodeString(childPath)
			t, ok := toBeStaged[k]
			if !ok {
				t = twins{
					childPath,
					[]string{},
				}
			}
			t.repoPaths = append(t.repoPaths, repoPath)
			toBeStaged[k] = t
		}

		return nil
	})
	return toBeStaged, err
}

func makeParentDirIfNeeded(ctx *cli.Context, ctl *client.Client, path string) error {
	parent := filepath.Dir(path)
	info, err := ctl.Stat(parent)
	if err != nil {
		if yes, _ := regexp.MatchString("No such file or directory:", err.Error()); yes {
			createParents := true
			err = ctl.Mkdir(parent, createParents)
		}
		return err
	}
	if info.IsDir {
		return nil
	}
	return fmt.Errorf("Cannot make dir from existing non dir node %s", parent)
}

func handleStageDirectory(ctx *cli.Context, ctl *client.Client, root, repoRoot string) error {
	// Links will be reflinked in the `man cp` sense,
	// i.e. resolved repoPaths will point to the same content and backend hash

	root = filepath.Clean(root)
	repoRoot = filepath.Clean(repoRoot)

	opt := walkOptions{
		dereference:     !ctx.Bool("no-dereference"),
		continueOnError: ctx.Bool("continue-on-error"),
	}

	toBeStaged, err := walk(root, repoRoot, 0, opt)
	if err != nil {
		return fmt.Errorf("failed to walk dir: %v: %v", root, err)
	}

	if len(toBeStaged) == 0 {
		// This might happen if ask to stage a symlink pointing to a dir
		// but Walk does not travel symlinks and we end up with empty list.
		return nil
	}

	width, err := terminal.Width()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to get terminal size: %s\n", err)
		width = 80
	}

	pbars := mpb.New(
		// override default (80) width
		mpb.WithWidth(int(width)),
		// override default 120ms refresh rate
		mpb.WithRefreshRate(250*time.Millisecond),
	)

	name := "ETA"
	bar := pbars.AddBar(
		int64(len(toBeStaged)),
		mpb.PrependDecorators(
			// display our name with one space on the right
			decor.Name(name, decor.WC{W: len(name) + 1, C: decor.DidentRight}),
			// replace ETA decorator with "done" message, OnComplete event
			decor.OnComplete(
				// ETA decorator with ewma age of 60, and width reservation of 4
				decor.EwmaETA(decor.ET_STYLE_GO, 60, decor.WC{W: 4}), "done",
			),
		),
		mpb.AppendDecorators(decor.Percentage()),
	)

	type stageList struct {
		local    string
		repoList []string
	}

	nWorkers := 20
	start := time.Now()
	jobs := make(chan twins, nWorkers)

	// Start a bunch of workers that will do the actual adding:
	for idx := 0; idx < nWorkers; idx++ {
		go func() {
			for {
				twinsSet, ok := <-jobs
				if !ok {
					return
				}

				firstToStage := ""
				for i, repoPath := range twinsSet.repoPaths {
					if i == 0 {
						firstToStage = repoPath
						// First occurrence is staged.
						// Stage creates all needed parent directories.
						if err := ctl.Stage(twinsSet.localPath, repoPath); err != nil {
							fmt.Fprintf(os.Stderr, "failed to stage '%s' as '%s': %v\n", twinsSet.localPath, repoPath, err)
							break
						}
						continue
					}
					// Copy does not create parent directories. We take care of it.
					if err := makeParentDirIfNeeded(ctx, ctl, repoPath); err != nil {
						fmt.Fprintf(os.Stderr, "failed to make the parent dir for '%s': %v\n", repoPath, err)
						break
					}
					if err := ctl.Copy(firstToStage, repoPath); err != nil {
						fmt.Fprintf(os.Stderr, "failed copy of '%s' to '%s': %v\n", firstToStage, repoPath, err)
						break
					}
				}

				// Notify the bar. The op time is used for the ETA.
				// The time is measured by "start" is NOT the time used to
				// stage a single file.  This would only work in a non-parallel
				// environment, because the ETA would assume that one file took
				// 2s, so 1000 files must take 2000s.  Instead it measures the
				// time between two time recordings, which are in the ideal
				// case around 1/n_workers * time_to_stage but it measures the
				// actual amount of parallelism that we achieve.
				bar.IncrBy(1, time.Since(start))
				start = time.Now()
			}
		}()
	}

	// Send the jobs onward:
	for _, v := range toBeStaged {
		jobs <- v
	}

	close(jobs)
	pbars.Wait()
	return nil
}

func handleCat(ctx *cli.Context, ctl *client.Client) error {
	path := "/"
	if len(ctx.Args()) >= 1 {
		path = ctx.Args().First()
	}

	info, err := ctl.Stat(path)
	if err != nil {
		return err
	}

	doOffline := ctx.Bool("offline")

	var stream io.ReadCloser
	if info.IsDir {
		stream, err = ctl.Tar(path, doOffline)
	} else if ctx.Bool("stream") {
		return ctl.CatOnClient(path, doOffline, os.Stdout)
	} else {
		stream, err = ctl.Cat(path, doOffline)
	}

	if err != nil {
		return err
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

func userPrefixMap(users []string) map[string]string {
	m := make(map[string]string)
	for _, user := range users {
		m[user] = user
	}

	tryAbbrev := func(abbrev string) bool {
		for _, short := range m {
			if short == abbrev {
				return false
			}
		}

		return true
	}

	for name := range m {
		atIdx := strings.Index(name, "@")
		if atIdx != -1 && tryAbbrev(name[:atIdx]) {
			m[name] = name[:atIdx]
			continue
		}

		slashIdx := strings.Index(name, "/")
		if slashIdx != -1 && tryAbbrev(name[:slashIdx]) {
			m[name] = name[:slashIdx]
			continue
		}
	}

	return m
}

func formatHint(hint client.Hint) string {
	return fmt.Sprintf("enc:%s-zip:%s", hint.EncryptionAlgo, hint.CompressionAlgo)
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

	tmpl, err := readFormatTemplate(ctx)
	if err != nil {
		return err
	}

	if tmpl != nil {
		for _, entry := range entries {
			if err := tmpl.Execute(os.Stdout, entry); err != nil {
				return err
			}
		}

		return nil
	}

	users := []string{}
	for _, entry := range entries {
		users = append(users, entry.User)
	}

	userMap := userPrefixMap(users)

	if len(entries) != 0 {
		userColumn := ""
		if len(userMap) > 1 {
			userColumn = "USER\t"
		}

		fmt.Fprintf(tabW, "SIZE\tBKEND\tMODTIME\t%sPATH\tPIN\tCACHED\tHINT\n", userColumn)
	}

	for _, entry := range entries {
		pinState := " " + pinStateToSymbol(entry.IsPinned, entry.IsExplicit)

		var coloredPath string
		if entry.IsDir {
			coloredPath = color.GreenString(entry.Path)
		} else {
			coloredPath = color.WhiteString(entry.Path)
		}

		userEntry := ""
		if len(userMap) > 1 {
			userEntry = color.GreenString(userMap[entry.User]) + "\t"
		}

		isCached, err := ctl.IsCached(entry.Path)
		if err != nil {
			return err
		}
		cachedState := "  " + pinStateToSymbol(isCached, false)

		fmt.Fprintf(
			tabW,
			"%s\t%s\t%s\t%s%s\t%s\t%s\t%s\n",
			colorForSize(entry.Size)(humanize.Bytes(entry.Size)),
			colorForSize(entry.Size)(humanize.Bytes(uint64(entry.CachedSize))),
			entry.ModTime.Format("2006-01-02 15:04:05 MST"),
			userEntry,
			coloredPath,
			pinState,
			cachedState,
			formatHint(entry.Hint),
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

func handleShow(ctx *cli.Context, ctl *client.Client) error {
	path := ctx.Args().First()
	isValidRef, cmt, err := ctl.CommitInfo(path)
	if err != nil {
		return err
	}

	if isValidRef {
		return handleShowCommit(ctx, ctl, cmt)
	}

	return handleShowFileOrDir(ctx, ctl, path)
}

func handleShowCommit(ctx *cli.Context, ctl *client.Client, cmt *client.Commit) error {
	tabW := tabwriter.NewWriter(
		os.Stdout, 0, 0, 2, ' ',
		tabwriter.StripEscape,
	)

	printPair := func(name string, val interface{}) {
		fmt.Fprintf(
			tabW,
			"%s\t%v\t\n",
			color.WhiteString(name),
			val,
		)
	}

	printPair("Path", cmt.Hash)
	printPair("Tags", strings.Join(cmt.Tags, ", "))
	printPair("ModTime", cmt.Date.Format(time.RFC3339))
	printPair("Message", cmt.Msg)
	tabW.Flush()

	self, err := ctl.Whoami()
	if err != nil {
		return err
	}

	diff, err := ctl.MakeDiff(
		self.CurrentUser,
		self.CurrentUser,
		cmt.Hash.B58String()+"^",
		cmt.Hash.B58String(),
		false,
	)

	if err != nil {
		return err
	}

	if !diff.IsEmpty() {
		fmt.Println()
		fmt.Println("Here's what changed in this commit:")
		fmt.Println()
		printDiffTree(diff, false)
	}

	return nil
}

func handleShowFileOrDir(ctx *cli.Context, ctl *client.Client, path string) error {
	info, err := ctl.Stat(path)
	if err != nil {
		return err
	}

	tmpl, err := readFormatTemplate(ctx)
	if err != nil {
		return err
	}

	if tmpl != nil {
		return tmpl.Execute(os.Stdout, info)
	}

	isCached, err := ctl.IsCached(path)
	if err != nil {
		return err
	}

	pinState := yesify(info.IsPinned)
	explicitState := yesify(info.IsExplicit)
	cachedState := yesify(isCached)

	nodeType := "file"
	if info.IsDir {
		nodeType = "directory"
	}

	tabW := tabwriter.NewWriter(
		os.Stdout, 0, 0, 2, ' ',
		tabwriter.StripEscape,
	)

	printPair := func(name string, val interface{}) {
		fmt.Fprintf(
			tabW,
			"%s\t%v\t\n",
			color.WhiteString(name),
			val,
		)
	}

	printPair("Path", info.Path)
	printPair("User", info.User)
	printPair("Type", nodeType)
	printPair("Size", fmt.Sprintf("%s (%d bytes)", humanize.Bytes(info.Size), info.Size))
	printPair("Backend Size", fmt.Sprintf("%s (%d bytes)", humanize.Bytes(uint64(info.CachedSize)), info.CachedSize))
	printPair("Inode", strconv.FormatUint(info.Inode, 10))
	printPair("Pinned", pinState)
	printPair("Explicit", explicitState)
	printPair("Cached", cachedState)
	printPair("IsRaw", yesify(info.IsRaw))
	printPair("ModTime", info.ModTime.Format(time.RFC3339))
	printPair("Tree Hash", info.TreeHash.B58String())
	printPair("Content Hash", info.ContentHash.B58String())
	printPair("Hint", formatHint(info.Hint))

	if !info.IsDir {
		printPair("Backend Hash", info.BackendHash.B58String())
	} else {
		printPair("Backend Hash", "-")
	}

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
		r, err := ctl.Cat(repoPath, false)
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

func handleTrashList(ctx *cli.Context, ctl *client.Client) error {
	root := "/"
	if firstArg := ctx.Args().First(); firstArg != "" {
		root = firstArg
	}

	nodes, err := ctl.DeletedNodes(root)
	if err != nil {
		return err
	}

	for _, node := range nodes {
		fmt.Println(node.Path)
	}

	return nil
}

func handleTrashRemove(ctx *cli.Context, ctl *client.Client) error {
	return ctl.Undelete(ctx.Args().First())
}
