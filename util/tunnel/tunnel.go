// Package tunnel implements an io.ReadWriter that encrypts it's data.
// Technically it performs a Elliptic Curve Diffie Hellman key exchange
// before the first read or write (or triggered manually using Exchange())
//
// All communication over the tunnel is encrypted with AES using CFB mode.
package tunnel

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"

	"github.com/disorganizer/brig/util/security"
	"golang.org/x/crypto/curve25519"
)

////////////////////////////////////////////////////////////////////
// The following code has been taken from go-ecdh:              //
// https://github.com/tang0th/go-ecdh/blob/master/curve25519.go //
// It's here to prevent another external dependency.            //
////////////////////////////////////////////////////////////////////

func generateKey(rand io.Reader) (crypto.PrivateKey, crypto.PublicKey, error) {
	var pub, priv [32]byte
	var err error

	_, err = io.ReadFull(rand, priv[:])
	if err != nil {
		return nil, nil, err
	}

	curve25519.ScalarBaseMult(&pub, &priv)
	return &priv, &pub, nil
}

func marshal(p crypto.PublicKey) []byte {
	pub := p.(*[32]byte)
	return pub[:]
}

func unmarshal(data []byte) (crypto.PublicKey, bool) {
	var pub [32]byte
	if len(data) != 32 {
		return nil, false
	}

	copy(pub[:], data)
	return &pub, true
}

func generateSharedSecret(privKey crypto.PrivateKey, pubKey crypto.PublicKey) ([]byte, error) {
	var priv, pub, secret *[32]byte

	priv = privKey.(*[32]byte)
	pub = pubKey.(*[32]byte)
	secret = new([32]byte)

	curve25519.ScalarMult(secret, priv, pub)
	return secret[:], nil
}

type ecdhTunnel struct {
	// Underlying ReadWriter
	ReadWriter io.ReadWriter

	// Elliptic Curve Diffie Hellman keys:
	privKey crypto.PrivateKey
	pubKey  crypto.PublicKey

	// CFB streaming ciphers for Read()/Write():
	streamW *cipher.StreamWriter
	streamR *cipher.StreamReader
}

// NewEllipticTunnel creates an io.ReadWriter that transparently encrypts all data.
func NewEllipticTunnel(rw io.ReadWriter) (io.ReadWriter, error) {
	tnl := &ecdhTunnel{
		ReadWriter: rw,
	}

	var err error
	tnl.privKey, tnl.pubKey, err = generateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	return tnl, nil
}

// Exchange triggers the Diffie Hellman key exchange manually.
func (tnl *ecdhTunnel) Exchange() error {
	if tnl.streamW != nil || tnl.streamR != nil {
		return nil
	}

	pubKeyBuf := marshal(tnl.pubKey)
	if _, err := tnl.ReadWriter.Write(pubKeyBuf); err != nil {
		return err
	}

	partnerBuf := make([]byte, len(pubKeyBuf))
	if _, err := tnl.ReadWriter.Read(partnerBuf); err != nil {
		return err
	}

	partnerKey, ok := unmarshal(partnerBuf)
	if !ok {
		return fmt.Errorf("Partner key unmarshal failed")
	}

	secret, err := generateSharedSecret(tnl.privKey, partnerKey)
	if err != nil {
		return err
	}

	// Transform the secret to a usable 32 byte key:
	key := security.Scrypt(secret, secret[:16], 32)
	inv := security.Scrypt(secret, secret[16:], aes.BlockSize)

	blockCipher, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	tnl.streamW = &cipher.StreamWriter{
		S: cipher.NewCFBEncrypter(blockCipher, inv),
		W: tnl.ReadWriter,
	}
	tnl.streamR = &cipher.StreamReader{
		S: cipher.NewCFBDecrypter(blockCipher, inv),
		R: tnl.ReadWriter,
	}
	return nil
}

// Read decrypts underlying data using CFB and will trigger a key exchange
// if this was not done yet for this session.
func (tnl *ecdhTunnel) Read(buf []byte) (int, error) {
	if err := tnl.Exchange(); err != nil {
		return 0, err
	}

	return tnl.streamR.Read(buf)
}

// Write encrypts incoming data using CFB and will trigger a key exchange
// if this was not done yet for this session.
func (tnl *ecdhTunnel) Write(buf []byte) (int, error) {
	if err := tnl.Exchange(); err != nil {
		return 0, err
	}

	n, e := tnl.streamW.Write(buf)
	return n, e
}
