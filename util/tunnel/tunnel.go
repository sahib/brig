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

	"github.com/tang0th/go-ecdh"
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

func NewEllipticTunnel(rw io.ReadWriter) (io.ReadWriter, error) {
	return NewEllipticTunnelWithCurve(rw, elliptic.P256())
}

func NewEllipticTunnelWithCurve(rw io.ReadWriter, curve elliptic.Curve) (io.ReadWriter, error) {
	tnl := &ecdhTunnel{
		ReadWriter: rw,
		ecdh:       ecdh.NewEllipticECDH(curve),
	}

	privKey, pubKey, err := tnl.ecdh.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	tnl.privKey = privKey
	tnl.pubKey = pubKey
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

	// Aim for AES 256
	if len(secret) < 32 {
		return fmt.Errorf("Secret too short")
	}

	blockCipher, err := aes.NewCipher(secret[:32])
	if err != nil {
		return err
	}

	// TODO: Is it okay to use the secret as IV?
	iv := secret[:aes.BlockSize]
	tnl.streamW = &cipher.StreamWriter{
		S: cipher.NewCFBEncrypter(blockCipher, iv),
		W: tnl.ReadWriter,
	}
	tnl.streamR = &cipher.StreamReader{
		S: cipher.NewCFBDecrypter(blockCipher, iv),
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
