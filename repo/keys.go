package repo

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/alokmenghrajani/gpgeez"
	"github.com/disorganizer/brig/util"
	"golang.org/x/crypto/openpgp"
)

func createKeyPair(owner, folder string, bits int) error {
	// Setting expiry time to zero is good enough for now.
	// (key wil never expire, not sure yet if expiring keys make sense for brig)
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

	if err := ioutil.WriteFile(prvPath, key.Secring(&cfg), 0600); err != nil {
		return err
	}

	return nil
}

func encryptAsymmetric(folder string, data []byte) ([]byte, error) {
	pubPath := filepath.Join(folder, "gpg.pub")
	fd, err := os.Open(pubPath)
	if err != nil {
		return nil, err
	}

	defer util.Closer(fd)

	ents, err := openpgp.ReadKeyRing(fd)
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

func readPubKey(folder string) ([]byte, error) {
	pubPath := filepath.Join(folder, "gpg.pub")
	return ioutil.ReadFile(pubPath)
}

type KeyPair struct {
	folder string
}

func newKeyPairHandle(folder string) *KeyPair {
	return &KeyPair{folder: folder}
}

func (kp *KeyPair) Encrypt(data []byte) ([]byte, error) {
	return encryptAsymmetric(kp.folder, data)
}

func (kp *KeyPair) Decrypt(data []byte) ([]byte, error) {
	return decryptAsymetric(kp.folder, data)
}

func (kp *KeyPair) PubKeyBytes() ([]byte, error) {
	return readPubKey(kp.folder)
}
