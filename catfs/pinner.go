package catfs

import (
	"bytes"
	"encoding/gob"
	"errors"

	// Because ipfs' package manager sucks a lot (sorry, but it does)
	// it imports badger with the import url below. This calls a few init()s,
	// which will panic when being called twice due to expvar defines e.g.
	// (i.e. when using the "correct" import github.com/dgraph-io/badger)
	//
	// So gx forces us to use their badger version for no good reason at all.
	"gx/ipfs/QmeAEa8FDWAmZJTL6YcM1oEndZ4MyhCr5rTsjYZQui1x1L/badger"

	c "github.com/sahib/brig/catfs/core"
	ie "github.com/sahib/brig/catfs/errors"
	n "github.com/sahib/brig/catfs/nodes"
	h "github.com/sahib/brig/util/hashlib"
)

// errNotPinnedSentinel is returned to signal an early exit in Walk()
var errNotPinnedSentinel = errors.New("not pinned")

// pinCacheEntry is one entry in the pin cache.
type pinCacheEntry struct {
	Inodes map[uint64]bool
}

// Pinner remembers which hashes are pinned and if they are pinned explicitly.
// Its API can be used to safely change the pinning state. It assumes that it
// is the only entitiy the pins & unpins nodes.
type Pinner struct {
	db  *badger.DB
	bk  FsBackend
	lkr *c.Linker
}

// NewPinner creates a new pin cache at `pinDbPath`, possibly erroring out.
// `lkr` and `bk` are used to make PinNode() and UnpinNode() work.
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

// Close the pinning cache.
func (pc *Pinner) Close() error {
	return pc.db.Close()
}

func getEntry(txn *badger.Txn, hash h.Hash) (*pinCacheEntry, error) {
	item, err := txn.Get(hash)
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, nil
		}

		return nil, err
	}

	value, err := item.Value()
	if err != nil {
		return nil, err
	}

	dec := gob.NewDecoder(bytes.NewReader(value))
	entry := pinCacheEntry{}
	return &entry, dec.Decode(&entry)
}

// remember the pin state of a certain hash.
// This does change anything in the backend but only changes the caching structure.
// Use with care to avoid data inconsistencies.
func (pc *Pinner) remember(inode uint64, hash h.Hash, isPinned, isExplicit bool) error {
	return pc.db.Update(func(txn *badger.Txn) error {
		buf := &bytes.Buffer{}
		enc := gob.NewEncoder(buf)

		oldEntry, err := getEntry(txn, hash)
		if err != nil {
			return err
		}

		var inodes map[uint64]bool
		if oldEntry != nil {
			inodes = oldEntry.Inodes
		} else {
			inodes = make(map[uint64]bool)
		}

		if !isPinned {
			delete(inodes, inode)
		} else {
			inodes[inode] = isExplicit
		}

		entry := pinCacheEntry{
			Inodes: inodes,
		}

		if err := enc.Encode(entry); err != nil {
			return err
		}

		return txn.Set(hash, buf.Bytes())
	})
}

func (pc *Pinner) IsPinned(inode uint64, hash h.Hash) (bool, bool, error) {
	entry := pinCacheEntry{}
	err := pc.db.View(func(txn *badger.Txn) error {
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

	if err != nil && err != badger.ErrKeyNotFound {
		return false, false, err
	}

	if err != badger.ErrKeyNotFound {
		// cache hit
		isExplicit, ok := entry.Inodes[inode]
		return ok, isExplicit, nil
	}

	// We do not have this information yet.
	// Create a new entry based on the backend information.

	// silence a key error, ok will be false then.
	isPinned, err := pc.bk.IsPinned(hash)
	if err != nil {
		return false, false, err
	}

	// remember the file to be pinned non-explicitly:
	if err := pc.remember(inode, hash, isPinned, false); err != nil {
		return false, false, err
	}

	return true, false, nil
}

////////////////////////////

func (pc *Pinner) Pin(inode uint64, hash h.Hash, explicit bool) error {
	isPinned, isExplicit, err := pc.IsPinned(inode, hash)
	if err != nil {
		return err
	}

	if isPinned {
		if isExplicit && !explicit {
			// will not "downgrade" an existing pin.
			return nil
		}
	} else {
		if err := pc.bk.Pin(hash); err != nil {
			return err
		}
	}

	return pc.remember(inode, hash, true, explicit)
}

func (pc *Pinner) Unpin(inode uint64, hash h.Hash, explicit bool) error {
	isPinned, isExplicit, err := pc.IsPinned(inode, hash)
	if err != nil {
		return err
	}

	if isPinned {
		if isExplicit && !explicit {
			return nil
		}

		if err := pc.bk.Unpin(hash); err != nil {
			return err
		}
	}

	return pc.remember(inode, hash, false, explicit)
}

////////////////////////////

// doPinOp recursively walks over all children of a node and pins or unpins them.
func (pc *Pinner) doPinOp(op func(uint64, h.Hash, bool) error, nd n.Node, explicit bool) error {
	return n.Walk(pc.lkr, nd, true, func(child n.Node) error {
		if child.Type() != n.NodeTypeFile {
			return nil
		}

		file, ok := child.(*n.File)
		if !ok {
			return ie.ErrBadNode
		}

		return op(file.Inode(), file.BackendHash(), explicit)
	})
}

// PinNode tries to pin the node referenced by `nd`.
// The difference to calling Pin(nd.BackendHash()) is,
// that this method will pin directories recursively, if given.
//
// If the file is already pinned exclusively and you want
// to pin it non-exclusive, this will be a no-op.
// In this case you have to unpin it first exclusively.
func (pc *Pinner) PinNode(nd n.Node, explicit bool) error {
	return pc.doPinOp(pc.Pin, nd, explicit)
}

// UnpinNode is the exact opposite of PinNode.
func (pc *Pinner) UnpinNode(nd n.Node, explicit bool) error {
	return pc.doPinOp(pc.Unpin, nd, explicit)
}

// IsNodePinned checks if all `nd` is pinned and if so, exlusively.
// If `nd` is a directory, it will only return true if all children
// are also pinned (same for second return value).
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

		isPinned, isExplicit, err := pc.IsPinned(file.Inode(), file.BackendHash())
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
