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
	"path"

	c "github.com/disorganizer/brig/catfs/core"
	ie "github.com/disorganizer/brig/catfs/errors"
	n "github.com/disorganizer/brig/catfs/nodes"
	e "github.com/pkg/errors"
)

const (
	ConflictStragetyMarker = iota
	ConflictStragetyIgnore
	ConflictStragetyUnknown
)

type ConflictStragey int

// SyncConfig gives you the possibility to configure the sync algorithm.
// The zero value of each option is the
type SyncConfig struct {
	ConflictStragey ConflictStragey
	IgnoreDeletes   bool
}

var (
	DefaultSyncConfig = &SyncConfig{}
)

type executor interface {
	handleAdd(src n.ModNode) error
	handleRemove(dst n.ModNode) error
	handleConflict(src, dst n.ModNode) error
	handleTypeConflict(src, dst n.ModNode) error
	handleMerge(src, dst n.ModNode, srcMask, dstMask ChangeType) error
}

////////////////////////////////
// DIFF METHOD IMPLEMENTATION //
////////////////////////////////

type Differ struct {
}

// TODO: Implement.

//////////////////////////////////////
// RESOLUTION METHOD IMPLEMENTATION //
//////////////////////////////////////

type Syncer struct {
	cfg    *SyncConfig
	lkrSrc *c.Linker
	lkrDst *c.Linker
}

func (sy *Syncer) add(src n.ModNode, srcParent, srcName string) error {
	fmt.Println("ADD", src, srcParent, srcName)
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
			sy.lkrDst.NextInode(),
		)

		if err != nil {
			return err
		}
	case n.NodeTypeFile:
		newDstFile, err := n.NewEmptyFile(
			parentDir,
			srcName,
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
		if err := parentDir.Add(sy.lkrDst, newDstFile); err != nil {
			return err
		}
	default:
		return fmt.Errorf("Unexpected node type in handleAdd")
	}
	fmt.Println("STAGE")

	return sy.lkrDst.StageNode(newDstNode)
}

func (sy *Syncer) handleAdd(src n.ModNode) error {
	return sy.add(src, path.Dir(src.Path()), src.Name())
}

func (sy *Syncer) handleRemove(dst n.ModNode) error {
	if sy.cfg.IgnoreDeletes {
		return nil
	}

	_, _, err := c.Remove(sy.lkrDst, dst, true, true)
	return err
}

func (sy *Syncer) handleConflict(src, dst n.ModNode) error {
	fmt.Println("CONFLICT", src, dst)
	if sy.cfg.ConflictStragey == ConflictStragetyIgnore {
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
	fmt.Println("Writing conflict file to ", dstDirname, conflictName)
	return sy.add(src, dstDirname, conflictName)
}

func (sy *Syncer) handleMerge(src, dst n.ModNode, srcMask, dstMask ChangeType) error {
	if src.Path() != dst.Path() {
		// Only move the file if it was only moved on the remote side.
		if srcMask&ChangeTypeMove != 0 && dstMask&ChangeTypeMove == 0 {
			// TODO: Sanity check that there's nothing that src.Path(),
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

	return sy.lkrDst.StageNode(dstFile)
}

func (sy *Syncer) handleTypeConflict(src, dst n.ModNode) error {
	// Simply do nothing.
	return nil
}

//////////////////////////////////////////////
// IMPLEMENTATION OF ACTUAL DECISION MAKING //
//////////////////////////////////////////////

type Resolver struct {
	lkrSrc *c.Linker
	lkrDst *c.Linker

	// cached attributes:
	dstMergeCmt *n.Commit
	srcMergeCmt *n.Commit

	exec executor
}

func NewResolver(lkrSrc, lkrDst *c.Linker, exec executor) *Resolver {
	return &Resolver{
		lkrSrc: lkrSrc,
		lkrDst: lkrDst,
		exec:   exec,
	}
}

func (rv *Resolver) Resolve() error {
	srcRoot, err := rv.lkrSrc.Root()
	if err != nil {
		return err
	}

	if err := rv.cacheLastCommonMerge(); err != nil {
		return e.Wrapf(err, "Error while finding last common merge")
	}

	srcHead, err := rv.lkrSrc.Head()
	if err != nil {
		return err
	}

	srcOwner, err := rv.lkrSrc.Owner()
	if err != nil {
		return err
	}

	mapper := NewMapper(rv.lkrSrc, rv.lkrDst, srcRoot)
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

	wasModified, err := rv.lkrDst.HaveStagedChanges()
	if err != nil {
		return err
	}

	// If something was changed, we should set the merge marker.
	if wasModified {
		// If something was changed, remember that we merged with src.
		// This avoids merging conflicting files a second time in the next resolve().
		if err := rv.lkrDst.SetMergeMarker(srcOwner, srcHead.Hash()); err != nil {
			return err
		}

		message := fmt.Sprintf("Merge with %s", srcOwner.ID())
		return rv.lkrDst.MakeCommit(srcOwner, message)
	}

	return nil
}

func (rv *Resolver) cacheLastCommonMerge() error {
	srcOwner, err := rv.lkrSrc.Owner()
	if err != nil {
		return err
	}

	dstHead, err := rv.lkrDst.Head()
	if err != nil {
		return err
	}

	for {
		with, srcRef := dstHead.MergeMarker()
		if with != nil && with.Equal(srcOwner) {
			srcHead, err := rv.lkrSrc.CommitByHash(srcRef)
			if err != nil {
				return err
			}

			rv.dstMergeCmt = dstHead
			rv.srcMergeCmt = srcHead
		}

		prevHeadNode, err := dstHead.Parent(rv.lkrDst)
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

		dstHead = newDstHead
	}

	return nil
}

// hasConflicts is always called when two nodes on both sides and they do not
// have the same hash.  In the best case, both have compatible changes and can
// be merged, otherwise a user defined conflict strategy has to be applied.
func (rv *Resolver) hasConflicts(src, dst n.ModNode) (bool, ChangeType, ChangeType, error) {
	srcHead, err := rv.lkrSrc.Head()
	if err != nil {
		return false, 0, 0, err
	}

	dstHead, err := rv.lkrDst.Head()
	if err != nil {
		return false, 0, 0, err
	}

	srcHist, err := History(rv.lkrSrc, src, srcHead, rv.srcMergeCmt)
	if err != nil {
		return false, 0, 0, err
	}

	dstHist, err := History(rv.lkrDst, dst, dstHead, rv.dstMergeCmt)
	if err != nil {
		return false, 0, 0, err
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
		// We can "fast forward" our node.
		// There are only remote changes for this file.
		fmt.Println("fast forward")
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

func (rv *Resolver) decide(pair MapPair) error {
	if pair.Src == nil && pair.Dst == nil {
		return fmt.Errorf("Received completely empty mapping; ignoring")
	}

	if pair.Src == nil {
		fmt.Println("Source was removed: ", pair.Dst.Path())
		return rv.exec.handleRemove(pair.Dst)
	}

	if pair.Dst == nil {
		fmt.Println("No such dest: ", pair.Src.Path())
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

	fmt.Println("HAS CONFLICT", hasConflicts, srcMask, dstMask)
	if hasConflicts {
		return rv.exec.handleConflict(pair.Src, pair.Dst)
	}

	// handleMerge needs the masks to decide what path / content to choose.
	return rv.exec.handleMerge(pair.Src, pair.Dst, srcMask, dstMask)
}

//////

func Sync(lkrSrc, lkrDst *c.Linker, cfg *SyncConfig) error {
	if cfg == nil {
		cfg = DefaultSyncConfig
	}

	syncer := &Syncer{
		cfg:    cfg,
		lkrSrc: lkrSrc,
		lkrDst: lkrDst,
	}

	rsv := NewResolver(lkrSrc, lkrDst, syncer)
	return rsv.Resolve()
}
