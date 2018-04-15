// Copyright © 2014 Lawrence E. Bakst. All rights reserved.

// This package contains a new set of interfacs for hash functions.
// It also implements the Go streaming hash interface as HashStream.
// It is an experiment.

package nhash

import (
	"io"
)

// Interface HashFunction requires 4 methods that return the
// size of the hasg function in bytes and bits. Probably wiil
// flush bits. Also the maximum number of bytes of seed needed.
type HashFunction interface {
	// Size returns the number of bytes Sum will return.
	Size() int

	// BlockSize returns the hash's underlying block size.
	// The Write method must be able to accept any amount
	// of data, but it may operate more efficiently if all writes
	// are a multiple of the block size.
	BlockSize() int

	// maximum number of seeds in bytes (should this be in bits?)
	NumSeedBytes() int

	// retunrs the number of bits the hash function outputs
	//HashSizeInBits() int
}

// HashStream is a streaming interface for hash functions.
type HashStream interface {
	HashFunction

	// Write (via the embedded io.Writer interface) adds more data to the running hash.
	// It never returns an error.
	io.Writer

	// Sum appends the current hash to b and returns the resulting slice.
	// It does not change the underlying hash state.
	Sum(b []byte) []byte

	// Reset resets the Hash to its initial state.
	Reset()
}

// Hash32 is a common interface implemented by the streaming 32-bit hash functions.
type Hash32 interface {
	HashStream
	Sum32() uint32
}

// Hash64 is a common interface implemented by the streaming 32-bit hash functions.
type Hash64 interface {
	HashStream
	Write64(h uint64) error
	Sum64() uint64
}

// *** Everything below here will be removed or chnaged as vargs was way too expensive. ***

// HashF32 is the interface that all non-streaming 32 bit hash functions implement.
type HashF32 interface {
	HashFunction
	Hash32(b []byte, seeds ...uint32) uint32
}

// HashF64 is the interface that all non-streaming 64 bit hash functions implement.
type HashF64 interface {
	HashFunction
	Hash64(b []byte, seeds ...uint64) uint64
	Hash64S(b []byte, seed uint64) uint64
}

// HashF128 is the interface that all non-streaming 128 bit hash functions implement.
type HashF128 interface {
	HashFunction
	Hash128(b []byte, seeds ...uint64) (uint64, uint64)
}

// HashGeneric is generic interface that non-streaming, typicall crytpo hash functions implement.
type HashGeneric interface {
	HashFunction

	// Hash takes "in" bytes of input, the hash is returned into byte slice "out"
	// change seeds to bytes ???
	Hash(in []byte, out []byte, seeds ...uint64) []byte
}
