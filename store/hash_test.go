package store

import (
	"encoding/json"
	"testing"

	"github.com/jbenet/go-multihash"
)

func TestJson(t *testing.T) {
	data := "QmYT9RV777QF1r4WCQ1PrtPcUcrQ5U6EFpnkkY9qhAupKx"
	mh, err := multihash.FromB58String(data)
	if err != nil {
		t.Errorf("Multihash failed to convert literal: %v", err)
		return
	}

	hash := &Hash{mh}
	jsonData, err := json.Marshal(hash)
	if err != nil {
		t.Errorf("hash to json conversion failed: %v", err)
		return
	}

	zombieHash := &Hash{}
	if err := json.Unmarshal(jsonData, zombieHash); err != nil {
		t.Errorf("json to hash conversion failed: %v", err)
		return
	}

	a, b := hash.B58String(), zombieHash.B58String()
	if a != b {
		t.Errorf("hashes differ after unmarshalling:")
		t.Errorf("\tExpected: %q", a)
		t.Errorf("\tGot:      %q", b)
		return
	}
}
