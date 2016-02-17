package ipfsutil_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/disorganizer/brig/util/ipfsutil"
	"github.com/disorganizer/brig/util/testwith"
)

var (
	TestPath = filepath.Join(os.TempDir(), "brig_test_ipfs_repo")
)

func TestStartDaemon(t *testing.T) {
	testwith.WithIpfs(t, TestPath, func(node *ipfsutil.Node) {
		if node.IpfsNode == nil {
			t.Errorf("withIpfs created an invalid Node.")
		}
	})
}

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
