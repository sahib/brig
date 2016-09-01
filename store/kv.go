package store

import (
	"os"
	"path/filepath"
	"time"

	"github.com/boltdb/bolt"
)

type KV interface {
	Bucket(name string) (Bucket, error)
	Close() error
}

type Bucket interface {
	Get(key string) ([]byte, error)
	Put(key string, data []byte) error
	Bucket(name string) (Bucket, error)
}

// Utility, not interface:
// GetPath(path string) ([]byte, error)
// SetPath(path string, data) error

// GetNode(path string, nd Node) error
// SetNode(path string, nd Node) error

type BoltKV struct {
	db *bolt.DB
}

func NewBoltKV(path string) (KV, error) {
	options := &bolt.Options{Timeout: 1 * time.Second}

	if err := os.MkdirAll(path, 0777); err != nil {
		return nil, err
	}

	db, err := bolt.Open(filepath.Join(path, "index.bolt"), 0600, options)
	if err != nil {
		return nil, err
	}

	return &BoltKV{db}, nil
}

func (kv *BoltKV) Bucket(name string) (Bucket, error) {
	err := kv.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(name))
		return err
	})

	if err != nil {
		return nil, err
	}

	return &BoltBucket{kv.db, []string{name}}, nil
}

func (kv *BoltKV) Close() error {
	return kv.db.Close()
}

type BoltBucket struct {
	db   *bolt.DB
	path []string
}

// dig to the correct bucket at bb.path
func (bb *BoltBucket) dig(path []string, writable bool, fn func(bucket *bolt.Bucket) error) error {
	if len(path) == 0 {
		return nil
	}

	tx, err := bb.db.Begin(writable)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	// Iterate down to the right bucket:
	buck := tx.Bucket([]byte(path[0]))
	for iter := path[1:]; len(iter) > 0; iter = iter[1:] {
		buck = buck.Bucket([]byte(iter[0]))
	}

	if err := fn(buck); err != nil {
		return err
	}

	if writable {
		// Commit the transaction and check for error.
		if err := tx.Commit(); err != nil {
			return err
		}
	}

	return nil
}

func (bb *BoltBucket) Get(key string) (data []byte, err error) {
	err = bb.dig(bb.path, false, func(bucket *bolt.Bucket) error {
		data = bucket.Get([]byte(key))
		return nil
	})

	// Silence `data` if some error happened:
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (bb *BoltBucket) Put(key string, data []byte) error {
	return bb.dig(bb.path, true, func(bucket *bolt.Bucket) error {
		return bucket.Put([]byte(key), data)
	})
}

func (bb *BoltBucket) Bucket(name string) (Bucket, error) {
	err := bb.dig(bb.path, true, func(bucket *bolt.Bucket) error {
		_, err := bucket.CreateBucketIfNotExists([]byte(name))
		return err
	})

	// Silence `data` if some error happened:
	if err != nil {
		return nil, err
	}

	return &BoltBucket{
		db:   bb.db,
		path: append(bb.path, name),
	}, nil
}
