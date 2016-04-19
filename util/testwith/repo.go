package testwith

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/disorganizer/brig/repo"
	"github.com/disorganizer/brig/util/testutil"
)

func WithAliceRepo(t *testing.T, f func(*repo.Repository)) {
	WithRepo(t, "alice", "alicepass", f)
}

func WithBobRepo(t *testing.T, f func(*repo.Repository)) {
	WithRepo(t, "bob", "bobpass", f)
}

func WithRepo(t *testing.T, user, pass string, f func(*repo.Repository)) {
	path, err := ioutil.TempDir("", "brig-repotest")
	if err != nil {
		t.Fatalf("Cannot create test repo: %v", err)
		return
	}
	WithRepoAtPath(t, path, user, pass, f)
}

func WithRepoAtPath(t *testing.T, path, user, pass string, f func(*repo.Repository)) {
	if err := os.RemoveAll(path); err != nil {
		t.Errorf("previous repo exists; cannot delete it though: %v", err)
		return
	}

	rep, err := repo.NewRepository(user, pass, path)
	if err != nil {
		t.Errorf("creating repo failed: %v", err)
		return
	}

	defer testutil.Remover(t, path)

	f(rep)

	if err := rep.Close(); err != nil {
		t.Errorf("closing repo failed: %v", err)
		return
	}
}
