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

	testDatabase(t, db1, db2)

	if err := db1.Close(); err != nil {
		t.Errorf("Failed to close db1: %v", err)
		return
	}

	if err := db2.Close(); err != nil {
		t.Errorf("Failed to close db2: %v", err)
		return
	}
}

func TestDiskvDatabase(t *testing.T) {
	testDir1, _ := ioutil.TempDir("", "brig-")
	testDir2, _ := ioutil.TempDir("", "brig-")

	defer func() {
		for _, dir := range []string{testDir1, testDir2} {
			if err := os.RemoveAll(dir); err != nil {
				t.Errorf("Failed to remove test dir %s: %s", dir, err)
			}
		}
	}()

	db1, err := NewDiskvDatabase(testDir1)
	if err != nil {
		t.Errorf("Failed to create db1: %v", err)
		return
	}

	db2, err := NewDiskvDatabase(testDir2)
	if err != nil {
		t.Errorf("Failed to create db2: %v", err)
		return
	}

	testDatabase(t, db1, db2)

	if err := db1.Close(); err != nil {
		t.Errorf("Failed to close db1: %v", err)
		return
	}

	if err := db2.Close(); err != nil {
		t.Errorf("Failed to close db2: %v", err)
		return
	}
}

func testDatabase(t *testing.T, db1, db2 Database) {
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
		if err := db1.Put([]byte{1, 2, 3}, "some/stuff", "x"); err != nil {
			t.Errorf("Failed set key: %v", err)
			return
		}

		data, err := db1.Get("some/stuff", "x")
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

		value, err := db2.Get("some/stuff", "x")
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
