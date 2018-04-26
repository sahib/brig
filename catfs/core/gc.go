package core

import (
	"github.com/sahib/brig/catfs/db"
	ie "github.com/sahib/brig/catfs/errors"
	n "github.com/sahib/brig/catfs/nodes"
	h "github.com/sahib/brig/util/hashlib"
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

func (gc *GarbageCollector) markMoveMap(key []string) error {
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
			gc.markMap[node.TreeHash().B58String()] = struct{}{}
		}

		return nil
	}

	return gc.kv.Keys(walker, key...)
}

func (gc *GarbageCollector) mark(cmt *n.Commit, recursive bool) error {
	if cmt == nil {
		return nil
	}

	root, err := gc.lkr.DirectoryByHash(cmt.Root())
	if err != nil {
		return err
	}

	gc.markMap[cmt.TreeHash().B58String()] = struct{}{}
	err = n.Walk(gc.lkr, root, true, func(child n.Node) error {
		gc.markMap[child.TreeHash().B58String()] = struct{}{}
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
			return ie.ErrBadNode
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

// TODO: write test that covers this.
//       we don't want to find hard bugs later because of gc going haywire.
func (gc *GarbageCollector) findAllMoveLocations(head *n.Commit) ([][]string, error) {
	locations := [][]string{
		{"stage", "moves"},
	}

	for {
		parent, err := head.Parent(gc.lkr)
		if err != nil {
			return nil, err
		}

		if parent == nil {
			break
		}

		parentCmt, ok := parent.(*n.Commit)
		if !ok {
			return nil, ie.ErrBadNode
		}

		head = parentCmt
		location := []string{"moves", head.TreeHash().B58String()}
		locations = append(locations, location)
	}

	return locations, nil
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
	moveMapLocations := [][]string{
		{"stage", "moves"},
	}

	if allObjects {
		moveMapLocations, err = gc.findAllMoveLocations(head)
		if err != nil {
			return err
		}
	}

	for _, location := range moveMapLocations {
		if err := gc.markMoveMap(location); err != nil {
			return err
		}
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
