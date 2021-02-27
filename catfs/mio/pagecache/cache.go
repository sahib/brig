package overlay

import (
	"github.com/sahib/brig/catfs/mio/pagecache/page"
)

type Cache interface {
	Evict(inode int32) error
	Lookup(inode, page int32) (*page.Page, error)
	Merge(inode, pageID, off int32, buf []byte) error
	Close() error
}
