package db

import (
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	badger "github.com/dgraph-io/badger/v3"
	"github.com/dustin/go-humanize"

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
	gcNeededAt time.Time
	needsGC    bool
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
		needsGC:  true,
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
	// The logic is quite convoluted:
	// It is not enough to ask for GC, badger DB needs to run some statistic
	// after last modification in DB (about a minute), if we don't wait
	// the GC call will be waisted in we might claim that there is nothing to GC.
	// Also to do GC correctly we need to Close DB before calling GC
	// (yep, badger is crazy in this regard).
	// Then we reopen DB and mark cleanable with GC,
	// then on Close() true GC happens (HDD space reclaimed),
	// and we need to reopen again.
	// Consequently we have a timer (to give time for stats update) and we needGC flag
	// which indicates that some updates were done on DB so we do not trash
	// underlying FS.
	// TODO: to do it really correct we need a queue which holds last DB modification
	// times. But it all become very convoluted since other processes mess with DB
	// Or repinner always does something in background and updates DB.
	if time.Now().Before(bdb.gcNeededAt) || !bdb.needsGC {
		// we are waiting for timer but also for updates
		// if there are no updated in DB, timer does not kick in
		log.Debugf("badger DB in %s stats are not updated yet, no need for GC", opts.Dir)
		return nil
	}
	bdb.mu.Lock()
	defer bdb.mu.Unlock()
	defer func() {
		bdb.gcNeededAt = time.Now()
		bdb.needsGC = false
	}()
	var errGC error
	var beforeGC uint64 = 0
	var total uint64 = 0
	var err error
	log.Infof("Performin GC for badger DB in %s", opts.Dir)
	for errGC == nil {
		// At the DB opening, a new vlog is created
		// and its size added to vlog size even if it is empty
		// The closing truncates vlog files so we get a better read of sizes
		err = bdb.db.Close() // GC is claimed only on close :(
		if err != nil {
			log.Fatalf("Could not close DB in %s", opts.Dir)
			return err
		}
		bdb.db, err = badger.Open(opts)
		if err != nil {
			log.Fatalf("Could not reopen DB in %s", opts.Dir)
			return err
		}
		// Badger people use os.FileInfo to get total size of lsm and vlog files
		// This is not actual but allocated of hdd size
		lsm, vlog := bdb.db.Size() // values could be stale even after Close/Open cycle, badger updates it once per minute
		// TODO: it might be simple just to count HDD allocation ourself as it done in badger internals
		//       for now I leave it alone, since it mostly for info messages
		newTotal := uint64(lsm) + uint64(vlog)
		log.Debugf("DB sizes in %s: lsm %s and vlog %s, total %s", opts.Dir, humanize.Bytes(uint64(lsm)), humanize.Bytes(uint64(vlog)), humanize.Bytes(uint64(newTotal)))
		if beforeGC == 0 {
			beforeGC = newTotal
		}
		if total != 0 {

			log.Infof("At this iteration DB size in %s decreased by %s", opts.Dir, humanize.Bytes(total-newTotal))
		}
		total = newTotal

		if errGC = bdb.db.RunValueLogGC(0.5); errGC != nil {
			if errGC == badger.ErrNoRewrite {
				log.Debugf("badger DB in %s has nothing for GC",opts.Dir)
			}
			continue
		}
	}
	if beforeGC > total {
		log.Infof("Total DB size in %s decreased by %s", opts.Dir, humanize.Bytes(beforeGC-total))
	} else {
		log.Infof("Total DB size in %s was not changed", opts.Dir)
	}
	return nil
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
	db.haveWrites = false
	if time.Now().After(db.gcNeededAt) && !db.needsGC {
		db.gcNeededAt = time.Now().Add(120 * time.Second) // badger updates stats once per minute, we give ample time for stats to settle
		db.needsGC = true
	}
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
