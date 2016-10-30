package store

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/disorganizer/brig/store/compress"
)

func TestStoreImportExport(t *testing.T) {
	data := []byte{1, 2, 3}

	withIpfsStore(t, "alice", func(alice *Store) {
		if err := alice.StageFromReader("/hello.go", bytes.NewReader(data), compress.AlgoNone); err != nil {
			t.Errorf("Failed to stage /hello.go: %v", err)
			return
		}

		aliceStore, err := alice.Export()
		if err != nil {
			t.Errorf("Exporting store failed: %v", err)
			return
		}

		aliceFile, err := alice.fs.LookupFile("/hello.go")
		if err != nil {
			t.Errorf("Failed to retrieve alice' /hello.go: %v", err)
			return
		}

		withIpfsStore(t, "bob", func(bob *Store) {
			if err := bob.Import(aliceStore); err != nil {
				t.Errorf("Import by bob failed: %v", err)
				return
			}

			err = bob.ViewFile("/hello.go", func(file *File) error {
				if !file.Hash().Equal(aliceFile.Hash()) {
					return fmt.Errorf("Hashes differ between alice und bob")
				}

				return nil
			})

			if err != nil {
				t.Errorf("Failed to retrieve /hello.go by bob: %v", err)
			}
		})
	})
}
