package vcs

import (
	"fmt"
	"path"

	e "github.com/pkg/errors"
	c "github.com/sahib/brig/catfs/core"
	ie "github.com/sahib/brig/catfs/errors"
	n "github.com/sahib/brig/catfs/nodes"
	log "github.com/sirupsen/logrus"
)

const (
	// ConflictStragetyMarker creates marker files for each conflict.
	ConflictStragetyMarker = iota

	// ConflictStragetyIgnore ignores conflicts totally.
	ConflictStragetyIgnore

	// ConflictStragetyEmbrace takes the version of the remote.
	ConflictStragetyEmbrace

	// ConflictStragetyUnknown should be used when the strategy is not clear.
	ConflictStragetyUnknown
)

// ConflictStrategy defines what conflict strategy to apply in case of
// nodes with different content hashes.
type ConflictStrategy int

func (cs ConflictStrategy) String() string {
	switch cs {
	case ConflictStragetyMarker:
		return "marker"
	case ConflictStragetyIgnore:
		return "ignore"
	case ConflictStragetyEmbrace:
		return "embrace"
	default:
		return "unknown"
	}
}

// ConflictStrategyFromString converts a string to a ConflictStrategy.
// It it is not valid, ConflictStragetyUnknown is returned.
func ConflictStrategyFromString(spec string) ConflictStrategy {
	switch spec {
	case "marker":
		return ConflictStragetyMarker
	case "ignore":
		return ConflictStragetyIgnore
	case "embrace":
		return ConflictStragetyEmbrace
	default:
		return ConflictStragetyUnknown
	}
}

// SyncOptions gives you the possibility to configure the sync algorithm.
type SyncOptions struct {
	ConflictStrategy          ConflictStrategy
	IgnoreDeletes             bool
	IgnoreMoves               bool
	Message                   string
	ReadOnlyFolders           map[string]bool
	ConflictStrategyPerFolder map[string]ConflictStrategy

	OnAdd      func(newNd n.ModNode) bool
	OnRemove   func(oldNd n.ModNode) bool
	OnMerge    func(src, dst n.ModNode) bool
	OnConflict func(src, dst n.ModNode) bool
}

var (
	defaultSyncConfig = &SyncOptions{}
)

type syncer struct {
	cfg    *SyncOptions
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
		newDstFile := n.NewEmptyFile(
			parentDir,
			srcName,
			src.User(),
			sy.lkrDst.NextInode(),
		)

		newDstNode = newDstFile

		srcFile, ok := src.(*n.File)
		if ok {
			newDstFile.SetContent(sy.lkrDst, srcFile.ContentHash())
			newDstFile.SetBackend(sy.lkrDst, srcFile.BackendHash())
			newDstFile.SetSize(srcFile.Size())
			newDstFile.SetKey(srcFile.Key())
		}

		if err := parentDir.Add(sy.lkrDst, newDstFile); err != nil {
			return err
		}

		return sy.lkrDst.StageNode(newDstNode)
	case n.NodeTypeGhost:
		// skipping addition of a ghost
		return nil
	default:
		return fmt.Errorf("Unexpected node type in handleAdd")
	}

	return nil
}

func isReadOnly(folders map[string]bool, nodePaths ...string) bool {
	for _, nodePath := range nodePaths {
		for {
			if folders[nodePath] {
				return true
			}

			newNodePath := path.Dir(nodePath)
			if newNodePath == nodePath {
				break
			}

			nodePath = newNodePath
		}
	}

	return false
}

func (sy *syncer) handleAdd(src n.ModNode) error {
	if isReadOnly(sy.cfg.ReadOnlyFolders, src.Path()) {
		return nil
	}

	log.Debugf("handling add: %s", src.Path())
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

	if isReadOnly(sy.cfg.ReadOnlyFolders, src.Path(), dst.Path()) {
		return nil
	}

	log.Debugf("handling move: %s -> %s", dst.Path(), src.Path())
	if _, err := c.Mkdir(sy.lkrDst, path.Dir(src.Path()), true); err != nil {
		return err
	}

	// Move our node (dst) to the path determined by src.
	return e.Wrapf(c.Move(sy.lkrDst, dst, src.Path()), "move")
}

func (sy *syncer) handleMissing(dst n.ModNode) error {
	// This is only called when a file in dst is missing on src.
	// No sync action is required.
	log.Debugf("handling missing: %s", dst.Path())
	return nil
}

func (sy *syncer) handleRemove(dst n.ModNode) error {
	if sy.cfg.IgnoreDeletes {
		return nil
	}

	if isReadOnly(sy.cfg.ReadOnlyFolders, dst.Path()) {
		return nil
	}

	log.Debugf("handling remove: %s", dst.Path())

	if sy.cfg.OnRemove != nil {
		if !sy.cfg.OnRemove(dst) {
			return nil
		}
	}

	_, _, err := c.Remove(sy.lkrDst, dst, true, true)
	return err
}

func (sy *syncer) getConflictStrategy(nd n.ModNode) ConflictStrategy {
	curr := nd.Path()

	// Shortcurt: If the per-folder feature is not used,
	// we can skip this whole loop below.
	if len(sy.cfg.ConflictStrategyPerFolder) == 0 {
		return sy.cfg.ConflictStrategy
	}

	log.Debugf("*** MAP %v", sy.cfg.ConflictStrategyPerFolder)

	for {
		cs, ok := sy.cfg.ConflictStrategyPerFolder[curr]
		if ok {
			return cs
		}

		parent := path.Dir(curr)
		if parent == curr {
			break
		}

		curr = parent
	}

	// No special strategy found for this folder
	return sy.cfg.ConflictStrategy
}

func (sy *syncer) handleConflict(src, dst n.ModNode, srcMask, dstMask ChangeType) error {
	cs := sy.getConflictStrategy(dst)

	if cs == ConflictStragetyIgnore {
		return nil
	}

	if cs == ConflictStragetyEmbrace {
		return sy.handleMerge(src, dst, srcMask, dstMask)
	}

	if isReadOnly(sy.cfg.ReadOnlyFolders, src.Path(), dst.Path()) {
		return nil
	}

	log.Debugf("handling conflict: %s <-> %s", src.Path(), dst.Path())

	// Find a path that we do not have yet.
	// stamp := time.Now().Format(time.RFC3339)
	conflictName := ""
	conflictNameTmpl := fmt.Sprintf("%s.conflict.%%d", dst.Name())

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
	if isReadOnly(sy.cfg.ReadOnlyFolders, src.Path(), dst.Path()) {
		return nil
	}

	if isReadOnly(sy.cfg.ReadOnlyFolders, src.Path(), dst.Path()) {
		return nil
	}

	if src.Path() != dst.Path() {
		// Only move the file if it was only moved on the remote side.
		if srcMask&ChangeTypeMove != 0 && dstMask&ChangeTypeMove == 0 {
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

	dstFile.SetContent(sy.lkrDst, srcFile.ContentHash())
	dstFile.SetBackend(sy.lkrDst, srcFile.BackendHash())
	dstFile.SetSize(srcFile.Size())
	dstFile.SetKey(srcFile.Key())

	if err := dstParent.Add(sy.lkrDst, dstFile); err != nil {
		return err
	}

	if sy.cfg.OnMerge != nil {
		if !sy.cfg.OnMerge(src, dst) {
			return nil
		}
	}

	return sy.lkrDst.StageNode(dstFile)
}

func (sy *syncer) handleTypeConflict(src, dst n.ModNode) error {
	log.Debugf("handling type conflict: %s <-> %s", src.Path(), dst.Path())

	// Simply do nothing.
	return nil
}

func (sy *syncer) handleConflictNode(src n.ModNode) error {
	log.Debugf("handling node conflict: %s", src.Path())

	// We don't care for files on the other side named "README.conflict.0" e.g.
	return nil
}

// Sync will synchronize the changes from `lkrSrc` to `lkrDst`,
// according to the options set in `cfg`. This is atomic.
// A new commit might be created with `message`, defaulting to a default message
// when an empty string was given.
func Sync(lkrSrc, lkrDst *c.Linker, cfg *SyncOptions) error {
	if cfg == nil {
		cfg = defaultSyncConfig
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
	return lkrDst.Atomic(func() (bool, error) {
		// This calls all the handleXXX() callbacks above.
		if err := resolver.resolve(); err != nil {
			return true, err
		}

		wasModified, err := lkrDst.HaveStagedChanges()
		if err != nil {
			return true, err
		}

		// If something was changed, we should set the merge marker
		// and also create a new commit.
		if wasModified {
			srcOwner, err := lkrSrc.Owner()
			if err != nil {
				return true, err
			}

			srcHead, err := lkrSrc.Head()
			if err != nil {
				return true, err
			}

			// If something was changed, remember that we merged with src.
			// This avoids merging conflicting files a second time in the next resolve().
			if err := lkrDst.SetMergeMarker(srcOwner, srcHead.TreeHash()); err != nil {
				return true, err
			}

			message := cfg.Message
			if message == "" {
				message = fmt.Sprintf("merge with »%s«", srcOwner)
			}

			if err := lkrDst.MakeCommit(srcOwner, message); err != nil {
				return true, err
			}
		}

		return false, nil
	})
}
