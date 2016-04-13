package security

// TODO: this is taken from (and slighly reduced):
//       https://github.com/ipfs/go-libp2p/blob/master/p2p/crypto/key.go

// Key represents a crypto key that can be compared to another key
type Key interface {
	// Bytes returns a serialized, storeable representation of this key
	Bytes() ([]byte, error)

	// Hash returns the hash of this key
	Hash() ([]byte, error)
}

// PrivKey represents a private key that can be used to generate a public key,
// sign data, and decrypt data that was encrypted with a public key
type PrivKey interface {
	Key

	// Cryptographically sign the given bytes
	Sign([]byte) ([]byte, error)

	// Decrypt the data in `b`.
	Decrypt(b []byte) ([]byte, error)
}

type PubKey interface {
	Key

	// Verify that 'sig' is the signed hash of 'data'
	Verify(data []byte, sig []byte) (bool, error)

	// Encrypt data in a way that can be decrypted by a paired private key
	Encrypt(data []byte) ([]byte, error)
}
