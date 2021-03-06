package overlay

import (
	"github.com/sahib/brig/catfs/mio/pagecache/page"
)

// Cache is the backing layer that stores pages in memory
// or whatever medium it choses to use.
type Cache interface {
	// Lookup returns a cached page, identified by `inode` and `page`.
	// If there is no such page page.ErrCacheMiss is returned.
	Lookup(inode int64, page uint32) (*page.Page, error)

	// Merge the existing cache contents with the new write
	// to `pageID`, starting at `pageOff` and with the contents of `buf`.
	Merge(inode int64, pageID, pageOff uint32, buf []byte) error

	// Evict clears cached pages for `inode`. `size` can be used
	// to clear only up to a certain size.
	Evict(inode, size int64) error

	// Close the cache and free up all resources.
	Close() error
}
