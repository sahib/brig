package cafs

import (
	"github.com/disorganizer/brig/cafs/db"
	n "github.com/disorganizer/brig/cafs/nodes"
	h "github.com/disorganizer/brig/util/hashlib"
)

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
	keyCh, err := gc.kv.Keys(key...)
	if err != nil {
		return err
	}

	for key := range keyCh {
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
				continue
			}

			// Actually get rid of the node:
			gc.lkr.MemIndexPurge(node)
			if err := gc.kv.Erase(key...); err != nil {
				return err
			}
		}
	}

	return nil
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
