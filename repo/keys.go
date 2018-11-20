package repo

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/alokmenghrajani/gpgeez"
	"github.com/sahib/brig/util"
	"golang.org/x/crypto/openpgp"
)

// create a new gpg key pair with self-signed subkeys
func createKeyPair(owner, folder string, bits int) error {
	// Setting expiry time to zero is good enough for now.
	// (key wil never expire; not sure yet if expiring keys make sense for brig)
	cfg := gpgeez.Config{
		Expiry: 0 * time.Second,
	}

	cfg.RSABits = bits
	comment := fmt.Sprintf("brig gpg key of %s", owner)
	key, err := gpgeez.CreateKey(owner, comment, owner, &cfg)
	if err != nil {
		return err
	}

	pubPath := filepath.Join(folder, "gpg.pub")
	prvPath := filepath.Join(folder, "gpg.prv")
	if err := ioutil.WriteFile(pubPath, key.Keyring(), 0600); err != nil {
		return err
	}

	return ioutil.WriteFile(prvPath, key.Secring(&cfg), 0600)
}

// encryptAsymmetric loads the pubkey from `folder` and encrypts `data` with it.
// This is not an efficient method and is not supposed to be used for large
// amounts of data.
func encryptAsymmetric(data, pubKey []byte) ([]byte, error) {
	ents, err := openpgp.ReadKeyRing(bytes.NewReader(pubKey))
	if err != nil {
		return nil, err
	}

	encBuf := &bytes.Buffer{}
	encW, err := openpgp.Encrypt(encBuf, ents, nil, nil, nil)
	if err != nil {
		return nil, err
	}

	if _, err := encW.Write(data); err != nil {
		return nil, err
	}

	if err := encW.Close(); err != nil {
		return nil, err
	}

	return encBuf.Bytes(), nil
}

// decryptAsymetric uses the private key from `folder` to decrypt `data`.
// This is not an efficient method and is not supposed to be used for large
// amounts of data.
func decryptAsymetric(folder string, data []byte) ([]byte, error) {
	prvPath := filepath.Join(folder, "gpg.prv")
	fd, err := os.Open(prvPath)
	if err != nil {
		return nil, err
	}

	defer util.Closer(fd)

	ents, err := openpgp.ReadKeyRing(fd)
	if err != nil {
		return nil, err
	}

	md, err := openpgp.ReadMessage(bytes.NewReader(data), ents, nil, nil)
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(md.UnverifiedBody)
}

// Keyring manages our own keypair and stores the last known
// pubkeys of other remotes.
type Keyring struct {
	folder string
}

func newKeyringHandle(folder string) *Keyring {
	return &Keyring{folder: folder}
}

// Encrypt `data` with `pubKey`.
// If it's desired to encrypt a message with our own pubkey,
// then use the PubKeyBytes() method to load one.
// This is not an efficient method and is not supposed to be used for large
// amounts of data.
func (kp *Keyring) Encrypt(data, pubKey []byte) ([]byte, error) {
	return encryptAsymmetric(data, pubKey)
}

// Decrypt decrypts a message encrypted with our public key.
// This is not an efficient method and is not supposed to be used for large
// amounts of data.
func (kp *Keyring) Decrypt(data []byte) ([]byte, error) {
	return decryptAsymetric(kp.folder, data)
}

// OwnPubKey returns an exported version of our own public key.
func (kp *Keyring) OwnPubKey() ([]byte, error) {
	pubPath := filepath.Join(kp.folder, "gpg.pub")
	return ioutil.ReadFile(pubPath)
}

// PubKeyFor returns the stored public key for a partner named `name`
func (kp *Keyring) PubKeyFor(name string) ([]byte, error) {
	path := filepath.Join(kp.folder, "pubkeys", filepath.Clean(name))
	return ioutil.ReadFile(path)
}

// SavePubKey stores a public key from a partner with the name `name`
func (kp *Keyring) SavePubKey(name string, pubKey []byte) error {
	base := filepath.Join(kp.folder, "pubkeys")
	if err := os.MkdirAll(base, 0700); err != nil {
		return err
	}

	pubKeyPath := filepath.Join(base, filepath.Clean(name))
	return ioutil.WriteFile(pubKeyPath, pubKey, 0600)
}
