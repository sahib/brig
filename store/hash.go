package store

import (
	"bytes"
	"strconv"

	"github.com/jbenet/go-multihash"
)

// Hash is like multihash.Multihash but also supports serializing to json.
// Otherwise all multihash features are supported.
type Hash struct {
	multihash.Multihash
}

// MarshalJSON converts a hash into a base58 string representation.
func (h *Hash) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(h.B58String())), nil
}

// UnmarshalJSON loads a base58 string representation of a hash
// and converts it to raw bytes.
func (h *Hash) UnmarshalJSON(data []byte) error {
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
	return h != nil && h.Multihash != nil
}

// Bytes returns the underlying bytes in the hash.
func (h *Hash) Bytes() []byte {
	return []byte(h.Multihash)
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
