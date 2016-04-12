package ipfsutil_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/disorganizer/brig/util/ipfsutil"
	"github.com/disorganizer/brig/util/testwith"
)

func TestAddCat(t *testing.T) {
	testwith.WithIpfs(t, func(node *ipfsutil.Node) {
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
	testwith.WithIpfs(t, func(node *ipfsutil.Node) {
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

func TestNet(t *testing.T) {
	testwith.WithIpfsAtPort(t, 4002, func(alice *ipfsutil.Node) {
		if err := alice.Online(); err != nil {
			t.Errorf("alice failed to go online: %v", err)
			return
		}

		fmt.Println("Alice is online.")
		testwith.WithIpfsAtPort(t, 4003, func(bob *ipfsutil.Node) {
			if err := bob.Online(); err != nil {
				t.Errorf("bob failed to go online: %v", err)
				return
			}

			fmt.Println("bob is online.")

			ls, err := bob.Listen()
			if err != nil {
				t.Errorf("Failed to listen on ipfs: %v", err)
				return
			}

			go func() {
				fmt.Println("Accept")
				conn, err := ls.Accept()
				if err != nil {
					t.Errorf("Accept() failed: %v", err)
					return
				}

				fmt.Println("Accept-Read")
				buf := make([]byte, 5)
				if n, err := conn.Read(buf); err != nil && n != len(buf) {
					t.Errorf("Listen-Read failed: %v (len: %d)", err, n)
					return
				}

				fmt.Println("Accept-Write")
				if _, err := conn.Write([]byte("World")); err != nil {
					t.Errorf("Liste-Write failed: %v", err)
					return
				}

				if err := conn.Close(); err != nil {
					t.Errorf("Listen-Close conn failed: %v", err)
					return
				}

				if err := ls.Close(); err != nil {
					t.Errorf("Closing listener failed: %v", err)
					return
				}
			}()

			// Alice sending data to bob:
			bobId, err := bob.Identity()
			if err != nil {
				t.Errorf("Could not get self identity: %v", err)
				return
			}

			fmt.Println("Dial", bobId)
			conn, err := alice.Dial(bobId)
			if err != nil {
				t.Errorf("Dial(self) did not work: %v", err)
				return
			}

			fmt.Println("Write")
			if _, err := conn.Write([]byte("Hello")); err != nil {
				t.Errorf("Write(self) failed: %v", err)
				return
			}

			buf := make([]byte, 5)
			if n, err := conn.Read(buf); err != nil && n != len(buf) {
				t.Errorf("Read(self) failed: %v (len: %d)", err, n)
				return
			}

			if !bytes.Equal(buf, []byte("World")) {
				t.Errorf("Read data does not match. Expected 'World'; got '%s'", buf)
				return
			}

			if err := conn.Close(); err != nil {
				t.Errorf("Closing conn failed: %v", err)
				return
			}

		})
	})
}
