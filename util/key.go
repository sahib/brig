package util

import (
	"golang.org/x/crypto/argon2"
)

// DeriveKey derives a key from password and salt being keyLen bytes long.
// It uses an established password derivation function.
func DeriveKey(pwd, salt []byte, keyLen int) []byte {
	// NOTE: These settings are below the recommendation of 64MB.
	//       This made adding files extremely slow since 90% of the time
	//       was spent hashing the content hash with argon2.

	// TODO: Do some in-depth research about argon2 and how we use it.
	//       Since we use it more for key stretching than actual salt based hashing
	//       I don't think this step is as critical as with normal passwords.
	//       Please tell me if I'm wrong though.
	return argon2.IDKey(pwd, salt, 1, 8*1024, 8, uint32(keyLen))
}
