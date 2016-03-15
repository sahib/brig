package ipfsutil_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/disorganizer/brig/util/ipfsutil"
	"github.com/disorganizer/brig/util/testwith"
)

var (
	TestPath = filepath.Join(os.TempDir(), "brig_test_ipfs_repo")
)

func TestAddCat(t *testing.T) {
	testwith.WithIpfs(t, TestPath, func(node *ipfsutil.Node) {
		// Dummy in-memory reader:
		origData := []byte("Hello World")
		buf := &bytes.Buffer{}
		buf.Write(origData)

		hash, err := ipfsutil.Add(node, buf)
		if err != nil {
			t.Errorf("Add of a simple file failed: %v", err)
			return
		}

		reader, err := ipfsutil.Cat(node, hash)
		if err != nil {
			t.Errorf("Could not cat simple file: %v", err)
			return
		}

		data, err := ioutil.ReadAll(reader)
		if err != nil {
			t.Errorf("Could not read back added data: %v", err)
			return
		}

		if err = reader.Close(); err != nil {
			t.Errorf("close(cat) failed: %v", err)
			return
		}

		if !bytes.Equal(data, origData) {
			t.Errorf("Data not equal: %v <- -> %v", string(data), string(origData))
		}
	})
}

func TestDHT(t *testing.T) {
	testwith.WithIpfs(t, TestPath, func(node *ipfsutil.Node) {
		if err := node.Online(); err != nil {
			t.Errorf("Failed to go online: %v", err)
			return
		}

		t.Logf("Im online.")

		data := []byte("Im a moose")
		mh, err := ipfsutil.AddBlock(node, data)
		if err != nil {
			t.Errorf("Adding block failed: %v", err)
			return
		}

		t.Logf("Added block.")

		peers, err := ipfsutil.Locate(node, mh, 1, 5*time.Second)

		t.Logf("Located.")

		if err != nil {
			t.Errorf("Looking up providers failed: %v", err)
			return
		}

		for _, peer := range peers {
			// TODO: check if
			fmt.Println(peer)
		}

		blockData, err := ipfsutil.CatBlock(node, mh, 1*time.Second)
		if err != nil {
			t.Errorf("Retrieving block failed: %v", err)
			return
		}

		if !bytes.Equal(data, blockData) {
			t.Errorf("Returned block data differs.")
			t.Errorf("\tExpect: %v", data)
			t.Errorf("\tGot:    %v", blockData)
			return
		}

		// Modify the hash and hope it there is none like that yet.
		mh[0] = 0
		_, err = ipfsutil.CatBlock(node, mh, 1*time.Second)
		if err != ipfsutil.ErrTimeout {
			t.Errorf("Oops, is there really a hash like that? %v", err)
			return
		}

	})
}
