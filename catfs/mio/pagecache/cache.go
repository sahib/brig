package overlay

import (
	"errors"
)

var (
	// not a real error, do not pass to outside.
	ErrCacheMiss = errors.New("cache miss")
)

//////

type Cache interface {
	Lookup(inode, page int32) (*Page, error)
	Merge(inode, pageID int32, off int32, buf []byte) error
	Close() error
}

// ///////
//
// type PageCache struct {
// 	bdb *badger.DB
// }
//
// func NewPageCache(path string) (*PageCache, error) {
// 	opts := badger.
// 		DefaultOptions(path).
// 		WithValueLogFileSize(10 * 1024 * 1024).
// 		WithMemTableSize(10 * 1024 * 1024).
// 		WithSyncWrites(false).
// 		WithLogger(nil)
//
// 	bdb, err := badger.Open(opts)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return &PageCache{
// 		bdb: bdb,
// 	}, nil
// }
//
// func (pc *PageCache) Lookup(id PageID) ([]byte, error) {
// 	return nil, nil
// }
//
// func (pc *PageCache) Store(id PageID, page []byte) error {
// 	return nil
// }
//
// func (pc *PageCache) Forget(inode int64) error {
// 	// either directly delete all or set a TTL.
// 	return nil
// }
