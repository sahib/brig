package store

import (
	"bytes"
	"fmt"
	"strconv"

	goipfsutil "github.com/ipfs/go-ipfs-util"
	"github.com/jbenet/go-multihash"
)

var (
	// EmptyHash is the hash Qm0000[...], representing an empty hash.
	EmptyHash *Hash
)

func init() {
	data := make([]byte, multihash.DefaultLengths[goipfsutil.DefaultIpfsHash])
	hash, err := multihash.Encode(data, goipfsutil.DefaultIpfsHash)

	// No point in living elsewhise...
	if err != nil {
		panic(fmt.Sprintf("Unable to create empty hash: %v", err))
	}

	EmptyHash = &Hash{hash}
}

// Hash is like multihash.Multihash but also supports serializing to json.
// Otherwise all multihash features are supported.
type Hash struct {
	multihash.Multihash
}

// UnmarshalJSON loads a base58 string representation of a hash
// and converts it to raw bytes.
func (h *Hash) UnmarshalJSON(data []byte) error {
	if h == nil {
		h = &Hash{}
	}

	unquoted, err := strconv.Unquote(string(data))
	if err != nil {
		return err
	}

	mh, err := multihash.FromB58String(unquoted)
	if err != nil {
		return err
	}

	h.Multihash = mh
	return nil
}

// Valid returns true if the hash contains a defined value.
func (h *Hash) Valid() bool {
	return h != nil && h.Multihash != nil && h != EmptyHash
}

// Bytes returns the underlying bytes in the hash.
func (h *Hash) Bytes() []byte {
	if h == nil || h.Multihash == nil {
		return []byte{}
	}

	return []byte(h.Multihash)
}

// Clone returns the same hash as `h`,
// but with a different underlying array.
func (h *Hash) Clone() *Hash {
	if h == nil {
		return nil
	}

	if h.Multihash == nil {
		return &Hash{nil}
	}

	cpy := make([]byte, len([]byte(h.Multihash)))
	copy(cpy, h.Multihash)
	return &Hash{cpy}
}

// Equal returns true if both hashes are equal.
// Nil hashes are considered equal.
func (h *Hash) Equal(other *Hash) bool {
	if other == h {
		return true
	}

	if h == nil && other == nil {
		return true
	}

	if h == nil || other == nil {
		return false
	}

	if other.Multihash == nil && h.Multihash == nil {
		return true
	}

	if other.Multihash == nil || h.Multihash == nil {
		return false
	}

	return bytes.Equal(h.Multihash, other.Multihash)
}

// Add hashes `data` and xors the resulting hash to `h`.
// The hash algorithm and length depends on what kind
// of hash `h` currently holds.
func (h *Hash) MixIn(data []byte) error {
	dec, err := multihash.Decode(h.Multihash)
	if err != nil {
		return err
	}

	dataMH, err := multihash.Sum(h.Multihash, dec.Code, dec.Length)
	if err != nil {
		return err
	}

	for i := 2; i < len(dataMH); i++ {
		h.Multihash[i] ^= dataMH[i]
	}

	return nil
}
