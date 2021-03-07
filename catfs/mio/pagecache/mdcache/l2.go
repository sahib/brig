package mdcache

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/sahib/brig/catfs/mio/pagecache/page"
	log "github.com/sirupsen/logrus"
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
	if c == nil {
		return nil
	}

	path := filepath.Join(c.dir, pk.String())
	return ioutil.WriteFile(path, p.AsBytes(), 0600)
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

func (c *l2cache) Del(pks []pageKey) {
	if c == nil {
		return
	}

	for _, pk := range pks {
		path := filepath.Join(c.dir, pk.String())
		if err := os.Remove(path); err != nil {
			// only log, we want to get rid of more old data.
			log.Warnf("page l2: failed to delete %s", path)
		}
	}
}

func (c *l2cache) Close() error {
	if c == nil {
		return nil
	}

	return os.RemoveAll(c.dir)
}
