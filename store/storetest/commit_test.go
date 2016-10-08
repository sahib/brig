package storetest

/*
func TestCommitting(t *testing.T) {
	withIpfsStore(t, "alice", func(st *store.Store) {
		head, err := st.Head()
		if err != nil {
			t.Errorf("Unable to peek head at beginning: %v", err)
			return
		}

		if len(head.Checkpoints) != 0 {
			t.Errorf("Initial commit has Checkpoints?")
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

		if len(head.Checkpoints) != 1 {
			t.Errorf("More or less Checkpoints than expected: %d", len(head.Checkpoints))
			return
		}

		checkpoint := head.Checkpoints[0]
		if checkpoint.Change != store.ChangeAdd {
			t.Errorf("Empty file was not added?")
			return
		}
	})
}

func TestStatus(t *testing.T) {
	withIpfsStore(t, "alice", func(st *store.Store) {
		status, err := st.Status()
		if err != nil {
			t.Errorf("Could not retrieve initial status: %v", err)
			return
		}

		// TODO: Check more than .Checkpoints (also Hash, Size, etc.)
		if len(status.Checkpoints) != 0 {
			t.Errorf("There are checkpoint after initial commit: %d", len(status.Checkpoints))
			return
		}

		if err := st.Touch("/hello.world"); err != nil {
			t.Errorf("Unable to touch hello.world: %v", err)
			return
		}

		status, err = st.Status()
		if err != nil {
			t.Errorf("Could not retrieve status with one added file: %v", err)
			return
		}

		ck := status.Checkpoints[0]
		if ck.Path != "/hello.world" {
			t.Errorf("Bad path after touching file: ", ck.Path)
			return
		}

		if ck.Change != store.ChangeAdd {
			t.Errorf("Bad change type after touching file: %s", ck.Change.String())
			return
		}

		if err := st.Remove("/hello.world", false); err != nil {
			t.Errorf("Unable to remove /hello.world again: %v", err)
			return
		}

		status, err = st.Status()
		if err != nil {
			t.Errorf("Could not retrieve status with one deleted file: %v", err)
			return
		}

		ck = status.Checkpoints[0]
		if ck.Path != "/hello.world" {
			t.Errorf("Bad path after removing file: ", ck.Path)
			return
		}

		if ck.Change != store.ChangeRemove {
			t.Errorf("Bad change type after removing file: %s", ck.Change.String())
			return
		}
	})
}
*/
