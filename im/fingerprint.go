package im

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/util"
	"gopkg.in/yaml.v2"
)

// FingerprintStore represents an arbitary store where fingerprints are stored.
type FingerprintStore interface {
	// Lookup returns the last known fingerprint related to this jid.
	Lookup(jid string) (string, error)

	// Remember stores the last known fingerprint of this jid.
	Remember(jid string, fingerprint string) error
}

// FormatFingerprint converts a raw byte string representation to a hex fingerprint.
func FormatFingerprint(raw []byte) string {
	// NOTE: This is a little stupid, but fits in one line:
	return strings.Replace(fmt.Sprintf("% X", raw), " ", ":", -1)
}

// FsFingerprintStore represents a FingerprintStore that saves it's contents to
// a YAML file on the filesystem at an absolute path.
type FsFingerprintStore struct {
	Path string
	keys map[string]string
}

// NewFsFingerprintStore returns a new, possibly empty, FingerprintStore
func NewFsFingerprintStore(path string) (*FsFingerprintStore, error) {
	k := FsFingerprintStore{Path: path}

	fd, err := os.Open(path)
	if err != nil {
		k.keys = make(map[string]string)
		return &k, nil
	}

	defer util.Closer(fd)

	data, err := ioutil.ReadAll(fd)
	if err != nil {
		return nil, err
	}

	return &k, yaml.Unmarshal(data, &k.keys)
}

// Lookup returns the last know fingerprint of this jid. No I/O is done.
func (k *FsFingerprintStore) Lookup(jid string) (string, error) {
	fingerprint, ok := k.keys[jid]
	if !ok {
		log.Warningf("No fingerprint known for `%v`.", jid)
	}

	return fingerprint, nil
}

// Remember stores the last knwon fingerprint to this jid. It rewrites the
// fingerprint database on the filesystem
func (k *FsFingerprintStore) Remember(jid string, fingerprint string) error {
	k.keys[jid] = fingerprint

	fd, err := os.OpenFile(k.Path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	defer util.Closer(fd)

	data, err := yaml.Marshal(&k.keys)
	if err != nil {
		return err
	}

	if _, err := fd.Write(data); err != nil {
		return err
	}

	return nil
}
