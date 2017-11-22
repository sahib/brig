package net

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"sync"
	"testing"
	"time"

	"golang.org/x/crypto/openpgp"

	"github.com/alokmenghrajani/gpgeez"
	"github.com/disorganizer/brig/net/peer"
	"github.com/disorganizer/brig/util/testutil"
)

const (
	dummyAddr = "127.0.0.1:7782" // Just a random high port
)

// create a new gpg key pair with self-signed subkeys
func createKeyPair(t *testing.T, bits int) ([]byte, []byte) {
	// Setting expiry time to zero is good enough for now.
	// (key wil never expire; not sure yet if expiring keys make sense for brig)
	cfg := gpgeez.Config{
		Expiry: 0 * time.Second,
	}

	cfg.RSABits = bits
	comment := fmt.Sprintf("brig gpg key of %s", "alice")
	key, err := gpgeez.CreateKey("alice", comment, "alice", &cfg)
	if err != nil {
		t.Fatalf("Failed to create gpg key pair: %v", err)
	}

	return key.Secring(&cfg), key.Keyring()
}

type DummyPrivKey []byte

func (pk DummyPrivKey) Decrypt(data []byte) ([]byte, error) {
	ents, err := openpgp.ReadKeyRing(bytes.NewReader(pk))
	if err != nil {
		return nil, err
	}

	md, err := openpgp.ReadMessage(bytes.NewReader(data), ents, nil, nil)
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(md.UnverifiedBody)
}

func withLoopbackConnection(t *testing.T, f func(a, b net.Conn)) {
	setup := make(chan bool)
	conCh := make(chan net.Conn)

	go func() {
		ls, err := net.Listen("tcp", dummyAddr)
		if err != nil {
			t.Errorf("Listening on dummy port failed: %v", err)
			setup <- false
			return
		}

		setup <- true
		defer func() {
			if err := ls.Close(); err != nil {
				t.Errorf("Closing listener failed: %v", err)
			}
			close(conCh)
		}()

		conn, err := ls.Accept()
		if err != nil {
			t.Errorf("Accepting on dummy listener failed: %v", err)
			return
		}

		conCh <- conn
	}()

	if <-setup == false {
		return
	}

	clientSide, err := net.Dial("tcp", dummyAddr)
	if err != nil {
		t.Errorf("Dialing self failed: %v", err)
		return
	}

	serverSide := <-conCh
	f(clientSide, serverSide)
}

func testAuthProcess(t *testing.T, size int64, privAli, privBob, pubAli, pubBob []byte) {
	withLoopbackConnection(t, func(a, b net.Conn) {
		authAli := NewAuthReadWriter(a, DummyPrivKey(privAli), pubAli, func(pubKey []byte) error {
			fpBob := peer.BuildFingerprint("bob", pubBob)
			if !fpBob.PubKeyMatches(pubKey) {
				return fmt.Errorf("bob has wrong public key")
			}

			return nil
		})
		authBob := NewAuthReadWriter(b, DummyPrivKey(privBob), pubBob, func(pubKey []byte) error {
			fpAli := peer.BuildFingerprint("ali", pubAli)
			if !fpAli.PubKeyMatches(pubKey) {
				return fmt.Errorf("alice has wrong public key")
			}

			return nil
		})

		expect := testutil.CreateDummyBuf(size)
		answer := make([]byte, len(expect))

		// Sort out connection based troubles quickly:
		// Just send a normal message over the conn.
		if _, err := a.Write(expect); err != nil {
			t.Errorf("Normal write failed: %v", err)
			return
		}

		if _, err := b.Read(answer); err != nil && (err != io.EOF && size == 0) {
			t.Errorf("Normal read failed: %v", err)
			return
		}

		if !bytes.Equal(expect, answer) {
			t.Errorf("Normal transmission failed; want `%s`, got `%s`", expect, answer)
			return
		}

		answer = make([]byte, len(expect))

		wg := &sync.WaitGroup{}

		go func() {
			wg.Add(1)
			defer wg.Done()

			if _, err := authAli.Write(expect); err != nil {
				t.Errorf("Auth Write failed: %v", err)
				return
			}
		}()

		if _, err := authBob.Read(answer); err != nil {
			t.Errorf("Auth Read failed: %v", err)
			return
		}

		if !bytes.Equal(expect, answer) {
			t.Errorf("auth transmission failed; want `%s`, got `%s`", expect, answer)
			return
		}

		wg.Wait()
	})
}

func TestAuthProcess(t *testing.T) {
	sizes := []int64{0, 255}
	for i := uint(0); i < 18; i++ {
		sizes = append(sizes, int64(1<<i))
	}

	privAli, pubAli := createKeyPair(t, 1024)
	privBob, pubBob := createKeyPair(t, 1024)

	for _, size := range sizes {
		t.Logf("Testing size %d", size)
		testAuthProcess(t, size, privAli, privBob, pubAli, pubBob)

		if t.Failed() {
			break
		}
	}
}
