package dircache

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/sahib/brig/catfs/mio/pagecache/page"
)

type l2cache struct {
	dir string
}

// NOTE: an empty (nil) l2cache is valid, but will not do anything. If an
// empty string for `dir` is given, such an empty l2cache will be returned.
func newL2Cache(dir string) (*l2cache, error) {
	if dir == "" {
		return nil, nil
	}

	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}

	for idx := 0; idx < 256; idx++ {
		shard := filepath.Join(dir, fmt.Sprintf("%02x", idx))
		if err := os.MkdirAll(shard, 0700); err != nil {
			return nil, err
		}
	}

	return &l2cache{dir: dir}, nil
}

func (c *l2cache) Set(pk pageKey, p *page.Page) error {
	return c.SetData(pk.String(), p.AsBytes())
}

func (c *l2cache) SetData(key string, pdata []byte) error {
	if c == nil {
		return nil
	}

	path := filepath.Join(c.dir, key)
	return ioutil.WriteFile(path, pdata, 0600)
}

func (c *l2cache) Get(pk pageKey) (*page.Page, error) {
	if c == nil {
		return nil, page.ErrCacheMiss
	}

	path := filepath.Join(c.dir, pk.String())
	pdata, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, page.ErrCacheMiss
	}

	return page.FromBytes(pdata)
}

func (c *l2cache) Del(pks []pageKey) error {
	if c == nil {
		return nil
	}

	for _, pk := range pks {
		path := filepath.Join(c.dir, pk.String())
		if err := os.Remove(path); err != nil {
			// only log, we want to get rid of more old data.
		}
	}

	return nil
}

func (c *l2cache) Close() error {
	if c == nil {
		return nil
	}

	return os.RemoveAll(c.dir)
}
