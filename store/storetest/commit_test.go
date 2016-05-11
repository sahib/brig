package storetest

import (
	"testing"

	"github.com/disorganizer/brig/store"
)

func TestCommitMarshalling(t *testing.T) {
	withIpfsStore(t, "alice", func(st *store.Store) {
		cm := NewEmptyCommit()
	})
}
