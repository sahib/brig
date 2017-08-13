package db

import (
	"errors"
	"io"
)

// TODO: Implement an actual fast KV store based on moss, boltdb or badger
//       if there is any performance problem later on.
//       For now, the filesystem based kv should suffice fine though.

var (
	// ErrNoSuchKey is returned when Get() was passed a non-existant key
	ErrNoSuchKey = errors.New("This key does not exist")
)

// Database is a key/value store that offers different buckets
// for storage. Keys are strings, values are arbitary untyped data.
type Database interface {
	// Get retrievies the key `key` out of bucket.
	// If no such key exists, it will return (nil, ErrNoSuchKey)
	Get(key ...string) ([]byte, error)

	// Set will set `val` to `key` in `bucket`.
	Put(val []byte, key ...string) error

	// Export backups all database content to `w` in
	// an implemenation specific format that can be read by Import.
	Export(w io.Writer) error

	// Import reads a previously exported db dump by Export from `r`.
	// Existing keys might be overwritten if the dump also contains them.
	Import(r io.Reader) error

	// Clear all contents below and including `key`.
	Clear(key ...string) error

	// Keys returns a channel that will yield all strings
	// this database currently knows of.
	Keys(prefix ...string) (<-chan []string, error)

	// TODO: Overthink this interface. (-> transactions)
	Erase(key ...string) error

	// Close closes the database. Since I/O may happen, an error is returned.
	Close() error
}
