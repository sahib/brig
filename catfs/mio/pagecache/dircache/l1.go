package dircache

import (
	"container/list"
	"errors"
	"fmt"

	"github.com/sahib/brig/catfs/mio/pagecache/page"
)

// L1 is a pure in-memory LRU cache which does no copying.
// I did go for LRU because it's insanely simple and easy to implement
// while still being quite effective.
//
// NOTE: We do not use one of the popular caching library here, since
// none of them seem to fit our use-case. We require the following properties:
//
// 1. We must notice when items get evicted (in order to write to l2)
// 2. We must be able to set a max memory bound.
// 3. We must avoid copying of pages due to performance reasons.
//
// The most popular libraries fail always one of the criterias:
//
// - fastcache: fails 1 and 3.
// - ristretto: fails 1.
// - bigcache: fails 3.
//
// Since we know what kind of data we cache, it is reasonable to implement
// a very basic LRU cache for L1. Therefore we just use sync.Map here.
// Oh, and the l1cache is not thread safe, but dircache.go does locking.

type l1item struct {
	Page *page.Page
	Link *list.Element
}

type l1cache struct {
	m         map[pageKey]l1item
	k         *list.List
	l2        *l2cache
	maxMemory int64
}

func NewL1Cache(l2 *l2cache, maxMemory int64) (*l1cache, error) {
	return &l1cache{
		maxMemory: maxMemory,
		l2:        l2,
		k:         list.New(),
	}, nil
}

func (c *l1cache) Set(pk pageKey, p *page.Page) error {
	c.m[pk] = l1item{
		Page: p,
		Link: c.k.PushBack(pk),
	}

	maxPages := c.maxMemory / (page.Size + page.Meta)
	if int64(len(c.m)) > maxPages {
		if c.l2 == nil {
			// just in case l2 cache was not given:
			return errors.New("cache is full")
		}

		oldPkIface := c.k.Remove(c.k.Front())
		oldPk, ok := oldPkIface.(pageKey)
		if !ok {
			return fmt.Errorf("non-pagekey type stored in l1 keys: %T", oldPkIface)
		}

		oldItem, ok := c.m[oldPk]
		delete(c.m, oldPk)
		if !ok {
			// c.m and c.k got out of sync.
			// this is very likely a bug.
			return fmt.Errorf("l1: key in key list, but not in map")
		}

		// move old page to more persistent cache layer:
		return c.l2.Set(oldPk, oldItem.Page)
	}

	return nil
}

func (c *l1cache) Get(pk pageKey) (*page.Page, error) {
	item, ok := c.m[pk]
	if !ok {
		return nil, page.ErrCacheMiss
	}

	// Sort recently fetched item to end of list:
	c.k.MoveToBack(item.Link)
	return item.Page, nil
}

func (c *l1cache) Del(pks []pageKey) error {
	for _, pk := range pks {
		delItem, ok := c.m[pk]
		if ok {
			c.k.Remove(delItem.Link)
			delete(c.m, pk)
		}
	}

	return nil
}

func (c *l1cache) Close() error {
	return nil
}
