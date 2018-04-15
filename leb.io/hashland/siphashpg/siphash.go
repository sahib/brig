// Copyright © 2014 Lawrence E. Bakst. All rights reserved.
// transliterated from the reference implementation cited below
// This version only works for LE

/*
SipHash reference C implementation

Copyright (c) 2012-2014 Jean-Philippe Aumasson <jeanphilippe.aumasson@gmail.com>
Copyright (c) 2012-2014 Daniel J. Bernstein <djb@cr.yp.to>

To the extent possible under law, the author(s) have dedicated all copyright
and related and neighboring rights to this software to the public domain
worldwide. This software is distributed without any warranty.

You should have received a copy of the CC0 Public Domain Dedication along with
this software. If not, see <http://creativecommons.org/publicdomain/zero/1.0/>.
*/

package siphashpg

import "fmt"
import "unsafe"
/* default: SipHash-2-4 */
const (
    Crounds  = 2
    Drounds = 4
)

func rotl(x, b uint64) uint64 {
	return ((x) << (b)) | ( (x) >> (64 - (b)))
}

func U8tou64le(p []byte) uint64 {
/*
	for k, v := range p {
		fmt.Printf("p[%d]=0x%x\n", k, v)
	}
*/
	return uint64(p[0]) | uint64(p[1]) << 8 | uint64(p[2]) << 16 | uint64(p[3]) << 24 | uint64(p[4]) << 32 | uint64(p[5]) << 40 | uint64(p[6]) << 48 | uint64(p[7]) << 56
}

func U64tou8le(v uint64) (r []byte) {
	r = make([]byte, 8, 8)
	r = r[:]
	for k, _ := range r {
		r[k] = byte(v&0xFF)
		v >>= 8
	}
	return
}

func siprounda(v0, v1, v2, v3 uint64) (uint64, uint64, uint64, uint64) {
	v0 += v1; v1=rotl(v1,13); v1 ^= v0; v0=rotl(v0,32);
	v2 += v3; v3=rotl(v3,16); v3 ^= v2;
	return v0, v1, v2, v3
}

func siproundb(v0, v1, v2, v3 uint64) (uint64, uint64, uint64, uint64) {
	v0 += v3; v3=rotl(v3,21); v3 ^= v0;
	v2 += v1; v1=rotl(v1,17); v1 ^= v2; v2=rotl(v2,32);
	return v0, v1, v2, v3
}

func TRACE(inlen int, v0, v1, v2, v3 uint64) {
	return
	fmt.Printf( "(%3d) v0 %08x %08x\n", inlen, v0 >> 32, v0&0xFFFFFFFF)
	fmt.Printf( "(%3d) v1 %08x %08x\n", inlen, v1 >> 32, v1&0xFFFFFFFF)
	fmt.Printf( "(%3d) v2 %08x %08x\n", inlen, v2 >> 32, v2&0xFFFFFFFF)
	fmt.Printf( "(%3d) v3 %08x %08x\n", inlen, v3 >> 32, v3&0xFFFFFFFF)
}

// This makes a new slice of uint64 that points to the same slice passed in as []byte.
// We should check alignment for architectures that don't handle unaligned reads.
// Fallback to a copy or maybe use encoding/binary?
// Not sure what the right thing to do is for little vs big endian?
// What are the right test vevtors for big-endian machines.
func sliceUI64(in []byte) []uint64 {
    return (*(*[]uint64)(unsafe.Pointer(&in)))[:len(in)/8]
}

// take input slice in and seeds k as well a compression and final rounds cr, dr
// return a 64 bit hash in ra and if dbl is true a 128 bit hash in ra and rb
func Siphash(in []byte, k []byte, cr, dr int, dbl bool) (ra, rb uint64) {
	var fast = true
	var v0, v1, v2, v3 uint64
	var sipround = func() {
		v0 += v1; v1=rotl(v1,13); v1 ^= v0; v0=rotl(v0,32);
		v2 += v3; v3=rotl(v3,16); v3 ^= v2;
		v0 += v3; v3=rotl(v3,21); v3 ^= v0;
		v2 += v1; v1=rotl(v1,17); v1 ^= v2; v2=rotl(v2,32);
	}

	if len(k) != 16 || cr <= 0 || dr <= 0 {
		panic("siphash")
	}
	// initialize state
	/* "somepseudorandomlygeneratedbytes" */
	v0 = uint64(0x736f6d6570736575)
	v1 = uint64(0x646f72616e646f6d)
	v2 = uint64(0x6c7967656e657261)
	v3 = uint64(0x7465646279746573)

	k0 := U8tou64le(k)
	k1 := U8tou64le(k[8:])
	b := uint64(len(in)) << 56
	//fmt.Printf("k=%v, k0=0x%08x, k1=0x%08x, b=0x%08x\n", k, k0, k1, b)

	v3 ^= k1
	v2 ^= k0
	v1 ^= k1
	v0 ^= k0

	if dbl {
		v1 ^= 0xee
	}

	k64 := sliceUI64(in)
	cnt := 0
	l := len(in)
	// peel off as many 64 bit words as we have
	if fast {
		for ; l >= 8; in = in[8:] {
	    	m := k64[cnt]
	    	cnt++
	    	l -= 8
	    	v3 ^= m

			//TRACE(len(in), v0, v1, v2, v3)
	    	for i := 0; i < cr; i++ {
	    		sipround()
	    		//v0, v1, v2, v3 = siprounda(v0, v1, v2, v3)
	    		//v0, v1, v2, v3 = siproundb(v0, v1, v2, v3)
			}
			v0 ^= m
		}
	} else {
		for ; len(in) >= 8; in = in[8:] {
	    	m := U8tou64le(in)
	    	v3 ^= m

			//TRACE(len(in), v0, v1, v2, v3)
	    	for i := 0; i < cr; i++ {
	    		sipround()
	    		//v0, v1, v2, v3 = siprounda(v0, v1, v2, v3)
	    		//v0, v1, v2, v3 = siproundb(v0, v1, v2, v3)
			}
			v0 ^= m
		}
	}

	//fmt.Printf("in=%v, len(in)=%d\n", in, len(in))
	// deal with the tail
	switch len(in) {
	case 7:
		b |= uint64(in[6]) << 48
		fallthrough
	case 6:
		b |= uint64(in[5]) << 40
		fallthrough
	case 5:
		b |= uint64(in[4]) << 32
		fallthrough
	case 4:
		b |= uint64(in[3]) << 24
		fallthrough
	case 3:
		b |= uint64(in[2]) << 16
		fallthrough
	case 2:
		b |= uint64(in[1]) <<  8
		fallthrough
	case 1:
		b |= uint64(in[0])
		break;
	case 0:
		break;
	default:
		//fmt.Printf("len(in)=%d\n", len(in))
		panic("siphash bad length")
	}
	v3 ^= b;

	//TRACE(len(in), v0, v1, v2, v3)
	for i := 0; i < cr; i++ {
		sipround()
		//v0, v1, v2, v3 = siprounda(v0, v1, v2, v3)
		//v0, v1, v2, v3 = siproundb(v0, v1, v2, v3)
	}
	v0 ^= b

	if dbl {
		v2 ^= 0xee
	} else {
		v2 ^= 0xff
	}

	//TRACE(len(in), v0, v1, v2, v3)
	for i := 0; i < dr; i++ {
		sipround()
		//v0, v1, v2, v3 = siprounda(v0, v1, v2, v3)
		//v0, v1, v2, v3 = siproundb(v0, v1, v2, v3)
	}
	b = v0 ^ v1 ^ v2  ^ v3
	ra = b

	// if 128 bit result desired run some more rounds and get another 64 bits
	if dbl {
  		v1 ^= 0xdd
		//TRACE(len(in), v0, v1, v2, v3)
		for i := 0; i < dr; i++ {
			sipround()
    		//v0, v1, v2, v3 = siprounda(v0, v1, v2, v3)
    		//v0, v1, v2, v3 = siproundb(v0, v1, v2, v3)
		}
		b = v0 ^ v1 ^ v2  ^ v3
		rb = b
	}
	return
}