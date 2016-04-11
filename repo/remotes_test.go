package repo

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/spf13/afero"
)

func TestRemote(t *testing.T) {
	fs := afero.NewMemMapFs()
	fd, err := fs.Open("remotes.yml")
	if err != nil {
		t.Errorf("Could not open in memory: %v", err)
		return
	}

	rms, err := NewYAMLRemotes(fd)
	if err != nil {
		t.Errorf("Creating yaml store failed: %v", err)
		return
	}

	defer rms.Close()

	rms.Insert(NewRemote("alice", "Qm123"))
	fmt.Println(ioutil.ReadAll(fd))
}
