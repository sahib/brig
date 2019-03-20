package db

import (
	"errors"
	"io"
)

var (
	// ErrNoSuchKey is returned when Get() was passed a non-existent key
	ErrNoSuchKey = errors.New("This key does not exist")
)

// Batch is an API object used to model a transaction.
type Batch interface {
	// Put sets `val` at `key`.
	Put(val []byte, key ...string)

	// Clear all contents below and including `key`.
	Clear(key ...string) error

	// Erase a key from the database.
	Erase(key ...string)

	// Flush the batch to the database.
	// Only now, all changes will be written to disk.
	Flush() error

	// Rollback will forget all changes without executing them.
	Rollback()

	// HaveWrites returns true when the batch contains something
	// we can write to the disk on Flush().
	HaveWrites() bool
}

// Database is a key/value store that offers different buckets
// for storage. Keys are strings, values are arbitrary untyped data.
type Database interface {
	// Get retrievies the key `key` out of bucket.
	// If no such key exists, it will return (nil, ErrNoSuchKey)
	// If a badge is currently open, Get() shall still return the
	// most current value currently set by the last Put() call
	// to `key`.
	Get(key ...string) ([]byte, error)

	// Keys iterates over all keys in the database If the error is returned by
	// `fn` the iteration stops and the error value is returned.
	// The keys are returned in lexical ordering.
	Keys(prefix ...string) ([][]string, error)

	// Batch returns a new Batch object, that will allow modifications
	// of the state. Batch() can be called recursive: The changes will
	// only be flushed to disk if batch.Flush() was called equal times
	// to the number Batch() was called.
	Batch() Batch

	// Export backups all database content to `w` in
	// an implemenation specific format that can be read by Import.
	Export(w io.Writer) error

	// Import reads a previously exported db dump by Export from `r`.
	// Existing keys might be overwritten if the dump also contains them.
	Import(r io.Reader) error

	// Close closes the database. Since I/O may happen, an error is returned.
	Close() error

	// Glob finds all existing keys in the store, starting with prefix.
	Glob(prefix []string) ([][]string, error)
}

// CopyKey is a helper method to copy a bunch of keys in `src` to `dst`.
func CopyKey(db Database, src, dst []string) error {
	data, err := db.Get(src...)
	if err != nil {
		return err
	}

	batch := db.Batch()
	batch.Put(data, dst...)
	return batch.Flush()
}
