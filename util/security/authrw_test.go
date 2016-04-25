package security

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"io"
	"net"
	"testing"

	"github.com/disorganizer/brig/util/testutil"
)

const (
	dummyAddr = "127.0.0.1:7782" // Just a random high port
)

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

type KeyPair struct {
	priv *rsa.PrivateKey
}

func (kp *KeyPair) Encrypt(data []byte) ([]byte, error) {
	pub := &kp.priv.PublicKey
	return rsa.EncryptOAEP(sha256.New(), rand.Reader, pub, data, nil)
}

func (kp *KeyPair) Decrypt(data []byte) ([]byte, error) {
	opts := &rsa.OAEPOptions{
		Hash:  crypto.SHA256,
		Label: nil,
	}

	return kp.priv.Decrypt(rand.Reader, data, opts)
}

func genKeyPair(t *testing.T) (Decrypter, Encrypter, error) {
	// Generate an dummy pub/priv keypair using ecdsa
	// Note: We just use the same pair for one auther.
	// In practice, this should be
	priv, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatalf("Unable to generate dummy RSA key: %v", err)
		return nil, nil, err
	}

	kp := &KeyPair{priv: priv}
	return kp, kp, nil
}

func testAuthProcess(t *testing.T, size int64) {
	withLoopbackConnection(t, func(a, b net.Conn) {
		aliPriv, aliPub, err := genKeyPair(t)
		if err != nil {
			return
		}

		bobPriv, bobPub, err := genKeyPair(t)
		if err != nil {
			return
		}

		autherA := NewAuthReadWriter(a, aliPriv, bobPub)
		autherB := NewAuthReadWriter(b, bobPriv, aliPub)

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

		go func() {
			if _, err := autherA.Write(expect); err != nil {
				t.Errorf("Auth Write failed: %v", err)
				return
			}
		}()

		if _, err := autherB.Read(answer); err != nil {
			t.Errorf("Auth Read failed: %v", err)
			return
		}

		if !bytes.Equal(expect, answer) {
			t.Errorf("auth transmission failed; want `%s`, got `%s`", expect, answer)
			return
		}
	})
}

func TestAuthProcess(t *testing.T) {
	sizes := []int64{0, 255}
	for i := uint(0); i < 18; i++ {
		sizes = append(sizes, int64(1<<i))
	}

	for _, size := range sizes {
		t.Logf("Testing size %d", size)
		testAuthProcess(t, size)

		if t.Failed() {
			break
		}
	}
}
