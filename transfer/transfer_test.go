package transfer

import (
	"bytes"
	"fmt"
	"sync"
	"testing"
)

// TODO: generally useful? Move to utils?
type SyncBuffer struct {
	sync.RWMutex
	buf bytes.Buffer
}

func (b *SyncBuffer) Read(p []byte) (int, error) {
	b.Lock()
	defer b.Unlock()

	return b.buf.Read(p)
}

func (b *SyncBuffer) Write(p []byte) (int, error) {
	b.Lock()
	defer b.Unlock()

	return b.buf.Write(p)
}

func TestCommunication(t *testing.T) {
	im := &SyncBuffer{}
	cl := NewClient(im)
	sv := NewServer(im)

	go func() {
		if err := sv.Serve(); err != nil {
			t.Fatalf("Serve failed with error: %v", err)
		}
	}()

	resp, err := cl.Send(&Command{ID: CmdClone})
	if err != nil {
		t.Errorf("Sending quit failed: %v", err)
		return
	}

	fmt.Println(resp)
}
