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
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"io"

	"github.com/disorganizer/brig/util/security"
	// TODO: Change back to tang0th when merged:
	//       https://github.com/tang0th/go-ecdh/pull/1
	"github.com/disorganizer/go-ecdh"
)

type ecdhTunnel struct {
	// Underlying ReadWriter
	ReadWriter io.ReadWriter

	// Elliptic Curve Diffie Hellman state and keys:
	ecdh    ecdh.ECDH
	privKey crypto.PrivateKey
	pubKey  crypto.PublicKey

	// CFB streaming ciphers for Read()/Write():
	streamW *cipher.StreamWriter
	streamR *cipher.StreamReader
}

// NewEllipticTunnel creates an io.ReadWriter that transparently encrypts all data.
func NewEllipticTunnel(rw io.ReadWriter) (io.ReadWriter, error) {
	// TODO: Find safe elliptic curve
	return newEllipticTunnelWithCurve(rw, elliptic.P521())
}

func newEllipticTunnelWithCurve(rw io.ReadWriter, curve elliptic.Curve) (io.ReadWriter, error) {
	tnl := &ecdhTunnel{
		ReadWriter: rw,
		ecdh:       ecdh.NewEllipticECDH(curve),
	}

	var err error
	tnl.privKey, tnl.pubKey, err = tnl.ecdh.GenerateKey(rand.Reader)
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

	pubKeyBuf := tnl.ecdh.Marshal(tnl.pubKey)
	if _, err := tnl.ReadWriter.Write(pubKeyBuf); err != nil {
		return err
	}

	partnerBuf := make([]byte, len(pubKeyBuf))
	if _, err := tnl.ReadWriter.Read(partnerBuf); err != nil {
		return err
	}

	partnerKey, ok := tnl.ecdh.Unmarshal(partnerBuf)
	if !ok {
		return fmt.Errorf("Partner key unmarshal failed")
	}

	secret, err := tnl.ecdh.GenerateSharedSecret(tnl.privKey, partnerKey)
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
