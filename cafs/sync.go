package cafs

// The sync algorithm tries to handle the following special cases:
// - Propagate moves (most of them, at least)
// - Propagate deletes (configurable?)

import (
	"errors"
	"fmt"

	n "github.com/disorganizer/brig/cafs/nodes"
	e "github.com/pkg/errors"
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
	cfg *SyncConfig
}

func NewSyncer(cfg *SyncConfig) *Syncer {
	if cfg == nil {
		cfg = DefaultSyncConfig
	}

	return &Syncer{
		cfg: cfg,
	}
}

func (syn *Syncer) syncFile(lkrSrc, lkrDst *Linker, srcCurr *n.File) error {
	dstCurr, err := lkrDst.LookupNode(srcCurr.Path())
	if err != nil && !n.IsNoSuchFileError(err) {
		return err
	}

	if dstCurr == nil {
		// We do not have this node yet, mark it for copying.
		fmt.Printf("Syncing remote file `%s`\n", srcCurr.Path())
		return nil
	}

	switch typ := dstCurr.Type(); typ {
	case n.NodeTypeFile:
		// We have two competing files. Let's figure out if the changes done to
		// them are compatible.
		fmt.Printf("Competing nodes: %s <-> %s\n", srcCurr.Path(), dstCurr.Path())
	case n.NodeTypeGhost:
		// Probably was moved or removed on the other side.
		fmt.Printf("Ghost and node: %s <-> %s\n", srcCurr.Path(), dstCurr.Path())
	default:
		return e.Wrapf(n.ErrBadNode, "Unexpected node type in syncFile: %v", typ)
	}

	return nil
}

func (sync *Syncer) syncDirectory(lkrSrc, lkrDst *Linker, srcCurr *n.Directory) error {
	// Possible cases here:
	// - lkrDst does not have this path (need to merge it)
	// - lkrDst has this path, but it's a ghost (need to merge with moved file, if any)
	// - lkrDst has this path, but it is not a directory (need to handle conflict?)
	// - lkrDst has this path, and it's a directory (attempt conflict resolution)
	//
	// It is guaranteed that srcCurr is *always* a directory.

	dstCurr, err := lkrDst.LookupDirectory(srcCurr.Path())
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

	// Both sides have this directory, but the content differs.
	// We need to figure out recursively what exactly is different.
	return srcCurr.VisitChildren(lkrSrc, func(nd n.Node) error {
		switch nd.Type() {
		case n.NodeTypeDirectory:
			srcChildDir, ok := nd.(*n.Directory)
			if !ok {
				return n.ErrBadNode
			}

			return sync.syncDirectory(lkrSrc, lkrDst, srcChildDir)
		case n.NodeTypeFile:
			srcChildFile, ok := nd.(*n.File)
			if !ok {
				return n.ErrBadNode
			}

			return sync.syncFile(lkrSrc, lkrDst, srcChildFile)
		case n.NodeTypeGhost:
			// TODO: Probably means that this node was removed on the remote side
			//       and we need to decide if we should propagate the remove.
			//       Find out if it's removed MoveMapping
			fmt.Printf("Ignoring ghost for now: %s\n", nd.Path())
			return nil
		}

		return n.ErrBadNode
	})
}

// Sync will apply brig's sync algorithm on lkrDst, getting it's data from lkrSrc.
// lkrSrc will not be modified in this process and will not be modified.
// Running Sync() will create a new merge commit, that takes note that we merged
// with lkrSrc at this point to avoid that conflicting files will not be processed again.
func (sync *Syncer) Sync(lkrSrc, lkrDst *Linker) error {
	srcRoot, err := lkrSrc.Root()
	if err != nil {
		return err
	}

	if err := sync.syncDirectory(lkrSrc, lkrDst, srcRoot); err != nil {
		return e.Wrap(err, "sync failed")
	}

	srcHead, err := lkrSrc.Head()
	if err != nil {
		return err
	}

	owner, err := lkrSrc.Owner()
	if err != nil {
		return err
	}

	if err := lkrDst.SetMergeMarker(owner, srcHead.Hash()); err != nil {
		return err
	}

	// TODO: Fill in the person in the commit message.
	return lkrDst.MakeCommit(n.AuthorOfStage(), "Merge with <person>")
}
