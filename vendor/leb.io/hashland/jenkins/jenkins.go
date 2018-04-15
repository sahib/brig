// Copyright © 2014 Lawrence E. Bakst. All rights reserved.

// This package contains various transliteration of Jenkins hash funcions.
// This includes lookup3.c and some other Jenkins functions that appear to be based on lookup2.c.
// See http://burtleburtle.net/bob/c/lookup3.c and http://burtleburtle.net/bob/hash/evahash.html

package jenkins

import (
	"fmt"
	"hash"
	"leb.io/hashland/nhash"
	"unsafe"
)

// Make sure interfaces are correctly implemented. Stolen from another implementation.
// I did something similar in another package to verify the interface but didn't know you could elide the variable in a var.
// What a cute wart it is.
var (
	//_ hash.Hash   = new(Digest)
	_ hash.Hash32   = new(State332c)
	_ nhash.HashF32 = new(State332c)
)

/*
uint32_t jenkins_one_at_a_time_hash(char *key, size_t len)
{
    uint32_t hash, i;
    for(hash = i = 0; i < len; ++i)
    {
        hash += key[i];
        hash += (hash << 10);
        hash ^= (hash >> 6);
    }
    hash += (hash << 3);
    hash ^= (hash >> 11);
    hash += (hash << 15);
    return hash;
}
*/

// This function mixes the state
func mix32(a, b, c uint32) (uint32, uint32, uint32) {
	a = a - b
	a = a - c
	a = a ^ (c >> 13)
	b = b - c
	b = b - a
	b = b ^ (a << 8)
	c = c - a
	c = c - b
	c = c ^ (b >> 13)
	a = a - b
	a = a - c
	a = a ^ (c >> 12)
	b = b - c
	b = b - a
	b = b ^ (a << 16)
	c = c - a
	c = c - b
	c = c ^ (b >> 5)
	a = a - b
	a = a - c
	a = a ^ (c >> 3)
	b = b - c
	b = b - a
	b = b ^ (a << 10)
	c = c - a
	c = c - b
	c = c ^ (b >> 15)
	return a, b, c
}

func mix32a(a, b, c uint32) (uint32, uint32, uint32) {
	a = a - b - c ^ (c >> 13)
	b = b - c - a ^ (a << 8)
	c = c - a - b ^ (b >> 13)
	a = a - b - c ^ (c >> 12)
	b = b - c - a ^ (a << 16)
	c = c - a - b ^ (b >> 5)
	a = a - b - c ^ (c >> 3)
	b = b - c - a ^ (a << 10)
	c = c - a - b ^ (b >> 15)
	return a, b, c
}

func Check() {
	var a, b, c uint32
	var init = func() {
		a, b, c = 0x12345678, 0x87654321, 0xdeadbeef
	}

	init()
	a = a - b
	a = a - c
	fmt.Printf("a=0x%08x\n", a)
	a = a ^ (c >> 13)
	fmt.Printf("a=0x%08x\n", a)
	fmt.Printf("a=0x%08x, b=0x%08x, c=0x%08x\n", a, b, c)
	init()
	a = ((a - b) - c)
	fmt.Printf("a=0x%08x\n", a)
	a = a ^ (c >> 13)
	fmt.Printf("a=0x%08x\n", a)
	fmt.Printf("a=0x%08x, b=0x%08x, c=0x%08x\n", a, b, c)
}

// This makes a new slice of uint64 that points to the same slice passed in as []byte.
// We should check alignment for architectures that don't handle unaligned reads.
// Fallback to a copy or maybe use encoding/binary?
// Not sure what the right thing to do is for little vs big endian?
// What are the right test vevtors for big-endian machines.
func sliceUI32(in []byte) []uint32 {
	return (*(*[]uint32)(unsafe.Pointer(&in)))[:len(in)/4]
}

// Jenkin's second generation 32 bit hash.
// Benchmarked with 4 byte key, inlining, and no store of hash at:
// benchmark32: 55 Mhashes/sec
// benchmark32: 219 MB/sec
func Hash232(k []byte, seed uint32) uint32 {
	var fast = true // fast is really much faster
	l := uint32(len(k))
	a := uint32(0x9e3779b9) // the golden ratio; an arbitrary value
	b := a
	c := seed // variable initialization of internal state

	if fast {
		k32 := sliceUI32(k)
		cnt := 0
		for ; l >= 12; l -= 12 {
			a += k32[0+cnt]
			b += k32[1+cnt]
			c += k32[2+cnt]
			a, b, c = mix32(a, b, c)
			k = k[12:]
			cnt += 3
		}
	} else {
		for ; l >= 12; l -= 12 {
			a += uint32(k[0]) + uint32(k[1])<<8 + uint32(k[2])<<16 + uint32(k[3])<<24
			b += uint32(k[4]) + uint32(k[5])<<8 + uint32(k[6])<<16 + uint32(k[7])<<24
			c += uint32(k[8]) + uint32(k[9])<<8 + uint32(k[10])<<16 + uint32(k[11])<<24
			a, b, c = mix32(a, b, c)
			k = k[12:]
		}
	}

	c += l
	switch l {
	case 11:
		c += uint32(k[10]) << 24
		fallthrough
	case 10:
		c += uint32(k[9]) << 16
		fallthrough
	case 9:
		c += uint32(k[8]) << 8
		fallthrough
	case 8:
		b += uint32(k[7]) << 24 // the first byte of c is reserved for the length
		fallthrough
	case 7:
		b += uint32(k[6]) << 16
		fallthrough
	case 6:
		b += uint32(k[5]) << 8
		fallthrough
	case 5:
		b += uint32(k[4])
		fallthrough
	case 4:
		a += uint32(k[3]) << 24
		fallthrough
	case 3:
		a += uint32(k[2]) << 16
		fallthrough
	case 2:
		a += uint32(k[1]) << 8
		fallthrough
	case 1:
		c += uint32(k[0])
		fallthrough
	case 0:
		break
	default:
		panic("HashWords32")
	}
	a, b, c = mix32(a, b, c)
	return c
}

// original mix function
func mix64(a, b, c uint64) (uint64, uint64, uint64) {
	a = a - b
	a = a - c
	a = a ^ (c >> 43)
	b = b - c
	b = b - a
	b = b ^ (a << 9)
	c = c - a
	c = c - b
	c = c ^ (b >> 8)
	a = a - b
	a = a - c
	a = a ^ (c >> 38)
	b = b - c
	b = b - a
	b = b ^ (a << 23)
	c = c - a
	c = c - b
	c = c ^ (b >> 5)
	a = a - b
	a = a - c
	a = a ^ (c >> 35)
	b = b - c
	b = b - a
	b = b ^ (a << 49)
	c = c - a
	c = c - b
	c = c ^ (b >> 11)
	a = a - b
	a = a - c
	a = a ^ (c >> 12)
	b = b - c
	b = b - a
	b = b ^ (a << 18)
	c = c - a
	c = c - b
	c = c ^ (b >> 22)
	return a, b, c
}

// restated mix function better for gofmt
func mix64alt(a, b, c uint64) (uint64, uint64, uint64) {
	a -= b - c ^ (c >> 43)
	b -= c - a ^ (a << 9)
	c -= a - b ^ (b >> 8)
	a -= b - c ^ (c >> 38)
	b -= c - a ^ (a << 23)
	c -= a - b ^ (b >> 5)
	a -= b - c ^ (c >> 35)
	b -= c - a ^ (a << 49)
	c -= a - b ^ (b >> 11)
	a -= b - c ^ (c >> 12)
	b -= c - a ^ (a << 18)
	c -= a - b ^ (b >> 22)
	return a, b, c
}

// the following functions can be inlined
func mix64a(a, b, c uint64) (uint64, uint64, uint64) {
	a -= b - c ^ (c >> 43)
	b -= c - a ^ (a << 9)
	return a, b, c
}

func mix64b(a, b, c uint64) (uint64, uint64, uint64) {
	c -= a - b ^ (b >> 8)
	a -= b - c ^ (c >> 38)
	return a, b, c
}

func mix64c(a, b, c uint64) (uint64, uint64, uint64) {
	b -= c - a ^ (a << 23)
	c -= a - b ^ (b >> 5)
	return a, b, c
}

func mix64d(a, b, c uint64) (uint64, uint64, uint64) {
	a -= b - c ^ (c >> 35)
	b -= c - a ^ (a << 49)
	return a, b, c
}

func mix64e(a, b, c uint64) (uint64, uint64, uint64) {
	c -= a - b ^ (b >> 11)
	a -= b - c ^ (c >> 12)
	return a, b, c
}

func mix64f(a, b, c uint64) (uint64, uint64, uint64) {
	b -= c - a ^ (a << 18)
	c -= a - b ^ (b >> 22)
	return a, b, c
}

// This makes a new slice of uint64 that points to the same slice passed in as []byte.
// We should check alignment for architectures that don't handle unaligned reads.
// Fallback to a copy or maybe use encoding/binary?
// Not sure what the right thing to do is for little vs big endian?
// What are the right test vevtors for big-endian machines.
func sliceUI64(in []byte) []uint64 {
	return (*(*[]uint64)(unsafe.Pointer(&in)))[:len(in)/8]
}

// Jenkin's second generation 64 bit hash.
// Benchmarked with 24 byte key, inlining, store of hash in memory (cache miss every 4 hashes) and fast=true at:
// benchmark64: 26 Mhashes/sec
// benchmark64: 623 MB/sec
func Hash264(k []byte, seed uint64) uint64 {
	var fast = true // fast is really much faster
	//fmt.Printf("k=%v\n", k)
	//fmt.Printf("length=%d, len(k)=%d\n", length, len(k))

	//The 64-bit golden ratio is 0x9e3779b97f4a7c13LL
	length := uint64(len(k))
	a := uint64(0x9e3779b97f4a7c13)
	b := a
	c := seed
	if fast {
		k64 := sliceUI64(k)
		cnt := 0
		for i := length; i >= 24; i -= 24 {
			a += k64[0+cnt]
			b += k64[1+cnt]
			c += k64[2+cnt]
			// inlining is slightly faster
			a, b, c = mix64a(a, b, c)
			a, b, c = mix64b(a, b, c)
			a, b, c = mix64c(a, b, c)
			a, b, c = mix64d(a, b, c)
			a, b, c = mix64e(a, b, c)
			a, b, c = mix64f(a, b, c)
			k = k[24:]
			cnt += 3
			length -= 24
		}
	} else {
		for i := length; i >= 24; i -= 24 {
			a += uint64(k[0]) | uint64(k[1])<<8 | uint64(k[2])<<16 | uint64(k[3])<<24 | uint64(k[4])<<32 | uint64(k[5])<<40 | uint64(k[6])<<48 | uint64(k[7])<<56
			b += uint64(k[8]) | uint64(k[9])<<8 | uint64(k[10])<<16 | uint64(k[11])<<24 | uint64(k[12])<<32 | uint64(k[13])<<40 | uint64(k[14])<<48 | uint64(k[15])<<56
			c += uint64(k[16]) | uint64(k[17])<<8 | uint64(k[18])<<16 | uint64(k[19])<<24 | uint64(k[20])<<32 | uint64(k[21])<<40 | uint64(k[22])<<48 | uint64(k[23])<<56
			a, b, c = mix64alt(a, b, c)
			k = k[24:]
			length -= 24
		}
	}
	c += length
	if len(k) > 23 {
		panic("Hash264")
	}
	switch length {
	case 23:
		c += uint64(k[22]) << 56
		fallthrough
	case 22:
		c += uint64(k[21]) << 48
		fallthrough
	case 21:
		c += uint64(k[20]) << 40
		fallthrough
	case 20:
		c += uint64(k[19]) << 32
		fallthrough
	case 19:
		c += uint64(k[18]) << 24
		fallthrough
	case 18:
		c += uint64(k[17]) << 16
		fallthrough
	case 17:
		c += uint64(k[16]) << 8
		fallthrough
	case 16:
		b += uint64(k[15]) << 56 // the first byte of c is reserved for the length
		fallthrough
	case 15:
		b += uint64(k[14]) << 48
		fallthrough
	case 14:
		b += uint64(k[13]) << 40
		fallthrough
	case 13:
		b += uint64(k[12]) << 32
		fallthrough
	case 12:
		b += uint64(k[11]) << 24
		fallthrough
	case 11:
		b += uint64(k[10]) << 16
		fallthrough
	case 10:
		b += uint64(k[9]) << 8
		fallthrough
	case 9:
		b += uint64(k[8])
		fallthrough
	case 8:
		a += uint64(k[7]) << 56
		fallthrough
	case 7:
		a += uint64(k[6]) << 48
		fallthrough
	case 6:
		a += uint64(k[5]) << 40
		fallthrough
	case 5:
		a += uint64(k[4]) << 32
		fallthrough
	case 4:
		a += uint64(k[3]) << 24
		fallthrough
	case 3:
		a += uint64(k[2]) << 16
		fallthrough
	case 2:
		a += uint64(k[1]) << 8
		fallthrough
	case 1:
		a += uint64(k[0])
	case 0:
		break
	default:
		panic("HashWords64")
	}
	a, b, c = mix64alt(a, b, c)
	return c
}

/*
func omix(a, b, c uint32) (uint32, uint32, uint32) {
	a -= c;  a ^= rot(c, 4);  c += b;
	b -= a;  b ^= rot(a, 6);  a += c;
	c -= b;  c ^= rot(b, 8);  b += a;
	a -= c;  a ^= rot(c,16);  c += b;
	b -= a;  b ^= rot(a,19);  a += c;
	c -= b;  c ^= rot(b, 4);  b += a;
	return a, b, c
}

func ofinal(a, b, c uint32) (uint32, uint32, uint32) {
	c ^= b; c -= rot(b,14);
	a ^= c; a -= rot(c,11);
	b ^= a; b -= rot(a,25);
	c ^= b; c -= rot(b,16);
	a ^= c; a -= rot(c,4);
	b ^= a; b -= rot(a,14);
	c ^= b; c -= rot(b,24);
	return a, b, c
}


func mix(a, b, c uint32) (uint32, uint32, uint32) {
	a -= c;  a ^= c << 4 | c >> (32 - 4);  c += b;
	b -= a;  b ^= a << 6 | a >> (32 - 6);  a += c;
	c -= b;  c ^= b << 8 | b >> (32 - 8);  b += a;
	a -= c;  a ^= c << 16 | c >> (32 - 16);  c += b;
	b -= a;  b ^= a << 19 | a >> (32 - 19);  a += c;
	c -= b;  c ^= b << 4 | b >> (32 - 4);  b += a;
	return a, b, c
}


func final(a, b, c uint32) (uint32, uint32, uint32) {
	c ^= b; c -= b << 14 | b >> (32 - 14);
	a ^= c; a -= c << 11 | c >> (32 - 11);
	b ^= a; b -= a << 25 | a >> (32 - 25);
	c ^= b; c -= b << 16 | b >> (32 - 16);
	a ^= c; a -= c << 4 | c >> (32 - 4);
	b ^= a; b -= a << 14 | a >> (32 - 14);
	c ^= b; c -= b << 24 | b >> (32 - 24);
	return a, b, c
}
*/

//var a, b, c uint32

func rot(x, k uint32) uint32 {
	return x<<k | x>>(32-k)
}

// current gc compilers can't inline long functions so we have to split mix into 2
func mix1(a, b, c uint32) (uint32, uint32, uint32) {
	a -= c
	a ^= rot(c, 4)
	c += b
	b -= a
	b ^= rot(a, 6)
	a += c
	c -= b
	c ^= rot(b, 8)
	b += a
	//a -= c;  a ^= c << 4 | c >> (32 - 4);  c += b;
	//b -= a;  b ^= a << 6 | a >> (32 - 6);  a += c;
	return a, b, c
}

func mix2(a, b, c uint32) (uint32, uint32, uint32) {
	a -= c
	a ^= rot(c, 16)
	c += b
	b -= a
	b ^= rot(a, 19)
	a += c
	c -= b
	c ^= rot(b, 4)
	b += a
	//	c -= b;  c ^= b << 8 | b >> (32 - 8);  b += a;
	//	a -= c;  a ^= c << 16 | c >> (32 - 16);  c += b;
	return a, b, c
}

/*
func mix3(a, b, c uint32) (uint32, uint32, uint32) {
	b -= a;  b ^= a << 19 | a >> (32 - 19);  a += c;
	c -= b;  c ^= b << 4 | b >> (32 - 4);  b += a;
	return a, b, c
}
*/

func final1(a, b, c uint32) (uint32, uint32, uint32) {
	c ^= b
	c -= rot(b, 14)
	a ^= c
	a -= rot(c, 11)
	b ^= a
	b -= rot(a, 25)
	c ^= b
	c -= rot(b, 16)
	//c ^= b; c -= b << 14 | b >> (32 - 14);
	//a ^= c; a -= c << 11 | c >> (32 - 11);
	//b ^= a; b -= a << 25 | a >> (32 - 25);
	//c ^= b; c -= b << 16 | b >> (32 - 16);
	return a, b, c
}

func final2(a, b, c uint32) (uint32, uint32, uint32) {
	a ^= c
	a -= rot(c, 4)
	b ^= a
	b -= rot(a, 14)
	c ^= b
	c -= rot(b, 24)
	//a ^= c; a -= c << 4 | c >> (32 - 4);
	//b ^= a; b -= a << 14 | a >> (32 - 14);
	//c ^= b; c -= b << 24 | b >> (32 - 24);
	return a, b, c
}

func HashWords332(k []uint32, seed uint32) uint32 {
	var a, b, c uint32

	length := uint32(len(k))
	a = 0xdeadbeef + length<<2 + seed
	b, c = a, a

	i := 0
	for ; length > 3; length -= 3 {
		a += k[i+0]
		b += k[i+1]
		c += k[i+2]
		a, b, c = mix1(a, b, c)
		a, b, c = mix2(a, b, c)
		i += 3
	}

	switch length {
	case 3:
		c += k[i+2]
		fallthrough
	case 2:
		b += k[i+1]
		fallthrough
	case 1:
		a += k[i+0]
		a, b, c = final1(a, b, c)
		a, b, c = final2(a, b, c)
	case 0:
		break
	}
	return c
}

func HashWordsLen(k []uint32, length int, seed uint32) uint32 {
	var a, b, c uint32

	//fmt.Printf("k=%v\n", k)
	//fmt.Printf("length=%d, len(k)=%d\n", length, len(k))
	if length > len(k)*4 {
		fmt.Printf("length=%d, len(k)=%d\n", length, len(k))
		panic("HashWords")
	}

	ul := uint32(len(k))
	a = 0xdeadbeef + ul<<2 + seed
	b, c = a, a

	i := 0
	//length := 0
	for ; length > 3; length -= 3 {
		a += k[i+0]
		b += k[i+1]
		c += k[i+2]
		a, b, c = mix1(a, b, c)
		a, b, c = mix2(a, b, c)
		//a, b, c = mix3(a, b, c)
		i += 3
	}

	//fmt.Printf("remaining length=%d, len(k)=%d, i=%d, k[i + 2]=%d, k[i + 1]=%d, k[i + 0]=%d\n", length, len(k), i, k[i + 2], k[i + 1], k[i + 0])
	switch length {
	case 3:
		c += k[i+2]
		fallthrough
	case 2:
		b += k[i+1]
		fallthrough
	case 1:
		a += k[i+0]
		a, b, c = final1(a, b, c)
		a, b, c = final2(a, b, c)
	case 0:
		break
	}
	//fmt.Printf("end\n")
	return c
}

// This is an example of how I could like to code hash functions like this.
// Using closures over the state and expecting thme to be inlined
func XHashWords(k []uint32, length int, seed uint32) uint32 {
	var a, b, c uint32
	var rot = func(x, k uint32) uint32 {
		return x<<k | x>>(32-k)
	}
	var mix = func() {
		a -= c
		a ^= rot(c, 4)
		c += b
		b -= a
		b ^= rot(a, 6)
		a += c
		c -= b
		c ^= rot(b, 8)
		b += a
		a -= c
		a ^= rot(c, 16)
		c += b
		b -= a
		b ^= rot(a, 19)
		a += c
		c -= b
		c ^= rot(b, 4)
		b += a
	}
	var final = func() {
		c ^= b
		c -= rot(b, 14)
		a ^= c
		a -= rot(c, 11)
		b ^= a
		b -= rot(a, 25)
		c ^= b
		c -= rot(b, 16)
		a ^= c
		a -= rot(c, 4)
		b ^= a
		b -= rot(a, 14)
		c ^= b
		c -= rot(b, 24)
	}
	ul := uint32(len(k))
	a = 0xdeadbeef + ul<<2 + seed
	b, c = a, a

	i := 0
	//length := 0
	for length = len(k); length > 3; length -= 3 {
		a += k[i+0]
		b += k[i+1]
		c += k[i+2]
		mix()
		i += 3
	}

	switch length {
	case 3:
		c += k[i+2]
		fallthrough
	case 2:
		b += k[i+1]
		fallthrough
	case 1:
		a += k[i+0]
		final()
	case 0:
		break
	}
	return c
}

// jenkins364: return 2 32-bit hash values.
// Returns two 32-bit hash values instead of just one.
// This is good enough for hash table lookup with 2^^64 buckets,
// or if you want a second hash if you're not happy with the first,
// or if you want a probably-unique 64-bit ID for the key.
// *pc is better mixed than *pb, so use *pc first.
// If you want a 64-bit value do something like "*pc + (((uint64_t)*pb)<<32)"
func Jenkins364(k []byte, length int, pc, pb uint32) (rpc, rpb uint32) {
	var a, b, c uint32

	//fmt.Printf("Jenkins364: k=%v, len(k)=%d\n", k, len(k))
	if length == 0 {
		length = len(k)
	}
	/*
		var rot = func(x, k uint32) uint32 {
			return x << k | x >> (32 - k)
		}

		var mix = func() {
			a -= c;  a ^= rot(c, 4); c += b;
			b -= a;  b ^= rot(a, 6);  a += c;
			c -= b;  c ^= rot(b, 8);  b += a;
			a -= c;  a ^= rot(c,16);  c += b;
			b -= a;  b ^= rot(a,19);  a += c;
			c -= b;  c ^= rot(b, 4);  b += a;
		}
		var final = func() {
			c ^= b; c -= rot(b,14);
			a ^= c; a -= rot(c,11);
			b ^= a; b -= rot(a,25);
			c ^= b; c -= rot(b,16);
			a ^= c; a -= rot(c,4);
			b ^= a; b -= rot(a,14);
			c ^= b; c -= rot(b,24);
		}
	*/
	ul := uint32(len(k))

	/* Set up the internal state */
	a = 0xdeadbeef + ul + pc
	b, c = a, a
	c += pb

	for ; length > 12; length -= 12 {
		//fmt.Printf("k=%q, length=%d\n", k, length)
		a += *(*uint32)(unsafe.Pointer(&k[0]))
		b += *(*uint32)(unsafe.Pointer(&k[4]))
		c += *(*uint32)(unsafe.Pointer(&k[8]))
		a, b, c = mix1(a, b, c)
		a, b, c = mix2(a, b, c)
		k = k[12:]
	}
	//fmt.Printf("k=%q, length=%d\n", k, length)

	/* handle the last (probably partial) block */
	/*
	 * "k[2]&0xffffff" actually reads beyond the end of the string, but
	 * then masks off the part it's not allowed to read.  Because the
	 * string is aligned, the masked-off tail is in the same word as the
	 * rest of the string.  Every machine with memory protection I've seen
	 * does it on word boundaries, so is OK with this.  But VALGRIND will
	 * still catch it and complain.  The masking trick does make the hash
	 * noticably faster for short strings (like English words).
	 */

	//fmt.Printf("length now=%d\n", length)
	switch length {
	case 12:
		a += *(*uint32)(unsafe.Pointer(&k[0]))
		b += *(*uint32)(unsafe.Pointer(&k[4]))
		c += *(*uint32)(unsafe.Pointer(&k[8]))
	case 11:
		c += uint32(k[10]) << 16
		fallthrough
	case 10:
		c += uint32(k[9]) << 8
		fallthrough
	case 9:
		c += uint32(k[8])
		fallthrough
	case 8:
		a += *(*uint32)(unsafe.Pointer(&k[0]))
		b += *(*uint32)(unsafe.Pointer(&k[4]))
		break
	case 7:
		b += uint32(k[6]) << 16
		fallthrough
	case 6:
		b += uint32(k[5]) << 8
		fallthrough
	case 5:
		b += uint32(k[4])
		fallthrough
	case 4:
		a += *(*uint32)(unsafe.Pointer(&k[0]))
		break
	case 3:
		a += uint32(k[2]) << 16
		fallthrough
	case 2:
		a += uint32(k[1]) << 8
		fallthrough
	case 1:
		a += uint32(k[0])
		break
	case 0:
		//fmt.Printf("case 0\n")
		return c, b /* zero length strings require no mixing */
	}
	a, b, c = final1(a, b, c)
	a, b, c = final2(a, b, c)
	return c, b
}

func HashString(s string, pc, pb uint32) (rpc, rpb uint32) {
	k := ([]byte)(s)
	rpc, rpb = Jenkins364(k, len(k), pc, pb)
	return
}

func HashBytesLength(k []byte, length int, seed uint32) uint32 {
	if length > len(k) {
		fmt.Printf("len(k)=%d, length=%d\n", len(k), length)
		panic("HashBytesLength")
	}
	ret, _ := Jenkins364(k, length, seed, 0)
	return ret
}

//
// Streaming interface and new interface

type State332 struct {
	hash uint32
	seed uint32
	pc   uint32
	pb   uint32
	clen int
	tail []byte
}

type State332c struct {
	State332
}

type State332b struct {
	State332
}

type State364 struct {
	hash uint64
	seed uint64
	pc   uint32
	pb   uint32
	clen int
	tail []byte
}

type State264 struct {
	hash uint64
	seed uint64
}

type State232 struct {
	hash uint32
	seed uint32
	clen int
	tail []byte
}

// const Size = 4

// Sum32 returns the 32 bit hash of data given the seed.
// This is code is what I started with before I added the hash.Hash and hash.Hash32 interfaces.
func Sum32(data []byte, seed uint32) uint32 {
	rpc, _ := Jenkins364(data, len(data), seed, seed)
	return rpc
}

// New returns a new hash.Hash32 interface that computes a 32 bit jenkins lookup3 hash.
func New(seed uint32) hash.Hash32 {
	s := new(State332c)
	s.seed = seed
	s.Reset()
	return s
}

// New returns a new nhash.HashF32 interface that computes a 32 bit jenkins lookup3 hash.
func New332c(seed uint32) nhash.HashF32 {
	s := new(State332c)
	s.seed = seed
	s.Reset()
	return s
}

/*
func New332b(seed uint32) nhash.HashF32 {
	s := new(State332b)
	s.seed = seed
	s.Reset()
	return s
}
*/

// Return the size of the resulting hash.
func (d *State332c) Size() int { return 4 }

// Return the blocksize of the hash which in this case is 1 byte.
func (d *State332c) BlockSize() int { return 1 }

// Return the maximum number of seed bypes required. In this case 2 x 32
func (d *State332c) NumSeedBytes() int {
	return 8
}

// Return the number of bits the hash function outputs.
func (d *State332c) HashSizeInBits() int {
	return 32
}

// Reset the hash state.
func (d *State332c) Reset() {
	d.pc = d.seed
	d.pb = d.seed
	d.clen = 0
	d.tail = nil
}

// Accept a byte stream p used for calculating the hash. For now this call is lazy and the actual hash calculations take place in Sum() and Sum32().
func (d *State332c) Write(p []byte) (nn int, err error) {
	l := len(p)
	d.clen += l
	d.tail = append(d.tail, p...)
	return l, nil
}

// Return the current hash as a byte slice.
func (d *State332c) Sum(b []byte) []byte {
	d.pc, d.pb = Jenkins364(d.tail, len(d.tail), d.pc, d.pb)
	d.hash = d.pc
	h := d.hash
	return append(b, byte(h>>24), byte(h>>16), byte(h>>8), byte(h))
}

// Return the current hash as a 32 bit unsigned type.
func (d *State332c) Sum32() uint32 {
	d.pc, d.pb = Jenkins364(d.tail, len(d.tail), d.pc, d.pb)
	d.hash = d.pc
	return d.hash
}

// Given b as input and an optional 32 bit seed return the Jenkins lookup3 hash c bits.
func (d *State332c) Hash32(b []byte, seeds ...uint32) uint32 {
	//fmt.Printf("len(b)=%d, b=%x\n", len(b), b)
	/*
		switch len(seeds) {
		case 2:
			d.pb = seeds[1]
			fallthrough
		case 1:
			d.pc = seeds[0]
		default:
			d.pc, d.pb = 0, 0
		}
	*/
	d.pc, d.pb = 0, 0

	d.pc, d.pb = Jenkins364(b, len(b), d.pc, d.pb)
	d.hash = d.pc
	return d.hash
}

// ----

// changed this from HashF64 to Hash64 to get streaming functions, what did I break? 7-22-15
func New364(seed uint64) nhash.Hash64 {
	s := new(State364)
	s.seed = seed
	s.Reset()
	return s
}

// Return the size of the resulting hash.
func (d *State364) Size() int { return 8 }

// Return the blocksize of the hash which in this case is 1 byte.
func (d *State364) BlockSize() int { return 1 }

// Return the maximum number of seed bypes required. In this case 2 x 32
func (d *State364) NumSeedBytes() int {
	return 8
}

// Return the number of bits the hash function outputs.
func (d *State364) HashSizeInBits() int {
	return 64
}

// Reset the hash state.
func (d *State364) Reset() {
	d.pc = uint32(d.seed)
	d.pb = uint32(d.seed >> 32)
	d.clen = 0
	d.tail = nil
}

// Accept a byte stream p used for calculating the hash. For now this call is lazy and the actual hash calculations take place in Sum() and Sum32().
func (d *State364) Write(p []byte) (nn int, err error) {
	l := len(p)
	d.clen += l
	d.tail = append(d.tail, p...)
	return l, nil
}

// Return the current hash as a byte slice.
func (d *State364) Sum(b []byte) []byte {
	d.pc, d.pb = Jenkins364(d.tail, len(d.tail), d.pc, d.pb)
	d.hash = uint64(d.pb<<32) | uint64(d.pc)
	h := d.hash
	return append(b, byte(h>>56), byte(h>>48), byte(h>>40), byte(h>>32), byte(h>>24), byte(h>>16), byte(h>>8), byte(h))
}

func (d *State364) Write64(h uint64) (err error) {
	d.clen += 8
	d.tail = append(d.tail, byte(h>>56), byte(h>>48), byte(h>>40), byte(h>>32), byte(h>>24), byte(h>>16), byte(h>>8), byte(h))
	return nil
}

// Return the current hash as a 64 bit unsigned type.
func (d *State364) Sum64() uint64 {
	d.pc, d.pb = Jenkins364(d.tail, len(d.tail), d.pc, d.pb)
	d.hash = uint64(d.pb)<<32 | uint64(d.pc)
	return d.hash
}

func (d *State364) Hash64(b []byte, seeds ...uint64) uint64 {
	switch len(seeds) {
	case 2:
		d.pc = uint32(seeds[0])
		d.pb = uint32(seeds[1])
	case 1:
		d.pc = uint32(seeds[0])
		d.pb = uint32(seeds[0] >> 32)
	}
	d.pc, d.pb = Jenkins364(b, len(b), d.pc, d.pb)
	//fmt.Printf("pc=0x%08x, pb=0x%08x\n", d.pc, d.pb)

	// pc is better mixed than pb so pc goes in the low order bits
	d.hash = uint64(d.pb)<<32 | uint64(d.pc)
	return d.hash
}

func (d *State364) Hash64S(b []byte, seed uint64) uint64 {
	d.pc, d.pb = uint32(seed), uint32(seed>>32)
	d.pc, d.pb = Jenkins364(b, len(b), d.pc, d.pb)
	//fmt.Printf("pc=0x%08x, pb=0x%08x\n", d.pc, d.pb)

	// pc is better mixed than pb so pc goes in the low order bits
	d.hash = uint64(d.pb)<<32 | uint64(d.pc)
	return d.hash
}

// ----

// New returns a new nhash.HashF32 interface that computes a 32 bit jenkins lookup3 hash.
func New232(seed uint32) nhash.HashF32 {
	s := new(State232)
	s.seed = seed
	s.Reset()
	return s
}

// Return the size of the resulting hash.
func (s *State232) Size() int { return 4 }

// Return the blocksize of the hash which in this case is 12 bytes at a time.
func (s *State232) BlockSize() int { return 12 }

// Return the maximum number of seed bypes required.
func (s *State232) NumSeedBytes() int {
	return 4
}

// Return the number of bits the hash function outputs.
func (s *State232) HashSizeInBits() int {
	return 32
}

// Reset the hash state.
func (s *State232) Reset() {
	s.seed = 0
	s.clen = 0
	s.tail = nil
}

// Accept a byte stream p used for calculating the hash. For now this call is lazy and the actual hash calculations take place in Sum() and Sum32().
func (s *State232) Write(p []byte) (nn int, err error) {
	l := len(p)
	s.clen += l
	s.tail = append(s.tail, p...)
	return l, nil
}

// Return the current hash as a byte slice.
func (s *State232) Sum(b []byte) []byte {
	s.hash = Hash232(s.tail, s.seed)
	h := s.hash
	return append(b, byte(h>>24), byte(h>>16), byte(h>>8), byte(h))
}

// Return the current hash as a 64 bit unsigned type.
func (s *State232) Sum32() uint32 {
	s.hash = Hash232(s.tail, s.seed)
	return s.hash
}

func (s *State232) Hash32(b []byte, seeds ...uint32) uint32 {
	//	fmt.Printf("Hash32: len(b)=%d, b=%x\n", len(b), b)
	if len(seeds) > 0 {
		s.seed = 0
	}
	s.hash = Hash232(b, s.seed)
	return s.hash
}

/*
	var mix = func() {
		a -= c;  a ^= c << 4 | c >> (32 - 4);  c += b;
		b -= a;  b ^= a << 6 | a >> (32 - 6);  a += c;
		c -= b;  c ^= b << 8 | b >> (32 - 8);  b += a;
		a -= c;  a ^= c << 16 | c >> (32 - 16);  c += b;
		b -= a;  b ^= a << 19 | a >> (32 - 19);  a += c;
		c -= b;  c ^= b << 4 | b >> (32 - 4);  b += a;
	}

	var final = func() {
		c ^= b; c -= b << 14 | b >> (32 - 14);
		a ^= c; a -= c << 11 | c >> (32 - 11);
		b ^= a; b -= a << 25 | a >> (32 - 25);
		c ^= b; c -= b << 16 | b >> (32 - 16);
		a ^= c; a -= c << 4 | c >> (32 - 4);
		b ^= a; b -= a << 14 | a >> (32 - 14);
		c ^= b; c -= b << 24 | b >> (32 - 24);
	}
*/
