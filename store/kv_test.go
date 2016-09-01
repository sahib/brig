package store

import (
	"bytes"
	"testing"
)

func TestKVBasic(t *testing.T) {
	kv, err := NewBoltKV("/tmp/bolt.kv.test")
	if err != nil {
		t.Errorf("Failed to open bolt kv: %v", err)
		return
	}

	buck, err := kv.Bucket("name")
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

	sub, err := buck.Bucket("sub")
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

	if err := kv.Close(); err != nil {
		t.Errorf("Failed to close bolt kv: %v", err)
		return
	}
}
