package dircache

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/sahib/brig/catfs/mio/pagecache/page"
)

type l2cache struct {
	dir string
}

func NewL2Cache(dir string) (*l2cache, error) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}

	// TODO: Implement git-like sharding by pre-mkdiring
	return &l2cache{
		dir: dir,
	}, nil
}

func (c *l2cache) Set(pk pageKey, p *page.Page) error {
	// TODO: possibly shard into different directories.
	return c.SetData(pk.String(), p.AsBytes())
}

func (c *l2cache) SetData(key string, pdata []byte) error {
	path := filepath.Join(c.dir, key)
	return ioutil.WriteFile(path, pdata, 0600)
}

func (c *l2cache) Get(pk pageKey) (*page.Page, error) {
	path := filepath.Join(c.dir, pk.String())
	pdata, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, page.ErrCacheMiss
	}

	return page.FromBytes(pdata)
}

func (c *l2cache) Del(pks []pageKey) error {
	for _, pk := range pks {
		path := filepath.Join(c.dir, pk.String())
		if err := os.Remove(path); err != nil {
			// only log, we want to get rid of more old data.
		}
	}

	return nil
}

func (c *l2cache) Close() error {
	return os.RemoveAll(c.dir)
}
