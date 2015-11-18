package util

import (
	"crypto/rand"

	"golang.org/x/crypto/scrypt"
)

func DeriveAESKey(jid, password string, keySize int) ([]byte, []byte, error) {
	salt := make([]byte, 32)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, nil, err
	}

	// Parameters to be changed in future
	// https://godoc.org/golang.org/x/crypto/scrypt
	dkey, err := scrypt.Key([]byte(password), salt, 16384, 8, 1, keySize)
	if err != nil {
		return nil, nil, err
	}

	return dkey, salt, nil
}
