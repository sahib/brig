package store

import (
	"bytes"
	"fmt"
	"os"
	"testing"
)

func withDummyKv(t *testing.T, fn func(kv KV)) {
	path := "/tmp/brig.test.bolt"

	// This may fail:
	os.Remove(path)

	kv, err := NewBoltKV(path)
	if err != nil {
		t.Errorf("Unable to create bolt kv: %v", err)
		return
	}

	defer os.Remove(path)

	fn(kv)

	if err := kv.Close(); err != nil {
		t.Errorf("Closing the kv failed: %v", err)
	}
}

func TestKVBasic(t *testing.T) {
	kv, err := NewBoltKV("/tmp/bolt.kv.test")
	if err != nil {
		t.Errorf("Failed to open bolt kv: %v", err)
		return
	}

	buck, err := kv.Bucket([]string{"name"})
	if err != nil {
		t.Errorf("Failed to create bucket")
		return
	}

	if err := buck.Put("key", []byte("data")); err != nil {
		t.Errorf("Failed to put data to bucket: %v", err)
		return
	}

	data, err := buck.Get("key")
	if err != nil {
		t.Errorf("Failed to get data from bucket: %v", err)
		return
	}

	if !bytes.Equal(data, []byte("data")) {
		t.Errorf("data is not equal")
		return
	}

	sub, err := buck.Bucket([]string{"sub"})
	if err != nil {
		t.Errorf("Could not create sub bucket: %v", err)
		return
	}

	if err := sub.Put("key", []byte("sub-data")); err != nil {
		t.Errorf("Failed to put data to sub-bucket: %v", err)
		return
	}

	subData, err := sub.Get("key")
	if err != nil {
		t.Errorf("Failed to get sub-data from bucket: %v", err)
		return
	}

	if !bytes.Equal(subData, []byte("sub-data")) {
		t.Errorf("sub-data is not equal")
		t.Errorf("Expected: sub-data Got: %s", subData)
		return
	}

	pathSubData, err := getPath(kv, "/name/key")
	if err != nil {
		t.Errorf("Sub data failed")
		return
	}
	fmt.Println(string(pathSubData))

	if err := kv.Close(); err != nil {
		t.Errorf("Failed to close bolt kv: %v", err)
		return
	}
}

func TestKVPaths(t *testing.T) {
	paths := []string{
		"stage/tree/root/.",
		"stage/tree/.",
		"stage/.",
	}

	data := []byte("Hello World")

	withDummyKv(t, func(kv KV) {
		for _, path := range paths {
			if err := putPath(kv, path, data); err != nil {
				t.Errorf("putPath() failed for %s: %v", path, err)
				return
			}

			getData, err := getPath(kv, path)
			if err != nil {
				t.Errorf("getPath() failed for %s: %v", path, err)
				return
			}

			if !bytes.Equal(getData, data) {
				t.Errorf("path-data differs for %s", path)
				t.Errorf("Want: '%s' Got: '%s'", data, getData)
				return
			}
		}
	})
}
