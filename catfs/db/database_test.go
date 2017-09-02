package db

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
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
			t.Errorf("Failed to flush key: %v", err)
			return
		}

		data, err := db1.Get(testKey...)
		if err != nil {
			t.Errorf("Failed get key: %v", err)
			return
		}

		if !bytes.Equal(data, []byte{1, 2, 3}) {
			t.Errorf("Data not equal")
			return
		}

		buf := &bytes.Buffer{}
		if eerr := db1.Export(buf); err != nil {
			t.Errorf("Export failed: %v", eerr)
			return
		}

		if ierr := db2.Import(buf); err != nil {
			t.Errorf("Import failed: %v", ierr)
			return
		}

		value, err := db2.Get(testKey...)
		if err != nil {
			t.Errorf("Failed to get value")
			return
		}

		if !bytes.Equal(value, []byte{1, 2, 3}) {
			t.Errorf("Wrong value after import")
			return
		}
	})
}
