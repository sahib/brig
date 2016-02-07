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
			return fmt.Errorf("index: No bucket named `%s`", name)
		}

		return handler(tx, bucket)
	}
}
