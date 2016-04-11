package security

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
)

// PublicKey is used as encryption key when passed to AuthReadWriter.
type PublicKey interface {
	Encrypt(data []byte) ([]byte, error)
}

// PrivateKey is used as decryption key when passed to AuthReadWriter.
type PrivateKey interface {
	Decrypt(data []byte) ([]byte, error)
}

// AuthReadWriter acts as a layer on top of a normal io.ReadWriteCloser
// that adds authentication of the communication partners.
// It does this by employing the following protocol:
//
//  Alice                        Bob
//         ->  Pub_Bob(rA)   ->
//         <-  Pub_Alice(rB) <-
//         <------  rA  <------
//         ------>  rB  ------>
//
// Where rA and rB are randomly generated 8 byte nonces.
// By being able to decrypt the nonce, alice and bob
// proved knowledge of the private key.
// If the protocol ran through succesfully, Read()
// and Write() will delegate to the underlying ReadWriteCloser,
// if not Close() will be called an Authorised() will return false.
type AuthReadWriter struct {
	rwc     io.ReadWriteCloser
	pubKey  PublicKey
	privKey PrivateKey

	authorised bool
}

// NewAuthReadWriter returns a new AuthReadWriter, authenticating rwc.
// `own` is our own private key, while `partner` is the partner's public key.
func NewAuthReadWriter(rwc io.ReadWriteCloser, own PrivateKey, partner PublicKey) *AuthReadWriter {
	return &AuthReadWriter{
		rwc:     rwc,
		privKey: own,
		pubKey:  partner,
	}
}

// Authorised will return true if the partner was succesfully authenticated.
// It will return false if no call to Read() or Write() was made.
func (ath *AuthReadWriter) Authorised() bool {
	return ath.authorised
}

// writeSizePack prefixes a datablock by it's binary size
func writeSizePack(w io.Writer, data []byte) (int, error) {
	pack := make([]byte, 8)
	binary.LittleEndian.PutUint64(pack, uint64(len(data)))

	if n, err := w.Write(append(pack, data...)); err != nil {
		return n, err
	}

	return len(pack) + len(data), nil
}

// readSizePack reads a 8 byte size prefix and return the following data block.
// If the block appears too large, it will error out.
func readSizePack(r io.Reader) ([]byte, error) {
	sizeBuf := make([]byte, 8)
	n, err := r.Read(sizeBuf)
	if err != nil {
		return nil, err
	}

	if n < 8 {
		return nil, fmt.Errorf("Insufficiently sized data: %d", n)
	}

	size := binary.LittleEndian.Uint64(sizeBuf)

	// Protect against unreasonable sizes:
	if size > 4096 {
		return nil, fmt.Errorf("Auth package is oversized: %d", size)
	}

	buf := make([]byte, size)
	if _, err := io.ReadAtLeast(r, buf, int(size)); err != nil {
		return nil, err
	}

	return buf, nil
}

// runAuth runs the protocol pointed out above.
func (ath *AuthReadWriter) runAuth() error {
	// Generate an 8 byte nonce:
	rA := make([]byte, 8)
	if _, err := io.ReadFull(rand.Reader, rA); err != nil {
		return err
	}

	// Send our challenge encrypted with remote's public key.
	chlForBob, err := ath.pubKey.Encrypt(rA)
	if err != nil {
		return err
	}

	if _, err := writeSizePack(ath.rwc, chlForBob); err != nil {
		return err
	}

	// Read their challenge (nonce encrypted with our pubkey)
	chlFromBob, err := readSizePack(ath.rwc)
	if err != nil {
		return err
	}

	respForBob, err := ath.privKey.Decrypt(chlFromBob)
	if err != nil {
		return err
	}

	// Send back our challenge-response
	if _, err := ath.rwc.Write(respForBob); err != nil {
		return err
	}

	// Read the response from bob to our challenge
	respFromBob := make([]byte, 8)
	if _, err := io.ReadFull(ath.rwc, respFromBob); err != nil {
		return err
	}

	if !bytes.Equal(respFromBob, rA) {
		return fmt.Errorf("Bad nonce; might communicate with imposter")
	}

	okSend := []byte("OK")
	if _, err := ath.rwc.Write(okSend); err != nil {
		return err
	}

	okBuf := make([]byte, 2)
	if _, err := io.ReadFull(ath.rwc, okBuf); err != nil {
		return err
	}

	if !bytes.Equal(okBuf, okSend) {
		return fmt.Errorf("Received no OK: %v", okBuf)
	}

	ath.authorised = true
	return nil
}

func (ath *AuthReadWriter) checkAuth() error {
	if !ath.Authorised() {
		if err := ath.runAuth(); err != nil {
			ath.rwc.Close()
			return err
		}
	}

	return nil
}

func (ath *AuthReadWriter) Read(buf []byte) (int, error) {
	if err := ath.checkAuth(); err != nil {
		return 0, err
	}

	return ath.rwc.Read(buf)
}

func (ath *AuthReadWriter) Write(buf []byte) (int, error) {
	if err := ath.checkAuth(); err != nil {
		return 0, err
	}

	return ath.rwc.Write(buf)
}
