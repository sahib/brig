package db

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMemoryDatabase(t *testing.T) {
	db1 := NewMemoryDatabase()
	db2 := NewMemoryDatabase()

	testDatabaseWithDifferentKeys(t, db1, db2)

	if err := db1.Close(); err != nil {
		t.Errorf("Failed to close db1: %v", err)
		return
	}

	if err := db2.Close(); err != nil {
		t.Errorf("Failed to close db2: %v", err)
		return
	}
}

func TestDiskDatabase(t *testing.T) {
	testDir1, _ := ioutil.TempDir("", "brig-")
	testDir2, _ := ioutil.TempDir("", "brig-")

	defer func() {
		for _, dir := range []string{testDir1, testDir2} {
			if err := os.RemoveAll(dir); err != nil {
				t.Errorf("Failed to remove test dir %s: %s", dir, err)
			}
		}
	}()

	db1, err := NewDiskDatabase(testDir1)
	if err != nil {
		t.Errorf("Failed to create db1: %v", err)
		return
	}

	db2, err := NewDiskDatabase(testDir2)
	if err != nil {
		t.Errorf("Failed to create db2: %v", err)
		return
	}

	testDatabaseWithDifferentKeys(t, db1, db2)

	if err := db1.Close(); err != nil {
		t.Errorf("Failed to close db1: %v", err)
		return
	}

	if err := db2.Close(); err != nil {
		t.Errorf("Failed to close db2: %v", err)
		return
	}
}

func testDatabaseWithDifferentKeys(t *testing.T, db1, db2 Database) {
	testKeys := [][]string{
		[]string{"some", "stuff", "x"},
		[]string{"some", "stuff", "."},
		[]string{"some", "stuff", "__NO_DOT__"},
		[]string{"some", "stuff", "DOT"},
		[]string{"some", "stuff", "x"},
	}

	for _, testKey := range testKeys {
		testDatabase(t, db1, db2, testKey)
	}
}

func testDatabase(t *testing.T, db1, db2 Database, testKey []string) {
	// TODO: add more testcases
	t.Run("access-invalid", func(t *testing.T) {
		val, err := db1.Get("hello", "world")
		if err != ErrNoSuchKey {
			t.Errorf("Not existant key yieled no ErrNoSuchKey: %v", err)
			return
		}

		if val != nil {
			t.Errorf("Not existing key still returned data")
			return
		}
	})

	t.Run("export", func(t *testing.T) {
		batch := db1.Batch()
		batch.Put([]byte{1, 2, 3}, testKey...)

		if err := batch.Flush(); err != nil {
			t.Fatalf("Failed to flush key: %v", err)
		}

		data, err := db1.Get(testKey...)
		if err != nil {
			t.Fatalf("Failed get key: %v", err)
		}

		if !bytes.Equal(data, []byte{1, 2, 3}) {
			t.Fatalf("Data not equal")
		}

		buf := &bytes.Buffer{}
		if eerr := db1.Export(buf); eerr != nil {
			t.Fatalf("Export failed: %v", eerr)
		}

		if ierr := db2.Import(buf); ierr != nil {
			t.Fatalf("Import failed: %v", ierr)
		}

		value, err := db2.Get(testKey...)
		if err != nil {
			t.Fatalf("Failed to get value: %v", err)
		}

		if !bytes.Equal(value, []byte{1, 2, 3}) {
			t.Fatalf("Wrong value after import")
		}
	})
}

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

func withMemDatabase(t *testing.T, fn func(db *MemoryDatabase)) {
	mdb := NewMemoryDatabase()

	fn(mdb)

	if err := mdb.Close(); err != nil {
		t.Errorf("Failed to close mdb: %v", err)
		return
	}
}

func TestGlob(t *testing.T) {
	t.Run("disk", func(t *testing.T) {
		withDiskDatabase(t, func(db *DiskDatabase) {
			testGlob(t, db)
		})
	})

	t.Run("memory", func(t *testing.T) {
		withMemDatabase(t, func(db *MemoryDatabase) {
			testGlob(t, db)
		})
	})
}

func testGlob(t *testing.T, db Database) {
	batch := db.Batch()
	batch.Put([]byte{1}, "a", "b", "pref_1")
	batch.Put([]byte{2}, "a", "b", "pref_2")
	batch.Put([]byte{3}, "a", "b", "prev_3")
	batch.Put([]byte{3}, "a", "b", "pref_dir", "x")
	if err := batch.Flush(); err != nil {
		t.Fatalf("Failed to create testdata: %v", err)
	}

	matches, err := db.Glob([]string{"a", "b", "pref_"})
	if err != nil {
		t.Fatalf("Failed to find matches: %v", err)
	}

	require.Equal(t, matches, [][]string{
		{"a", "b", "pref_1"},
		{"a", "b", "pref_2"},
	})
}
