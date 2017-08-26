package cafs

import (
	"github.com/disorganizer/brig/cafs/db"
	n "github.com/disorganizer/brig/cafs/nodes"
	h "github.com/disorganizer/brig/util/hashlib"
)

// TODO: Make sure to not gc nodes that are present in MoveMapping.
//       (at least not in staging)

type GarbageCollector struct {
	lkr      *Linker
	kv       db.Database
	notifier func(nd n.Node) bool
	markMap  map[string]struct{}
}

func NewGarbageCollector(lkr *Linker, kv db.Database, kc func(nd n.Node) bool) *GarbageCollector {
	return &GarbageCollector{
		lkr:      lkr,
		kv:       kv,
		notifier: kc,
	}
}

func (gc *GarbageCollector) mark(cmt *n.Commit, recursive bool) error {
	if cmt == nil {
		return nil
	}

	root, err := gc.lkr.DirectoryByHash(cmt.Root())
	if err != nil {
		return err
	}

	gc.markMap[cmt.Hash().B58String()] = struct{}{}
	err = n.Walk(gc.lkr, root, true, func(child n.Node) error {
		gc.markMap[child.Hash().B58String()] = struct{}{}
		return nil
	})

	if err != nil {
		return err
	}

	parent, err := cmt.Parent(gc.lkr)
	if err != nil {
		return err
	}

	if recursive && parent != nil {
		parentCmt, ok := parent.(*n.Commit)
		if !ok {
			return n.ErrBadNode
		}

		return gc.mark(parentCmt, recursive)
	}

	return nil
}

func (gc *GarbageCollector) sweep(key []string) error {
	batch := gc.kv.Batch()
	sweeper := func(key []string) error {
		b58Hash := key[len(key)-1]
		if _, ok := gc.markMap[b58Hash]; !ok {
			hash, err := h.FromB58String(b58Hash)
			if err != nil {
				return err
			}

			node, err := gc.lkr.NodeByHash(hash)
			if err != nil {
				return err
			}

			// Allow the gc caller to check if he really
			// wants to delete this node.
			if gc.notifier != nil && !gc.notifier(node) {
				return nil
			}

			// Actually get rid of the node:
			gc.lkr.MemIndexPurge(node)
			batch.Erase(key...)
		}

		return nil
	}

	if err := gc.kv.Keys(sweeper, key...); err != nil {
		return err
	}

	return batch.Flush()
}

func (gc *GarbageCollector) Run(allObjects bool) error {
	gc.markMap = make(map[string]struct{})
	head, err := gc.lkr.Status()
	if err != nil {
		return err
	}

	if err := gc.mark(head, allObjects); err != nil {
		return err
	}

	if err := gc.sweep([]string{"stage", "objects"}); err != nil {
		return err
	}

	if allObjects {
		if err := gc.sweep([]string{"objects"}); err != nil {
			return err
		}
	}

	return nil
}
