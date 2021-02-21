package dircache

import (
	"fmt"

	"github.com/sahib/brig/catfs/mio/pagecache/page"
	log "github.com/sirupsen/logrus"
)

type Options struct {
	MaxMemoryUsage int64
	SwapDirectory  string
}

type DirCache struct {
	l1 *l1cache
	l2 *l2cache
}

type pageKey struct {
	inode   int32
	pageIdx int32
}

func (pk pageKey) String() string {
	// TODO: encode as hex, but insert a / after the first two characters.
	// this will help sharding.
	return fmt.Sprintf("%d-%d", pk.inode, pk.pageIdx)
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
	return dc.get(pageKey{inode: inode, pageIdx: pageIdx})
}

func (dc *DirCache) get(pk pageKey) (*page.Page, error) {
	p, err := dc.l1.Get(pk)
	if err != nil {
		if err == page.ErrCacheMiss {
			return dc.l2.Get(pk)
		}

		return nil, err
	}

	return p, nil
}

func (dc *DirCache) Merge(inode, pageIdx, off int32, write []byte) error {
	if len(write) == 0 {
		// empty write deserve no extra computation.
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
		// No cached page yet. Let's add an empty page.
		var extents []page.Extent
		if len(write) != page.Size {
			extents = append(extents, page.Extent{
				OffLo: off,
				OffHi: off + int32(len(write)),
			})
		}

		p = &page.Page{
			Data:    make([]byte, page.Size),
			Extents: extents,
		}
	} else {
		// TODO: Do merging magic.
	}

	return dc.l1.Set(pk, p)
}

func (dc *DirCache) Evict(inode int32, size int64) error {
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
	if err := dc.l1.Close(); err != nil {
		log.WithError(err).Warnf("failed to reset l1 cache")
	}

	return dc.l2.Close()
}
