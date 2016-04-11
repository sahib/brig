// Package security implements utility function for often used
// security operations. At this very moment this includes:
//
// - Key derivation function using scrypt (DeriveAESKey)
package security

import (
	"golang.org/x/crypto/scrypt"
)

// Scrypt wraps scrypt.Key with the standard parameters.
// keyLen is in bytes, not bits.
func Scrypt(pwd, salt []byte, keyLen int) []byte {
	// Parameters to be changed in future
	// https://godoc.org/golang.org/x/crypto/scrypt
	key, err := scrypt.Key(pwd, salt, 16384, 8, 1, keyLen)
	if err != nil {
		panic("Bad scrypt parameters: " + err.Error())
	}

	return key
}
