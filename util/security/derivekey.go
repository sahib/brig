// Package security implements utility function for often used
// security operations. At this very moment this includes:
//
// - Key derivation function using scrypt (DeriveAESKey)
package security

import (
	"crypto/rand"

	"golang.org/x/crypto/scrypt"
)

// Scrypt() wraps scrypt.Key with the standard parameters.
func Scrypt(pwd, salt []byte, keyLen int) []byte {
	// Parameters to be changed in future
	// https://godoc.org/golang.org/x/crypto/scrypt
	key, err := scrypt.Key(pwd, salt, 16384, 8, 1, keyLen)
	if err != nil {
		panic("Bad scrypt parameters: " + err.Error())
	}

	return key
}

func DeriveAESKey(jid, password string, keySize int) ([]byte, []byte, error) {
	salt := make([]byte, 32)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, nil, err
	}

	return Scrypt([]byte(password), salt, keySize), salt, nil
}
