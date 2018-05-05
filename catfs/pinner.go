package catfs

import (
	"bytes"
	"encoding/gob"
	"errors"

	"github.com/dgraph-io/badger"
	c "github.com/sahib/brig/catfs/core"
	ie "github.com/sahib/brig/catfs/errors"
	n "github.com/sahib/brig/catfs/nodes"
	h "github.com/sahib/brig/util/hashlib"
)

// errNotPinnedSentinel is returned to signal an early exit in Walk()
var errNotPinnedSentinel = errors.New("not pinned")

type pinCacheEntry struct {
	IsExplicit bool
	IsPinned   bool
}

type Pinner struct {
	db  *badger.DB
	bk  FsBackend
	lkr *c.Linker
}

func NewPinner(pinDbPath string, lkr *c.Linker, bk FsBackend) (*Pinner, error) {
	opts := badger.DefaultOptions
	opts.Dir = pinDbPath
	opts.ValueDir = pinDbPath

	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	return &Pinner{db: db, lkr: lkr, bk: bk}, nil
}

func (pc *Pinner) Close() error {
	return pc.db.Close()
}

func (pc *Pinner) Remember(content h.Hash, isPinned, isExplicit bool) error {
	return pc.db.Update(func(txn *badger.Txn) error {
		buf := &bytes.Buffer{}
		enc := gob.NewEncoder(buf)
		if err := enc.Encode(pinCacheEntry{isExplicit, isPinned}); err != nil {
			return err
		}

		return txn.Set(content, buf.Bytes())
	})
}

func (pc *Pinner) IsPinned(hash h.Hash) (isPinned bool, isExplicit bool, err error) {
	entry := pinCacheEntry{}
	err = pc.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(hash)
		if err != nil {
			return err
		}

		value, err := item.Value()
		if err != nil {
			return err
		}

		dec := gob.NewDecoder(bytes.NewReader(value))
		return dec.Decode(&entry)
	})

	if err != badger.ErrKeyNotFound {
		isPinned = entry.IsPinned
		isExplicit = entry.IsExplicit
	}

	// silence a key error, ok will be false then.
	isPinned, err = pc.bk.IsPinned(hash)
	if err != nil {
		return
	}

	isExplicit = false
	pc.Remember(hash, isPinned, isExplicit)
	return
}

////////////////////////////

func (pc *Pinner) Pin(hash h.Hash, explicit bool) error {
	isPinned, isExplicit, err := pc.IsPinned(hash)
	if err != nil {
		return err
	}

	if isPinned {
		if isExplicit && !explicit {
			// will not "downgrade" an existing pin.
			return nil
		}
	}

	if !isPinned {
		if err := pc.bk.Pin(hash); err != nil {
			return err
		}
	}

	return pc.Remember(hash, isPinned, explicit)
}

func (pc *Pinner) Unpin(hash h.Hash, explicit bool) error {
	isPinned, isExplicit, err := pc.IsPinned(hash)
	if err != nil {
		return err
	}

	if isPinned {
		if isExplicit && !explicit {
			return nil
		}
	}

	if isPinned {
		if err := pc.bk.Unpin(hash); err != nil {
			return err
		}
	}

	return pc.Remember(hash, isPinned, explicit)
}

////////////////////////////

func (pc *Pinner) doPinOp(op func(h.Hash, bool) error, nd n.Node, explicit bool) error {
	return n.Walk(pc.lkr, nd, true, func(child n.Node) error {
		if child.Type() == n.NodeTypeFile {
			file, ok := child.(*n.File)
			if !ok {
				return ie.ErrBadNode
			}

			if err := op(file.BackendHash(), explicit); err != nil {
				return err
			}
		}

		return nil
	})
}

func (pc *Pinner) PinNode(nd n.Node, explicit bool) error {
	return pc.doPinOp(pc.Pin, nd, explicit)
}

func (pc *Pinner) UnpinNode(nd n.Node, explicit bool) error {
	return pc.doPinOp(pc.Unpin, nd, explicit)
}

func (pc *Pinner) IsNodePinned(nd n.Node) (bool, bool, error) {
	pinCount := 0
	explicitCount := 0
	totalCount := 0

	err := n.Walk(pc.lkr, nd, true, func(child n.Node) error {
		if child.Type() != n.NodeTypeFile {
			return nil
		}

		file, ok := child.(*n.File)
		if !ok {
			return ie.ErrBadNode
		}

		totalCount++

		isPinned, isExplicit, err := pc.IsPinned(file.BackendHash())
		if err != nil {
			return err
		}

		if isExplicit {
			explicitCount++
		}

		if isPinned {
			// Make sure that we do not count empty directories
			// as pinned nodes.
			pinCount++
		} else {
			// Return a special error here to stop Walk() iterating.
			// One file is enough to stop IsPinned() from being true.
			return errNotPinnedSentinel
		}

		return nil
	})

	if err != nil && err != errNotPinnedSentinel {
		return false, false, err
	}

	if err == errNotPinnedSentinel {
		return false, false, nil
	}

	return pinCount > 0, explicitCount == totalCount, nil
}
