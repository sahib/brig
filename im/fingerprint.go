package im

import (
	"bytes"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

// FingerprintStore represents an arbitary store where fingerprints are stored.
type FingerprintStore interface {
	// Lookup returns the last known fingerprint related to this jid.
	Lookup(jid string) ([]byte, error)

	// Remember stores the last known fingerprint of this jid.
	Remember(jid string, fingerprint []byte) error

	// Match checks if the current fingerprint matches the last one.
	Match(jid string, current []byte) bool
}

// FsFingerprintStore represents a FingerprintStore that saves it's contents to
// a YAML file on the filesystem at an absolute path.
type FsFingerprintStore struct {
	Path string
	keys map[string][]byte
}

// NewFsFingerprintStore returns a new, possibly empty, FingerprintStore
func NewFsFingerprintStore(path string) (*FsFingerprintStore, error) {
	k := FsFingerprintStore{Path: path}

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

// Lookup returns the last know fingerprint of this jid. No I/O is done.
func (k *FsFingerprintStore) Lookup(jid string) ([]byte, error) {
	return k.keys[jid], nil
}

// Remember stores the last knwon fingerprint to this jid. It rewrites the
// fingerprint database on the filesystem
func (k *FsFingerprintStore) Remember(jid string, fingerprint []byte) error {
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

// Match does a Lookup and compares it with the current fingerprint for
// convinience.
func (k *FsFingerprintStore) Match(jid string, current []byte) bool {
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
