package catfs

import (
	"fmt"
	"path"
	"strings"

	n "github.com/disorganizer/brig/catfs/nodes"
	e "github.com/pkg/errors"
)

const (
	// ChangeTypeNone means that a node did not change (compared to HEAD)
	ChangeTypeNone = ChangeType(0)
	// ChangeTypeAdd says that the node was initially added after HEAD.
	ChangeTypeAdd = ChangeType(1 << iota)
	// ChangeTypeModify says that the the node was modified after HEAD
	ChangeTypeModify
	// ChangeTypeMove says that the node was moved after HEAD.
	// Note that Move and Modify may happen at the same time.
	ChangeTypeMove
	// ChangeTypeRemove says that the node was removed after HEAD.
	ChangeTypeRemove
)

// ChangeType is a mask of possible state change events.
type ChangeType uint8

// String will convert a ChangeType to a human readable form
func (ct ChangeType) String() string {
	v := []string{}

	if ct&ChangeTypeAdd != 0 {
		v = append(v, "added")
	}
	if ct&ChangeTypeModify != 0 {
		v = append(v, "modified")
	}
	if ct&ChangeTypeMove != 0 {
		v = append(v, "moved")
	}
	if ct&ChangeTypeRemove != 0 {
		v = append(v, "removed")
	}

	if len(v) == 0 {
		return "none"
	}

	return strings.Join(v, "|")
}

func (ct ChangeType) IsCompatible(ot ChangeType) bool {
	modifyMask := ChangeTypeAdd | ChangeTypeModify

	if ct&modifyMask != 0 && ot&modifyMask != 0 {
		return false
	}

	return true
}

///////////////////////////

// NodeState represents a single change of a node between two commits.
// TODO: Rename. NodeState is clumsy.
type NodeState struct {
	// Mask is a bitmask of changes that were made.
	Mask ChangeType
	// Head is the commit that was the current HEAD when this change happened.
	// Note that this is NOT the commit that contains the change, but the commit before.
	Head *n.Commit
	// Curr is the node with the attributes at a specific state
	Curr n.ModNode
}

func (ns *NodeState) String() string {
	return fmt.Sprintf("<%s:%s>", ns.Curr.Path(), ns.Mask)
}

// HistoryWalker provides a way to iterate over all changes a single Node had.
// It is capable of tracking a file even over multiple moves.
//
// The API is loosely modeled after bufio.Scanner and can be used like this:
//
// 	head, _ := lkr.Head()
// 	nd, _ := lkr.LookupFile("/x")
// 	walker := NewHistoryWalker(lkr, head, nd)
// 	for walker.Next() {
// 		walker.Change()
// 	}
//
// 	if walker.Error() != nil {
// 		// Handle errors.
// 	}
type HistoryWalker struct {
	lkr    *Linker
	head   *n.Commit
	curr   n.ModNode
	next   n.ModNode
	err    error
	state  *NodeState
	isLast bool
}

// NewHistoryWalker will return a new HistoryWalker that will yield changes of
// `node` starting from the state in `cmt` until the root commit if desired.
// Note that it is not checked that `node` is actually part of `cmt`.
func NewHistoryWalker(lkr *Linker, cmt *n.Commit, node n.ModNode) *HistoryWalker {
	return &HistoryWalker{
		lkr:  lkr,
		head: cmt,
		curr: node,
	}
}

// maskFromState figures out the change mask based on the current state
func (hw *HistoryWalker) maskFromState() ChangeType {
	mask := ChangeType(0)

	// Initial state; no succesor known yet to compare too.
	if hw.next == nil {
		return mask
	}

	isGhostCurr := hw.curr.Type() == n.NodeTypeGhost
	isGhostNext := hw.next.Type() == n.NodeTypeGhost

	currHash, err := n.ContentHash(hw.curr)
	if err != nil {
		return ChangeTypeNone
	}

	nextHash, err := n.ContentHash(hw.next)
	if err != nil {
		return ChangeTypeNone
	}

	// If the hash differs, there's likely a modification going on.
	if !currHash.Equal(nextHash) {
		mask |= ChangeTypeModify
	}

	if hw.next.Path() != hw.curr.Path() {
		mask |= ChangeTypeMove
	} else {
		// If paths did not move, but the current node is a ghost,
		// then it means that the node was removed in this commit.
		if isGhostCurr && !isGhostNext {
			mask |= ChangeTypeAdd
		}
	}

	// If the next node is a ghost it was deleted after this HEAD.
	if hw.next.Type() == n.NodeTypeGhost {
		// Be safe and check that it was not a move:
		if hw.next.Path() == hw.curr.Path() {
			mask |= ChangeTypeRemove
		}
	}
	return mask
}

// Next advances the walker to the next commit.
// Call State() to get the current state after.
// If there are no commits left or an error happended,
// false is returned. True otherwise. You should check
// after a failing Next() if an error happended via Err()
func (hw *HistoryWalker) Next() bool {
	if hw.err != nil {
		return false
	}

	if hw.head == nil {
		return false
	}

	if hw.isLast {
		hw.state = &NodeState{
			Head: hw.head,
			Mask: ChangeTypeAdd,
			Curr: hw.curr,
		}
		hw.head = nil
		return true
	}

	// Pack up the current state:
	hw.state = &NodeState{
		Head: hw.head,
		Mask: hw.maskFromState(),
		Curr: hw.curr,
	}

	// Check if this node participated in a move:
	prev, direction, err := hw.lkr.MoveMapping(hw.head, hw.curr)
	if err != nil {
		hw.err = err
		return false
	}

	if prev != nil && prev.Type() == n.NodeTypeGhost {
		prevGhost, ok := prev.(*n.Ghost)
		if !ok {
			hw.err = n.ErrBadNode
			return false
		}

		prev = prevGhost.OldNode()
	}

	// Advance to the previous commit:
	prevHead, err := hw.head.Parent(hw.lkr)
	if err != nil {
		hw.err = err
		return false
	}

	// We ran out of commits to check.
	if prevHead == nil {
		hw.head = nil
		return true
	}

	prevHeadCommit, ok := prevHead.(*n.Commit)
	if !ok {
		hw.err = e.Wrap(n.ErrBadNode, "history: bad commit")
		return false
	}

	// Assumption here: The move mapping should only store one move per commit.
	// i.e: for move(a, b); a and b should always be in different commits.
	// This is enforced by the logic in MakeCommit()
	if prev == nil || direction != MoveDirSrcToDst {
		// No valid move mapping found, node was probably not moved.
		// Assume that we can reach it directly via it's path.
		currRoot, err := hw.lkr.DirectoryByHash(prevHeadCommit.Root())
		if err != nil {
			hw.err = e.Wrap(err, "Cannot find previous root directory")
			return false
		}

		prev, err = currRoot.Lookup(hw.lkr, hw.curr.Path())
		if n.IsNoSuchFileError(err) {
			// The file did not exist in the previous commit (no ghost!)
			// It must have been added in this commit.
			hw.isLast = true
			return true
		}

		if err != nil {
			hw.err = e.Wrap(err, "history: prev root lookup failed")
			return false
		}
	}

	prevModNode, ok := prev.(n.ModNode)
	if !ok {
		hw.err = e.Wrap(n.ErrBadNode, "history: bad mod node")
		return false
	}

	// Swap for the next call to Next():
	hw.curr, hw.next = prevModNode, hw.curr
	hw.head = prevHeadCommit
	return true
}

// State returns the current change state.
// Note that the change may have ChangeTypeNone as Mask if nothing changed.
// If you only want states where it actually changed, just filter those.
func (hw *HistoryWalker) State() *NodeState {
	return hw.state
}

// Err returns the last happened error or nil if none.
func (hw *HistoryWalker) Err() error {
	return hw.err
}

// History returns a list of `nd`'s states starting with the commit in `start`
// and stopping at `stop`. `stop` can be nil; in this case all commits will be
// iterated. The returned list has the most recent change upfront, and the
// latest change as lsat element.
func History(lkr *Linker, nd n.ModNode, start, stop *n.Commit) ([]*NodeState, error) {
	states := make([]*NodeState, 0)
	walker := NewHistoryWalker(lkr, start, nd)

	for walker.Next() {
		state := walker.State()
		if stop != nil && state.Head.Hash().Equal(stop.Hash()) {
			break
		}

		states = append(states, state)
	}

	if err := walker.Err(); err != nil {
		return nil, err
	}

	return states, nil
}

////////////////////////////////////////

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
	lkrSrc, lkrDst *Linker
	srcRoot        n.Node
	fn             func(pair MapPair) error
	visited        map[string]n.Node
}

func (ma *Mapper) mapFile(srcCurr *n.File, dstFilePath string) error {
	// Check if we already visited this file.
	if _, ok := ma.visited[srcCurr.Path()]; ok {
		return nil
	}

	// Remember that we visited this node.
	ma.visited[srcCurr.Path()] = srcCurr

	dstCurr, err := ma.lkrDst.LookupNode(dstFilePath)
	if err != nil && !n.IsNoSuchFileError(err) {
		return err
	}

	if dstCurr == nil {
		// We do not have this node yet, mark it for copying.
		return ma.fn(MapPair{
			Src:          srcCurr,
			Dst:          nil,
			TypeMismatch: false,
		})
	}

	switch typ := dstCurr.Type(); typ {
	case n.NodeTypeDirectory:
		// Our node seems to be a directory and theirs a file.
		// That's not something we can fix.
		dstDir, ok := dstCurr.(*n.Directory)
		if !ok {
			return n.ErrBadNode
		}

		// File and Directory don't go well together.
		return ma.fn(MapPair{
			Src:          srcCurr,
			Dst:          dstDir,
			TypeMismatch: true,
		})
	case n.NodeTypeFile:
		// We have two competing files. Let's figure out if the changes done to
		// them are compatible.
		dstFile, ok := dstCurr.(*n.File)
		if !ok {
			return n.ErrBadNode
		}

		// We still have the slight chance that both files
		// are equal and thus we do not need to do any resolving.
		if dstFile.Content().Equal(srcCurr.Content()) {
			return nil
		}

		return ma.fn(MapPair{
			Src:          srcCurr,
			Dst:          dstFile,
			TypeMismatch: false,
		})
	case n.NodeTypeGhost:
		// It's still possible that the file was moved on our side.
		aliveDstCurr, err := ma.ghostToAlive(ma.lkrDst, dstCurr)
		if err != nil {
			return err
		}

		isTypeMismatch := false
		if aliveDstCurr != nil && aliveDstCurr.Type() != n.NodeTypeFile {
			isTypeMismatch = true
		}

		return ma.fn(MapPair{
			Src:          srcCurr,
			Dst:          aliveDstCurr,
			TypeMismatch: isTypeMismatch,
		})

		return nil
	default:
		return e.Wrapf(n.ErrBadNode, "Unexpected node type in syncFile: %v", typ)
	}
}

func (ma *Mapper) mapDirectory(srcCurr *n.Directory, dstPath string) error {
	if _, ok := ma.visited[srcCurr.Path()]; ok {
		return nil
	}

	ma.visited[srcCurr.Path()] = srcCurr
	dstCurrNd, err := ma.lkrDst.LookupModNode(dstPath)
	if err != nil && !n.IsNoSuchFileError(err) {
		return err
	}

	if dstCurrNd == nil {
		// We never heard of this directory apparently. Go sync it.
		return ma.fn(MapPair{
			Src:          srcCurr,
			Dst:          nil,
			TypeMismatch: false,
		})
	}

	// Special case: The node might have been moved on dst's side.
	// We might notice this, if dst type is a ghost.
	if dstCurrNd.Type() == n.NodeTypeGhost {
		aliveDstCurr, err := ma.ghostToAlive(ma.lkrDst, dstCurrNd)
		if err != nil {
			return err
		}

		// No sibling found for this ghost.
		if aliveDstCurr == nil {
			return ma.fn(MapPair{
				Src:          srcCurr,
				Dst:          nil,
				TypeMismatch: false,
			})
		}

		localBackCheck, err := ma.lkrSrc.LookupNode(aliveDstCurr.Path())
		if err != nil && !n.IsNoSuchFileError(err) {
			return err
		}

		if localBackCheck == nil || localBackCheck.Type() == n.NodeTypeGhost {
			// Delete the guard again, due to to recursive call.
			// TODO: This feels a bit hacky.
			delete(ma.visited, srcCurr.Path())
			return ma.mapDirectory(srcCurr, aliveDstCurr.Path())
		}

		// TODO: Make this to report() again to save a few lines.
		return ma.fn(MapPair{
			Src:          srcCurr,
			Dst:          nil,
			TypeMismatch: false,
		})
	}

	if dstCurrNd.Type() != n.NodeTypeDirectory {
		return ma.fn(MapPair{
			Src:          srcCurr,
			Dst:          dstCurrNd,
			TypeMismatch: true,
		})
	}

	dstCurr, ok := dstCurrNd.(*n.Directory)
	if !ok {
		return n.ErrBadNode
	}

	// Check if we're lucky and the directory hash is equal:
	if srcCurr.Hash().Equal(dstCurr.Hash()) {
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
				return n.ErrBadNode
			}

			if err := ma.mapDirectory(srcChildDir, childDstPath); err != nil {
				return err
			}
		case n.NodeTypeFile:
			srcChildFile, ok := srcChild.(*n.File)
			if !ok {
				return n.ErrBadNode
			}

			if err := ma.mapFile(srcChildFile, childDstPath); err != nil {
				return err
			}
		case n.NodeTypeGhost:
			// remote ghosts are ignored, since they were handled beforehand.
		default:
			return n.ErrBadNode
		}
	}

	return nil
}

// ghostToAlive receives a `nd` and tries to find
func (ma *Mapper) ghostToAlive(lkr *Linker, nd n.Node) (n.ModNode, error) {
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
	if moveDir != MoveDirDstToSrc {
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

	reacheable, err := lkr.LookupNode(mostRecent.Path())
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
		return nil, n.ErrBadNode
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

		aliveSrcNd, err := ma.ghostToAlive(ma.lkrSrc, srcNd)
		if err != nil {
			return err
		}

		if aliveSrcNd == nil {
			// It's a ghost, but it has no living counterpart.
			// This node *might* have been removed on the remote side.
			// Try to see if we have a node at this path, the next step
			// of sync then needs to decide if the node needs to be removed.
			dstNd, err := ma.lkrDst.LookupNode(srcNd.Path())
			if err != nil && !n.IsNoSuchFileError(err) {
				return err
			}

			if dstNd != nil && dstNd.Type() != n.NodeTypeGhost {
				dstModNd, ok := dstNd.(n.ModNode)
				if !ok {
					return n.ErrBadNode
				}

				return ma.fn(MapPair{
					Src:          nil,
					Dst:          dstModNd,
					TypeMismatch: false,
				})
			}

			// Not does not exist on both sides, nothing to report.
			return nil
		}

		// At this point we know that the ghost related to a moved file.
		// Check if we have a file at the same place.
		dstNd, err := ma.lkrDst.LookupNode(aliveSrcNd.Path())
		if err != nil && !n.IsNoSuchFileError(err) {
			return err
		}

		if dstNd != nil && dstNd.Type() != n.NodeTypeGhost {
			// The node already exists in our place. No way we can really merge
			// it cleanly, so just handle the ghost as normal file and potentially
			// apply the normal conflict resolution later on.
			return nil
		}

		dstRefNd, err := ma.lkrDst.LookupNode(srcNd.Path())
		if err != nil && !n.IsNoSuchFileError(err) {
			return err
		}

		if dstRefNd != nil {
			if dstRefNd.Type() == n.NodeTypeGhost {
				aliveOrig, err := ma.ghostToAlive(ma.lkrDst, dstRefNd)
				if err != nil {
					return err
				}

				dstRefNd = aliveOrig
			}
		}

		dstRefModNd, ok := dstRefNd.(n.ModNode)
		if !ok {
			return e.Wrapf(n.ErrBadNode, "dstRefModNd is not a file or directory: %v", dstRefNd)
		}

		switch aliveSrcNd.Type() {
		case n.NodeTypeFile:
			// Mark those both ghosts and original node as visited.
			ma.visited[aliveSrcNd.Path()] = aliveSrcNd
			ma.visited[srcNd.Path()] = srcNd

			if !aliveSrcNd.Hash().Equal(dstRefNd.Hash()) {
				return ma.fn(MapPair{
					Src:          aliveSrcNd,
					Dst:          dstRefModNd,
					TypeMismatch: (dstRefNd.Type() != aliveSrcNd.Type()),
				})
			}

			return nil
		case n.NodeTypeDirectory:
			ma.visited[srcNd.Path()] = srcNd
			if dstRefNd.Type() != n.NodeTypeDirectory {
				return ma.fn(MapPair{
					Src:          aliveSrcNd,
					Dst:          dstRefModNd,
					TypeMismatch: true,
				})
			}

			aliveSrcDir, ok := aliveSrcNd.(*n.Directory)
			if !ok {
				return n.ErrBadNode
			}

			movedSrcDirs = append(movedSrcDirs, ghostDir{
				srcDir:  aliveSrcDir,
				dstPath: dstRefNd.Path(),
			})

			return nil
		default:
			return e.Wrapf(n.ErrBadNode, "Unexpected type in handle ghosts: %v", err)
		}
	})

	if err != nil {
		return err
	}

	// Handle moved paths after handling single files.
	// (mapDirectory assumes that moved files in it were already handled).
	for _, movedSrcDir := range movedSrcDirs {
		if err := ma.mapDirectory(movedSrcDir.srcDir, movedSrcDir.dstPath); err != nil {
			return err
		}
	}

	return nil
}

func NewMapper(lkrSrc, lkrDst *Linker, srcRoot n.Node) *Mapper {
	return &Mapper{
		lkrSrc:  lkrSrc,
		lkrDst:  lkrDst,
		srcRoot: srcRoot,
		visited: make(map[string]n.Node),
	}
}

// Diff calls `fn` for each pairing that was found. Equal files and
// directories are not reported.  Most directories are also not reported, but
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

	switch ma.srcRoot.Type() {
	case n.NodeTypeDirectory:
		dir, ok := ma.srcRoot.(*n.Directory)
		if !ok {
			return n.ErrBadNode
		}

		return ma.mapDirectory(dir, dir.Path())
	case n.NodeTypeFile:
		file, ok := ma.srcRoot.(*n.File)
		if !ok {
			return n.ErrBadNode
		}

		return ma.mapFile(file, file.Path())
	case n.NodeTypeGhost:
		return nil
	default:
		return e.Wrapf(n.ErrBadNode, "Unexpected type in route(): %v", ma.srcRoot)
	}
}
