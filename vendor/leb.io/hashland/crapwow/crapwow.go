// Copyright © 2014 Lawrence E. Bakst. All rights reserved.

package crapwow

import (
	"leb.io/hashland/nhash"
	"unsafe"
)

// this makes a new slice of uint32 that points to the same slice passed in as []byte
// we should check alignment for architectures that don't handle unaligned reads
// and fallback to a copy maybe using encoding/binary.
// One question is what are the right test vevtors for big-endian machines.
func sliceUI32(in []byte) []uint32 {
	return (*(*[]uint32)(unsafe.Pointer(&in)))[:len(in)/4]
}

func CrapWow(key []byte, seed uint32) uint32 {
	const m uint32 = 0x57559429
	const n uint32 = 0x5052acdb
	var l = len(key)
	var h = uint32(l)
	var k = h + seed + n
	var p uint64
	/*
		var cwfold = func(a, b, lo, hi uint32) {
			p = uint32(a) * uint64(b)
			lo ^= uint32(p)
			hi ^= uint32(p >> 32)
		}
	*/
	var cwmixa = func(in uint32) {
		p = uint64(in) * uint64(m)
		k ^= uint32(p)
		h ^= uint32(p >> 32)
	}

	var cwmixb = func(in uint32) {
		p = uint64(in) * uint64(n)
		h ^= uint32(p)
		k ^= uint32(p >> 32)
	}

	key4 := sliceUI32(key)
	for l >= 8 {
		cwmixb(key4[0])
		cwmixa(key4[1])
		key4 = key4[2:]
		key = key[8:]
		l -= 8
	}
	if l >= 4 {
		cwmixb(key4[0])
		key4 = key4[1:]
		key = key[4:]
		l -= 4
	}
	switch l {
	case 3:
		tmp := uint32(key[2])<<16 | uint32(key[1])<<8 | uint32(key[0])
		cwmixa(tmp & ((1 << (uint32(l) * 8)) - 1))
	case 2:
		tmp := uint32(key[1])<<8 | uint32(key[0])
		cwmixa(tmp & ((1 << (uint32(l) * 8)) - 1))
	case 1:
		tmp := uint32(key[0])
		cwmixa(tmp & ((1 << (uint32(l) * 8)) - 1))
	}
	cwmixb(h ^ (k + n))
	return k ^ h
}

type State struct {
	hash uint32
	seed uint32
}

// New returns a new hash.HashF32 interface that computes a 32 bit CrapWow hash.
func New(seed uint32) nhash.HashF32 {
	s := new(State)
	s.seed = seed
	return s
}

// The size of an jenkins3 32 bit hash in bytes.
const Size = 4

// Return the size of the resulting hash.
func (s *State) Size() int { return Size }

// Return the blocksize of the hash which in this case is 8 bytes.
func (s *State) BlockSize() int { return 8 }

// Return the maximum number of seed bypes required.
func (s *State) NumSeedBytes() int {
	return 4
}

// retunrs the number of bits the hash function outputs
func (s *State) HashSizeInBits() int {
	return 32
}

func (s *State) Hash32(b []byte, seeds ...uint32) uint32 {
	if len(seeds) > 0 {
		s.seed = seeds[0]
	}
	s.hash = CrapWow(b, s.seed)
	return s.hash
}
