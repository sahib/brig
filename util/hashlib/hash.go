package hashlib

import (
	"bytes"
	"fmt"
	"strconv"
	"testing"

	goipfsutil "github.com/ipfs/go-ipfs-util"
	"github.com/multiformats/go-multihash"
)

var (
	// EmptyHash is the hash of empty data, hashed with the default hash of ipfs.
	EmptyHash Hash
)

func init() {
	data := make([]byte, multihash.DefaultLengths[goipfsutil.DefaultIpfsHash])
	hash, err := multihash.Encode(data, goipfsutil.DefaultIpfsHash)

	// No point in living elsewhise...
	if err != nil {
		panic(fmt.Sprintf("Unable to create empty hash: %v", err))
	}

	EmptyHash = Hash(hash)
}

// Hash is like multihash.Multihash but also supports serializing to json.
// It's methods are nil-value safe.
type Hash []byte

func (h Hash) String() string {
	return h.B58String()
}

func (h Hash) B58String() string {
	if h == nil {
		return "<empty hash>"
	}

	return multihash.Multihash(h).B58String()
}

// Create a new Hash from a b58 string.
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
		h = EmptyHash
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
	return h != nil && !bytes.Equal(h, EmptyHash)
}

// Bytes returns the underlying bytes in the hash.
func (h Hash) Bytes() []byte {
	if h == nil {
		return EmptyHash
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

func (h Hash) Xor(o Hash) error {
	decH, err := multihash.Decode(h)
	if err != nil {
		fmt.Println("Decode self failed")
		return err
	}

	decO, err := multihash.Decode(o)
	if err != nil {
		fmt.Println("Decode other failed", o)
		return err
	}

	if decO.Length != decH.Length {
		return fmt.Errorf("xor: hashs have different lengths: %d != %d", decH.Length, decO.Length)
	}

	for i := 0; i < decH.Length; i++ {
		decH.Digest[i] ^= decO.Digest[i]
	}

	mh, err := multihash.Encode(decH.Digest, decH.Code)
	if err != nil {
		fmt.Println("encode failed")
		return err
	}

	copy(h, mh)
	return nil
}

func Sum(data []byte) Hash {
	mh, err := multihash.Sum(
		data, multihash.BLAKE2B_MAX, multihash.DefaultLengths[multihash.BLAKE2B_MAX],
	)

	if err != nil {
		panic(fmt.Sprintf("Failed to calculate basic hash value. Something is wrong: %s", err))
	}

	return Hash(mh)
}

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
	data := make([]byte, multihash.DefaultLengths[multihash.BLAKE2B_MAX])
	for idx := range data {
		data[idx] = seed
	}

	hash, err := multihash.Encode(data, multihash.BLAKE2B_MAX)
	if err != nil {
		t.Fatalf("Failed to create dummy hash: %v", err)
		return nil
	}

	return Hash(hash)
}

func (h Hash) ShortB58() string {
	full := h.B58String()
	if len(full) > 10 {
		return full[:10]
	}

	return full
}
