package storetest

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/disorganizer/brig/id"
	"github.com/disorganizer/brig/store"
	"github.com/disorganizer/brig/util/ipfsutil"
	"github.com/disorganizer/brig/util/testutil"
	"github.com/disorganizer/brig/util/testwith"
)

func withStore(t *testing.T, ID id.ID, IPFS *ipfsutil.Node, fn func(st *store.Store)) {
	tempDir, err := ioutil.TempDir("", "brig-store-")
	if err != nil {
		t.Fatalf("Could not create temp dir `%s`: %v", tempDir, err)
		return
	}

	defer testutil.Remover(t, tempDir)

	st, err := store.Open(tempDir, id.NewPeer(ID, "QmIMACOW"), IPFS)
	if err != nil {
		t.Fatalf("Could not create store: %v", err)
		return
	}

	fn(st)
}

func withIpfsStore(t *testing.T, ID id.ID, fn func(st *store.Store)) {
	testwith.WithIpfs(t, func(nd *ipfsutil.Node) {
		withStore(t, ID, nd, fn)
	})
}

func TestHash(t *testing.T) {

	withIpfsStore(t, "alice", func(st *store.Store) {
		creator := func(path string, data []byte) []byte {
			if err := st.AddFromReader(path, bytes.NewReader(data)); err != nil {
				t.Fatalf("Adding `%s` failed: %v", path, err)
				return nil
			}

			return st.Root.Lookup(path).Hash().Bytes()
		}

		hash1 := creator("/child.go", []byte("Hello"))
		hash2 := creator("/russia/piotr.go", []byte("World"))

		if len(hash1) != len(hash2) {
			t.Errorf("Hash lengths differ")
			return
		}

		rootHash := st.Root.Hash().Bytes()
		if len(rootHash) != len(hash1) {
			t.Errorf("Root hash length changed")
			return
		}

		if !bytes.Equal(hash1[:2], rootHash[:2]) {
			t.Errorf("Root hash used different hash algorithm")
			return
		}

		// Check if root hash is really the xor of the two others:
		// (but skip the "Qm" in the beginning)
		for idx := 2; idx < len(hash1); idx++ {
			if hash1[idx]^hash2[idx] != rootHash[idx] {
				t.Errorf(
					"Hash differs at idx `%d`: %d^%d != %d",
					idx,
					hash1[idx]^hash2[idx],
					rootHash[idx],
				)
			}
		}

		// Remove the file with `hash1`
		if err := st.Remove("/child.go", false); err != nil {
			t.Errorf("Removing child failed: %v", err)
			return
		}

		// Check if roothash equals the file with `hash2`
		rootHash = []byte(string(st.Root.Hash().Bytes()))

		if !bytes.Equal(hash2, rootHash) {
			t.Errorf("Root hash is not the same as single member")
			return
		}

		// Try to modify the russian file:
		hash3 := creator("/russia/piotr.go", []byte("Comrade!"))

		newRootHash := st.Root.Hash().Bytes()
		if !bytes.Equal(hash3, newRootHash) {
			t.Errorf("Modifying leads to dirty traces of old hashes")
			return
		}
	})
}
