package catfs

import (
	"github.com/disorganizer/brig/catfs/db"
	n "github.com/disorganizer/brig/catfs/nodes"
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

func (gc *GarbageCollector) markMoveMaps() error {
	walker := func(key []string) error {
		data, err := gc.kv.Get(key...)
		if err != nil {
			return err
		}

		node, _, err := gc.lkr.parseMoveMappingLine(string(data))
		if err != nil {
			return err
		}

		if node != nil {
			gc.markMap[node.Hash().B58String()] = struct{}{}
		}

		return nil
	}

	return gc.kv.Keys(walker, "stage", "moves")
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

	// Staging might contain moved files that are not reachable anymore,
	// but still are referenced by the move mapping.
	// Keep them for now, they will die most likely on MakeCommit()
	if err := gc.markMoveMaps(); err != nil {
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
