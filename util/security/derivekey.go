// Package security implements utility function for often used
// security operations. At this very moment this includes:
//
// - Key derivation function using scrypt (DeriveAESKey)
package security

import (
	"golang.org/x/crypto/scrypt"
)

// DeriveKey derives a key from password and salt being keyLen bytes long.
// It uses an established password derivation function.
func DeriveKey(pwd, salt []byte, keyLen int) []byte {
	// Parameters to be changed in future
	// https://godoc.org/golang.org/x/crypto/scrypt

	// TODO: 16384 is awfully slow. Consider increasing it later
	//       or consider using argon2 or something if it's faster.
	// key, err := scrypt.Key(pwd, salt, 16384, 8, 1, keyLen)
	key, err := scrypt.Key(pwd, salt, 4096, 8, 1, keyLen)
	if err != nil {
		panic("Bad scrypt parameters: " + err.Error())
	}

	return key
}
