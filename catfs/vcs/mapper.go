package vcs

// NOTE ON CODING STYLE:
// If you modify something in here, make sure to always
// incude "src" or "dst" in the symbol name to indicate
// to which side of the sync/diff this symbol belongs!

import (
	"fmt"
	"path"

	e "github.com/pkg/errors"
	c "github.com/sahib/brig/catfs/core"
	ie "github.com/sahib/brig/catfs/errors"
	n "github.com/sahib/brig/catfs/nodes"
)

// MapPair is a pair of nodes (a file or a directory)
// One of Src and Dst might be nil:
// - If Src is nil, the node was removed on the remote side.
// - If Dst is nil, the node was added on the remote side.
//
// Both shall never be nil at the same time.
//
// If TypeMismatch is true, nodes have a different type
// and need conflict resolution.
type MapPair struct {
	Src          n.ModNode
	Dst          n.ModNode
	TypeMismatch bool
}

type Mapper struct {
	lkrSrc, lkrDst *c.Linker
	srcRoot        n.Node
	srcHead        *n.Commit
	dstHead        *n.Commit
	fn             func(pair MapPair) error
	srcVisited     map[string]n.Node
	dstVisited     map[string]n.Node
}

func (ma *Mapper) report(src, dst n.ModNode, typeMismatch bool) error {
	if dst != nil {
		ma.dstVisited[dst.Path()] = dst
	}

	return ma.fn(MapPair{
		Src:          src,
		Dst:          dst,
		TypeMismatch: typeMismatch,
	})
}

func (ma *Mapper) mapFile(srcCurr *n.File, dstFilePath string) error {
	// Check if we already visited this file.
	if _, ok := ma.srcVisited[srcCurr.Path()]; ok {
		return nil
	}

	// Remember that we visited this node.
	ma.srcVisited[srcCurr.Path()] = srcCurr

	dstCurr, err := ma.lkrDst.LookupNodeAt(ma.dstHead, dstFilePath)
	if err != nil && !ie.IsNoSuchFileError(err) {
		return err
	}

	if dstCurr == nil {
		// We do not have this node yet, mark it for copying.
		return ma.report(srcCurr, nil, false)
	}

	ma.dstVisited[dstCurr.Path()] = dstCurr

	switch typ := dstCurr.Type(); typ {
	case n.NodeTypeDirectory:
		// Our node seems to be a directory and theirs a file.
		// That's not something we can fix.
		dstDir, ok := dstCurr.(*n.Directory)
		if !ok {
			return ie.ErrBadNode
		}

		// File and Directory don't go well together.
		return ma.report(srcCurr, dstDir, true)
	case n.NodeTypeFile:
		// We have two competing files. Let's figure out if the changes done to
		// them are compatible.
		dstFile, ok := dstCurr.(*n.File)
		if !ok {
			return ie.ErrBadNode
		}

		// We still have the slight chance that both files
		// are equal and thus we do not need to do any resolving.
		if dstFile.Content().Equal(srcCurr.Content()) {
			return nil
		}

		return ma.report(srcCurr, dstFile, false)
	case n.NodeTypeGhost:
		// It's still possible that the file was moved on our side.
		aliveDstCurr, err := ma.ghostToAlive(ma.lkrDst, ma.dstHead, dstCurr)
		if err != nil {
			return err
		}

		isTypeMismatch := false
		if aliveDstCurr != nil && aliveDstCurr.Type() != n.NodeTypeFile {
			isTypeMismatch = true
		}

		return ma.report(srcCurr, aliveDstCurr, isTypeMismatch)
	default:
		return e.Wrapf(ie.ErrBadNode, "Unexpected node type in syncFile: %v", typ)
	}
}

func (ma *Mapper) mapDirectory(srcCurr *n.Directory, dstPath string, force bool) error {
	if !force {
		if _, ok := ma.srcVisited[srcCurr.Path()]; ok {
			return nil
		}
	}

	ma.srcVisited[srcCurr.Path()] = srcCurr
	dstCurrNd, err := ma.lkrDst.LookupModNodeAt(ma.dstHead, dstPath)
	if err != nil && !ie.IsNoSuchFileError(err) {
		return err
	}

	if dstCurrNd == nil {
		// We never heard of this directory apparently. Go sync it.
		return ma.report(srcCurr, nil, false)
	}

	// Special case: The node might have been moved on dst's side.
	// We might notice this, if dst type is a ghost.
	if dstCurrNd.Type() == n.NodeTypeGhost {
		aliveDstCurr, err := ma.ghostToAlive(ma.lkrDst, ma.dstHead, dstCurrNd)
		if err != nil {
			return err
		}

		// No sibling found for this ghost.
		if aliveDstCurr == nil {
			return ma.report(srcCurr, nil, false)
		}

		localBackCheck, err := ma.lkrSrc.LookupNodeAt(ma.srcHead, aliveDstCurr.Path())
		if err != nil && !ie.IsNoSuchFileError(err) {
			return err
		}

		if localBackCheck == nil || localBackCheck.Type() == n.NodeTypeGhost {
			// Delete the guard again, due to the recursive call.
			return ma.mapDirectory(srcCurr, aliveDstCurr.Path(), true)
		}

		return ma.report(srcCurr, nil, false)
	}

	if dstCurrNd.Type() != n.NodeTypeDirectory {
		return ma.report(srcCurr, dstCurrNd, true)
	}

	dstCurr, ok := dstCurrNd.(*n.Directory)
	if !ok {
		return ie.ErrBadNode
	}

	// Check if we're lucky and the directory hash is equal:
	if srcCurr.Hash().Equal(dstCurr.Hash()) {
		// Remember that we visited this subtree.
		ma.dstVisited[dstCurr.Path()] = dstCurr
		return nil
	}

	// Both sides have this directory, but the content differs.
	// We need to figure out recursively what exactly is different.
	srcChildren, err := srcCurr.ChildrenSorted(ma.lkrSrc)
	if err != nil {
		return err
	}

	for _, srcChild := range srcChildren {
		childDstPath := path.Join(dstPath, srcChild.Name())
		switch srcChild.Type() {
		case n.NodeTypeDirectory:
			srcChildDir, ok := srcChild.(*n.Directory)
			if !ok {
				return ie.ErrBadNode
			}

			if err := ma.mapDirectory(srcChildDir, childDstPath, false); err != nil {
				return err
			}
		case n.NodeTypeFile:
			srcChildFile, ok := srcChild.(*n.File)
			if !ok {
				return ie.ErrBadNode
			}

			if err := ma.mapFile(srcChildFile, childDstPath); err != nil {
				return err
			}
		case n.NodeTypeGhost:
			// remote ghosts are ignored, since they were handled beforehand.
		default:
			return ie.ErrBadNode
		}
	}

	return nil
}

// ghostToAlive receives a `nd` and tries to find
func (ma *Mapper) ghostToAlive(lkr *c.Linker, head *n.Commit, nd n.Node) (n.ModNode, error) {
	partnerNd, moveDir, err := lkr.MoveEntryPoint(nd)
	if err != nil {
		return nil, err
	}

	// No move partner found.
	if partnerNd == nil {
		return nil, nil
	}

	// We want to go forward in history.
	// In theory, the other direction should not happen.
	if moveDir != c.MoveDirDstToSrc {
		return nil, nil
	}

	// Go forward to the most recent version of this node.
	// This is no guarantee yet that this node is reachable
	// from the head commit (it might have been removed...)
	mostRecent, err := lkr.NodeByInode(partnerNd.Inode())
	if err != nil {
		return nil, err
	}

	if mostRecent == nil {
		err = fmt.Errorf("sync: No such node with inode %d", partnerNd.Inode())
		return nil, err
	}

	// This should usually not happen, but just to be sure.
	if mostRecent.Type() == n.NodeTypeGhost {
		return nil, nil
	}

	reacheable, err := lkr.LookupNodeAt(head, mostRecent.Path())
	if err != nil {
		return nil, err
	}

	if reacheable.Inode() != mostRecent.Inode() {
		// The node is still reachable, but it was changed
		// (i.e. by removing and re-adding it -> different inode)
		return nil, nil
	}

	reacheableModNd, ok := reacheable.(n.ModNode)
	if !ok {
		return nil, ie.ErrBadNode
	}

	return reacheableModNd, nil
}

func (ma *Mapper) handleGhosts() error {
	type ghostDir struct {
		// source directory.
		srcDir *n.Directory

		// mapped path in lkrDst
		dstPath string
	}

	movedSrcDirs := []ghostDir{}

	err := n.Walk(ma.lkrSrc, ma.srcRoot, true, func(srcNd n.Node) error {
		// Ignore everything that is not a ghost.
		if srcNd.Type() != n.NodeTypeGhost {
			return nil
		}

		aliveSrcNd, err := ma.ghostToAlive(ma.lkrSrc, ma.srcHead, srcNd)
		if err != nil {
			return err
		}

		if aliveSrcNd == nil {
			// It's a ghost, but it has no living counterpart.
			// This node *might* have been removed on the remote side.
			// Try to see if we have a node at this path, the next step
			// of sync then needs to decide if the node needs to be removed.
			dstNd, err := ma.lkrDst.LookupNodeAt(ma.dstHead, srcNd.Path())
			if err != nil && !ie.IsNoSuchFileError(err) {
				return err
			}

			if dstNd != nil && dstNd.Type() != n.NodeTypeGhost {
				dstModNd, ok := dstNd.(n.ModNode)
				if !ok {
					return ie.ErrBadNode
				}

				return ma.report(nil, dstModNd, false)
			}

			// Not does not exist on both sides, nothing to report.
			return nil
		}

		// At this point we know that the ghost related to a moved file.
		// Check if we have a file at the same place.
		dstNd, err := ma.lkrDst.LookupNodeAt(ma.dstHead, aliveSrcNd.Path())
		if err != nil && !ie.IsNoSuchFileError(err) {
			return err
		}

		if dstNd != nil && dstNd.Type() != n.NodeTypeGhost {
			// The node already exists in our place. No way we can really merge
			// it cleanly, so just handle the ghost as normal file and potentially
			// apply the normal conflict resolution later on.
			return nil
		}

		dstRefNd, err := ma.lkrDst.LookupNodeAt(ma.dstHead, srcNd.Path())
		if err != nil && !ie.IsNoSuchFileError(err) {
			return err
		}

		if dstRefNd != nil {
			if dstRefNd.Type() == n.NodeTypeGhost {
				aliveOrig, err := ma.ghostToAlive(ma.lkrDst, ma.dstHead, dstRefNd)
				if err != nil {
					return err
				}

				dstRefNd = aliveOrig
			}
		}

		dstRefModNd, ok := dstRefNd.(n.ModNode)
		if !ok {
			return e.Wrapf(ie.ErrBadNode, "dstRefModNd is not a file or directory: %v", dstRefNd)
		}

		switch aliveSrcNd.Type() {
		case n.NodeTypeFile:
			// Mark those both ghosts and original node as visited.
			ma.srcVisited[aliveSrcNd.Path()] = aliveSrcNd
			ma.srcVisited[srcNd.Path()] = srcNd

			if !aliveSrcNd.Hash().Equal(dstRefNd.Hash()) {
				mismatch := dstRefNd.Type() != aliveSrcNd.Type()
				return ma.report(aliveSrcNd, dstRefModNd, mismatch)
			}

			return nil
		case n.NodeTypeDirectory:
			ma.srcVisited[srcNd.Path()] = srcNd
			if dstRefNd.Type() != n.NodeTypeDirectory {
				return ma.report(aliveSrcNd, dstRefModNd, true)
			}

			aliveSrcDir, ok := aliveSrcNd.(*n.Directory)
			if !ok {
				return ie.ErrBadNode
			}

			movedSrcDirs = append(movedSrcDirs, ghostDir{
				srcDir:  aliveSrcDir,
				dstPath: dstRefNd.Path(),
			})

			return nil
		default:
			return e.Wrapf(ie.ErrBadNode, "Unexpected type in handle ghosts: %v", err)
		}
	})

	if err != nil {
		return err
	}

	// Handle moved paths after handling single files.
	// (mapDirectory assumes that moved files in it were already handled).
	for _, movedSrcDir := range movedSrcDirs {
		if err := ma.mapDirectory(movedSrcDir.srcDir, movedSrcDir.dstPath, false); err != nil {
			return err
		}
	}

	return nil
}

func NewMapper(lkrSrc, lkrDst *c.Linker, srcHead, dstHead *n.Commit, srcRoot n.Node) (*Mapper, error) {
	var err error

	if srcHead == nil {
		srcHead, err = lkrSrc.Head()
		if err != nil {
			return nil, err
		}
	}

	if dstHead == nil {
		dstHead, err = lkrDst.Head()
		if err != nil {
			return nil, err
		}
	}

	return &Mapper{
		lkrSrc:     lkrSrc,
		lkrDst:     lkrDst,
		srcHead:    srcHead,
		dstHead:    dstHead,
		srcRoot:    srcRoot,
		srcVisited: make(map[string]n.Node),
		dstVisited: make(map[string]n.Node),
	}, nil
}

// extractDstLeftovers goes over all nodes in dst that were not covered
// yet by previous measures. It will report any dst node without a match then.
func (ma *Mapper) extractDstLeftovers(dstRoot *n.Directory) error {
	if _, visited := ma.dstVisited[dstRoot.Path()]; visited {
		return nil
	}

	children, err := dstRoot.ChildrenSorted(ma.lkrDst)
	if err != nil {
		return err
	}

	for _, dstChild := range children {
		if _, visited := ma.dstVisited[dstChild.Path()]; visited {
			continue
		}

		switch dstChild.Type() {
		case n.NodeTypeDirectory:
			dstDir, ok := dstChild.(*n.Directory)
			if !ok {
				return ie.ErrBadNode
			}

			if err := ma.extractDstLeftovers(dstDir); err != nil {
				return err
			}
		case n.NodeTypeFile:
			dstFile, ok := dstChild.(*n.File)
			if !ok {
				return ie.ErrBadNode
			}

			if err := ma.report(nil, dstFile, false); err != nil {
				return err
			}
		case n.NodeTypeGhost:
			// Those were already handled (or are not important)
		}
	}

	return nil
}

// Diff calls `fn` for each pairing that was found. Equal files and
// directories are not reported. Most directories are also not reported, but
// if they are empty and not present on our side they will. No ghosts will be
// reported.
//
// Some implementation background for the curious reader:
//
// In the simplest case a filesystem is a tree and the assumption can be made
// that one node that lives at the same path on both sides is the same "file"
// (i.e. in terms of "this is the file that the user wants to synchronize with").
//
// With ghosts though, we have nodes that can indicate a removed or a moved file.
// Due to moved files the filesystem tree becomes a graph and the mapping
// algorithm (that is the base of Mapper) needs to do a depth first search
// and thus needs to remember already visited nodes.
//
// Since moved nodes also takes priorty we need to iterate over all ghosts first,
// and mark their respective counterparts or report that they were removed on
// the remote side (i.e. no counterpart exists.). Only after that we cycle
// through all other nodes and assume that files living at the same path
// reference the same "file". At this point we can treat the file graph
// as tree again by ignoring all ghosts.
//
// A special case is when a file was moved on one side but, a file exists
// already on the other side. In this case the already existing files wins.
//
// Some examples of the described behaviours can be found in the tests of Mapper.
// TODO: write down some examples from notebook.
func (ma *Mapper) Map(fn func(pair MapPair) error) error {
	ma.fn = fn
	if err := ma.handleGhosts(); err != nil {
		return err
	}

	// TODO: Add step for extracting all non visited nodes from dst.

	switch ma.srcRoot.Type() {
	case n.NodeTypeDirectory:
		dir, ok := ma.srcRoot.(*n.Directory)
		if !ok {
			return ie.ErrBadNode
		}

		if err := ma.mapDirectory(dir, dir.Path(), false); err != nil {
			return err
		}

		dstRoot, err := ma.lkrDst.DirectoryByHash(ma.dstHead.Root())
		if err != nil {
			return err
		}

		return ma.extractDstLeftovers(dstRoot)
	case n.NodeTypeFile:
		file, ok := ma.srcRoot.(*n.File)
		if !ok {
			return ie.ErrBadNode
		}

		return ma.mapFile(file, file.Path())
	case n.NodeTypeGhost:
		return nil
	default:
		return e.Wrapf(ie.ErrBadNode, "Unexpected type in route(): %v", ma.srcRoot)
	}

}
