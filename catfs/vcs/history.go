package vcs

import (
	"fmt"
	"path"
	"strings"

	e "github.com/pkg/errors"
	c "github.com/sahib/brig/catfs/core"
	ie "github.com/sahib/brig/catfs/errors"
	n "github.com/sahib/brig/catfs/nodes"
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

// rule: do not loose content,
//       but we may loose metadata.
//
//   |  a  c  r  m
// ---------------
// a |  n  n  n  y
// c |  n  n  n  y
// r |  y  y  y  y
// m |  y  y  y  y
func (ct ChangeType) IsCompatible(ot ChangeType) bool {
	modifyMask := ChangeTypeAdd | ChangeTypeModify

	if ct&modifyMask != 0 && ot&modifyMask != 0 {
		return false
	}

	return true
}

///////////////////////////

// Change represents a single change of a node between two commits.
type Change struct {
	// Mask is a bitmask of changes that were made.
	// It describes the change that was made between `Next` to `Head`
	// and which is part of `Head`.
	Mask ChangeType

	// Head is the commit that was the current HEAD when this change happened.
	// Note that this is NOT the commit that contains the change, but the commit before.
	Head *n.Commit

	// Next is the commit that comes before `Head`.
	Next *n.Commit

	// Curr is the node with the attributes at a specific state
	Curr n.ModNode

	// ReferToPath is only filled for ghosts that were the source
	// of a move. It's the path of the node it was moved to.
	ReferToPath string
}

func (ns *Change) String() string {
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
	lkr   *c.Linker
	head  *n.Commit
	curr  n.ModNode
	next  n.ModNode
	err   error
	state *Change
}

// NewHistoryWalker will return a new HistoryWalker that will yield changes of
// `node` starting from the state in `cmt` until the root commit if desired.
// Note that it is not checked that `node` is actually part of `cmt`.
func NewHistoryWalker(lkr *c.Linker, cmt *n.Commit, node n.ModNode) *HistoryWalker {
	return &HistoryWalker{
		lkr:  lkr,
		head: cmt,
		curr: node,
	}
}

// maskFromState figures out the change mask based on the current state
func (hw *HistoryWalker) maskFromState(curr, next n.ModNode) ChangeType {
	mask := ChangeType(0)

	// Initial state; no succesor known yet to compare too.
	if next == nil {
		return mask
	}

	isGhostCurr := curr.Type() == n.NodeTypeGhost
	isGhostNext := next.Type() == n.NodeTypeGhost

	currHash, err := n.ContentHash(curr)
	if err != nil {
		return ChangeTypeNone
	}

	nextHash, err := n.ContentHash(next)
	if err != nil {
		return ChangeTypeNone
	}

	// If the hash differs, there's likely a modification going on.
	if !currHash.Equal(nextHash) {
		mask |= ChangeTypeModify
	}

	if next.Path() != curr.Path() {
		mask |= ChangeTypeMove
	} else {
		// If paths did not move, but the current node is a ghost,
		// then it means that the node was removed in this commit.
		if isGhostCurr && !isGhostNext {
			mask |= ChangeTypeRemove
		}

		if !isGhostCurr && isGhostNext {
			mask |= ChangeTypeAdd
		}
	}

	return mask
}

func ParentDirectoryForCommit(lkr *c.Linker, cmt *n.Commit, curr n.Node) (*n.Directory, error) {
	nextDirPath := path.Dir(curr.Path())
	if nextDirPath == "/" {
		return nil, nil
	}

	root, err := lkr.DirectoryByHash(cmt.Root())
	if err != nil {
		return nil, err
	}

	nd, err := root.Lookup(lkr, nextDirPath)
	if err != nil {
		return nil, err
	}

	dir, ok := nd.(*n.Directory)
	if !ok {
		return nil, ie.ErrBadNode
	}

	return dir, nil
}

// Check if a node was moved and if so, return the coressponding other half.
// If it was not moved, this method will return nil, MoveDirNone, nil.
//
// This is also supposed to work with moved directories (keep in mind that moving directories
// will only create a ghost for the moved directory itself, not the children of it):
//
// $ tree .
// a/
//  b/
//   c  # a file.
// $ mv a f
//
// For this case we need to go over the parent directories of c (b and f) to find the ghost dir "a".
// From there we can resolve back to "c".
func findMovePartner(lkr *c.Linker, head *n.Commit, curr n.Node) (n.Node, c.MoveDir, error) {
	prev, direction, err := lkr.MoveMapping(head, curr)
	if err != nil {
		return nil, c.MoveDirNone, err
	}

	if prev != nil {
		return prev, direction, nil
	}

	childPath := []string{curr.Name()}

	for {
		parentDir, err := ParentDirectoryForCommit(lkr, head, curr)
		if err != nil {
			return nil, c.MoveDirNone, e.Wrap(err, "bad parent dir")
		}

		if parentDir == nil {
			return nil, c.MoveDirNone, nil
		}

		prevDirNd, direction, err := lkr.MoveMapping(head, parentDir)
		if err != nil {
			return nil, c.MoveDirNone, nil
		}

		// Advance for next round:
		curr = parentDir

		if prevDirNd == nil {
			// This was not moved; remember step for final lookup:
			childPath = append([]string{parentDir.Name()}, childPath...)
			continue
		}

		// At this point we know that the dir `parentDir` was moved.
		// Now we have to find the old version of the node.
		// This for loop will end now anyways.

		var prevDir *n.Directory

		switch prevDirNd.Type() {
		case n.NodeTypeDirectory:
			// This case will probably not happen not very often.
			// Most of the time the old node in a mapping is a ghost.
			var ok bool
			prevDir, ok = prevDirNd.(*n.Directory)
			if !ok {
				return nil, c.MoveDirNone, ie.ErrBadNode
			}
		case n.NodeTypeGhost:
			// If it's a ghost we need to unpack it.
			prevDirGhost, ok := prevDirNd.(*n.Ghost)
			if !ok {
				return nil, c.MoveDirNone, e.Wrap(
					ie.ErrBadNode,
					"bad previous dir",
				)
			}

			prevDir, err = prevDirGhost.OldDirectory()
			if err != nil {
				return nil, c.MoveDirNone, e.Wrap(err, "bad old directory")
			}
		default:
			return nil, c.MoveDirNone, fmt.Errorf("unexpected file node")
		}

		// By the current logic, the path is still reachable in
		// the directory the same way before.
		child, err := prevDir.Lookup(lkr, strings.Join(childPath, "/"))
		if err != nil {
			return nil, c.MoveDirNone, err
		}

		return child, direction, nil
	}

	return nil, c.MoveDirNone, fmt.Errorf("How did we end up here?")
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

	// Check if this node participated in a move:
	prev, direction, err := findMovePartner(hw.lkr, hw.head, hw.curr)
	if err != nil {
		hw.err = err
		return false
	}

	// Unpack the old ghost before doing anything with it:
	referToPath := ""
	if prev != nil {
		referToPath = prev.Path()

		if prev.Type() == n.NodeTypeGhost {
			prevGhost, ok := prev.(*n.Ghost)
			if !ok {
				hw.err = ie.ErrBadNode
				return false
			}

			prev = prevGhost.OldNode()
		}
	}

	// Advance to the previous commit:
	prevHead, err := hw.head.Parent(hw.lkr)
	if err != nil {
		hw.err = err
		return false
	}

	// We ran out of commits to check.
	if prevHead == nil {
		hw.state = &Change{
			Head: hw.head,
			Mask: ChangeTypeAdd,
			Curr: hw.curr,
			Next: nil,
		}
		hw.head = nil
		return true
	}

	prevHeadCommit, ok := prevHead.(*n.Commit)
	if !ok {
		hw.err = e.Wrap(ie.ErrBadNode, "history: bad commit")
		return false
	}

	// Assumption here: The move mapping should only store one move per commit.
	// i.e: for move(a, b); a and b should always be in different commits.
	// This is enforced by the logic in MakeCommit()
	if prev == nil || direction != c.MoveDirSrcToDst {
		// No valid move mapping found, node was probably not moved.
		// Assume that we can reach it directly via it's path.
		currRoot, err := hw.lkr.DirectoryByHash(prevHeadCommit.Root())
		if err != nil {
			hw.err = e.Wrap(err, "Cannot find previous root directory")
			return false
		}

		prev, err = currRoot.Lookup(hw.lkr, hw.curr.Path())
		if ie.IsNoSuchFileError(err) {
			// The file did not exist in the previous commit (no ghost!)
			// It must have been added in this commit.
			hw.state = &Change{
				Head: hw.head,
				Mask: ChangeTypeAdd,
				Curr: hw.curr,
				Next: prevHeadCommit,
			}
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
		hw.err = e.Wrap(ie.ErrBadNode, "history: bad mod node")
		return false
	}

	// Pack up the current state:
	hw.state = &Change{
		Head: hw.head,
		Mask: hw.maskFromState(hw.curr, prevModNode),
		Curr: hw.curr,
		Next: prevHeadCommit,
	}

	// Special case: A ghost that still has a move partner.
	// This means the node here was moved to `prev` in this commit.
	if hw.curr.Type() == n.NodeTypeGhost && direction == c.MoveDirDstToSrc {
		// Indicate that this node was indeed removed,
		// but still lives somewhere else.
		hw.state.Mask |= ChangeTypeMove
		hw.state.ReferToPath = referToPath
	}

	// Swap for the next call to Next():
	hw.curr, hw.next = prevModNode, hw.curr
	hw.head = prevHeadCommit
	return true
}

// State returns the current change state.
// Note that the change may have ChangeTypeNone as Mask if nothing changed.
// If you only want states where it actually changed, just filter those.
func (hw *HistoryWalker) State() *Change {
	return hw.state
}

// Err returns the last happened error or nil if none.
func (hw *HistoryWalker) Err() error {
	return hw.err
}

// History returns a list of `nd`'s states starting with the commit in `start`
// and stopping at `stop`. `stop` can be nil; in this case all commits will be
// iterated. The returned list has the most recent change upfront, and the
// latest change as last element.
func History(lkr *c.Linker, nd n.ModNode, start, stop *n.Commit) ([]*Change, error) {
	states := make([]*Change, 0)
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
