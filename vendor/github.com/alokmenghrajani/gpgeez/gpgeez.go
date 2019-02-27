// Package gpgeez is a wrapper around golang.org/x/crypto/openpgp
package gpgeez

import (
	"bytes"
	"time"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
)

// Config for generating keys.
type Config struct {
	packet.Config
	// Expiry is the duration that the generated key will be valid for.
	Expiry time.Duration
}

// Key represents an OpenPGP key.
type Key struct {
	openpgp.Entity
}

// Values from https://tools.ietf.org/html/rfc4880#section-9
const (
	md5       = 1
	sha1      = 2
	ripemd160 = 3
	sha256    = 8
	sha384    = 9
	sha512    = 10
	sha224    = 11
)

// CreateKey creates an OpenPGP key which is similar to running gpg --gen-key
// on the command line. In other words, this method returns a primary signing
// key and an encryption subkey with expected self-signatures.
//
// There are a few differences:
//
// • GnuPG sets key server preference to 0x80, no-modify (see https://tools.ietf.org/html/rfc4880#section-5.2.3.17).
//
// • GnuPG sets features to 0x01, modification detection (see https://tools.ietf.org/html/rfc4880#page-36).
//
// • GnuPG sets the digest algorithm to SHA1. Go defaults to SHA256.
//
// • GnuPG includes Bzip2 as a compression method. Go currently doesn't support Bzip2, so that option isn't set.
//
// • Issuer key ID is hashed subpkt instead of subpkt, and contains a primary user ID sub packet.
//
// You can see these differences for yourself by comparing the output of:
//  go run example/create_key.go | gpg --homedir /tmp --list-packets
// with:
//  gpg --homedir /tmp --gen-key
//  gpg --homedir /tmp -a --export | gpg --homedir /tmp --list-packets
//
// Or just look at
// https://github.com/alokmenghrajani/gpgeez/blob/master/gpgeez_test.pl
//
// Some useful links:
// https://godoc.org/golang.org/x/crypto/openpgp,
// https://davesteele.github.io/gpg/2014/09/20/anatomy-of-a-gpg-key,
// https://github.com/golang/go/issues/12153
func CreateKey(name, comment, email string, config *Config) (*Key, error) {
	// Create the key
	key, err := openpgp.NewEntity(name, comment, email, &config.Config)
	if err != nil {
		return nil, err
	}

	// Set expiry and algorithms. Self-sign the identity.
	dur := uint32(config.Expiry.Seconds())
	for _, id := range key.Identities {
		id.SelfSignature.KeyLifetimeSecs = &dur

		id.SelfSignature.PreferredSymmetric = []uint8{
			uint8(packet.CipherAES256),
			uint8(packet.CipherAES192),
			uint8(packet.CipherAES128),
			uint8(packet.CipherCAST5),
			uint8(packet.Cipher3DES),
		}

		id.SelfSignature.PreferredHash = []uint8{
			sha256,
			sha1,
			sha384,
			sha512,
			sha224,
		}

		id.SelfSignature.PreferredCompression = []uint8{
			uint8(packet.CompressionZLIB),
			uint8(packet.CompressionZIP),
		}

		err := id.SelfSignature.SignUserId(id.UserId.Id, key.PrimaryKey, key.PrivateKey, &config.Config)
		if err != nil {
			return nil, err
		}
	}

	// Self-sign the Subkeys
	for _, subkey := range key.Subkeys {
		subkey.Sig.KeyLifetimeSecs = &dur
		err := subkey.Sig.SignKey(subkey.PublicKey, key.PrivateKey, &config.Config)
		if err != nil {
			return nil, err
		}
	}

	r := Key{*key}
	return &r, nil
}

// Armor returns the public part of a key in armored format.
func (key *Key) Armor() (string, error) {
	buf := new(bytes.Buffer)
	armor, err := armor.Encode(buf, openpgp.PublicKeyType, nil)
	if err != nil {
		return "", err
	}
	key.Serialize(armor)
	armor.Close()

	return buf.String(), nil
}

// ArmorPrivate returns the private part of a key in armored format.
//
// Note: if you want to protect the string against varous low-level attacks,
// you should look at https://github.com/stouset/go.secrets and
// https://github.com/worr/secstring and then re-implement this function.
func (key *Key) ArmorPrivate(config *Config) (string, error) {
	buf := new(bytes.Buffer)
	armor, err := armor.Encode(buf, openpgp.PrivateKeyType, nil)
	if err != nil {
		return "", err
	}
	c := config.Config
	key.SerializePrivate(armor, &c)
	armor.Close()

	return buf.String(), nil
}

// A keyring is simply one (or more) keys in binary format.
func (key *Key) Keyring() []byte {
	buf := new(bytes.Buffer)
	key.Serialize(buf)
	return buf.Bytes()
}

// A secring is simply one (or more) keys in binary format.
func (key *Key) Secring(config *Config) []byte {
	buf := new(bytes.Buffer)
	c := config.Config
	key.SerializePrivate(buf, &c)
	return buf.Bytes()
}
