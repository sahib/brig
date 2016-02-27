package store

import (
	"testing"

	"github.com/jbenet/go-multihash"
)

func dummyStore(t *testing.T) *Store {
	store := &Store{}
	rootDir, err := newDirUnlocked(store, "/")
	if err != nil {
		t.Errorf("newDir failed: %v", err)
		return nil
	}

	store.Root = rootDir
	return store
}

func dummyFile(t *testing.T, store *Store, path string) *File {
	child := &File{
		Metadata: &Metadata{
			size: 27,
			kind: FileTypeRegular,
		},
		store:   store,
		RWMutex: store.Root.RWMutex,
	}

	// Insert self to root:
	child.insert(store.Root, path)
	return child
}

func dummyHash(t *testing.T, store *Store, seed byte) *Hash {
	hash, err := multihash.Encode(
		[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, seed},
		multihash.SHA1,
	)

	if err != nil {
		t.Errorf("Could not convert dummy bytes to multihash: %v", err)
		return nil
	}

	return &Hash{hash}
}

func TestHash(t *testing.T) {
	store := dummyStore(t)
	if store == nil {
		return
	}

	file := dummyFile(t, store, "/child.go")
	file.hash = dummyHash(t, store, 1)

	other := dummyFile(t, store, "/russia/piotr.go")
	other.hash = dummyHash(t, store, 2)

	rootHash := store.Root.Hash().Bytes()
	t.Logf("FIXME: %v", rootHash)

	// TODO: Fix test; sync() is now doing the xor.
	// if rootHash[len(rootHash)-1] != 1^2 {
	// 	t.Errorf("Root dir has not XOR'd children checksum properly:")
	// 	t.Errorf("\tEXPECTED: 3 at end; GOT: %v", rootHash)
	// 	return
	// }

}
