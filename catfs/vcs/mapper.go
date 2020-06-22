package vcs

// NOTE ON CODING STYLE:
// If you modify something in here, make sure to always
// incude "src" or "dst" in the symbol name to indicate
// to which side of the sync/diff this symbol belongs!
// Too many hours have been spent on confused debugging.

import (
	"fmt"
	"path"

	e "github.com/pkg/errors"
	c "github.com/sahib/brig/catfs/core"
	ie "github.com/sahib/brig/catfs/errors"
	n "github.com/sahib/brig/catfs/nodes"
	"github.com/sahib/brig/util/trie"
	log "github.com/sirupsen/logrus"
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
//
// If SrcWasRemoved is true, the node was deleted on the
// remote's side and we might need to propagate this remove.
// Otherwise, if src is nil, dst can be considered as missing
// file on src's side.
//
// If SrcWasMoved is true, the two nodes were purely moved,
// but not modified otherwise.
type MapPair struct {
	Src n.ModNode
	Dst n.ModNode

	SrcWasRemoved bool
	SrcWasMoved   bool
	TypeMismatch  bool
}

// flags that are set during the mapper run.
// The zero value of this struct should mean "disabled".
type flags struct {
	// The node was visited on the source side.
	// This should prohibit duplicate visits.
	srcVisited bool

	// The file was already reported/tested equal on src side.
	srcHandled bool

	// The file was already reported/tested equal on dst side.
	dstHandled bool

	// The directory consists completely of other src reports.
	srcComplete bool

	// The directory consists completely of other dst reports.
	dstComplete bool
}

// Mapper holds the state for the mapping algorithm.
type Mapper struct {
	lkrSrc, lkrDst *c.Linker
	srcRoot        n.Node
	srcHead        *n.Commit
	dstHead        *n.Commit
	flagsRoot      *trie.Node
	fn             func(pair MapPair) error
}

func (ma *Mapper) getFlags(path string) *flags {
	child := ma.flagsRoot.Lookup(path)
	if child == nil {
		child = ma.flagsRoot.InsertWithData(path, &flags{})
	}

	if child.Data == nil {
		child.Data = &flags{}
	}

	return child.Data.(*flags)
}

func (ma *Mapper) setSrcVisited(nd n.Node) {
	ma.getFlags(nd.Path()).srcVisited = true
}

func (ma *Mapper) setSrcHandled(nd n.Node) {
	ma.getFlags(nd.Path()).srcHandled = true
}

func (ma *Mapper) setDstHandled(nd n.Node) {
	ma.getFlags(nd.Path()).dstHandled = true
}

func (ma *Mapper) setSrcComplete(nd n.Node) {
	ma.getFlags(nd.Path()).srcComplete = true
}

func (ma *Mapper) setDstComplete(nd n.Node) {
	ma.getFlags(nd.Path()).dstComplete = true
}

func (ma *Mapper) isSrcVisited(nd n.Node) bool {
	return ma.getFlags(nd.Path()).srcVisited
}

func (ma *Mapper) isSrcHandled(nd n.Node) bool {
	return ma.getFlags(nd.Path()).srcHandled
}

func (ma *Mapper) isDstHandled(nd n.Node) bool {
	return ma.getFlags(nd.Path()).dstHandled
}

func (ma *Mapper) isSrcComplete(nd n.Node) bool {
	return ma.getFlags(nd.Path()).srcComplete
}

func (ma *Mapper) isDstComplete(nd n.Node) bool {
	return ma.getFlags(nd.Path()).dstComplete
}

////////////////////

func (ma *Mapper) report(src, dst n.ModNode, typeMismatch, isRemove, isMove bool) error {
	if src != nil {
		ma.setSrcHandled(src)
	}

	if dst != nil {
		ma.setDstHandled(dst)
	}

	debug("=> report", src, dst)
	return ma.fn(MapPair{
		Src:           src,
		Dst:           dst,
		TypeMismatch:  typeMismatch,
		SrcWasRemoved: isRemove,
		SrcWasMoved:   isMove,
	})
}

func (ma *Mapper) reportByType(src, dst n.ModNode) error {
	if src == nil || dst == nil {
		return ma.report(src, dst, false, false, false)
	}

	isTypeMismatch := src.Type() != dst.Type()

	if src.ContentHash().Equal(dst.ContentHash()) {
		// If the files are equal, but the location changed,
		// the file were moved.
		if src.Path() != dst.Path() {
			return ma.report(src, dst, isTypeMismatch, false, true)
		}

		// The files appear to be equal.
		// We need to remember to not output them again.
		ma.setSrcHandled(src)
		ma.setDstHandled(dst)
		return nil
	}

	return ma.report(src, dst, isTypeMismatch, false, false)
}

func (ma *Mapper) mapFile(srcCurr *n.File, dstFilePath string) error {
	// Check if we already visited this file.
	if ma.isSrcVisited(srcCurr) {
		return nil
	}

	debug("map file", srcCurr.Path(), dstFilePath)

	// Remember that we visited this node.
	ma.setSrcVisited(srcCurr)

	dstCurr, err := ma.lkrDst.LookupNodeAt(ma.dstHead, dstFilePath)
	if err != nil && !ie.IsNoSuchFileError(err) {
		return err
	}

	if dstCurr == nil {
		// We do not have this node yet, mark it for copying.
		return ma.report(srcCurr, nil, false, false, false)
	}

	switch typ := dstCurr.Type(); typ {
	case n.NodeTypeDirectory:
		// Our node seems to be a directory and theirs a file.
		// That's not something we can fix.
		dstDir, ok := dstCurr.(*n.Directory)
		if !ok {
			return ie.ErrBadNode
		}

		// File and Directory don't go well together.
		return ma.report(srcCurr, dstDir, true, false, false)
	case n.NodeTypeFile:
		// We have two competing files.
		dstFile, ok := dstCurr.(*n.File)
		if !ok {
			return ie.ErrBadNode
		}

		return ma.reportByType(srcCurr, dstFile)
	case n.NodeTypeGhost:
		// It's still possible that the file was moved on our side.
		aliveDstCurr, err := ma.ghostToAlive(ma.lkrDst, ma.dstHead, dstCurr)
		if err != nil {
			return err
		}

		return ma.reportByType(srcCurr, aliveDstCurr)
	default:
		return e.Wrapf(ie.ErrBadNode, "Unexpected node type in syncFile: %v", typ)
	}
}

func (ma *Mapper) mapDirectoryContents(srcCurr *n.Directory, dstPath string) error {
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
			ma.setSrcHandled(srcChildDir)
			dstCurrNd, err := ma.lkrDst.LookupModNodeAt(ma.dstHead, childDstPath)
			if err == nil {
				ma.setDstHandled(dstCurrNd)
			}
		case n.NodeTypeFile:
			srcChildFile, ok := srcChild.(*n.File)
			if !ok {
				return ie.ErrBadNode
			}

			if err := ma.mapFile(srcChildFile, childDstPath); err != nil {
				return err
			}
			ma.setSrcHandled(srcChildFile)
		case n.NodeTypeGhost:
			// remote ghosts are ignored, since they were handled beforehand.
		default:
			return ie.ErrBadNode
		}
	}

	return nil
}

func (ma *Mapper) mapDirectory(srcCurr *n.Directory, dstPath string, force bool) error {
	if !force {
		if ma.isSrcVisited(srcCurr) {
			return nil
		}
	}

	log.Debugf("mapping dir %s <-> %s", srcCurr.Path(), dstPath)

	ma.setSrcVisited(srcCurr)
	dstCurrNd, err := ma.lkrDst.LookupModNodeAt(ma.dstHead, dstPath)
	if err != nil && !ie.IsNoSuchFileError(err) {
		return err
	}

	if dstCurrNd == nil {
		// We never heard of this directory apparently. Go sync it.
		return ma.report(srcCurr, nil, false, false, false)
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
			return ma.report(srcCurr, nil, false, false, false)
		}

		localBackCheck, err := ma.lkrSrc.LookupNodeAt(ma.srcHead, aliveDstCurr.Path())
		if err != nil && !ie.IsNoSuchFileError(err) {
			return err
		}

		if localBackCheck == nil || localBackCheck.Type() == n.NodeTypeGhost {
			// Delete the guard again, due to the recursive call.
			return ma.mapDirectory(srcCurr, aliveDstCurr.Path(), true)
		}

		return ma.report(srcCurr, nil, false, false, false)
	}

	if dstCurrNd.Type() != n.NodeTypeDirectory {
		return ma.report(srcCurr, dstCurrNd, true, false, false)
	}

	dstCurr, ok := dstCurrNd.(*n.Directory)
	if !ok {
		return ie.ErrBadNode
	}

	// Check if we're lucky and the directory hash is equal:
	if srcCurr.ContentHash().Equal(dstCurr.ContentHash()) {
		// Remember that we visited this subtree.
		ma.setSrcHandled(srcCurr)
		ma.setDstHandled(dstCurr)

		log.Debugf(
			"%s and %s have the same content; skipping",
			srcCurr.Path(),
			dstCurr.Path(),
		)

		if srcCurr.Path() != dstCurr.Path() {
			return ma.report(srcCurr, dstCurr, false, false, true)
		}

		// If they even have the same tree hash, we can be sure that both
		// use the same path layout even. No work to do in this case.
		if srcCurr.TreeHash().Equal(dstCurr.TreeHash()) {
			return nil
		}
	}

	// Both sides have this directory, but the content differs.
	// We need to figure out recursively what exactly is different.
	return ma.mapDirectoryContents(srcCurr, dstPath)
}

func (ma *Mapper) ghostToAlive(lkr *c.Linker, head *n.Commit, nd n.Node) (n.ModNode, error) {
	partnerNd, _, err := lkr.MoveEntryPoint(nd)
	if err != nil {
		return nil, e.Wrap(err, "move entry point")
	}

	// No move partner found.
	if partnerNd == nil {
		return nil, nil
	}

	// We want to go forward in history.
	// In theory, the other direction should not happen,
	// since we're always operating on ghosts here.
	// if moveDir != c.MoveDirDstToSrc {
	// 	log.Debugf("bad move direction")
	// 	return nil, nil
	// }

	// Go forward to the most recent version of this node.
	// This is no guarantee yet that this node is reachable
	// from the head commit (it might have been removed...)
	mostRecent, err := lkr.NodeByInode(partnerNd.Inode())
	if err != nil {
		return nil, err
	}

	if mostRecent == nil {
		err = fmt.Errorf("mapper: No such node with inode %d", partnerNd.Inode())
		return nil, err
	}

	// This should usually not happen, but just to be sure.
	if mostRecent.Type() == n.NodeTypeGhost {
		return nil, nil
	}

	reacheable, err := lkr.LookupNodeAt(head, mostRecent.Path())
	if err != nil && !ie.IsNoSuchFileError(err) {
		return nil, e.Wrapf(err, "ghost2alive: lookupAt")
	}

	if reacheable == nil {
		return nil, nil
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

type ghostDir struct {
	// source directory.
	srcDir *n.Directory

	// mapped path in lkrDst
	dstPath string
}

func (ma *Mapper) handleGhostsWithoutAliveNd(srcNd n.Node) error {
	dstNd, err := ma.lkrDst.LookupNodeAt(ma.dstHead, srcNd.Path())
	if err != nil && !ie.IsNoSuchFileError(err) {
		return err
	}

	// Check if we maybe already removed or moved the node:
	if dstNd != nil && dstNd.Type() != n.NodeTypeGhost {
		dstModNd, ok := dstNd.(n.ModNode)
		if !ok {
			return ie.ErrBadNode
		}

		// Report that the file is missing on src's side.
		return ma.report(nil, dstModNd, false, true, false)
	}

	// does not exist on both sides, nothing to report.
	return nil
}

func (ma *Mapper) extractGhostDirs() ([]ghostDir, error) {
	movedSrcDirs := []ghostDir{}
	return movedSrcDirs, n.Walk(ma.lkrSrc, ma.srcRoot, true, func(srcNd n.Node) error {
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
			return ma.handleGhostsWithoutAliveNd(srcNd)
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
			// Node maybe also moved. If so, try to resolve it to the full node:
			if dstRefNd.Type() == n.NodeTypeGhost {
				aliveOrig, err := ma.ghostToAlive(ma.lkrDst, ma.dstHead, dstRefNd)
				if err != nil {
					return err
				}

				dstRefNd = aliveOrig
			}
		}

		// The node was removed on dst:
		// We will detect the removal later.
		if dstRefNd == nil {
			return nil
		}

		dstRefModNd, ok := dstRefNd.(n.ModNode)
		if !ok {
			return e.Wrapf(ie.ErrBadNode, "dstRefModNd is not a file or directory: %v", dstRefNd)
		}

		switch aliveSrcNd.Type() {
		case n.NodeTypeFile:
			// Mark those both ghosts and original node as visited.
			err = ma.mapFile(aliveSrcNd.(*n.File), dstRefModNd.Path())
			ma.setSrcVisited(aliveSrcNd)
			ma.setSrcVisited(srcNd)
			return err
		case n.NodeTypeDirectory:
			// ma.setSrcVisited(srcNd)
			if dstRefNd.Type() != n.NodeTypeDirectory {
				return ma.report(aliveSrcNd, dstRefModNd, true, false, false)
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
}

func (ma *Mapper) handleGhosts() error {
	movedSrcDirs, err := ma.extractGhostDirs()
	if err != nil {
		return err
	}

	// Handle moved paths after handling single files.
	// (mapDirectory assumes that moved files in it were already handled).
	for _, movedSrcDir := range movedSrcDirs {
		log.Debugf("map: %v %v", movedSrcDir.srcDir.Path(), movedSrcDir.dstPath)
		if err := ma.mapDirectory(movedSrcDir.srcDir, movedSrcDir.dstPath, false); err != nil {
			return err
		}
	}

	return nil
}

// NewMapper creates a new mapper object that is capable of finding pairs of
// nodes between lkrDst and lkrSrc.
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
		lkrSrc:    lkrSrc,
		lkrDst:    lkrDst,
		srcHead:   srcHead,
		dstHead:   dstHead,
		srcRoot:   srcRoot,
		flagsRoot: trie.NewNodeWithData(&flags{}),
	}, nil
}

func (ma *Mapper) nodeIsHandled(nd n.Node, srcToDst bool) bool {
	if srcToDst {
		return ma.isSrcHandled(nd)
	}

	return ma.isDstHandled(nd)
}

func (ma *Mapper) isComplete(lkr *c.Linker, root n.Node, srcToDst bool) (bool, error) {
	// If the file was already handled: ignore it completely.
	if ma.nodeIsHandled(root, srcToDst) {
		return false, nil
	}

	if root.Type() != n.NodeTypeDirectory {
		return true, nil
	}

	dir, ok := root.(*n.Directory)
	if !ok {
		return false, ie.ErrBadNode
	}

	children, err := dir.ChildrenSorted(lkr)
	if err != nil {
		return false, err
	}

	nComplete := 0
	for _, child := range children {
		if ma.nodeIsHandled(child, srcToDst) {
			continue
		}

		isComplete, err := ma.isComplete(lkr, child, srcToDst)
		if err != nil {
			return false, err
		}

		if isComplete {
			nComplete++
		}
	}

	// If all children were not handled & are complete we copy the flag.
	if nComplete == len(children) {
		if srcToDst {
			ma.setSrcComplete(root)
		} else {
			ma.setDstComplete(root)
		}

		return true, nil
	}

	return false, nil
}

// extractLeftovers goes over all nodes in src that were not covered
// yet by previous measures. It will report any src node without a match then.
func (ma *Mapper) extractLeftovers(lkr *c.Linker, root *n.Directory, srcToDst bool) error {
	if ma.nodeIsHandled(root, srcToDst) {
		return nil
	}

	if _, err := ma.isComplete(lkr, root, srcToDst); err != nil {
		return err
	}

	// Implement a basic walk/DFS with filtering:
	children, err := root.ChildrenSorted(lkr)
	if err != nil {
		return err
	}

	for _, child := range children {
		debug(fmt.Sprintf("extract: %v", child.Path()))
		if ma.nodeIsHandled(child, srcToDst) {
			debug(fmt.Sprintf("node is handled: %v", child.Path()))
			continue
		}

		switch child.Type() {
		case n.NodeTypeDirectory:
			dir, ok := child.(*n.Directory)
			if !ok {
				return ie.ErrBadNode
			}

			var complete bool
			if srcToDst {
				complete = ma.isSrcComplete(dir)
			} else {
				complete = ma.isDstComplete(dir)
			}

			debug(fmt.Sprintf("is complete: %v %v", child.Path(), complete))
			if complete {
				if srcToDst {
					err = ma.report(dir, nil, false, false, false)
				} else {
					err = ma.report(nil, dir, false, false, false)
				}

				if err != nil {
					return err
				}
			} else {
				if err := ma.extractLeftovers(lkr, dir, srcToDst); err != nil {
					return err
				}
			}
		case n.NodeTypeFile:
			file, ok := child.(*n.File)
			if !ok {
				return ie.ErrBadNode
			}

			// Report the leftover:
			if srcToDst {
				err = ma.report(file, nil, false, false, false)
			} else {
				err = ma.report(nil, file, false, false, false)
			}

			if err != nil {
				return err
			}
		case n.NodeTypeGhost:
			// Those were already handled (or are not important)
		}
	}

	return nil
}

// Map calls `fn` for each pairing that was found. Equal files and
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
func (ma *Mapper) Map(fn func(pair MapPair) error) error {
	ma.fn = fn
	log.Debugf("mapping ghosts")
	if err := ma.handleGhosts(); err != nil {
		return err
	}

	log.Debugf("mapping non-ghosts")

	switch ma.srcRoot.Type() {
	case n.NodeTypeDirectory:
		dir, ok := ma.srcRoot.(*n.Directory)
		if !ok {
			return ie.ErrBadNode
		}

		if err := ma.mapDirectory(dir, dir.Path(), false); err != nil {
			return err
		}

		// Get root directories:
		// (only get them now since, in theory, mapFn could have changed things)
		srcRoot, err := ma.lkrSrc.DirectoryByHash(ma.srcHead.Root())
		if err != nil {
			return err
		}

		dstRoot, err := ma.lkrDst.DirectoryByHash(ma.dstHead.Root())
		if err != nil {
			return err
		}
		debug("-- Extract leftover src")

		// Extract things in "src" that were not mapped yet.
		// These are files that can be added to our inventory,
		// since we have notthing that mapped to them.
		if err := ma.extractLeftovers(ma.lkrSrc, srcRoot, true); err != nil {
			return err
		}
		debug("-- Extract leftover dst")

		// Check for files that we have, but dst does not.
		// We call those files "missing".
		return ma.extractLeftovers(ma.lkrDst, dstRoot, false)
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
