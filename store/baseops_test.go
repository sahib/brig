package store

import (
	"fmt"
	"testing"

	goipfsutil "github.com/ipfs/go-ipfs-util"
	multihash "github.com/jbenet/go-multihash"
)

func TestStoreBaseOpMkdir(t *testing.T) {
	paths := []string{
		"/home/sahib/b/c/d/e/f",
		"/home/sahib/b/x/d/e/f",
	}

	withDummyKv(t, func(kv KV) {
		fs := NewFilesystem(kv)

		for _, path := range paths {
			dir, err := mkdir(fs, path, true)
			if err != nil {
				t.Errorf("Failed to mkdir parents of %s: %v", path, err)
				return
			}

			dirPath := NodePath(dir)
			if dirPath != path {
				t.Errorf("`%s` was inserted at `%s` :(", path, dirPath)
				return
			}

			fmt.Println(fs.LookupNode("/home/sahib/music.txt"))
		}
	})
}

func dummyHash(t *testing.T, seed byte) *Hash {
	data := make([]byte, multihash.DefaultLengths[goipfsutil.DefaultIpfsHash])

	for idx := range data {
		data[idx] = seed
	}

	hash, err := multihash.Encode(data, goipfsutil.DefaultIpfsHash)

	if err != nil {
		t.Fatalf("Failed to create dummy hash: %v", err)
		return nil
	}

	return &Hash{hash}
}

func TestStoreBaseOpCreateFile(t *testing.T) {
	dummyKey := make([]byte, 32)
	dummyPath := "/home/sahib/music.txt"

	withDummyKv(t, func(kv KV) {
		fs := NewFilesystem(kv)

		par, err := mkdir(fs, "/home/sahib", true)
		if err != nil {
			t.Errorf("Failed to create base dir: %v", err)
			return
		}

		file, err := touchFile(fs, dummyPath, dummyHash(t, 0), dummyKey, 17, "alice")
		if err != nil {
			t.Errorf("Failed to create file: %v", err)
			return
		}

		filePar, err := file.Parent()
		if err != nil {
			t.Errorf("Getting parent of file failed: %v", err)
			return
		}

		if !filePar.Hash().Equal(par.Hash()) {
			t.Errorf(
				"Hashes of parents differ: %s != %s",
				filePar.Hash().B58String(),
				par.Hash().B58String(),
			)
			return
		}

		printTree(fs)

		modFile, err := touchFile(fs, dummyPath, dummyHash(t, 1), dummyKey, 19, "alice")
		if err != nil {
			t.Errorf("Modification was not possible: %v", err)
			return
		}

		if !modFile.Hash().Equal(file.Hash()) {
			t.Errorf("Hashes of new and old do differ, despite being the same instance.")
		}

		printTree(fs)

		resolveModFile, err := fs.ResolveFile(dummyPath)
		if err != nil {
			t.Errorf("Failed to resolve modified file: %v", err)
			return
		}

		if resolveModFile.Size() != 19 {
			t.Errorf("Modified file did not update size")
			return
		}

		if !resolveModFile.Hash().Equal(dummyHash(t, 1)) {
			t.Errorf("Modified file has not the new hash")
		}
	})
}
