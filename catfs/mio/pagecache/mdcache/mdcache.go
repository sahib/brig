package mdcache

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/sahib/brig/catfs/mio/pagecache/page"
	log "github.com/sirupsen/logrus"
)

type Options struct {
	MaxMemoryUsage int64
	SwapDirectory  string

	// TODO: Those need to be still implemented.

	// L1Compress will compress the memory with snappy compression and
	// decompress on fetch. Reduces memory, but increases CPU usage.
	L1Compress bool

	// L2Compress will compress on-disk pages with snappy and decompress them
	// on load. Reduces storage, but increases CPU usage if you're swapping.
	// Since swapping is slow anyways this is recommended.
	L2Compress bool

	// L1CacheMissRefill will propagate
	// data from L2 to L1 if it could be found
	// successfully.
	L1CacheMissRefill bool
}

type cacheLayer interface {
	Get(pk pageKey) (*page.Page, error)
	Set(pk pageKey, p *page.Page) error
	Del(pks []pageKey) error
	Close() error
}

// MDCache is a leveled Memory/Disk cache combination.
type MDCache struct {
	mu   sync.Mutex
	l1   cacheLayer
	l2   cacheLayer
	opts Options
}

type pageKey struct {
	inode   int64
	pageIdx uint32
}

func (pk pageKey) String() string {
	// TODO: That's pointless right now, it starts with 00 most of the time.

	// NOTE: Could be implemented with less allocations, but effect is
	// probably not noticeable. Only used in slow l2 cache anyways.
	s := []byte(fmt.Sprintf("%08x-%08x", pk.inode, pk.pageIdx))

	// Idea here is to split numbers 00-FF into own directories.
	// This can yield better lookup performance, depending on the fs.
	return filepath.Join(string(s[:2]), string(s[2:]))
}

func NewDirCache(opts Options) (*MDCache, error) {
	l2, err := newL2Cache(opts.SwapDirectory)
	if err != nil {
		return nil, err
	}

	var l2Iface cacheLayer = l2
	if l2 == nil {
		// special case: when we don't have a l2 cache
		// then use another memory cache as backing,
		// with infinite memory.
		maxMemory := int64(^uint64(0) >> 1)
		l2Iface, _ = newL1Cache(nil, maxMemory)
	}

	l1, err := newL1Cache(l2Iface, opts.MaxMemoryUsage)
	if err != nil {
		return nil, err
	}

	return &MDCache{
		l1:   l1,
		l2:   l2Iface,
		opts: opts,
	}, nil
}

func (dc *MDCache) Lookup(inode int64, pageIdx uint32) (*page.Page, error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	return dc.get(pageKey{inode: inode, pageIdx: pageIdx})
}

func (dc *MDCache) get(pk pageKey) (*page.Page, error) {
	p, err := dc.l1.Get(pk)
	switch err {
	case nil:
		return p, nil
	case page.ErrCacheMiss:
		p, err = dc.l2.Get(pk)
		if err != nil {
			return p, err
		}

		if dc.opts.L1CacheMissRefill {
			// propagate back to l1 cache:
			if err := dc.l1.Set(pk, p); err != nil {
				return p, err
			}
		}

		return p, err
	default:
		return nil, err
	}
}

func (dc *MDCache) Merge(inode int64, pageIdx, off uint32, write []byte) error {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	if len(write) == 0 {
		// empty write deserves no extra computation.
		return nil
	}

	if off+uint32(len(write)) > page.Size {
		return fmt.Errorf("merge: write overflows page bounds")
	}

	pk := pageKey{inode: inode, pageIdx: pageIdx}
	p, err := dc.get(pk)
	if err != nil && err != page.ErrCacheMiss {
		return err
	}

	if p == nil {
		// Page was not cached yet.
		// Create an almost empty page.
		p = page.New(off, write)
	}

	p.Overlay(off, write)
	return dc.l1.Set(pk, p)
}

func (dc *MDCache) Evict(inode, size int64) error {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	// Figure out all possible indices from size:
	pks := []pageKey{}
	pageHi := uint32(size / page.Size)
	if size%page.Size > 0 {
		pageHi++
	}

	for pageIdx := uint32(0); pageIdx < pageHi; pageIdx++ {
		pks = append(pks, pageKey{inode: inode, pageIdx: pageIdx})
	}

	if err := dc.l1.Del(pks); err != nil {
		log.WithError(err).Warnf("l1 delete failed for %v", pks)
	}

	// TODO: This will spam logs in case of no page:
	if err := dc.l2.Del(pks); err != nil {
		log.WithError(err).Warnf("l2 delete failed for %v", pks)
	}

	return nil
}

func (dc *MDCache) Close() error {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	if err := dc.l1.Close(); err != nil {
		log.WithError(err).Warnf("failed to reset l1 cache")
	}

	return dc.l2.Close()
}
