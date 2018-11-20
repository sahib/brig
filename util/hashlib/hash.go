package hashlib

import (
	"bytes"
	"fmt"
	"hash"
	"strconv"
	"testing"

	goipfsutil "github.com/ipfs/go-ipfs-util"
	"github.com/multiformats/go-multihash"
	"golang.org/x/crypto/sha3"
)

const (
	internalHashAlgo = multihash.SHA3_256
)

var (
	// EmptyBackendHash is a hash containing only zeros, using IPFS's default hash.
	EmptyBackendHash Hash

	// EmptyInternalHash is a hash containing only zeros, using brig's default hash.
	EmptyInternalHash Hash
)

func init() {
	data := make([]byte, multihash.DefaultLengths[goipfsutil.DefaultIpfsHash])
	hash, err := multihash.Encode(data, goipfsutil.DefaultIpfsHash)
	if err != nil {
		panic(fmt.Sprintf("Unable to create empty hash: %v", err))
	}

	EmptyBackendHash = Hash(hash)

	data = make([]byte, multihash.DefaultLengths[internalHashAlgo])
	hash, err = multihash.Encode(data, internalHashAlgo)
	if err != nil {
		panic(fmt.Sprintf("Unable to create empty content hash: %v", err))
	}

	EmptyInternalHash = Hash(hash)
}

// Hash is like multihash.Multihash but also supports serializing to json.
// It's methods are nil-value safe.
type Hash []byte

func (h Hash) String() string {
	return h.B58String()
}

// B58String formats the hash as base58 string.
func (h Hash) B58String() string {
	if h == nil {
		return "<empty hash>"
	}

	return multihash.Multihash(h).B58String()
}

// ShortB58 produces a shorter version (12 bytes long) of B58String()
func (h Hash) ShortB58() string {
	full := h.B58String()
	if len(full) > 12 {
		return full[:12]
	}

	return full
}

// FromB58String creates a new Hash from a base58 string.
// (This is shorthand for importing/using &Hash{multihash.FromB58String("xxx")}
func FromB58String(b58 string) (Hash, error) {
	mh, err := multihash.FromB58String(b58)
	if err != nil {
		return nil, err
	}

	return Hash(mh), nil
}

// UnmarshalJSON loads a base58 string representation of a hash
// and converts it to raw bytes.
func (h Hash) UnmarshalJSON(data []byte) error {
	if h == nil {
		h = EmptyBackendHash
	}

	unquoted, err := strconv.Unquote(string(data))
	if err != nil {
		return err
	}

	mh, err := multihash.FromB58String(unquoted)
	if err != nil {
		return err
	}

	copy(h, mh)
	return nil
}

// Valid returns true if the hash contains a defined value.
func (h Hash) Valid() bool {
	return h != nil && !bytes.Equal(h, EmptyBackendHash)
}

// Bytes returns the underlying bytes in the hash.
func (h Hash) Bytes() []byte {
	if h == nil {
		return EmptyBackendHash
	}

	return []byte(h)
}

// Clone returns the same hash as `h`,
// but with a different underlying array.
func (h Hash) Clone() Hash {
	if h == nil {
		return nil
	}

	cpy := make(Hash, len([]byte(h)))
	copy(cpy, h)
	return Hash(cpy)
}

// Equal returns true if both hashes are equal.
// Nil hashes are considered equal.
func (h Hash) Equal(other Hash) bool {
	if h == nil || other == nil {
		return h == nil && other == nil
	}

	return bytes.Equal(h, other)
}

// Mix produces a hash of both passed hashes.
func (h Hash) Mix(o Hash) Hash {
	buf := make([]byte, len(h)+len(o))
	copy(buf, h)
	copy(buf[len(h):], o)
	return Sum(buf)
}

// Sum hashes `data` with the internal hashing algorithm.
func Sum(data []byte) Hash {
	return sum(data, internalHashAlgo, multihash.DefaultLengths[internalHashAlgo])
}

// SumWithBackendHash creates a hash with the same algorithm the backend uses.
func SumWithBackendHash(data []byte) Hash {
	return sum(
		data,
		goipfsutil.DefaultIpfsHash,
		multihash.DefaultLengths[goipfsutil.DefaultIpfsHash],
	)
}

func sum(data []byte, code uint64, length int) Hash {
	mh, err := multihash.Sum(data, code, length)
	if err != nil {
		panic(fmt.Sprintf("failed to calculate basic hash value; something is wrong: %s", err))
	}

	return Hash(mh)
}

// Cast checks if `data` is a suitable hash and converts it.
func Cast(data []byte) (Hash, error) {
	mh, err := multihash.Cast(data)
	if err != nil {
		return nil, err
	}

	return Hash(mh), nil
}

// TestDummy returns a blake2b hash based on `seed`.
// The same `seed` will always generate the same hash.
func TestDummy(t *testing.T, seed byte) Hash {
	data := make([]byte, multihash.DefaultLengths[internalHashAlgo])
	for idx := range data {
		data[idx] = seed
	}

	hash, err := multihash.Encode(data, internalHashAlgo)
	if err != nil {
		t.Fatalf("Failed to create dummy hash: %v", err)
		return nil
	}

	return Hash(hash)
}

// HashWriter is a io.Writer that supports being written to.
type HashWriter struct {
	hash hash.Hash
}

// NewHashWriter returns a new HashWriter.
// Currently it is always sha3-256.
func NewHashWriter() *HashWriter {
	return &HashWriter{hash: sha3.New256()}
}

// Finalize returns the final hash of the written data.
func (hw *HashWriter) Finalize() Hash {
	sum := hw.hash.Sum(nil)
	hash, err := multihash.Encode(sum, internalHashAlgo)
	if err != nil {
		// If this does not work, there's something serious wrong.
		panic(fmt.Sprintf("failed to encode final hash: %v", err))
	}

	return hash
}

func (hw *HashWriter) Write(buf []byte) (int, error) {
	return hw.hash.Write(buf)
}
