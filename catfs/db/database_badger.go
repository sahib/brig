package db

import (
	"fmt"
	"gx/ipfs/QmeAEa8FDWAmZJTL6YcM1oEndZ4MyhCr5rTsjYZQui1x1L/badger"
	"io"
	"strings"
	"sync"
)

// TODO: Implement an actual fast KV store based on moss, boltdb or badger
//       if there is any performance problem later on.
//       For now, the filesystem based kv should suffice fine though.

type BadgerDatabase struct {
	mu         sync.Mutex
	db         *badger.DB
	txn        *badger.Txn
	haveWrites bool
}

func NewBadgerDatabase(path string) (*BadgerDatabase, error) {
	opts := badger.DefaultOptions
	opts.Dir = path
	opts.ValueDir = path

	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	return &BadgerDatabase{
		db: db,
	}, nil
}

func (db *BadgerDatabase) Get(key ...string) ([]byte, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	data := []byte{}
	err := db.db.View(func(txn *badger.Txn) error {
		keyPath := strings.Join(key, ".")
		fmt.Println("PATH", keyPath)
		item, err := txn.Get([]byte(keyPath))
		if err == badger.ErrKeyNotFound {
			return ErrNoSuchKey
		}

		if err != nil {
			return err
		}

		data, err = item.Value()
		return err
	})

	if err != nil {
		return nil, err
	}

	return data, nil
}

func (db *BadgerDatabase) Keys(fn func(key []string) error, prefix ...string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	return db.db.View(func(txn *badger.Txn) error {
		iter := txn.NewIterator(badger.IteratorOptions{})
		defer iter.Close()

		for iter.Valid() {
			item := iter.Item()

			fullKey := string(item.Key())
			if err := fn(strings.Split(fullKey, ".")); err != nil {
				return err
			}

			iter.Next()
		}

		return nil
	})
}

func (db *BadgerDatabase) Export(w io.Writer) error {
	_, err := db.db.Backup(w, 0)
	return err
}

func (db *BadgerDatabase) Import(r io.Reader) error {
	return db.db.Load(r)
}

func (db *BadgerDatabase) Glob(prefix []string) ([][]string, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	results := [][]string{}
	err := db.db.View(func(txn *badger.Txn) error {
		iter := txn.NewIterator(badger.IteratorOptions{})
		defer iter.Close()

		fullPrefix := strings.Join(prefix, ".")
		fmt.Println("go key", fullPrefix)

		for iter.Valid() {
			item := iter.Item()

			fullKey := string(item.Key())
			fmt.Println("key", fullKey)
			if strings.HasPrefix(fullKey, fullPrefix) {
				results = append(results, strings.Split(fullKey, "."))
			}

			iter.Next()
		}

		return nil
	})

	return results, err
}

func (db *BadgerDatabase) Batch() Batch {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.txn == nil {
		db.txn = db.db.NewTransaction(true)
	}

	return db
}

func (db *BadgerDatabase) Put(val []byte, key ...string) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.haveWrites = true

	fullKey := []byte(strings.Join(key, "."))
	db.txn.Set(fullKey, val)

	// TODO: handle error?
}

func (db *BadgerDatabase) Clear(key ...string) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.haveWrites = true

	iter := db.txn.NewIterator(badger.IteratorOptions{})
	defer iter.Close()

	keys := [][]byte{}
	for iter.Valid() {
		item := iter.Item()
		keys = append(keys, item.Key())
		iter.Next()
	}

	for _, key := range keys {
		if err := db.txn.Delete(key); err != nil {
			fmt.Println("TODO: delete failed", err)
			// TODO: handle error?
		}
	}
}

func (db *BadgerDatabase) Erase(key ...string) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.haveWrites = true

	fullKey := []byte(strings.Join(key, "."))
	db.txn.Delete(fullKey)
}

func (db *BadgerDatabase) Flush() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if err := db.txn.Commit(nil); err != nil {
		return err
	}

	db.txn = nil
	db.haveWrites = false
	return nil
}

func (db *BadgerDatabase) Rollback() {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.txn.Discard()
	db.txn = nil
	db.haveWrites = false
}

func (db *BadgerDatabase) HaveWrites() bool {
	db.mu.Lock()
	defer db.mu.Unlock()

	return db.haveWrites
}

func (db *BadgerDatabase) Close() error {
	return db.db.Close()
}
