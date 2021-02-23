package dircache

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
}

type DirCache struct {
	mu sync.Mutex
	l1 *l1cache
	l2 *l2cache
}

type pageKey struct {
	inode   int32
	pageIdx int32
}

func (pk pageKey) String() string {
	// NOTE: Could be implemented with less allocations, but effect is
	// probably not noticeable. Only used in slow l2 cache anyways.
	s := []byte(fmt.Sprintf("%08x-%08x", pk.inode, pk.pageIdx))

	// Idea here is to split numbers 00-FF into own directories.
	// This can yield better lookup performance, depending on the fs.
	return filepath.Join(string(s[:2]), string(s[2:]))
}

func NewDirCache(opts Options) (*DirCache, error) {
	l2, err := NewL2Cache(opts.SwapDirectory)
	if err != nil {
		return nil, err
	}

	l1, err := NewL1Cache(l2, opts.MaxMemoryUsage)
	if err != nil {
		return nil, err
	}

	return &DirCache{
		l1: l1,
		l2: l2,
	}, nil
}

func (dc *DirCache) Lookup(inode, pageIdx int32) (*page.Page, error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	return dc.get(pageKey{inode: inode, pageIdx: pageIdx})
}

func (dc *DirCache) get(pk pageKey) (*page.Page, error) {
	p, err := dc.l1.Get(pk)
	if err != nil {
		if err == page.ErrCacheMiss {
			// TODO: attempt propagate to l1?
			return dc.l2.Get(pk)
		}

		return nil, err
	}

	return p, nil
}

func (dc *DirCache) Merge(inode, pageIdx, off int32, write []byte) error {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	if len(write) == 0 {
		// empty write deserves no extra computation.
		return nil
	}

	if off+int32(len(write)) > page.Size {
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
	} else {
		p.AddExtent(off, write)
	}

	return dc.l1.Set(pk, p)
}

func (dc *DirCache) Evict(inode int32, size int64) error {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	pks := []pageKey{}
	pageHi := int32(size / page.Size)
	for pageIdx := int32(0); pageIdx <= pageHi; pageIdx++ {
		pks = append(pks, pageKey{inode: inode, pageIdx: pageIdx})
	}

	if err := dc.l1.Del(pks); err != nil {
		log.WithError(err).Warnf("l1 delete failed for %v", pks)
	}

	if err := dc.l2.Del(pks); err != nil {
		log.WithError(err).Warnf("l2 delete failed for %v", pks)
	}

	return nil
}

func (dc *DirCache) Close() error {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	if err := dc.l1.Close(); err != nil {
		log.WithError(err).Warnf("failed to reset l1 cache")
	}

	return dc.l2.Close()
}
