package vcs

import (
	c "github.com/sahib/brig/catfs/core"
	n "github.com/sahib/brig/catfs/nodes"
)

type DiffPair struct {
	Src     n.ModNode
	Dst     n.ModNode
	SrcMask ChangeType
	DstMask ChangeType
}

type Diff struct {
	cfg *SyncConfig

	// Nodes that were added from remote.
	Added []n.ModNode

	// Nodes that will be removed on remote side.
	Removed []n.ModNode

	// Nodes from remote that were ignored.
	Ignored []n.ModNode

	// Merged contains nodes where sync is able to combine changes
	// on both sides (i.e. one side moved, another modified)
	Merged []DiffPair

	// Conflict contains nodes where sync was not able to combine
	// the changes made on both sides.
	Conflict []DiffPair
}

func (df *Diff) handleAdd(src n.ModNode) error {
	df.Added = append(df.Added, src)
	return nil
}
func (df *Diff) handleRemove(dst n.ModNode) error {
	if df.cfg.IgnoreDeletes {
		df.Ignored = append(df.Ignored, dst)
		return nil
	}

	df.Removed = append(df.Removed, dst)
	return nil
}

func (df *Diff) handleTypeConflict(src, dst n.ModNode) error {
	df.Ignored = append(df.Ignored, dst)
	return nil
}

func (df *Diff) handleConflict(src, dst n.ModNode, srcMask, dstMask ChangeType) error {
	df.Conflict = append(df.Conflict, DiffPair{
		Src:     src,
		Dst:     dst,
		SrcMask: srcMask,
		DstMask: dstMask,
	})

	return nil
}

func (df *Diff) handleMerge(src, dst n.ModNode, srcMask, dstMask ChangeType) error {
	df.Merged = append(df.Merged, DiffPair{
		Src:     src,
		Dst:     dst,
		SrcMask: srcMask,
		DstMask: dstMask,
	})

	return nil
}

// Diff show the differences between two linkers.
//
// Internally it works like Sync() but does not modify anything and just
// merely records what the algorithm decided to do.
func MakeDiff(lkrSrc, lkrDst *c.Linker, headSrc, headDst *n.Commit, cfg *SyncConfig) (*Diff, error) {
	if cfg == nil {
		cfg = DefaultSyncConfig
	}

	diff := &Diff{cfg: cfg}
	rsv, err := newResolver(lkrSrc, lkrDst, headSrc, headDst, diff)
	if err != nil {
		return nil, err
	}

	if err := rsv.resolve(); err != nil {
		return nil, err
	}

	return diff, nil
}
