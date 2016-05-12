package storetest

import (
	"strings"
	"testing"

	"github.com/disorganizer/brig/store"
)

func TestCommitting(t *testing.T) {
	withIpfsStore(t, "alice", func(st *store.Store) {
		head, err := st.Head()
		if err != nil {
			t.Errorf("Unable to peek head at beginning: %v", err)
			return
		}

		if len(head.Changes) != 0 {
			t.Errorf("Initial commit has changes?")
			return
		}

		if !strings.Contains(strings.ToLower(head.Message), "initial") {
			t.Errorf("Initial commit does not name itself like that.")
			return
		}

		// Stage is empty after initial commit:
		if err := st.MakeCommit("empty."); err != store.ErrEmptyStage {
			t.Errorf("Empty commits are now allowed: %v", err)
			return
		}

		if err := st.Touch("/hello.world"); err != nil {
			t.Errorf("Unable to touch empty file: %v", err)
			return
		}

		if err := st.MakeCommit("testing around"); err != nil {
			t.Errorf("Could not commit: %v", err)
			return
		}

		head, err = st.Head()
		if err != nil {
			t.Errorf("Unable to peek head the second time: %v", err)
			return
		}

		if head.Message != "testing around" {
			t.Errorf("Bad commit message: %s", head.Message)
			return
		}

		if len(head.Changes) != 1 {
			t.Errorf("More or less changes than expected: %d", len(head.Changes))
			return
		}

		file := st.Root.Lookup("/hello.world")
		checkpoint, ok := head.Changes[file]
		if !ok {
			t.Errorf("No such file in changeset: %v", file)
			return
		}

		if checkpoint.Change != store.ChangeAdd {
			t.Errorf("Empty file was not added?")
			return
		}
	})
}
