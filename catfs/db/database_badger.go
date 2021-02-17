package db

import (
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	badger "github.com/dgraph-io/badger/v3"

	log "github.com/sirupsen/logrus"
)

// BadgerDatabase is a database implementation based on BadgerDB
type BadgerDatabase struct {
	mu         sync.Mutex
	isStopped  int64
	db         *badger.DB
	txn        *badger.Txn
	refCount   int
	haveWrites bool
	writeTimes []time.Time
	gcTicker   *time.Ticker
}

// NewBadgerDatabase creates a new badger database.
func NewBadgerDatabase(path string) (*BadgerDatabase, error) {
	opts := badger.DefaultOptions(path).
		WithValueLogFileSize(10 * 1024 * 1024). //default is 2GB we should not need 2GB
		WithMemTableSize(10 * 1024 * 1024).     //default is 64MB
		WithNumVersionsToKeep(1).               // it is default but it's better to force it
		WithCompactL0OnClose(true).
		WithSyncWrites(false).
		WithLogger(nil)

	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	gcTicker := time.NewTicker(5 * time.Minute)

	bdb := &BadgerDatabase{
		db:       db,
		gcTicker: gcTicker,
	}

	go func() {
		for range gcTicker.C {
			if atomic.LoadInt64(&bdb.isStopped) > 0 {
				return
			}

			err := bdb.runGC()
			if err != nil {
				log.WithError(err).Error("badger GC failed")
			}
		}
	}()

	return bdb, nil
}

func (bdb *BadgerDatabase) runGC() error {
	opts := bdb.db.Opts()
	bdb.mu.Lock()
	defer bdb.mu.Unlock()
	log.Debugf("Performing GC for badger DB in %s", opts.Dir)
	tStart := time.Now()
	defer func() {
		log.Debugf("GC collection on %s took %v", opts.Dir, time.Now().Sub(tStart))
	}()
	// we will go through array of write times to see if it is time to run GC
	var gcStatsUpdateDelay = 5 * time.Minute
	var deadlineMet = false
	n := 0
	for _, t := range bdb.writeTimes {
		if time.Now().Before(t.Add(gcStatsUpdateDelay)) {
			bdb.writeTimes[n] = t
			n++
		} else {
			deadlineMet = true
		}
	}
	bdb.writeTimes = bdb.writeTimes[:n]
	if !deadlineMet {
		log.Debugf("DB in %s has no new stats for GC", opts.Dir)
		return nil
	}
	// In large DB, GC will happen automatically, because compaction will find garbage
	// but we are to small and compactors do not run (150 MB is small).
	// So we need to run Flatten
	bdb.db.Flatten(5)
	// Very likely Flatten will not do much because the hard coded priority is too small.
	// At this point, we hope that there is something for GC
	var errGC error
	var success = false
	for errGC == nil {
		errGC = bdb.db.RunValueLogGC(0.5) // cleans DB online and it is safe to rerun on success
		if errGC == nil && !success {
			success = true
		}
	}
	if success {
		log.Debugf("Cleaned some garbage for DB in %s", opts.Dir)
		return nil
	}
	// Now we have a dilemma: we could trust badger GC mechanism and stop here.
	// But unfortunately for our typical size (even as high as 150 MB)
	// compaction, even with Flatten(), does not kick in.
	// The only way to truly force compaction (to update stats for GC) is to close DB
	// see Note in https://github.com/dgraph-io/badger/issues/767#issuecomment-485713746
	// After Close() the GC on a next run will have updated statistic
	// Actually even Close() does not guaranteed success, it requires more than a minute
	// to update stats after DB was modified. But eventually GC stats will be ready.
	if bdb.txn != nil {
		// someone still using DB, we will try to Close/Open next time
		return nil
	}
	err := bdb.db.Close()
	if err != nil {
		// something prevent the Close, no worries we will try another time
		log.Errorf("Could not close DB in %s", opts.Dir)
		return err
	}
	var cnt = 10
	for cnt > 0 {
		cnt--
		bdb.db, err = badger.Open(opts)
		if err == nil {
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	log.Fatalf("Could not reopen DB in %s", opts.Dir)
	return err
}

func (db *BadgerDatabase) view(fn func(txn *badger.Txn) error) error {
	// If we have an open transaction, retrieve the values from there.
	// Otherwise we would not be able to retrieve in-memory values.
	if db.txn != nil {
		return fn(db.txn)
	}

	// If no transaction is running (no Batch()-call), use a fresh view txn.
	return db.db.View(fn)
}

// Get is the badger implementation of Database.Get.
func (db *BadgerDatabase) Get(key ...string) ([]byte, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	data := []byte{}
	err := db.view(func(txn *badger.Txn) error {
		if db.txn != nil {
			txn = db.txn
		}

		keyPath := strings.Join(key, ".")
		item, err := txn.Get([]byte(keyPath))
		if err == badger.ErrKeyNotFound {
			return ErrNoSuchKey
		}

		if err != nil {
			return err
		}

		data, err = item.ValueCopy(nil)
		return err
	})

	if err != nil {
		return nil, err
	}

	return data, nil
}

// Keys is the badger implementation of Database.Keys.
func (db *BadgerDatabase) Keys(prefix ...string) ([][]string, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	keys := [][]string{}
	return keys, db.view(func(txn *badger.Txn) error {
		iter := txn.NewIterator(badger.IteratorOptions{})
		defer iter.Close()

		for iter.Rewind(); iter.Valid(); iter.Next() {
			item := iter.Item()

			fullKey := string(item.Key())
			splitKey := strings.Split(fullKey, ".")

			hasPrefix := len(prefix) <= len(splitKey)
			for i := 0; hasPrefix && i < len(prefix) && i < len(splitKey); i++ {
				if prefix[i] != splitKey[i] {
					hasPrefix = false
				}
			}

			if hasPrefix {
				keys = append(keys, strings.Split(fullKey, "."))
			}
		}

		return nil
	})
}

// Export is the badger implementation of Database.Export.
func (db *BadgerDatabase) Export(w io.Writer) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	_, err := db.db.Backup(w, 0)
	return err
}

// Import is the badger implementation of Database.Import.
func (db *BadgerDatabase) Import(r io.Reader) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	return db.db.Load(r, 1)
}

// Glob is the badger implementation of Database.Glob
func (db *BadgerDatabase) Glob(prefix []string) ([][]string, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	fullPrefix := strings.Join(prefix, ".")

	results := [][]string{}
	err := db.view(func(txn *badger.Txn) error {
		iter := txn.NewIterator(badger.IteratorOptions{})
		defer iter.Close()

		for iter.Seek([]byte(fullPrefix)); iter.Valid(); iter.Next() {
			fullKey := string(iter.Item().Key())
			if !strings.HasPrefix(fullKey, fullPrefix) {
				break
			}

			// Don't do recursive globbing:
			leftOver := fullKey[len(fullPrefix):]
			if !strings.Contains(leftOver, ".") {
				results = append(results, strings.Split(fullKey, "."))
			}
		}

		return nil
	})

	return results, err
}

// Batch is the badger implementation of Database.Batch
func (db *BadgerDatabase) Batch() Batch {
	db.mu.Lock()
	defer db.mu.Unlock()

	return db.batch()
}

func (db *BadgerDatabase) batch() Batch {
	if db.txn == nil {
		db.txn = db.db.NewTransaction(true)
	}

	db.refCount++
	return db
}

// Put is the badger implementation of Database.Put
func (db *BadgerDatabase) Put(val []byte, key ...string) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.haveWrites = true

	fullKey := []byte(strings.Join(key, "."))

	err := db.withRetry(func() error {
		return db.txn.Set(fullKey, val)
	})

	if err != nil {
		log.Warningf("badger: failed to set key %s: %v", fullKey, err)
	}
}

func (db *BadgerDatabase) withRetry(fn func() error) error {
	if err := fn(); err != badger.ErrTxnTooBig {
		// This also returns nil.
		return err
	}

	// Commit previous (almost too big) transaction:
	if err := db.txn.Commit(); err != nil {
		// Something seems pretty wrong.
		return err
	}

	db.txn = db.db.NewTransaction(true)

	// If this fails again, we're out of luck.
	return fn()
}

// Clear is the badger implementation of Database.Clear
func (db *BadgerDatabase) Clear(key ...string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.haveWrites = true

	iter := db.txn.NewIterator(badger.IteratorOptions{})
	prefix := strings.Join(key, ".")

	keys := [][]byte{}
	for iter.Rewind(); iter.Valid(); iter.Next() {
		item := iter.Item()

		key := make([]byte, len(item.Key()))
		copy(key, item.Key())
		keys = append(keys, key)
	}

	// This has to happen here, since withRetry might call
	// txn.Discard() which will complain about open iterators.
	// (I previously used a defer which executed too late)
	iter.Close()

	for _, key := range keys {
		if !strings.HasPrefix(string(key), prefix) {
			continue
		}

		err := db.withRetry(func() error {
			return db.txn.Delete(key)
		})

		if err != nil {
			return err
		}
	}

	return nil
}

// Erase is the badger implementation of Database.Erase
func (db *BadgerDatabase) Erase(key ...string) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.haveWrites = true

	fullKey := []byte(strings.Join(key, "."))
	err := db.withRetry(func() error {
		return db.txn.Delete(fullKey)
	})

	if err != nil {
		log.Warningf("badger: failed to del key %s: %v", fullKey, err)
	}
}

// Flush is the badger implementation of Database.Flush
func (db *BadgerDatabase) Flush() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.refCount--
	if db.refCount > 0 {
		return nil
	}

	if db.refCount < 0 {
		log.Errorf("negative batch ref count: %d", db.refCount)
		return nil
	}

	defer db.txn.Discard()
	if err := db.txn.Commit(); err != nil {
		return err
	}

	db.txn = nil
	if db.haveWrites {
		db.writeTimes = append(db.writeTimes, time.Now())
	}
	db.haveWrites = false
	return nil
}

// Rollback is the badger implementation of Database.Rollback
func (db *BadgerDatabase) Rollback() {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.refCount--
	if db.refCount > 0 {
		return
	}

	if db.refCount < 0 {
		log.Errorf("negative batch ref count: %d", db.refCount)
		return
	}

	db.txn.Discard()
	db.txn = nil
	db.haveWrites = false
	db.refCount = 0
}

// HaveWrites is the badger implementation of Database.HaveWrites
func (db *BadgerDatabase) HaveWrites() bool {
	db.mu.Lock()
	defer db.mu.Unlock()

	return db.haveWrites
}

// Close is the badger implementation of Database.Close
func (db *BadgerDatabase) Close() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.gcTicker.Stop()
	atomic.StoreInt64(&db.isStopped, 1)

	// With an open transaction it would deadlock:
	if db.txn != nil {
		db.txn.Discard()
		db.txn = nil
		db.haveWrites = false
	}

	if db.db != nil {
		oldDb := db.db
		db.db = nil
		if err := oldDb.Close(); err != nil {
			return err
		}
	}

	return nil
}
