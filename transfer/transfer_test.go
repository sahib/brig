package transfer

import (
	"fmt"
	"github.com/disorganizer/brig/util"
	"testing"
)

func TestCommunication(t *testing.T) {
	im := &util.SyncBuffer{}
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
