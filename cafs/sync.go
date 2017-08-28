package cafs

// The sync algorithm tries to handle the following special cases:
// - Propagate moves (most of them, at least)
// - Propagate deletes (configurable?)

import (
	"errors"
	// "fmt"
	// n "github.com/disorganizer/brig/cafs/nodes"
	// e "github.com/pkg/errors"
)

var (
	ErrConflict = errors.New("Conflicting changes")
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

var DefaultSyncConfig = &SyncConfig{}

// type SyncStats struct {
// 	Merged int
// 	Size   int
// }
//
// func resolve(lkrSrc, lkrDst *Linker, src, dst n.Node) error {
// 	return ErrConflict
// }
//
// func conflict(cfg *SyncConfig, lkrSrc, lkrDst *Linker, src, dst n.Node) error {
// 	switch cfg.ConflictStragey {
// 	case ConflictStragetyIgnore:
// 		return nil
// 	case ConflictStragetyMarker:
// 		// Import src as src.Path() + owner name
// 		return nil
// 	default:
// 		return fmt.Errorf("Unknown conflict strategy: %v", cfg.ConflictStragey)
// 	}
// 	return nil
// }
//
// func syncFile(lkrSrc, lkrDst *Linker, cfg *SyncConfig, curr *n.File) error {
// 	movedNode, moveDirection, err := lkrSrc.MoveMapping(curr)
// 	if err != nil {
// 		return err
// 	}
//
// 	if movedNode.Type() == n.NodeTypeGhost && moveDirection == MoveDirDstToSrc {
// 		ghost, ok := movedNode.(*n.Ghost)
// 		if !ok {
// 			return n.ErrBadNode
// 		}
//
// 		oldFile, err := ghost.OldFile()
// 		if err != nil {
// 			return err
// 		}
//
// 		return syncFile(lkrSrc, lkrDst, cfg, oldFile)
// 	}
//
// 	srcNode, err := lkrDst.LookupNode(curr.Path())
// 	if err != nil && !n.IsNoSuchFileError(err) {
// 		return err
// 	}
//
// 	if srcNode == nil {
// 		// We do not have this node yet, mark it for copying.
// 		fmt.Printf("Syncing remote file `%s`\n", curr.Path())
// 		return nil
// 	}
//
// 	switch typ := srcNode.Type(); typ {
// 	case n.NodeTypeFile:
// 		// We have two competing files. Let's figure out if the changes done to
// 		// them are compatible.
// 		fmt.Printf("Competing nodes: %s <-> %s\n", curr.Path(), srcNode.Path())
// 	case n.NodeTypeGhost:
// 		fmt.Printf("Ghost and node: %s <-> %s\n", curr.Path(), srcNode.Path())
// 	default:
// 		return e.Wrapf(n.ErrBadNode, "Unexpected node type in syncFile: %v", typ)
// 	}
//
// 	return nil
// }
//
// func syncDirectory(lkrSrc, lkrDst *Linker, cfg *SyncConfig, curr *n.Directory) error {
// 	movedNode, moveDirection, err := lkrDst.MoveMapping(curr)
// 	if err != nil {
// 		return err
// 	}
//
// 	if movedNode.Type() == n.NodeTypeGhost && moveDirection == MoveDirDstToSrc {
// 		// Apparently, this node was removed on the remote side.
// 		// Try to sync the directory again with the old destination.
// 		ghost, ok := movedNode.(*n.Ghost)
// 		if !ok {
// 			return n.ErrBadNode
// 		}
//
// 		oldDirectory, err := ghost.OldDirectory()
// 		if err != nil {
// 			return err
// 		}
//
// 		return syncDirectory(lkrSrc, lkrDst, cfg, oldDirectory)
// 	}
//
// 	// TODO: Do we need to see if curr.Path() in lkrSrc is a ghost?
// 	//       And if so, retrieve the old directory before assigning srcCurr?
// 	srcCurr, err := lkrSrc.LookupDirectory(curr.Path())
// 	if err != nil && !n.IsNoSuchFileError(err) {
// 		return err
// 	}
//
// 	if srcCurr == nil {
// 		// We do not have this directory apparently. Go sync it.
// 		fmt.Printf("Marked remote directory `%s` for syncing.\n", curr.Path())
// 		return nil
// 	}
//
// 	// Check if we're lucky and the directory hash is equal:
// 	if srcCurr.Hash().Equal(curr.Hash()) {
// 		fmt.Printf(
// 			"src (%s) and dest (%s) are equal; no sync needed",
// 			srcCurr.Path(), curr.Path()
// 		)
// 		return nil
// 	}
//
// 	// Recurse into sub nodes:
// 	return curr.VisitChildren(lkrDst, func(nd n.Node) error {
// 		switch nd.Type() {
// 		case n.NodeTypeDirectory:
// 			childDir, ok := nd.(*n.Directory)
// 			if !ok {
// 				return n.ErrBadNode
// 			}
//
// 			return syncDirectory(lkrSrc, lkrDst, cfg, childDir)
// 		case n.NodeTypeFile:
// 			childFile, ok := nd.(*n.File)
// 			if !ok {
// 				return n.ErrBadNode
// 			}
//
// 			return syncFile(lkrSrc, lkrDst, cfg, childFile)
// 		case n.NodeTypeGhost:
// 			fmt.Printf("Ignoring ghost: %s\n", nd.Path())
// 			return nil
// 		}
// 	})
// }
//
// // sync will apply brig's sync algorithm, merging files from lkrDst to lkrSrc.
// // Exact behaviour can be configured via `cfg`.
// // If the merge was succesful, statistics about the number of merged files
// // will be returned.
// func sync(lkrSrc, lkrDst *Linker, cfg *SyncConfig) error {
// 	if cfg == nil {
// 		cfg = DefaultSyncConfig
// 	}
//
// 	srcRoot, err := lkrSrc.Root()
// 	if err != nil {
// 		return err
// 	}
//
// 	return syncDirectory(lkrSrc, lkrDst, cfg, srcRoot)
//
// 	// // TODO: Only walk over leaf nodes (files, ghosts, dirs without children)
// 	// err = n.Walk(lkrDst, root, true, func(child n.Node) error {
// 	// 	srcNode, err := lkrSrc.LookupNode(child.Path())
// 	// 	if err != nil && !n.IsNoSuchFileError(err) {
// 	// 		return err
// 	// 	}
//
// 	// 	if srcNode == nil {
// 	// 		// We do not have this file yet:
// 	// 		// Go add it.
// 	// 		return nil
// 	// 	}
//
// 	// 	// Check if we can resolve the two files ourselves.
// 	// 	switch err = resolve(lkrSrc, lkrDst, srcNode, child); err {
// 	// 	case ErrConflict:
// 	// 		// We need to perform conflict resolution.
// 	// 		return conflict(cfg, lkrSrc, lkrDst, srcNode, child)
// 	// 	case nil:
// 	// 		// resolve() was able to resolve this conflict.
// 	// 	default:
// 	// 		// Some other error happened in resolve()
// 	// 		return err
// 	// 	}
//
// 	// 	return nil
// 	// })
//
// 	// if err != nil {
// 	// 	return fmt.Errorf("sync: Walk failed: %v", err)
// 	// }
//
// 	// return nil
// }
