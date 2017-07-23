package db

import (
	"errors"
	"io"
)

var (
	ErrNoSuchKey = errors.New("This key does not exist in this bucket")
)

// Database is a key/value store that offers different buckets
// for storage. Keys are strings, values are arbitary untyped data.
type Database interface {
	// Get retrievies the key `key` out of bucket.
	// If no such key exists, it will return (nil, ErrNoSuchKey)
	Get(bucket string, key string) ([]byte, error)

	// Set will set `val` to `key` in `bucket`.
	Set(bucket string, key string, val []byte) error

	// Export backups all database content to `w` in
	// an implemenation specific format that can be read by Import.
	Export(w io.Writer) error

	// Import reads a previously exported db dump by Export from `r`.
	// Existing keys might be overwritten if the dump also contains them.
	Import(r io.Reader) error

	// Close closes the database. Since I/O may happen, an error is returned.
	Close() error
}
