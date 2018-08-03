package db

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func withDiskDatabase(t *testing.T, fn func(db *DiskDatabase)) {
	testDir, _ := ioutil.TempDir("", "brig-")
	defer os.RemoveAll(testDir)

	db, err := NewDiskDatabase(testDir)
	if err != nil {
		t.Errorf("Failed to create db1: %v", err)
		return
	}

	fn(db)

	if err := db.Close(); err != nil {
		t.Errorf("Failed to close db: %v", err)
		return
	}
}

func withBadgerDatabase(t *testing.T, fn func(db *BadgerDatabase)) {
	testDir, _ := ioutil.TempDir("", "brig-")
	defer os.RemoveAll(testDir)

	db, err := NewBadgerDatabase(testDir)
	if err != nil {
		t.Errorf("Failed to create db1: %v", err)
		return
	}

	fn(db)

	if err := db.Close(); err != nil {
		t.Errorf("Failed to close db: %v", err)
		return
	}
}

func withMemDatabase(t *testing.T, fn func(db *MemoryDatabase)) {
	mdb := NewMemoryDatabase()

	fn(mdb)

	if err := mdb.Close(); err != nil {
		t.Errorf("Failed to close mdb: %v", err)
		return
	}
}

func withDbByName(t *testing.T, name string, fn func(db Database)) {
	switch name {
	case "memory":
		withMemDatabase(t, func(db *MemoryDatabase) {
			fn(db)
		})
	case "disk":
		withDiskDatabase(t, func(db *DiskDatabase) {
			fn(db)
		})
	case "badger":
		withBadgerDatabase(t, func(db *BadgerDatabase) {
			fn(db)
		})
	default:
		panic("bad db name: " + name)
	}
}

func withDbsByName(t *testing.T, name string, fn func(db1, db2 Database)) {
	withDbByName(t, name, func(db1 Database) {
		withDbByName(t, name, func(db2 Database) {
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
				withDbsByName(t, name, func(db1, db2 Database) {
					tc.test(t, db1, db2)
				})
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
		},
	}

	t.Run("single", func(t *testing.T) {
		for _, tc := range tcs {
			t.Run(tc.name, func(t *testing.T) {
				withDbByName(t, name, func(db Database) {
					tc.test(t, db)
				})
			})
		}
	})
}

///////////

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
		[]string{"some", "stuff", "x"},
		[]string{"some", "stuff", "."},
		[]string{".", ".", "."},
		[]string{"some", "stuff", "__NO_DOT__"},
		[]string{"some", "stuff", "DOT"},
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
	batch.Clear()

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
	batch.Clear("a")

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
		[]string{"some", "stuff", "x"},
		[]string{"some", "stuff", "."},
		[]string{"some", "stuff", "__NO_DOT__"},
		[]string{"some", "stuff", "DOT"},
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
