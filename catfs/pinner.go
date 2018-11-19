package catfs

import (
	"errors"

	capnp "github.com/sahib/brig/catfs/capnp"
	c "github.com/sahib/brig/catfs/core"
	"github.com/sahib/brig/catfs/db"
	ie "github.com/sahib/brig/catfs/errors"
	n "github.com/sahib/brig/catfs/nodes"
	h "github.com/sahib/brig/util/hashlib"
	capnp_lib "zombiezen.com/go/capnproto2"
)

// errNotPinnedSentinel is returned to signal an early exit in Walk()
var errNotPinnedSentinel = errors.New("not pinned")

// pinCacheEntry is one entry in the pin cache.
type pinCacheEntry struct {
	Inodes map[uint64]bool
}

func capnpToPinCacheEntry(data []byte) (*pinCacheEntry, error) {
	msg, err := capnp_lib.Unmarshal(data)
	if err != nil {
		return nil, err
	}

	capEntry, err := capnp.ReadRootPinEntry(msg)
	if err != nil {
		return nil, err
	}

	capPins, err := capEntry.Pins()
	if err != nil {
		return nil, err
	}

	entry := &pinCacheEntry{
		Inodes: make(map[uint64]bool),
	}

	for idx := 0; idx < capPins.Len(); idx++ {
		capPin := capPins.At(idx)
		entry.Inodes[capPin.Inode()] = capPin.IsPinned()
	}

	return entry, nil
}

func pinEnryToCapnpData(entry *pinCacheEntry) ([]byte, error) {
	msg, seg, err := capnp_lib.NewMessage(capnp_lib.SingleSegment(nil))
	if err != nil {
		return nil, err
	}

	capEntry, err := capnp.NewRootPinEntry(seg)
	if err != nil {
		return nil, err
	}

	capPinList, err := capnp.NewPin_List(seg, int32(len(entry.Inodes)))
	if err != nil {
		return nil, err
	}

	idx := 0
	for inode, isPinned := range entry.Inodes {
		capPin, err := capnp.NewPin(seg)
		if err != nil {
			return nil, err
		}

		capPin.SetInode(inode)
		capPin.SetIsPinned(isPinned)

		if err := capPinList.Set(idx, capPin); err != nil {
			return nil, err
		}

		idx++
	}

	if err := capEntry.SetPins(capPinList); err != nil {
		return nil, err
	}

	return msg.Marshal()
}

// Pinner remembers which hashes are pinned and if they are pinned explicitly.
// Its API can be used to safely change the pinning state. It assumes that it
// is the only entitiy the pins & unpins nodes.
type Pinner struct {
	bk  FsBackend
	lkr *c.Linker
}

// NewPinner creates a new pin cache at `pinDbPath`, possibly erroring out.
// `lkr` and `bk` are used to make PinNode() and UnpinNode() work.
func NewPinner(lkr *c.Linker, bk FsBackend) (*Pinner, error) {
	return &Pinner{lkr: lkr, bk: bk}, nil
}

// Close the pinning cache.
func (pc *Pinner) Close() error {
	// currently a no-op
	return nil
}

func getEntry(kv db.Database, hash h.Hash) (*pinCacheEntry, error) {
	data, err := kv.Get("pins", hash.B58String())
	if err != nil {
		if err == db.ErrNoSuchKey {
			return nil, nil
		}

		return nil, err
	}

	return capnpToPinCacheEntry(data)
}

// remember the pin state of a certain hash.
// This does change anything in the backend but only changes the caching structure.
// Use with care to avoid data inconsistencies.
func (pc *Pinner) remember(inode uint64, hash h.Hash, isPinned, isExplicit bool) error {
	return pc.lkr.AtomicWithBatch(func(batch db.Batch) (bool, error) {
		oldEntry, err := getEntry(pc.lkr.KV(), hash)
		if err != nil {
			return true, err
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

		data, err := pinEnryToCapnpData(&entry)
		if err != nil {
			return true, err
		}

		batch.Put(data, "pins", hash.B58String())
		return false, nil
	})
}

func (pc *Pinner) IsPinned(inode uint64, hash h.Hash) (bool, bool, error) {
	data, err := pc.lkr.KV().Get("pins", hash.B58String())
	if err != nil && err != db.ErrNoSuchKey {
		return false, false, err
	}

	if err == nil {
		// cache hit
		entry, err := capnpToPinCacheEntry(data)
		if err != nil {
			return false, false, err
		}

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

	return isPinned, false, nil
}

////////////////////////////

// Pin will remember the node at `inode` with hash `hash` as `explicit`ly pinned.
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
