package security

import (
	"bytes"
	"crypto/aes"
	"crypto/hmac"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"

	"golang.org/x/crypto/sha3"
)

const (
	nonceSize      = 62
	MaxMessageSize = 16 * 1024 * 1024
)

// AuthReadWriter acts as a layer on top of a normal io.ReadWriteCloser
// that adds authentication of the communication partners.
// It does this by employing the following protocol:
//
//  Alice                        Bob
//         ->  Pub_Bob(rA)   ->
//         <-  Pub_Alice(rB) <-
//         <---  sha3(rA)  <--
//         --->  sha3(rB)  --->
//
// Where rA and rB are randomly generated $nonceSize byte nonces.
// By being able to decrypt the nonce, alice and bob
// proved knowledge of the private key. The response to the challenge
// is exchanged as sha3-hash (512 bit) of the original nonce.
// This way an attacker is not able to read the nonces.
//
// If the protocol ran through succesfully, a symmetric session key is generated
// by xoring both nonces.  Read() and Write() will then proceed to use
// it to encrypt all messages written over it using CFB mode.
type AuthReadWriter struct {
	rwc        io.ReadWriteCloser
	pubKey     Encrypter
	privKey    Decrypter
	crypted    io.ReadWriter
	symkey     []byte
	authorised bool
	readBuf    *bytes.Buffer
}

// NewAuthReadWriter returns a new AuthReadWriter, authenticating rwc.
// `own` is our own private key, while `partner` is the partner's public key.
func NewAuthReadWriter(rwc io.ReadWriteCloser, own Decrypter, partner Encrypter) *AuthReadWriter {
	return &AuthReadWriter{
		rwc:     rwc,
		privKey: own,
		pubKey:  partner,
		readBuf: &bytes.Buffer{},
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
	if _, err := io.ReadFull(r, sizeBuf); err != nil {
		return nil, err
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
	// Generate our own nonce:
	rA := make([]byte, nonceSize)
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

	// nonceFromBob is their nonce:
	nonceFromBob, err := ath.privKey.Decrypt(chlFromBob)
	if err != nil {
		return err
	}

	if len(nonceFromBob) != nonceSize {
		return fmt.Errorf(
			"Bad nonce size from partner: %d (not %d)",
			len(nonceFromBob),
			nonceSize,
		)
	}

	// Send back our challenge-response
	respHash := sha3.Sum512(nonceFromBob)
	if _, err := ath.rwc.Write(respHash[:]); err != nil {
		return err
	}

	// Read the response from bob to our challenge
	hashFromBob := make([]byte, 512/8)
	if _, err := io.ReadFull(ath.rwc, hashFromBob); err != nil {
		return err
	}

	ownHash := sha3.Sum512(rA)
	if !bytes.Equal(hashFromBob, ownHash[:]) {
		return fmt.Errorf("Bad nonce; might communicate with imposter")
	}

	keySource := make([]byte, nonceSize)
	for i := range keySource {
		keySource[i] = nonceFromBob[i] ^ rA[i]
	}

	key := DeriveKey(keySource, keySource[:nonceSize/2], 32)
	inv := DeriveKey(keySource, keySource[nonceSize/2:], aes.BlockSize)

	rw, err := WrapReadWriter(inv, key, ath.rwc)
	if err != nil {
		return err
	}

	ath.symkey = key
	ath.crypted = rw
	ath.authorised = true
	return nil
}

func (ath *AuthReadWriter) Trigger() error {
	if !ath.Authorised() {
		if err := ath.runAuth(); err != nil {
			ath.rwc.Close()
			return err
		}
	}

	return nil
}

func (ath *AuthReadWriter) readMessage() ([]byte, error) {
	header := make([]byte, 28+4)

	if _, err := io.ReadFull(ath.rwc, header); err != nil {
		return nil, err
	}

	size := binary.LittleEndian.Uint32(header[28:])
	if size > MaxMessageSize {
		return nil, fmt.Errorf("Message too large (%d/%d)", size, MaxMessageSize)
	}

	buf := make([]byte, size)

	if _, err := io.ReadAtLeast(ath.crypted, buf, int(size)); err != nil {
		return nil, err
	}

	macWriter := hmac.New(sha3.New224, ath.symkey)
	if _, err := macWriter.Write(buf); err != nil {
		return nil, err
	}

	mac := macWriter.Sum(nil)
	if !hmac.Equal(mac, header[:28]) {
		return nil, fmt.Errorf("Mac differs in received metadata message")
	}

	return buf, nil
}

func (ath *AuthReadWriter) Read(buf []byte) (int, error) {
	if err := ath.Trigger(); err != nil {
		return 0, err
	}

	n := 0
	bufLen := len(buf)

	for {
		if ath.readBuf.Len() > 0 {
			bn, berr := ath.readBuf.Read(buf)
			if berr != nil && berr != io.EOF {
				return n, berr
			}

			n += bn
			buf = buf[bn:]

			if berr == io.EOF {
				break
			}
		}

		if n >= bufLen {
			return n, nil
		}

		msg, err := ath.readMessage()
		if err != nil {
			return n, err
		}

		if _, werr := ath.readBuf.Write(msg); werr != nil {
			return n, err
		}
	}

	return n, nil
}

func (ath *AuthReadWriter) Write(buf []byte) (int, error) {
	if err := ath.Trigger(); err != nil {
		return 0, err
	}

	macWriter := hmac.New(sha3.New224, ath.symkey)
	if _, err := macWriter.Write(buf); err != nil {
		return -1, err
	}

	mac := macWriter.Sum(nil)
	n, err := ath.rwc.Write(mac)
	if err != nil {
		return -2, err
	}

	if n != len(mac) {
		return -3, fmt.Errorf(
			"Unable to write full mac. Should be %d; was %d",
			len(mac),
			n,
		)
	}

	// Note: this assumes that `crypted` does not pad the data.
	sizeBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(sizeBuf, uint32(len(buf)))

	n, err = ath.rwc.Write(sizeBuf)
	if err != nil {
		return -4, err
	}

	if n != len(sizeBuf) {
		return -5, fmt.Errorf(
			"Unable to write full size buf. Should be %d; was %d",
			len(sizeBuf),
			n,
		)
	}

	return ath.crypted.Write(buf)
}
