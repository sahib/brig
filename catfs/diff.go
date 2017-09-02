package catfs

import (
	"strings"

	n "github.com/disorganizer/brig/cafs/nodes"
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

// NodeState represents a single change of a node between two commits.
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
