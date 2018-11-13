package util

import (
	"golang.org/x/crypto/scrypt"
)

// DeriveKey derives a key from password and salt being keyLen bytes long.
// It uses an established password derivation function.
func DeriveKey(pwd, salt []byte, keyLen int) []byte {
	// Parameters to be changed in future
	// https://godoc.org/golang.org/x/crypto/scrypt
	key, err := scrypt.Key(pwd, salt, 32768, 8, 1, keyLen)
	if err != nil {
		panic("Bad scrypt parameters: " + err.Error())
	}

	return key
}
