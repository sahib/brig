package vcs

import (
	"fmt"
	"path"

	c "github.com/sahib/brig/catfs/core"
	ie "github.com/sahib/brig/catfs/errors"
	n "github.com/sahib/brig/catfs/nodes"
)

const (
	ConflictStragetyMarker = iota
	ConflictStragetyIgnore
	ConflictStragetyUnknown
)

type ConflictStrategy int

func (cs ConflictStrategy) String() string {
	switch cs {
	case ConflictStragetyMarker:
		return "marker"
	case ConflictStragetyIgnore:
		return "ignore"
	default:
		return "unknown"
	}
}

func ConflictStrategyFromString(spec string) ConflictStrategy {
	switch spec {
	case "marker":
		return ConflictStragetyMarker
	case "ignore":
		return ConflictStragetyIgnore
	default:
		return ConflictStragetyUnknown
	}
}

// SyncConfig gives you the possibility to configure the sync algorithm.
type SyncConfig struct {
	ConflictStrategy ConflictStrategy
	IgnoreDeletes    bool
	IgnoreMoves      bool

	OnAdd      func(newNd n.ModNode) bool
	OnRemove   func(oldNd n.ModNode) bool
	OnMerge    func(src, dst n.ModNode) bool
	OnConflict func(src, dst n.ModNode) bool
}

var (
	DefaultSyncConfig = &SyncConfig{}
)

type syncer struct {
	cfg    *SyncConfig
	lkrSrc *c.Linker
	lkrDst *c.Linker
}

func (sy *syncer) add(src n.ModNode, srcParent, srcName string) error {
	var newDstNode n.ModNode
	var err error

	parentDir, err := sy.lkrDst.LookupDirectory(srcParent)
	if err != nil {
		return err
	}

	switch src.Type() {
	case n.NodeTypeDirectory:
		newDstNode, err = n.NewEmptyDirectory(
			sy.lkrDst,
			parentDir,
			srcName,
			src.User(),
			sy.lkrDst.NextInode(),
		)

		if err != nil {
			return err
		}

		if err := sy.lkrDst.StageNode(newDstNode); err != nil {
			return err
		}

		srcDir, ok := src.(*n.Directory)
		if !ok {
			return ie.ErrBadNode
		}

		children, err := srcDir.ChildrenSorted(sy.lkrSrc)
		if err != nil {
			return err
		}

		for _, child := range children {
			childModNode, ok := child.(n.ModNode)
			if !ok {
				continue
			}

			if err := sy.add(childModNode, srcDir.Path(), child.Name()); err != nil {
				return err
			}
		}
	case n.NodeTypeFile:
		newDstFile, err := n.NewEmptyFile(
			parentDir,
			srcName,
			src.User(),
			sy.lkrDst.NextInode(),
		)

		if err != nil {
			return err
		}

		newDstNode = newDstFile

		srcFile, ok := src.(*n.File)
		if ok {
			newDstFile.SetContent(sy.lkrDst, srcFile.Content())
			newDstFile.SetSize(srcFile.Size())
			newDstFile.SetKey(srcFile.Key())
		}

		// TODO: This is inconsistent:
		// NewEmptyDirectory calls Add(), NewEmptyFile does not
		// Fix the API here a bit...
		if err := parentDir.Add(sy.lkrDst, newDstFile); err != nil {
			return err
		}

		return sy.lkrDst.StageNode(newDstNode)
	default:
		return fmt.Errorf("Unexpected node type in handleAdd")
	}

	return nil
}

func (sy *syncer) handleAdd(src n.ModNode) error {
	if sy.cfg.OnAdd != nil {
		if !sy.cfg.OnAdd(src) {
			return nil
		}
	}

	return sy.add(src, path.Dir(src.Path()), src.Name())
}

func (sy *syncer) handleMove(src, dst n.ModNode) error {
	if sy.cfg.IgnoreMoves {
		return nil
	}

	// Move our node (dst) to the path determined by src.
	return c.Move(sy.lkrDst, dst, src.Path())
}

func (sy *syncer) handleMissing(dst n.ModNode) error {
	// This is only called when a file in dst is missing on src.
	// No sync action is required.
	return nil
}

func (sy *syncer) handleRemove(dst n.ModNode) error {
	if sy.cfg.IgnoreDeletes {
		return nil
	}

	// We should check if dst really exists for us.
	if sy.cfg.OnRemove != nil {
		if !sy.cfg.OnRemove(dst) {
			return nil
		}
	}

	_, _, err := c.Remove(sy.lkrDst, dst, true, true)
	return err
}

func (sy *syncer) handleConflict(src, dst n.ModNode, srcMask, dstMask ChangeType) error {
	if sy.cfg.ConflictStrategy == ConflictStragetyIgnore {
		return nil
	}

	// Find a path that we do not have yet.
	// stamp := time.Now().Format(time.RFC3339)
	conflictNameTmpl := fmt.Sprintf("%s.conflict.%%d", dst.Name())
	conflictName := ""

	// Fix the unlikely case that there is already a node at the conflict path:
	for tries := 0; tries < 100; tries++ {
		conflictName = fmt.Sprintf(conflictNameTmpl, tries)
		dstNd, err := sy.lkrDst.LookupNode(conflictName)
		if err != nil && !ie.IsNoSuchFileError(err) {
			return err
		}

		if dstNd == nil {
			break
		}
	}

	dstDirname := path.Dir(dst.Path())

	if sy.cfg.OnConflict != nil {
		if !sy.cfg.OnConflict(src, dst) {
			return nil
		}
	}

	return sy.add(src, dstDirname, conflictName)
}

func (sy *syncer) handleMerge(src, dst n.ModNode, srcMask, dstMask ChangeType) error {
	if src.Path() != dst.Path() {
		// Only move the file if it was only moved on the remote side.
		if srcMask&ChangeTypeMove != 0 && dstMask&ChangeTypeMove == 0 {
			// TODO: Sanity check that there's nothing at src.Path(),
			//       but Mapper should already have checked that.
			if err := c.Move(sy.lkrDst, dst, src.Path()); err != nil {
				return err
			}
		}
	}

	// If src did not change, there's no need to sync the content.
	// If src has no changes, we know that dst must have changes,
	// otherwise it would have been reported as conflict.
	if srcMask&ChangeTypeModify == 0 && srcMask&ChangeTypeAdd == 0 {
		return nil
	}

	dstParent, err := n.ParentDirectory(sy.lkrDst, dst)
	if err != nil {
		return err
	}

	if err := dstParent.RemoveChild(sy.lkrSrc, dst); err != nil {
		return err
	}

	dstFile, ok := dst.(*n.File)
	if !ok {
		return ie.ErrBadNode
	}

	srcFile, ok := src.(*n.File)
	if !ok {
		return ie.ErrBadNode
	}

	dstFile.SetContent(sy.lkrDst, srcFile.Content())
	dstFile.SetSize(srcFile.Size())
	dstFile.SetKey(srcFile.Key())

	if sy.cfg.OnMerge != nil {
		if !sy.cfg.OnMerge(src, dst) {
			return nil
		}
	}

	return sy.lkrDst.StageNode(dstFile)
}

func (sy *syncer) handleTypeConflict(src, dst n.ModNode) error {
	// Simply do nothing.
	return nil
}

func Sync(lkrSrc, lkrDst *c.Linker, cfg *SyncConfig) error {
	if cfg == nil {
		cfg = DefaultSyncConfig
	}

	syncer := &syncer{
		cfg:    cfg,
		lkrSrc: lkrSrc,
		lkrDst: lkrDst,
	}

	resolver, err := newResolver(lkrSrc, lkrDst, nil, nil, syncer)
	if err != nil {
		return err
	}

	// Make sure the complete sync goes through in one disk transaction.
	return lkrDst.Atomic(func() error {
		// This calls all the handleXXX() callbacks above.
		if err := resolver.resolve(); err != nil {
			return err
		}

		wasModified, err := lkrDst.HaveStagedChanges()
		if err != nil {
			return err
		}

		// If something was changed, we should set the merge marker
		// and also create a new commit.
		if wasModified {
			srcOwner, err := lkrSrc.Owner()
			if err != nil {
				return err
			}

			srcHead, err := lkrSrc.Head()
			if err != nil {
				return err
			}

			// If something was changed, remember that we merged with src.
			// This avoids merging conflicting files a second time in the next resolve().
			if err := lkrDst.SetMergeMarker(srcOwner, srcHead.Hash()); err != nil {
				return err
			}

			message := fmt.Sprintf("merge with %s", srcOwner)
			return lkrDst.MakeCommit(srcOwner, message)
		}

		return nil
	})
}
