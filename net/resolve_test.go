package net

import (
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/sahib/brig/backend"
	"github.com/sahib/brig/repo"
)

func withServer(who string, t *testing.T, fn func(bk backend.Backend)) {
	tmpFolder, err := ioutil.TempDir("", "brig-resolve-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	defer os.RemoveAll(tmpFolder)

	if err := repo.Init(tmpFolder, who, "xxx", "mock"); err != nil {
		t.Fatalf("Failed to init repo at: %v", err)
	}

	rp, err := repo.Open(tmpFolder, "xxx")
	if err != nil {
		t.Fatalf("Failed to open repository: %v", err)
	}

	bk, err := backend.FromName("mock", "")
	if err != nil {
		t.Fatalf("Failed to get mock backend (wtf?): %v", err)
	}

	srv, err := NewServer(rp, bk)
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	wg := &sync.WaitGroup{}
	go func() {
		wg.Add(1)
		defer wg.Done()

		if err := srv.Serve(); err != nil {
			t.Fatalf("Failed to start serving process: %v", err)
		}

		if err := srv.Close(); err != nil {
			t.Fatalf("Failed to close server properly: %v", err)
		}
	}()

	// Allow a short time for the server go routine to fully boot up.
	time.Sleep(50 * time.Millisecond)

	// Actually execute the test...
	fn(testUnit{
		rp: rp,
		bk: bk,
	})

	if err := rp.Close("xxx"); err != nil {
		t.Fatalf("Failed to close repo: %v", err)
	}

	// Quit the server.
	srv.Quit()
	wg.Wait()
}

func TestResolve(t *testing.T) {
	withServer("alice", t, func(bk backend.Backend) {
	})
}
