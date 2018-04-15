// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Hashing algorithm inspired by
//   xxhash: https://code.google.com/p/xxhash/
// cityhash: https://code.google.com/p/cityhash/

package gomap

import "unsafe"

func Hash32(b []byte, seed uint32) uint32 {
	const (
		// Constants for multiplication: four random odd 32-bit numbers.
		m1 = 3168982561
		m2 = 3339683297
		m3 = 832293441
		m4 = 2336365089
	)

	p := unsafe.Pointer(&b[0])
	s := len(b)
	h := uint32(seed + uint32(s))
tail:
	switch {
	case s == 0:
	case s < 4:
		w := uint32(*(*byte)(p))
		w += uint32(*(*byte)(add(p, s>>1))) << 8
		w += uint32(*(*byte)(add(p, s-1))) << 16
		h ^= w * m1
	case s == 4:
		h ^= readUnaligned32(p) * m1
	case s <= 8:
		h ^= readUnaligned32(p) * m1
		h = rotl_15(h) * m2
		h = rotl_11(h)
		h ^= readUnaligned32(add(p, s-4)) * m1
	case s <= 16:
		h ^= readUnaligned32(p) * m1
		h = rotl_15(h) * m2
		h = rotl_11(h)
		h ^= readUnaligned32(add(p, 4)) * m1
		h = rotl_15(h) * m2
		h = rotl_11(h)
		h ^= readUnaligned32(add(p, s-8)) * m1
		h = rotl_15(h) * m2
		h = rotl_11(h)
		h ^= readUnaligned32(add(p, s-4)) * m1
	default:
		v1 := h
		v2 := h + m1
		v3 := h + m2
		v4 := h + m3
		for s >= 16 {
			v1 ^= readUnaligned32(p) * m1
			v1 = rotl_15(v1) * m2
			p = add(p, 4)
			v2 ^= readUnaligned32(p) * m1
			v2 = rotl_15(v2) * m2
			p = add(p, 4)
			v3 ^= readUnaligned32(p) * m1
			v3 = rotl_15(v3) * m2
			p = add(p, 4)
			v4 ^= readUnaligned32(p) * m1
			v4 = rotl_15(v4) * m2
			p = add(p, 4)
			s -= 16
		}
		h = rotl_11(v1)*m1 + rotl_11(v2)*m2 + rotl_11(v3)*m3 + rotl_11(v4)*m4
		goto tail
	}
	h ^= h >> 17
	h *= m3
	h ^= h >> 13
	h *= m4
	h ^= h >> 16
	return uint32(h)
}

// Note: in order to get the compiler to issue rotl instructions, we
// need to constant fold the shift amount by hand.  TODO: convince
// the compiler to issue rotl instructions after inlining.
func rotl_15(x uint32) uint32 {
	return (x << 15) | (x >> (32 - 15))
}
func rotl_11(x uint32) uint32 {
	return (x << 11) | (x >> (32 - 11))
}
