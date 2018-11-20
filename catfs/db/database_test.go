package db

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/sahib/brig/util/testutil"
	"github.com/stretchr/testify/require"
)

func withDiskDatabase(fn func(db *DiskDatabase)) error {
	testDir, _ := ioutil.TempDir("", "brig-")
	defer os.RemoveAll(testDir)

	db, err := NewDiskDatabase(testDir)
	if err != nil {
		return err
	}

	fn(db)
	return db.Close()
}

func withBadgerDatabase(fn func(db *BadgerDatabase)) error {
	testDir, _ := ioutil.TempDir("", "brig-")
	defer os.RemoveAll(testDir)

	db, err := NewBadgerDatabase(testDir)
	if err != nil {
		return err
	}

	fn(db)
	return db.Close()
}

func withMemDatabase(fn func(db *MemoryDatabase)) error {
	mdb := NewMemoryDatabase()
	fn(mdb)
	return mdb.Close()
}

func withDbByName(name string, fn func(db Database)) error {
	switch name {
	case "memory":
		return withMemDatabase(func(db *MemoryDatabase) {
			fn(db)
		})
	case "disk":
		return withDiskDatabase(func(db *DiskDatabase) {
			fn(db)
		})
	case "badger":
		return withBadgerDatabase(func(db *BadgerDatabase) {
			fn(db)
		})
	default:
		panic("bad db name: " + name)
	}
}

func withDbsByName(name string, fn func(db1, db2 Database)) error {
	return withDbByName(name, func(db1 Database) {
		withDbByName(name, func(db2 Database) {
			fn(db1, db2)
		})
	})
}

//////////

func TestDatabase(t *testing.T) {
	t.Run("memory", func(t *testing.T) {
		testDatabaseOneDb(t, "memory")
		testDatabaseTwoDbs(t, "memory")
	})
	t.Run("disk", func(t *testing.T) {
		testDatabaseOneDb(t, "disk")
		testDatabaseTwoDbs(t, "disk")
	})
	t.Run("badger", func(t *testing.T) {
		testDatabaseOneDb(t, "badger")
		testDatabaseTwoDbs(t, "badger")
	})
}

//////////

func testDatabaseTwoDbs(t *testing.T, name string) {
	tcs := []struct {
		name string
		test func(t *testing.T, db1, db2 Database)
	}{
		{
			name: "export-import",
			test: testExportImport,
		},
	}

	t.Run("double", func(t *testing.T) {
		for _, tc := range tcs {
			t.Run(tc.name, func(t *testing.T) {
				require.Nil(t, withDbsByName(name, func(db1, db2 Database) {
					tc.test(t, db1, db2)
				}))
			})
		}
	})
}

func testDatabaseOneDb(t *testing.T, name string) {
	tcs := []struct {
		name string
		test func(t *testing.T, db Database)
	}{
		{
			name: "put-and-get",
			test: testPutAndGet,
		}, {
			name: "glob",
			test: testGlob,
		}, {
			name: "clear",
			test: testClear,
		}, {
			name: "clear-prefix",
			test: testClearPrefix,
		}, {
			name: "invalid-access",
			test: testInvalidAccess,
		}, {
			name: "recursive-batch",
			test: testRecursiveBatch,
		}, {
			name: "rollback",
			test: testRollback,
		}, {
			name: "erase",
			test: testErase,
		}, {
			name: "keys",
			test: testKeys,
		},
	}

	t.Run("single", func(t *testing.T) {
		for _, tc := range tcs {
			t.Run(tc.name, func(t *testing.T) {
				require.Nil(t, withDbByName(name, func(db Database) {
					tc.test(t, db)
				}))
			})
		}
	})
}

///////////

func testErase(t *testing.T, db Database) {
	batch := db.Batch()
	batch.Put([]byte{1}, "existing_key")
	batch.Flush()

	batch = db.Batch()
	batch.Erase("existing_key")

	_, err := db.Get("existing_key")
	require.Equal(t, ErrNoSuchKey, err)

	batch.Flush()

	_, err = db.Get("existing_key")
	require.Equal(t, ErrNoSuchKey, err)
}

func testKeys(t *testing.T, db Database) {
	batch := db.Batch()
	for i := 0; i < 15; i++ {
		batch.Put([]byte{byte(i)}, fmt.Sprintf("%d", i))
	}
	batch.Flush()

	extractKeys := func(prefixes []string) []string {
		keys := []string{}
		err := db.Keys(func(key []string) error {
			keys = append(keys, strings.Join(key, "."))
			return nil
		}, prefixes...)
		require.Nil(t, err)
		return keys
	}

	keys := extractKeys(nil)
	require.Equal(t,
		[]string{
			"0", "1", "10", "11", "12", "13", "14",
			"2", "3", "4", "5", "6", "7", "8", "9",
		},
		keys,
	)

	keys = extractKeys([]string{"1"})
	require.Equal(t,
		[]string{"1"},
		keys,
	)

	errSentinel := errors.New("weird error")
	err := db.Keys(func(key []string) error { return errSentinel })
	require.Equal(t, errSentinel, err)
}

func testRollback(t *testing.T, db Database) {
	batch := db.Batch()
	batch.Put([]byte{1}, "existing_key")
	batch.Flush()

	batch = db.Batch()
	batch.Put([]byte{2}, "existing_key")
	batch.Put([]byte{2}, "some_key")

	data, err := db.Get("some_key")
	require.Nil(t, err)
	require.Equal(t, []byte{2}, data)

	batch.Rollback()

	data, err = db.Get("existing_key")
	require.Nil(t, err)
	require.Equal(t, []byte{1}, data)

	data, err = db.Get("some_key")
	require.Equal(t, ErrNoSuchKey, err)
	require.Nil(t, data)
}

func testRecursiveBatch(t *testing.T, db Database) {
	batch1 := db.Batch()
	batch2 := db.Batch()

	batch2.Put([]byte{1}, "batch2_key")
	val, err := db.Get("batch2_key")

	require.Nil(t, err)
	require.Equal(t, []byte{1}, val)

	require.True(t, batch1.HaveWrites())
	require.True(t, batch2.HaveWrites())
	require.Nil(t, batch2.Flush())

	require.True(t, batch1.HaveWrites())
	require.True(t, batch2.HaveWrites())

	require.Nil(t, batch1.Flush())
	require.False(t, batch1.HaveWrites())
	require.False(t, batch2.HaveWrites())
}

func testPutAndGet(t *testing.T, db Database) {
	testKeys := [][]string{
		{"some", "stuff", "x"},
		{"some", "stuff", "."},
		{".", ".", "."},
		{"some", "stuff", "__NO_DOT__"},
		{"some", "stuff", "DOT"},
	}

	for _, key := range testKeys {
		t.Run(strings.Join(key, "."), func(t *testing.T) {
			batch := db.Batch()
			batch.Put([]byte("hello"), key...)
			require.Nil(t, batch.Flush())

			data, err := db.Get(key...)
			require.Nil(t, err)
			require.Equal(t, []byte("hello"), data)
		})
	}
}

func testInvalidAccess(t *testing.T, db Database) {
	val, err := db.Get("hello", "world")
	require.Equal(t, ErrNoSuchKey, err)
	require.Nil(t, val)
}

func testClear(t *testing.T, db Database) {
	batch := db.Batch()
	for i := 0; i < 100; i++ {
		batch.Put([]byte{1}, "a", "b", "c", fmt.Sprintf("%d", i))
	}

	require.Nil(t, batch.Flush())

	batch = db.Batch()
	require.Nil(t, batch.Clear())

	// before flush:
	for i := 0; i < 100; i++ {
		_, err := db.Get("a", "b", "c", fmt.Sprintf("%d", i))
		require.Equal(t, ErrNoSuchKey, err)
	}

	require.Nil(t, batch.Flush())

	// after flush:
	for i := 0; i < 100; i++ {
		_, err := db.Get("a", "b", "c", fmt.Sprintf("%d", i))
		require.Equal(t, ErrNoSuchKey, err)
	}
}

func testClearPrefix(t *testing.T, db Database) {
	batch := db.Batch()
	for i := 0; i < 10; i++ {
		batch.Put([]byte{1}, "a", "b", "c", fmt.Sprintf("%d", i))
	}

	for i := 0; i < 10; i++ {
		batch.Put([]byte{1}, "x", "y", "z", fmt.Sprintf("%d", i))
	}

	require.Nil(t, batch.Flush())

	batch = db.Batch()
	require.Nil(t, batch.Clear("a"))

	// before flush:
	for i := 0; i < 10; i++ {
		_, err := db.Get("a", "b", "c", fmt.Sprintf("%d", i))
		require.Equal(t, ErrNoSuchKey, err)
	}

	for i := 0; i < 10; i++ {
		data, err := db.Get("x", "y", "z", fmt.Sprintf("%d", i))
		require.Nil(t, err)
		require.Equal(t, []byte{1}, data)
	}

	require.Nil(t, batch.Flush())

	// after flush:
	for i := 0; i < 10; i++ {
		_, err := db.Get("a", "b", "c", fmt.Sprintf("%d", i))
		require.Equal(t, ErrNoSuchKey, err)
	}

	for i := 0; i < 10; i++ {
		data, err := db.Get("x", "y", "z", fmt.Sprintf("%d", i))
		require.Nil(t, err)
		require.Equal(t, []byte{1}, data)
	}
}

func testGlob(t *testing.T, db Database) {
	batch := db.Batch()
	batch.Put([]byte{1}, "a", "b", "pref_1")
	batch.Put([]byte{2}, "a", "b", "pref_2")
	batch.Put([]byte{3}, "a", "b", "prev_3")
	batch.Put([]byte{4}, "a", "b", "pref_dir", "x")

	err := batch.Flush()
	require.Nil(t, err)

	matches, err := db.Glob([]string{"a", "b", "pref_"})
	require.Nil(t, err)

	require.Equal(t, [][]string{
		{"a", "b", "pref_1"},
		{"a", "b", "pref_2"},
	}, matches)
}

func testExportImport(t *testing.T, db1, db2 Database) {
	testKeys := [][]string{
		{"some", "stuff", "x"},
		{"some", "stuff", "."},
		{"some", "stuff", "__NO_DOT__"},
		{"some", "stuff", "DOT"},
	}

	batch := db1.Batch()
	for _, key := range testKeys {
		batch.Put([]byte{1, 2, 3}, key...)
	}

	require.Nil(t, batch.Flush())

	for _, key := range testKeys {
		data, err := db1.Get(key...)
		require.Nil(t, err)
		require.Equal(t, []byte{1, 2, 3}, data)
	}

	buf := &bytes.Buffer{}
	require.Nil(t, db1.Export(buf))
	require.Nil(t, db2.Import(buf))

	for _, key := range testKeys {
		data, err := db2.Get(key...)
		require.Nil(t, err)
		require.Equal(t, []byte{1, 2, 3}, data)
	}
}

func BenchmarkDatabase(b *testing.B) {
	benchmarks := map[string]func(*testing.B, Database){
		"put": benchmarkDatabasePut,
		"get": benchmarkDatabaseGet,
	}

	for benchmarkName, benchmark := range benchmarks {
		b.Run(benchmarkName, func(b *testing.B) {
			for _, name := range []string{"badger", "memory", "disk"} {
				b.Run(name, func(b *testing.B) {
					withDbByName(name, func(db Database) {
						b.ResetTimer()
						benchmark(b, db)
						b.StopTimer()
					})
				})
			}
		})
	}
}

func benchmarkDatabasePut(b *testing.B, db Database) {
	batch := db.Batch()
	for i := 0; i < b.N; i++ {
		keyName := fmt.Sprintf("key_%d", i%(1024*1024))
		batch.Put(testutil.CreateDummyBuf(128), keyName)
	}
	batch.Flush()
}

func benchmarkDatabaseGet(b *testing.B, db Database) {
	batch := db.Batch()
	for i := 0; i < b.N; i++ {
		keyName := fmt.Sprintf("key_%d", i%(1024*1024))
		batch.Put(testutil.CreateDummyBuf(128), "prefix", keyName)
	}
	batch.Flush()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		keyName := fmt.Sprintf("key_%d", i%(1024*1024))
		db.Get("prefix", keyName)
	}
}
