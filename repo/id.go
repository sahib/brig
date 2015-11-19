package repo

import (
	"github.com/cathalgarvey/go-minilock"
)

// EncryptMinilockMsg encrypts a given plaintext for multiple receivers.
func EncryptMinilockMsg(jid, pass, plaintext string, mid ...string) (string, error) {
	ciphertext, err := minilock.EncryptFileContentsWithStrings("Minilock Filename.", []byte(plaintext), jid, pass, false, mid...)
	if err != nil {
		return "", nil
	}
	return string(ciphertext), nil
}

// DecryptMinilockMsg decrypts a given ciphertext.
func DecryptMinilockMsg(jid, pass, ciphertext string) (string, error) {
	userKey, err := minilock.GenerateKey(jid, pass)
	if err != nil {
		return "", nil
	}
	_, _, plaintext, _ := minilock.DecryptFileContents([]byte(ciphertext), userKey)
	return string(plaintext), nil
}

// GenerateMinilockID generates a base58-encoded pubkey + 1-byte blake2s checksum as a string
func GenerateMinilockID(jid, pass string) (string, error) {
	keys, err := minilock.GenerateKey(jid, pass)
	if err != nil {
		return "", err
	}
	return keys.EncodeID()
}
