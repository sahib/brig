package im

import (
	"bytes"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

type KeyStore interface {
	Lookup(jid string) ([]byte, error)
	Remember(jid string, fingerprint []byte) error
	Match(jid string, current []byte) bool
}

type FsKeyStore struct {
	Path string
	keys map[string][]byte
}

func NewFsKeyStore(path string) (*FsKeyStore, error) {
	k := FsKeyStore{Path: path}

	fd, err := os.Open(path)
	if err != nil {
		k.keys = make(map[string][]byte)
		return &k, nil
	}

	defer fd.Close()

	data, err := ioutil.ReadAll(fd)
	if err != nil {
		return nil, err
	}

	return &k, yaml.Unmarshal(data, &k.keys)
}

func (k *FsKeyStore) Lookup(jid string) ([]byte, error) {
	return k.keys[jid], nil
}

func (k *FsKeyStore) Remember(jid string, fingerprint []byte) error {
	k.keys[jid] = fingerprint[:]

	fd, err := os.OpenFile(k.Path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	defer fd.Close()

	data, err := yaml.Marshal(&k.keys)
	if err != nil {
		return err
	}

	if _, err := fd.Write(data); err != nil {
		return err
	}

	return nil
}

func (k *FsKeyStore) Match(jid string, current []byte) bool {
	old, err := k.Lookup(jid)
	if err != nil {
		return false
	}

	// TODO: Later this should be only done by the initial auth module.
	if old == nil {
		return true
	}

	return bytes.Equal(old, current)
}
