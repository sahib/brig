package vcs

// This package implements brig's sync algorithm which I called, in a burst of
// modesty, "bright". (Not because it's or I'm very bright, but because it
// starts with brig...)
//
// The sync algorithm tries to handle the following special cases:
// - Propagate moves (most of them, at least)
// - Propagate deletes (configurable?)
//
// Terminology:
// - Destination (short "dst") is used to reference our own storage.
// - Source (short: "src") is used to reference the remote storage.
//
// The sync algorithm can be roughly divided in 4 stages:
// - Stage 1: "Move Marking":
//   Iterate over all ghosts in the tree and check if they were either moved
//   (has sibling) or removed (has no sibling). In case of directories, the
//   second mapping stage is already executed.
//
// - Stage 2: "Mapping":
//   Finding pairs of files that possibly adding, merging or conflict handling.
//   Equal files will already be sorted out at this point. Every already
//   visited node in the remote linker will be marked. The mapping algorithm
//   starts at the root node and uses the attributes of the merkle trees
//   (same hash = same content) to skip over same parts.
//
// - Stage 3: "Resolving":
//   For each file a decision needs to be made. This decison defines the next step
//   and can be one of the following.
//
//   - The file was added on the remote, we should add it to -> Add them.
//   - The file has compatible changes on the both sides. -> Merge them.
//   - The file was incompatible changes on both sides -> Do conflict resolution.
//
//   This the part where most configuration can be done.
//
// - Stage 4: "Handling"
//   TODO: Define exactly.
//
// Everything except Stage 4 is read-only. If a user wants to only show the diff
// between two linkers, he just prints what would be done instead of actually doing it.
// This makes the diff and sync implementation share most of it's code.

import (
	"fmt"

	c "github.com/disorganizer/brig/catfs/core"
	ie "github.com/disorganizer/brig/catfs/errors"
	n "github.com/disorganizer/brig/catfs/nodes"
	"github.com/disorganizer/brig/util"
	e "github.com/pkg/errors"
)

// executor is the interface that executes the actual action
// needed to perform the sync (see "phase 4" on top of this file)
type executor interface {
	handleAdd(src n.ModNode) error
	handleRemove(dst n.ModNode) error
	handleTypeConflict(src, dst n.ModNode) error
	handleMerge(src, dst n.ModNode, srcMask, dstMask ChangeType) error
	handleConflict(src, dst n.ModNode, srcMask, dstMask ChangeType) error
}

//////////////////////////////////////////////
// IMPLEMENTATION OF ACTUAL DECISION MAKING //
//////////////////////////////////////////////

type resolver struct {
	lkrSrc *c.Linker
	lkrDst *c.Linker

	// What points should be resolved
	dstHead *n.Commit
	srcHead *n.Commit

	// cached attributes:
	dstMergeCmt *n.Commit
	srcMergeCmt *n.Commit

	// actual executor based on the decision
	exec executor
}

func newResolver(lkrSrc, lkrDst *c.Linker, srcHead, dstHead *n.Commit, exec executor) (*resolver, error) {
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

	return &resolver{
		lkrSrc:  lkrSrc,
		lkrDst:  lkrDst,
		srcHead: srcHead,
		dstHead: dstHead,
		exec:    exec,
	}, nil
}

func (rv *resolver) resolve() error {
	srcRoot, err := rv.lkrSrc.Root()
	if err != nil {
		return err
	}

	if err := rv.cacheLastCommonMerge(); err != nil {
		return e.Wrapf(err, "Error while finding last common merge")
	}

	mapper, err := NewMapper(rv.lkrSrc, rv.lkrDst, rv.srcHead, rv.dstHead, srcRoot)
	if err != nil {
		return err
	}

	mappings := []MapPair{}

	err = mapper.Map(func(pair MapPair) error {
		mappings = append(mappings, pair)
		return nil
	})

	if err != nil {
		return err
	}

	for _, pair := range mappings {
		if err := rv.decide(pair); err != nil {
			return err
		}
	}

	return nil
}

func (rv *resolver) cacheLastCommonMerge() error {
	srcOwner, err := rv.lkrSrc.Owner()
	if err != nil {
		return err
	}

	currHead := rv.dstHead

	for currHead != nil {
		with, srcRef := currHead.MergeMarker()
		if with == srcOwner {
			srcHead, err := rv.lkrSrc.CommitByHash(srcRef)
			if err != nil {
				return err
			}

			rv.dstMergeCmt = currHead
			rv.srcMergeCmt = srcHead
		}

		prevHeadNode, err := currHead.Parent(rv.lkrDst)
		if err != nil {
			return err
		}

		if prevHeadNode == nil {
			break
		}

		newDstHead, ok := prevHeadNode.(*n.Commit)
		if !ok {
			return ie.ErrBadNode
		}

		currHead = newDstHead
	}

	return nil
}

// hasConflicts is always called when two nodes on both sides and they do not
// have the same hash.  In the best case, both have compatible changes and can
// be merged, otherwise a user defined conflict strategy has to be applied.
func (rv *resolver) hasConflicts(src, dst n.ModNode) (bool, ChangeType, ChangeType, error) {
	srcHist, err := History(rv.lkrSrc, src, rv.srcHead, rv.srcMergeCmt)
	if err != nil {
		return false, 0, 0, err
	}

	dstHist, err := History(rv.lkrDst, dst, rv.dstHead, rv.dstMergeCmt)
	if err != nil {
		return false, 0, 0, err
	}

	var srcMask, dstMask ChangeType
	srcRoot := len(srcHist)
	dstRoot := len(dstHist)

	for srcIdx, srcChange := range srcHist {
		for dstIdx, dstChange := range dstHist {
			srcMask |= srcChange.Mask
			dstMask |= dstChange.Mask

			if srcChange.Curr.Hash().Equal(dstChange.Curr.Hash()) {
				srcRoot = srcIdx + 1
				dstRoot = dstIdx + 1
			}
		}
	}

	// Make sure that enough commits are on both sides
	// (or assume that
	srcChanges := srcHist[:util.Clamp(srcRoot, 0, len(srcHist)-1)]
	dstChanges := dstHist[:util.Clamp(dstRoot, 0, len(dstHist)-1)]

	// Handle a few lucky cases:
	if len(srcChanges) > 0 && len(dstChanges) == 0 {
		// We can "fast forward" our node.
		// There are only remote changes for this file.
		fmt.Println("fast forward", src.Path(), dst.Path())
		return false, 0, 0, nil

	}
	if len(srcChanges) == 0 && len(dstChanges) > 0 {
		// Only our side has changes. We can consider this node as merged.
		fmt.Println("fast ignore")
		return false, 0, 0, nil
	}
	if len(srcChanges) == 0 && len(dstChanges) == 0 {
		// This should not happen:
		// Both sides have no changes and still the hash is different...
		fmt.Println("BUG: both sides have no changes...")
		return false, 0, 0, nil
	}

	// Both sides have changes. Now we need to figure out if they are compatible.
	// We do this simply by OR-ing all changes on both side to an individual mask
	// and check if those can be applied on top of dst's current state.
	// TODO: Define this really.
	if !dstMask.IsCompatible(srcMask) {
		// The changes are not compatible.
		// We need to apply a conflict resolution strategy.
		fmt.Println("Incompatible changes", srcChanges, dstChanges)
		return true, srcMask, dstMask, nil
	}

	if srcMask&ChangeTypeMove != 0 && dst.Path() != src.Path() {
		fmt.Println("NOTE: File has moved...")
	}

	fmt.Println("no conflicts", srcChanges, dstChanges)
	// No conflict. We can merge src and dst.
	return false, srcMask, dstMask, nil
}

func (rv *resolver) decide(pair MapPair) error {
	if pair.Src == nil && pair.Dst == nil {
		return fmt.Errorf("Received completely empty mapping; ignoring")
	}

	if pair.Src == nil {
		return rv.exec.handleRemove(pair.Dst)
	}

	if pair.Dst == nil {
		return rv.exec.handleAdd(pair.Src)
	}

	if pair.TypeMismatch {
		fmt.Printf(
			"%s is a %s and %s a %s; ignoring",
			pair.Src.Path(), pair.Src.Type(),
			pair.Dst.Path(), pair.Dst.Type(),
		)
		return rv.exec.handleTypeConflict(pair.Src, pair.Dst)
	}

	hasConflicts, srcMask, dstMask, err := rv.hasConflicts(pair.Src, pair.Dst)
	if err != nil {
		return err
	}

	if hasConflicts {
		return rv.exec.handleConflict(pair.Src, pair.Dst, srcMask, dstMask)
	}

	// handleMerge needs the masks to decide what path / content to choose.
	return rv.exec.handleMerge(pair.Src, pair.Dst, srcMask, dstMask)
}
