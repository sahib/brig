package vcs

import (
	"fmt"
	"strings"

	e "github.com/pkg/errors"
	c "github.com/sahib/brig/catfs/core"
	ie "github.com/sahib/brig/catfs/errors"
	n "github.com/sahib/brig/catfs/nodes"
	capnp_patch "github.com/sahib/brig/catfs/vcs/capnp"
	capnp "zombiezen.com/go/capnproto2"
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
	return ct&modifyMask == 0 || ot&modifyMask == 0
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

func (ch *Change) String() string {
	return fmt.Sprintf("<%s:%s>", ch.Curr.Path(), ch.Mask)
}

func (ch *Change) Replay(lkr *c.Linker) error {
	// TODO: Implement.
	// 	if ch.Mask&ChangeTypeMove != 0 {
	// 		c.Move(lkr, ch.Curr, ch.ReferToPath)
	// 	}
	//
	// 	if ch.Mask&ChangeTypeAdd != 0 || ch.Mask&ChangeTypeModify {
	// 		return lkr.StageNode(ch.Curr)
	// 	}
	//
	return nil
}

func (ch *Change) toCapnpChange(capCh *capnp_patch.Change) error {
	currData, err := n.MarshalNode(ch.Curr)
	if err != nil {
		return err
	}

	headData, err := n.MarshalNode(ch.Head)
	if err != nil {
		return err
	}

	nextData, err := n.MarshalNode(ch.Next)
	if err != nil {
		return err
	}

	if err := capCh.SetCurr(currData); err != nil {
		return err
	}

	if err := capCh.SetHead(headData); err != nil {
		return err
	}

	if err := capCh.SetNext(nextData); err != nil {
		return err
	}

	if err := capCh.SetReferToPath(ch.ReferToPath); err != nil {
		return err
	}

	capCh.SetMask(uint64(ch.Mask))
	return nil

}

func (ch *Change) ToCapnp() (*capnp.Message, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, err
	}

	capCh, err := capnp_patch.NewRootChange(seg)
	if err != nil {
		return nil, err
	}

	if err := ch.toCapnpChange(&capCh); err != nil {
		return nil, err
	}

	return msg, nil
}

func (ch *Change) fromCapnpChange(capCh capnp_patch.Change) error {
	currData, err := capCh.Curr()
	if err != nil {
		return err
	}

	headData, err := capCh.Head()
	if err != nil {
		return err
	}

	nextData, err := capCh.Next()
	if err != nil {
		return err
	}

	curr, err := n.UnmarshalNode(currData)
	if err != nil {
		return err
	}

	head, err := n.UnmarshalNode(headData)
	if err != nil {
		return err
	}

	next, err := n.UnmarshalNode(nextData)
	if err != nil {
		return err
	}

	referToPath, err := capCh.ReferToPath()
	if err != nil {
		return err
	}

	var ok bool
	ch.Curr, ok = curr.(n.ModNode)
	if !ok {
		return e.Wrapf(ie.ErrBadNode, "change: from-capnp: curr")
	}

	ch.Head, ok = head.(*n.Commit)
	if !ok {
		return e.Wrapf(ie.ErrBadNode, "change: from-capnp: head")
	}

	ch.Next, ok = next.(*n.Commit)
	if !ok {
		return e.Wrapf(ie.ErrBadNode, "change: from-capnp: next")
	}

	ch.ReferToPath = referToPath
	ch.Mask = ChangeType(capCh.Mask())
	return nil
}

func (ch *Change) FromCapnp(msg *capnp.Message) error {
	capCh, err := capnp_patch.ReadRootChange(msg)
	if err != nil {
		return err
	}

	return ch.fromCapnpChange(capCh)
}

func CombineChanges(changes []*Change) *Change {
	if len(changes) == 0 {
		return nil
	}

	ch := &Change{
		Mask:        ChangeType(0),
		Head:        changes[0].Head,
		Next:        changes[0].Next,
		Curr:        changes[0].Curr,
		ReferToPath: changes[0].ReferToPath,
	}

	for _, change := range changes {
		ch.Mask |= change.Mask
	}

	return ch
}
