package store

import (
	"fmt"

	"github.com/boltdb/bolt"
)

type bucketHandler func(tx *bolt.Tx, b *bolt.Bucket) error

// withBucket wraps a bolt handler closure and passes a named bucket
// as extra parameter. Error handling is done universally for convinience.
func withBucket(name string, handler bucketHandler) func(tx *bolt.Tx) error {
	return func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(name))
		if bucket == nil {
			return ErrNoSuchBucket{name}
		}

		return handler(tx, bucket)
	}
}

func (s *Store) updateWithBucket(name string, handler bucketHandler) error {
	return s.db.Update(withBucket(name, handler))
}

func (s *Store) viewWithBucket(name string, handler bucketHandler) error {
	return s.db.View(withBucket(name, handler))
}

type ErrNoSuchBucket struct {
	Name string
}

func (e ErrNoSuchBucket) Error() string {
	return fmt.Sprintf("bolt: no bucket named `%s`", e.Name)
}
