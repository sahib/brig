package mdcache

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/golang/snappy"
	"github.com/sahib/brig/catfs/mio/pagecache/page"
)

// NOTE: Room for improvement:
// Introduce heuristic for snappy compression:
// When an inode does not yield any noticeable compression,
// disable it for this specific inode. Protects against
// wasting CPU on already compressed data.

type l2cache struct {
	mu       sync.Mutex
	dir      string
	compress bool
	zipBuf   []byte
}

// NOTE: an empty (nil) l2cache is valid, but will not do anything. If an
// empty string for `dir` is given, such an empty l2cache will be returned.
func newL2Cache(dir string, compress bool) (*l2cache, error) {
	if dir == "" {
		return nil, nil
	}

	var zipBuf []byte
	if compress {
		zipBuf = make([]byte, snappy.MaxEncodedLen(page.Size))
	}

	return &l2cache{
		dir:      dir,
		compress: compress,
		zipBuf:   zipBuf,
	}, nil
}

func (c *l2cache) Set(pk pageKey, p *page.Page) error {
	if c == nil {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	data := p.AsBytes()
	if c.compress {
		data = snappy.Encode(c.zipBuf, p.AsBytes())
	}

	path := filepath.Join(c.dir, pk.String())
	return ioutil.WriteFile(path, data, 0600)
}

func (c *l2cache) Get(pk pageKey) (*page.Page, error) {
	if c == nil {
		return nil, page.ErrCacheMiss
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	path := filepath.Join(c.dir, pk.String())
	pdata, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, page.ErrCacheMiss
	}

	if c.compress {
		pdata, err = snappy.Decode(c.zipBuf, pdata)
		if err != nil {
			return nil, err
		}
	}

	return page.FromBytes(pdata)
}

func (c *l2cache) Del(pks []pageKey) {
	if c == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	for _, pk := range pks {
		path := filepath.Join(c.dir, pk.String())

		// no error handling, just get rid of things.
		// we can't do anything if it could not be deleted.
		os.Remove(path)
	}
}

func (c *l2cache) Close() error {
	if c == nil {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	return os.RemoveAll(c.dir)
}
