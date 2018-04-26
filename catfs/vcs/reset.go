package vcs

import (
	"errors"
	"fmt"
	"path"

	e "github.com/pkg/errors"
	c "github.com/sahib/brig/catfs/core"
	ie "github.com/sahib/brig/catfs/errors"
	n "github.com/sahib/brig/catfs/nodes"
)

func findPathAt(lkr *c.Linker, cmt *n.Commit, path string) (string, error) {
	nd, err := lkr.LookupModNode(path)
	if err != nil && !ie.IsNoSuchFileError(err) {
		return "", err
	}

	if ie.IsNoSuchFileError(err) {
		// The file does not exist in the current commit,
		// so user probably knows that it had this path before.
		return path, nil
	}

	status, err := lkr.Status()
	if err != nil {
		return "", err
	}

	walker := NewHistoryWalker(lkr, status, nd)
	for walker.Next() {
		state := walker.State()
		if state.Head.TreeHash().Equal(cmt.TreeHash()) {
			return state.Curr.Path(), nil
		}
	}

	if err := walker.Err(); err != nil {
		return "", err
	}

	// Take the current path as best guess.
	return path, nil
}

func clearPath(lkr *c.Linker, ndPath string) (*n.Directory, error) {
	nd, err := lkr.LookupModNode(ndPath)
	isNoSuchFile := ie.IsNoSuchFileError(err)

	if err != nil && !isNoSuchFile {
		return nil, err
	}

	var par *n.Directory
	if ndPath != "/" {
		par, err = lkr.LookupDirectory(path.Dir(ndPath))
		if err != nil {
			return nil, err
		}
	}

	if par == nil {
		return nil, fmt.Errorf(
			"checkout by commit if you want to checkout previous roots",
		)
	}

	// The node does currently not exist (and the user wants to bring it back)
	if isNoSuchFile {
		return par, nil
	}

	err = n.Walk(lkr, nd, true, func(child n.Node) error {
		lkr.MemIndexPurge(child)
		return nil
	})
	if err != nil {
		return nil, err
	}

	if err := par.RemoveChild(lkr, nd); err != nil {
		return nil, err
	}

	lkr.MemIndexPurge(nd)
	return par, lkr.StageNode(par)
}

// ResetFile resets a certain file to the state it had in cmt. If the file
// did not exist back then, it will be deleted. `nd` is usually retrieved by
// calling ResolveNode() and sorts.
//
// A special case occurs when the file was moved we reset to.
// In this case the state of the old node (at the old path)
// is being written to the node at the new path.
// This is the more obvious choice to the user when he types:
//
//    $ brig reset HEAD^ i-was-somewhere-else-before   # name does not change.
//
func ResetFile(lkr *c.Linker, cmt *n.Commit, currPath string) error {
	root, err := lkr.DirectoryByHash(cmt.Root())
	if err != nil {
		return err
	}

	if root == nil {
		return errors.New("no root to reset to")
	}

	// Find out the old path of `currPath` at `cmt`.
	// It might have changed due to moves.
	oldPath, err := findPathAt(lkr, cmt, currPath)
	if err != nil {
		return err
	}

	oldNode, err := root.Lookup(lkr, oldPath)
	if err != nil && !ie.IsNoSuchFileError(err) {
		return err
	}

	// Make sure that all write related action happen in one go:
	return lkr.Atomic(func() error {
		// Remove the node that is present at the current path:
		par, err := clearPath(lkr, currPath)
		if err != nil {
			return err
		}

		// old Node might not have yet existed back then.
		// If so, simply do not re-add it.
		if oldNode != nil {
			oldModNode, ok := oldNode.(n.ModNode)
			if !ok {
				return e.Wrapf(ie.ErrBadNode, "reset file")
			}

			// If the old node was at a different location,
			// we need to modify its path.
			oldModNode.SetName(path.Base(currPath))
			oldModNode.SetParent(lkr, par)
			oldModNode.NotifyMove(lkr, oldModNode.Path())

			if err := par.Add(lkr, oldNode); err != nil {
				return err
			}

			if err := lkr.StageNode(oldNode); err != nil {
				return err
			}
		}

		return nil
	})
}
