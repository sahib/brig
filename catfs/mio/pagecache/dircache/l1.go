package dircache

import (
	"time"

	"github.com/allegro/bigcache"
	"github.com/sahib/brig/catfs/mio/pagecache/page"
	log "github.com/sirupsen/logrus"
)

// NOTE: We use the bigcache library here. I initially wanted to go with
// ristretto, since I heard good things about it. It seems though that it does
// not guarantee that keys are actually inserted into the cache and there is no
// efficient way to figure out if the entry was inserted or not.
//
// This is however a crucial property we need from our cache. When the memory
// cache evicts an item due to memory constraints then we need it to the
// persistent cache of l2. Bigcache seems to be the only cache library offering
// this. A little drawback is that it requires (de-)serialization of the
// cached items. But we can implement this is in a not so costly way.
//
// One possible way could also be to use sync.Map and implement our own
// eviction logic on top of that. Not sure if that's worth the trouble though.

type l1cache struct {
	big *bigcache.BigCache
}

func NewL1Cache(l2 *l2cache, maxMemoryMB int64) (*l1cache, error) {
	// NOTE: bigcache requires setting an expiry time.
	//       Just set it to the highest possible time to effectively disable it.
	expiry := time.Duration(^uint64(0) >> 1)

	cfg := bigcache.DefaultConfig(expiry)
	cfg.CleanWindow = 0
	cfg.HardMaxCacheSize = int(maxMemoryMB)

	cfg.OnRemoveWithReason = func(key string, entry []byte, reason bigcache.RemoveReason) {
		if reason == bigcache.Deleted {
			// if we deleted on purpose we should not write to l2
			// of course.
			return
		}

		if err := l2.SetData(key, entry); err != nil {
			log.WithError(err).Warnf("failed to move »%s« to l2", key)
		}
	}

	big, err := bigcache.NewBigCache(cfg)
	if err != nil {
		return nil, err
	}

	return &l1cache{
		big: big,
	}, nil
}

func (c *l1cache) Set(pk pageKey, p *page.Page) error {
	// NOTE: Item should be written always.
	//       If an item has to go, OnRemoveWithReason will be called.
	return c.big.Set(pk.String(), p.AsBytes())
}

func (c *l1cache) Get(pk pageKey) (*page.Page, error) {
	pdata, err := c.big.Get(pk.String())
	if err != nil {
		if err == bigcache.ErrEntryNotFound {
			return nil, page.ErrCacheMiss
		}

		return nil, err
	}

	return page.FromBytes(pdata)
}

func (c *l1cache) Del(pks []pageKey) error {
	for _, pk := range pks {
		c.big.Delete(pk.String())
	}

	return nil
}

func (c *l1cache) Close() error {
	return c.big.Reset()
}
