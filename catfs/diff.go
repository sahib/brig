package catfs

import (
	"fmt"
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
	// TODO:a thing this through.
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
	lkr   *Linker
	head  *n.Commit
	curr  n.ModNode
	next  n.ModNode
	err   error
	state *NodeState
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
			hw.state.Mask = ChangeTypeAdd
			hw.head = nil
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

type DiffPair struct {
	Src        n.ModNode
	Dst        n.ModNode
	IsConflict bool
}

type Differ struct {
	lkrSrc, lkrDst *Linker
	srcRoot        n.Node
	fn             func(pair DiffPair) error
	visited        map[string]n.ModNode
}

func NewDiffer(lkrSrc, lkrDst *Linker, srcRoot n.Node, fn func(pair DiffPair) error) (*Differ, error) {
	if srcRoot == nil {
		srcRootDir, err := lkrSrc.Root()
		if err != nil {
			return nil, err
		}

		srcRoot = srcRootDir
	}

	return &Differ{
		lkrSrc:  lkrSrc,
		lkrDst:  lkrDst,
		srcRoot: srcRoot,
	}, nil
}

func (df *Differ) route(nd n.Node) error {
	switch nd.Type() {
	case n.NodeTypeDirectory:
		dir, ok := nd.(*n.Directory)
		if !ok {
			return n.ErrBadNode
		}

		return df.diffDirectory(dir)
	case n.NodeTypeFile:
		file, ok := nd.(*n.File)
		if !ok {
			return n.ErrBadNode
		}

		return df.diffFile(file)
	case n.NodeTypeGhost:
		ghost, ok := nd.(*n.Ghost)
		if !ok {
			return n.ErrBadNode
		}

		return df.diffGhost(ghost)
	default:
		return e.Wrapf(n.ErrBadNode, "Unexpected type in route(): %v", nd)
	}
}

func (df *Differ) report(pair DiffPair) error {
	if _, ok := df.visited[pair.Src.Path()]; ok {
		return nil
	}

	return df.fn(pair)
}

func (df *Differ) diffGhost(srcCurr *n.Ghost) error {
	// We encountered a ghost on src's side. We need to figure out.
	// if that means that the node there was removed or moved elsewhere.
	var err error
	var mapSrcNd n.Node = nil // XXX
	if err != nil {
		return err
	}

	if mapSrcNd == nil {
		// It was a removal on src's side.
		return nil
	}

	return df.route(mapSrcNd)
}

func (df *Differ) diffFile(srcCurr *n.File) error {
	dstCurr, err := df.lkrDst.LookupNode(srcCurr.Path())
	if err != nil && !n.IsNoSuchFileError(err) {
		return err
	}

	if dstCurr == nil {
		// We do not have this node yet, mark it for copying.
		return df.report(DiffPair{
			Src:        srcCurr,
			Dst:        nil,
			IsConflict: false,
		})
	}

	if dstCurr.Hash().Equal(srcCurr.Hash()) {
		fmt.Printf("File remote:%s - local:%s do not differ.", srcCurr.Path(), dstCurr.Path())
		return nil
	}

	switch typ := dstCurr.Type(); typ {
	case n.NodeTypeDirectory:
		// Our node seems to be a directory and theirs a file.
		// That's not something we can fix.
		dstDir, ok := dstCurr.(*n.File)
		if !ok {
			return n.ErrBadNode
		}

		return df.report(DiffPair{
			Src:        srcCurr,
			Dst:        dstDir,
			IsConflict: true,
		})
	case n.NodeTypeFile:
		// We have two competing files. Let's figure out if the changes done to
		// them are compatible.
		dstFile, ok := dstCurr.(*n.File)
		if !ok {
			return n.ErrBadNode
		}

		return df.report(DiffPair{
			Src:        srcCurr,
			Dst:        dstFile,
			IsConflict: false,
		})
	case n.NodeTypeGhost:
		// Probably was moved or removed on our side.
		// TODO: Check which case happened and try to resolve to
		// mapDstNd, err := df.findMapNode(dstCurr)
		var mapDstNd n.Node = nil
		if err != nil {
			return err
		}

		if mapDstNd == nil {
			// Was removed on our side...
			// Nothing to do really.
			return nil
		}

		mapModNd, ok := mapDstNd.(n.ModNode)
		if !ok {
			return e.Wrap(n.ErrBadNode, "XXX: Hmm... still a bug?")
		}

		isConflict := !(mapModNd.Type() == n.NodeTypeFile)
		return df.report(DiffPair{
			Src:        srcCurr,
			Dst:        mapModNd,
			IsConflict: isConflict,
		})
	default:
		return e.Wrapf(n.ErrBadNode, "Unexpected node type in syncFile: %v", typ)
	}

	return nil
}

func (df *Differ) diffDirectory(srcCurr *n.Directory) error {
	// Possible cases here:
	// - lkrDst does not have this path (need to merge it)
	// - lkrDst has this path, but it's a ghost (need to merge with moved file, if any)
	// - lkrDst has this path, but it is not a directory (need to handle conflict?)
	// - lkrDst has this path, and it's a directory (attempt conflict resolution)
	//
	// It is guaranteed that srcCurr is *always* a directory.
	dstCurr, err := df.lkrDst.LookupDirectory(srcCurr.Path())
	if err != nil && !n.IsNoSuchFileError(err) {
		return err
	}

	if dstCurr == nil {
		// We never heard of this directory apparently. Go sync it.
		fmt.Printf("Marked remote directory `%s` for syncing.\n", srcCurr.Path())
		return nil
	}

	// Check if we're lucky and the directory hash is equal:
	if srcCurr.Hash().Equal(dstCurr.Hash()) {
		fmt.Printf(
			"src (%s) and dest (%s) are equal; no sync needed",
			srcCurr.Path(),
			dstCurr.Path(),
		)
		return nil
	}

	// Check if it's an empty directory.
	// If so, we should check if we should adopt it.
	if srcCurr.NChildren(df.lkrSrc) == 0 {
		return df.report(DiffPair{
			Src:        srcCurr,
			Dst:        nil,
			IsConflict: false,
		})
	}

	// Both sides have this directory, but the content differs.
	// We need to figure out recursively what exactly is different.
	return srcCurr.VisitChildren(df.lkrSrc, func(srcNd n.Node) error {
		switch srcNd.Type() {
		case n.NodeTypeDirectory:
			srcChildDir, ok := srcNd.(*n.Directory)
			if !ok {
				return n.ErrBadNode
			}

			return df.diffDirectory(srcChildDir)
		case n.NodeTypeFile:
			srcChildFile, ok := srcNd.(*n.File)
			if !ok {
				return n.ErrBadNode
			}

			return df.diffFile(srcChildFile)
		case n.NodeTypeGhost:
			// TODO: mapDstNo, err := some clever mapping.
			var mapSrcNd n.Node = nil
			if err != nil {
				return err
			}

			if mapSrcNd == nil {
				// Was removed on our side...
				// Nothing to do really.
				return nil
			}

			mapSrcModNd, ok := mapSrcNd.(n.ModNode)
			if !ok {
				return e.Wrap(n.ErrBadNode, "XXX2: Hmm... still a bug?")
			}

			df.visited[mapSrcModNd.Path()] = mapSrcModNd

			if mapSrcDir, ok := mapSrcModNd.(*n.Directory); ok {
				return df.diffDirectory(mapSrcDir)
			}

			// srcCurr is a directory and mapSrcNd or some reason a file.
			// This cannot really work out.
			return df.report(DiffPair{
				Src:        srcCurr,
				Dst:        mapSrcModNd,
				IsConflict: true,
			})
		default:
			return n.ErrBadNode
		}

	})
}

func (df *Differ) Diff() error {
	return df.route(df.srcRoot)
}
