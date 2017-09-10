package catfs

// This package implements brig's sync which I called,
// in a burst of modesty, "bright".
// (Not because it's very bright, but it starts with brig...)
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
// This makes the diff and sync implemenation share most of it's code.

import (
	"errors"
	"fmt"

	n "github.com/disorganizer/brig/catfs/nodes"
)

const (
	ConflictStragetyUnknown = iota
	ConflictStragetyIgnore
	ConflictStragetyMarker
)

type ConflictStragey int

type SyncConfig struct {
	ConflictStragey  ConflictStragey
	PropagateDeletes bool
}

var (
	ErrConflict       = errors.New("Conflicting changes")
	DefaultSyncConfig = &SyncConfig{}
)

type Syncer struct {
	cfg    *SyncConfig
	lkrSrc *Linker
	lkrDst *Linker

	// Mapping from src to our own nodes.
	mapping map[*n.Node]*n.Node

	// cached attributes:
	dstMergeCmt *n.Commit
	srcMergeCmt *n.Commit
}

func NewSyncer(lkrSrc, lkrDst *Linker, cfg *SyncConfig) *Syncer {
	if cfg == nil {
		cfg = DefaultSyncConfig
	}

	return &Syncer{
		cfg:    cfg,
		lkrSrc: lkrSrc,
		lkrDst: lkrDst,
	}
}

func (sy *Syncer) cacheLastCommonMerge() error {
	srcOwner, err := sy.lkrSrc.Owner()
	if err != nil {
		return err
	}

	dstHead, err := sy.lkrDst.Head()
	if err != nil {
		return err
	}

	for {
		with, srcRef := dstHead.MergeMarker()
		if with != nil && with.Equal(srcOwner) {
			srcHead, err := sy.lkrSrc.CommitByHash(srcRef)
			if err != nil {
				return err
			}

			sy.dstMergeCmt = dstHead
			sy.srcMergeCmt = srcHead
		}

		prevHeadNode, err := dstHead.Parent(sy.lkrDst)
		if err != nil {
			return err
		}

		if prevHeadNode == nil {
			break
		}

		newDstHead, ok := prevHeadNode.(*n.Commit)
		if !ok {
			return n.ErrBadNode
		}

		dstHead = newDstHead
	}

	return nil
}

func (sy *Syncer) merge(src, dst *n.File) error {
	dstParent, err := n.ParentDirectory(sy.lkrDst, dst)
	if err != nil {
		return err
	}

	if err := dstParent.RemoveChild(sy.lkrSrc, dst); err != nil {
		return err
	}

	return nil
}

// resolve is always called when two nodes on both sides and they do not have the same hash.
// In the best case, both have compatible changes and can be merged, otherwise a user
// defined conflict strategy has to be applied.
func (sy *Syncer) resolve(src, dst *n.File) error {
	srcHead, err := sy.lkrSrc.Head()
	if err != nil {
		return err
	}

	dstHead, err := sy.lkrDst.Head()
	if err != nil {
		return err
	}

	srcHist, err := History(sy.lkrSrc, src, srcHead, sy.srcMergeCmt)
	if err != nil {
		return err
	}

	dstHist, err := History(sy.lkrDst, dst, dstHead, sy.dstMergeCmt)
	if err != nil {
		return err
	}

	var srcMask, dstMask ChangeType
	srcRoot := len(srcHist)
	dstRoot := len(srcHist)

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

	srcChanges := srcHist[:srcRoot]
	dstChanges := dstHist[:dstRoot]

	// Handle a few lucky cases:
	if len(srcChanges) > 0 && len(dstChanges) == 0 {
		// We can "fast forward" our node. There are only remote changes for this file.
		fmt.Println("fast forward")
		return nil

	}
	if len(srcChanges) == 0 && len(dstChanges) > 0 {
		// Only our side has changes. We can consider this node as merged.
		fmt.Println("fast ignore")
		return nil
	}
	if len(srcChanges) == 0 && len(dstChanges) == 0 {
		// This should not happen:
		// Both sides have no changes and still the hash is different...
		fmt.Println("BUG: both sides have no changes...")
		return nil
	}

	// Both sides have changes. Now we need to figure out if they are compatible.
	// We do this simply by OR-ing all changes on both side to an individual mask
	// and check if those can be applied on top of dst's current state.
	// TODO: Define this really.
	if !dstMask.IsCompatible(srcMask) {
		// The changes are not compatible.
		// We need to apply a conflict resolution strategy.
		fmt.Println("Incompatible changes")
		return ErrConflict
	}

	if srcMask&ChangeTypeMove != 0 && dst.Path() != src.Path() {
		fmt.Println("NOTE: File has moved...")
	}

	// No conflict. We can merge src and dst.
	return nil
}

// func (sy *Syncer) mapDirectory(srcCurr *n.Directory) error {
// 	// Possible cases here:
// 	// - lkrDst does not have this path (need to merge it)
// 	// - lkrDst has this path, but it's a ghost (need to merge with moved file, if any)
// 	// - lkrDst has this path, but it is not a directory (need to handle conflict?)
// 	// - lkrDst has this path, and it's a directory (attempt conflict resolution)
// 	//
// 	// It is guaranteed that srcCurr is *always* a directory.
//
// 	dstCurr, err := sy.lkrDst.LookupDirectory(srcCurr.Path())
// 	if err != nil && !n.IsNoSuchFileError(err) {
// 		return err
// 	}
//
// 	if dstCurr == nil {
// 		// We never heard of this directory apparently. Go sync it.
// 		fmt.Printf("Marked remote directory `%s` for syncing.\n", srcCurr.Path())
// 		return nil
// 	}
//
// 	// Check if we're lucky and the directory hash is equal:
// 	if srcCurr.Hash().Equal(dstCurr.Hash()) {
// 		fmt.Printf(
// 			"src (%s) and dest (%s) are equal; no sync needed",
// 			srcCurr.Path(),
// 			dstCurr.Path(),
// 		)
// 		return nil
// 	}
//
// 	// Check if it's an empty directory.
// 	// If so, we should check if we should adopt it.
// 	if srcCurr.NChildren(sy.lkrSrc) == 0 {
// 		sy.mapping[srcCurr] = dstCurr
// 	}
//
// 	// Both sides have this directory, but the content differs.
// 	// We need to figure out recursively what exactly is different.
// 	return srcCurr.VisitChildren(sy.lkrSrc, func(srcNd n.Node) error {
// 		switch srcNd.Type() {
// 		case n.NodeTypeDirectory:
// 			srcChildDir, ok := srcNd.(*n.Directory)
// 			if !ok {
// 				return n.ErrBadNode
// 			}
//
// 			return sy.mapDirectory(srcChildDir)
// 		case n.NodeTypeFile:
// 			srcChildFile, ok := srcNd.(*n.File)
// 			if !ok {
// 				return n.ErrBadNode
// 			}
//
// 			return sy.mapFile(srcChildFile)
// 		case n.NodeTypeGhost:
// 			// TODO: Probably means that this node was removed on the remote side
// 			//       and we need to decide if we should propagate the remove.
// 			//       Find out if it's removed MoveMapping
// 			mapDstNd, err := sy.findMoveMap(srcNd)
// 			if err != nil {
// 				return err
// 			}
//
// 			if mappedNode == nil {
// 				// It was a removal on src's side.
// 				sy.mapping[srcNd] = nil
// 				return nil
// 			}
//
// 			fmt.Printf("Ignoring ghost for now: %s\n", srcNd.Path())
// 			return nil
// 		}
//
// 		return n.ErrBadNode
// 	})
// }
//
// // Sync will apply brig's sync algorithm on lkrDst, getting it's data from lkrSrc.
// // lkrSrc will not be modified in this process and will not be modified.
// // Running Sync() will create a new merge commit, that takes note that we merged
// // with lkrSrc at this point to avoid that conflicting files will not be processed again.
func (sy *Syncer) Sync() error {
	return nil
	// 	srcRoot, err := sy.lkrSrc.Root()
	// 	if err != nil {
	// 		return err
	// 	}
	//
	// 	if err := sy.cacheLastCommonMerge(); err != nil {
	// 		return e.Wrapf(err, "Error while finding last common merge")
	// 	}
	//
	// 	if err := sy.mapDirectory(srcRoot); err != nil {
	// 		return e.Wrap(err, "sync failed")
	// 	}
	//
	// 	srcHead, err := sy.lkrSrc.Head()
	// 	if err != nil {
	// 		return err
	// 	}
	//
	// 	srcOwner, err := sy.lkrSrc.Owner()
	// 	if err != nil {
	// 		return err
	// 	}
	//
	// 	if len(sy.mapping) > 0 {
	// 		// If something was changed, remember that we merged with src.
	// 		// This avoids merging conflicting files a second time in the next resolve().
	// 		if err := sy.lkrDst.SetMergeMarker(srcOwner, srcHead.Hash()); err != nil {
	// 			return err
	// 		}
	//
	// 		message := fmt.Sprintf("Merge with %s", srcOwner.ID())
	// 		return sy.lkrDst.MakeCommit(srcOwner, message)
	// 	}
	//
	// 	return nil
}
