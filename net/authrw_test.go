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
	"github.com/sahib/brig/net/peer"
	"github.com/sahib/brig/util/testutil"
	"github.com/stretchr/testify/require"
)

// create a new gpg key pair with self-signed subkeys
func createKeyPair(t *testing.T, bits int) ([]byte, []byte) {
	cfg := gpgeez.Config{Expiry: 0 * time.Second}

	cfg.RSABits = bits
	comment := fmt.Sprintf("brig gpg key of %s", "alice")
	key, err := gpgeez.CreateKey("alice", comment, "alice", &cfg)
	if err != nil {
		t.Fatalf("Failed to create gpg key pair: %v", err)
	}

	return key.Secring(&cfg), key.Keyring()
}

// Do not use repo.Keyring, simply re-implement for this test's purpose.
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
	// Test with a real connection and not use a simple net.Pipe().
	// net.Pipe() does not provide any buffering like real world connection do.
	// authrw relies on this...

	setup := make(chan bool)
	conCh := make(chan net.Conn)
	waitForTestCase := make(chan bool)

	ls, err := testutil.RandomLocalListener()
	if err != nil {
		t.Errorf("Listening on dummy port failed: %v", err)
		setup <- false
		return
	}

	go func() {
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

		// We should wait until we quit this function.
		// Otherwise the defer above will be executed.
		<-waitForTestCase
	}()

	if <-setup == false {
		return
	}

	clientSide, err := net.Dial("tcp", ls.Addr().String())
	if err != nil {
		t.Errorf("Dialing self failed: %v", err)
		return
	}

	select {
	case serverSide := <-conCh:
		// This is needed to write big messages in one go:
		// clientSide.(*net.TCPConn).SetWriteBuffer(1024 * 1024)
		clientSide.(*net.TCPConn).SetReadBuffer(1024 * 1024)
		// serverSide.(*net.TCPConn).SetWriteBuffer(1024 * 1024)
		serverSide.(*net.TCPConn).SetReadBuffer(1024 * 1024)

		f(clientSide, serverSide)
	case <-time.After(5 * time.Second):
		t.Fatalf("test took too long")
	}

	waitForTestCase <- true
}

func testAuthProcess(t *testing.T, size int64, privAli, privBob, pubAli, pubBob []byte) {
	withLoopbackConnection(t, func(a, b net.Conn) {
		authAli := NewAuthReadWriter(a, DummyPrivKey(privAli), pubAli, "ali", func(pubKey []byte) error {
			fpBob := peer.BuildFingerprint("bob", pubBob)
			if !fpBob.PubKeyMatches(pubKey) {
				return fmt.Errorf("bob has wrong public key")
			}

			return nil
		})
		authBob := NewAuthReadWriter(b, DummyPrivKey(privBob), pubBob, "bob", func(pubKey []byte) error {
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
		n, err := a.Write(expect)
		if err != nil {
			t.Errorf("Normal write failed: %v", err)
			return
		}

		require.Equal(t, len(expect), n)

		n, err = a.Read(answer)
		if err != nil && (err != io.EOF && size == 0) {
			t.Errorf("Normal read failed: %v", err)
			return
		}

		require.Equal(t, len(answer), n)

		if !bytes.Equal(expect, answer) {
			t.Errorf("Normal transmission failed; want `%v`, got `%v`", len(expect), len(answer))
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

		require.Equal(t, authBob.RemoteName(), "ali")
		require.Equal(t, authAli.RemoteName(), "bob")

		if !bytes.Equal(expect, answer) {
			t.Errorf("auth transmission failed; want `%s`, got `%s`", expect, answer)
			return
		}

		wg.Wait()
	})
}

func TestAuthProcess(t *testing.T) {
	t.Parallel()

	sizes := []int64{0, 255}
	// TODO: This breaks for size 131072 - let's find out why.

	for i := uint(0); i < 18; i++ {
		sizes = append(sizes, int64(1<<i))
	}

	privAli, pubAli := createKeyPair(t, 1024)
	privBob, pubBob := createKeyPair(t, 1024)

	for _, size := range sizes {
		t.Run(fmt.Sprintf("size-%d", size), func(t *testing.T) {
			testAuthProcess(t, size, privAli, privBob, pubAli, pubBob)
		})
	}
}
