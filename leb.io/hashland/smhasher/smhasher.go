// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smhasher

import (
	"fmt"
	"leb.io/hashland/hashf"
	"math"
	"math/rand"
	"strings"
	"time"
)

// stubs for now
func HaveGoodHash() bool {
	return true
}

func Short() bool {
	return false
}

func SetBytes(amt int64) {

}

/*
func BytesHash(b []byte, aseed uintptr) uintptr {
	seed := uint32(aseed)
	pc, pb := jenkins.Jenkins364(b, len(b), seed, ^seed)
	return uintptr(uint64(pb)<<32 | uint64(pc))
}

func StringHash (s string, aseed uintptr) uintptr {
	seed := uint32(aseed)
	pc, pb := jenkins.HashString(s, seed, ^seed)
	return uintptr(uint64(pb)<<32 | uint64(pc))
}

func Int32Hash(i uint32, seed uint32) uintptr {
	b := make([]byte, 4, 4)
	b = b[:]
	b[0], b[1], b[2], b[3] = byte(i&0xFF), byte((i>>8)&0xFF), byte((i>>16)&0xFF), byte((i>>24)&0xFF)
	pc, pb := jenkins.Jenkins364(b, len(b), seed, seed)
	return uintptr(uint64(pb)<<32 | uint64(pc))
}

func Int64Hash(i uint64, seed uint32) uintptr {
	b := make([]byte, 8, 8)
	b = b[:]
	b[0], b[1], b[2], b[3], b[4], b[5], b[6], b[7] = byte(i&0xFF), byte((i>>8)&0xFF), byte((i>>16)&0xFF), byte((i>>24)&0xFF), byte((i>>32)&0xFF), byte((i>>40)&0xFF), byte((i>>48)&0xFF),
	byte((i>>56)&0xFF)
	pc, pb := jenkins.Jenkins364(b, len(b), seed, seed)
	return uintptr(uint64(pb)<<32 | uint64(pc))
}
*/

/*
func BytesHash(b []byte, aseed uintptr) uintptr {
	seed := uint64(aseed)
	h := jenkins.Hash264(b, seed)
	return uintptr(h)
}

func StringHash (s string, aseed uintptr) uintptr {
	seed := uint64(aseed)
	b := make([]byte, len(s), len(s))
	b = b[:]
	copy(b, s)
	h := jenkins.Hash264(b, seed)
	return uintptr(h)
}

func Int32Hash(i uint32, aseed uintptr) uintptr {
	seed := uint64(aseed)
	b := make([]byte, 4, 4)
	b = b[:]
	b[0], b[1], b[2], b[3] = byte(i&0xFF), byte((i>>8)&0xFF), byte((i>>16)&0xFF), byte((i>>24)&0xFF)
	h := jenkins.Hash264(b, seed)
	return uintptr(h)
}

func Int64Hash(i uint64, aseed uintptr) uintptr {
	seed := uint64(aseed)
	b := make([]byte, 8, 8)
	b = b[:]
	b[0], b[1], b[2], b[3], b[4], b[5], b[6], b[7] = byte(i&0xFF), byte((i>>8)&0xFF), byte((i>>16)&0xFF), byte((i>>24)&0xFF), byte((i>>32)&0xFF), byte((i>>40)&0xFF), byte((i>>48)&0xFF),
	byte((i>>56)&0xFF)
	h := jenkins.Hash264(b, seed)
	return uintptr(h)
}
*/

func BytesHash(b []byte, aseed uintptr) uintptr {
	seed := uint64(aseed)
	h := hashf.Hashf(b, seed)
	return uintptr(h)
}

func StringHash(s string, aseed uintptr) uintptr {
	seed := uint64(aseed)
	b := make([]byte, len(s), len(s))
	b = b[:]
	copy(b, s)
	h := hashf.Hashf(b, seed)
	return uintptr(h)
}

func Int32Hash(i uint32, aseed uintptr) uintptr {
	seed := uint64(aseed)
	b := make([]byte, 4, 4)
	b = b[:]
	b[0], b[1], b[2], b[3] = byte(i&0xFF), byte((i>>8)&0xFF), byte((i>>16)&0xFF), byte((i>>24)&0xFF)
	h := hashf.Hashf(b, seed)
	return uintptr(h)
}

func Int64Hash(i uint64, aseed uintptr) uintptr {
	seed := uint64(aseed)
	b := make([]byte, 8, 8)
	b = b[:]
	b[0], b[1], b[2], b[3], b[4], b[5], b[6], b[7] = byte(i&0xFF), byte((i>>8)&0xFF), byte((i>>16)&0xFF), byte((i>>24)&0xFF), byte((i>>32)&0xFF), byte((i>>40)&0xFF), byte((i>>48)&0xFF),
		byte((i>>56)&0xFF)
	h := hashf.Hashf(b, seed)
	return uintptr(h)
}

// Smhasher is a torture test for hash functions.
// https://code.google.com/p/smhasher/
// This code is a port of some of the Smhasher tests to Go.
//
// The current AES hash function passes Smhasher.  Our fallback
// hash functions don't, so we only enable the difficult tests when
// we know the AES implementation is available.

// Sanity checks.
// hash should not depend on values outside key.
// hash should not depend on alignment.
func TestSmhasherSanity(t *TState) (ret bool) {
	r := rand.New(rand.NewSource(1234))
	const REP = 10
	const KEYMAX = 128
	const PAD = 16
	const OFFMAX = 16
	for k := 0; k < REP; k++ {
		for n := 0; n < KEYMAX; n++ {
			for i := 0; i < OFFMAX; i++ {
				var b [KEYMAX + OFFMAX + 2*PAD]byte
				var c [KEYMAX + OFFMAX + 2*PAD]byte
				randBytes(r, b[:])
				randBytes(r, c[:])
				copy(c[PAD+i:PAD+i+n], b[PAD:PAD+n])
				if BytesHash(b[PAD:PAD+n], 0) != BytesHash(c[PAD+i:PAD+i+n], 0) {
					fmt.Printf("hash depends on bytes outside key")
					ret = true
				}
			}
		}
	}
	return
}

type HashSet struct {
	m map[uintptr]struct{} // set of hashes added
	n int                  // number of hashes added
}

func newHashSet() *HashSet {
	return &HashSet{make(map[uintptr]struct{}), 0}
}
func (s *HashSet) add(h uintptr) {
	s.m[h] = struct{}{}
	s.n++
}
func (s *HashSet) addS(x string) {
	s.add(StringHash(x, 0))
}
func (s *HashSet) addB(x []byte) {
	s.add(BytesHash(x, 0))
}
func (s *HashSet) addS_seed(x string, seed uintptr) {
	s.add(StringHash(x, seed))
}
func (s *HashSet) check() (ret bool) {
	const SLOP = 10.0
	collisions := s.n - len(s.m)
	//
	//fmt.Printf("check: %d/%d\n", len(s.m), s.n)
	pairs := int64(s.n) * int64(s.n-1) / 2
	expected := float64(pairs) / math.Pow(2.0, float64(hashSize))
	stddev := math.Sqrt(expected)
	if float64(collisions) > expected+SLOP*3*stddev {
		fmt.Printf("check: unexpected number of collisions: got=%d mean=%f stddev=%f\n", collisions, expected, stddev)
		ret = true
	}
	return
}

// a string plus adding zeros must make distinct hashes
func TestSmhasherAppendedZeros(t *TState) bool {
	s := "hello" + strings.Repeat("\x00", 256)
	h := newHashSet()
	for i := 0; i <= len(s); i++ {
		h.addS(s[:i])
	}
	return h.check()
}

// All 0-3 byte strings have distinct hashes.
func TestSmhasherSmallKeys(t *TState) bool {
	h := newHashSet()
	var b [3]byte
	for i := 0; i < 256; i++ {
		b[0] = byte(i)
		h.addB(b[:1])
		for j := 0; j < 256; j++ {
			b[1] = byte(j)
			h.addB(b[:2])
			if !Short() {
				for k := 0; k < 256; k++ {
					b[2] = byte(k)
					h.addB(b[:3])
				}
			}
		}
	}
	return h.check()
}

// Different length strings of all zeros have distinct hashes.
func TestSmhasherZeros(t *TState) bool {
	N := 256 * 1024
	if Short() {
		N = 1024
	}
	h := newHashSet()
	b := make([]byte, N)
	for i := 0; i <= N; i++ {
		h.addB(b[:i])
	}
	return h.check()
}

// Strings with up to two nonzero bytes all have distinct hashes.
func TestSmhasherTwoNonzero(t *TState) bool {
	if Short() {
		//t.Skip("Skipping in short mode")
	}
	h := newHashSet()
	for n := 2; n <= 16; n++ {
		twoNonZero(h, n)
	}
	return h.check()
}
func twoNonZero(h *HashSet, n int) {
	b := make([]byte, n)

	// all zero
	h.addB(b[:])

	// one non-zero byte
	for i := 0; i < n; i++ {
		for x := 1; x < 256; x++ {
			b[i] = byte(x)
			h.addB(b[:])
			b[i] = 0
		}
	}

	// two non-zero bytes
	for i := 0; i < n; i++ {
		for x := 1; x < 256; x++ {
			b[i] = byte(x)
			for j := i + 1; j < n; j++ {
				for y := 1; y < 256; y++ {
					b[j] = byte(y)
					h.addB(b[:])
					b[j] = 0
				}
			}
			b[i] = 0
		}
	}
}

// Test strings with repeats, like "abcdabcdabcdabcd..."
func TestSmhasherCyclic(t *TState) (ret bool) {
	if Short() {
		//t.Skip("Skipping in short mode")
	}
	if !HaveGoodHash() {
		//t.Skip("fallback hash not good enough for this test")
	}
	r := rand.New(rand.NewSource(1234))
	const REPEAT = 8
	const N = 1000000
	for n := 4; n <= 12; n++ {
		h := newHashSet()
		b := make([]byte, REPEAT*n)
		for i := 0; i < N; i++ {
			b[0] = byte(i * 79 % 97)
			b[1] = byte(i * 43 % 137)
			b[2] = byte(i * 151 % 197)
			b[3] = byte(i * 199 % 251)
			randBytes(r, b[4:n])
			for j := n; j < n*REPEAT; j++ {
				b[j] = b[j-n]
			}
			h.addB(b)
		}
		if h.check() {
			ret = true
		}
	}
	return
}

type pair struct {
	n int
	k int
}

var pairs = []pair{pair{32, 6}, pair{40, 6}, pair{48, 5}, pair{56, 5}, pair{64, 5}, pair{96, 4}, pair{256, 3}, pair{2048, 2}}

// Test strings with only a few bits set
func TestSmhasherSparse(t *TState) (ret bool) {
	if Short() {
		//t.Skip("Skipping in short mode")
	}
	for _, v := range pairs {
		if sparse(v.n, v.k) {
			ret = true
		}
	}
	return
}
func sparse(n int, k int) bool {
	b := make([]byte, n/8)
	h := newHashSet()
	setbits(h, b, 0, k)
	return h.check()
}

// set up to k bits at index i and greater
func setbits(h *HashSet, b []byte, i int, k int) {
	h.addB(b)
	if k == 0 {
		return
	}
	for j := i; j < len(b)*8; j++ {
		b[j/8] |= byte(1 << uint(j&7))
		setbits(h, b, j+1, k-1)
		b[j/8] &= byte(^(1 << uint(j&7)))
	}
}

type Permutation struct {
	n int
	s []uint32
}

var Permutations = []Permutation{
	Permutation{8, []uint32{0, 1, 2, 3, 4, 5, 6, 7}},
	Permutation{8, []uint32{0, 1 << 29, 2 << 29, 3 << 29, 4 << 29, 5 << 29, 6 << 29, 7 << 29}},
	Permutation{20, []uint32{0, 1}},
	Permutation{20, []uint32{0, 1 << 31}},
	Permutation{6, []uint32{0, 1, 2, 3, 4, 5, 6, 7, 1 << 29, 2 << 29, 3 << 29, 4 << 29, 5 << 29, 6 << 29, 7 << 29}},
}

// Test all possible combinations of n blocks from the set s.
// "permutation" is a bad name here, but it is what Smhasher uses.
func TestSmhasherPermutation(t *TState) bool {
	if Short() {
		//t.Skip("Skipping in short mode")
	}
	if !HaveGoodHash() {
		//t.Skip("fallback hash not good enough for this test")
	}
	for _, v := range Permutations {
		fmt.Printf("\n\t\tn=%d, s=%v", v.n, v.s)
		permutation(v.s, v.n)
	}
	return false
}
func permutation(s []uint32, n int) {
	b := make([]byte, n*4)
	h := newHashSet()
	genPerm(h, b, s, 0)
	h.check()
}
func genPerm(h *HashSet, b []byte, s []uint32, n int) {
	h.addB(b[:n])
	if n == len(b) {
		return
	}
	for _, v := range s {
		b[n] = byte(v)
		b[n+1] = byte(v >> 8)
		b[n+2] = byte(v >> 16)
		b[n+3] = byte(v >> 24)
		genPerm(h, b, s, n+4)
	}
}

type Key interface {
	clear()              // set bits all to 0
	random(r *rand.Rand) // set key to something random
	bits() int           // how many bits key has
	flipBit(i int)       // flip bit i of the key
	hash() uintptr       // hash the key
	name() string        // for error reporting
}

type BytesKey struct {
	b []byte
}

func (k *BytesKey) clear() {
	for i := range k.b {
		k.b[i] = 0
	}
}
func (k *BytesKey) random(r *rand.Rand) {
	randBytes(r, k.b)
}
func (k *BytesKey) bits() int {
	return len(k.b) * 8
}
func (k *BytesKey) flipBit(i int) {
	k.b[i>>3] ^= byte(1 << uint(i&7))
}
func (k *BytesKey) hash() uintptr {
	return BytesHash(k.b, 0)
}
func (k *BytesKey) name() string {
	return fmt.Sprintf("bytes%d", len(k.b))
}

type Int32Key struct {
	i uint32
}

func (k *Int32Key) clear() {
	k.i = 0
}
func (k *Int32Key) random(r *rand.Rand) {
	k.i = r.Uint32()
}
func (k *Int32Key) bits() int {
	return 32
}
func (k *Int32Key) flipBit(i int) {
	k.i ^= 1 << uint(i)
}
func (k *Int32Key) hash() uintptr {
	return Int32Hash(k.i, 0)
}
func (k *Int32Key) name() string {
	return "int32"
}

type Int64Key struct {
	i uint64
}

func (k *Int64Key) clear() {
	k.i = 0
}
func (k *Int64Key) random(r *rand.Rand) {
	k.i = uint64(r.Uint32()) + uint64(r.Uint32())<<32
}
func (k *Int64Key) bits() int {
	return 64
}
func (k *Int64Key) flipBit(i int) {
	k.i ^= 1 << uint(i)
}
func (k *Int64Key) hash() uintptr {
	return Int64Hash(k.i, 0)
}
func (k *Int64Key) name() string {
	return "int64"
}

// Flipping a single bit of a key should flip each output bit with 50% probability.
func TestSmhasherAvalanche(t *TState) bool {
	if !HaveGoodHash() {
		//t.Skip("fallback hash not good enough for this test")
	}
	if Short() {
		//t.Skip("Skipping in short mode")
	}
	avalancheTest1(&BytesKey{make([]byte, 2)})
	avalancheTest1(&BytesKey{make([]byte, 4)})
	avalancheTest1(&BytesKey{make([]byte, 8)})
	avalancheTest1(&BytesKey{make([]byte, 16)})
	avalancheTest1(&BytesKey{make([]byte, 32)})
	avalancheTest1(&BytesKey{make([]byte, 200)})
	avalancheTest1(&Int32Key{})
	avalancheTest1(&Int64Key{})
	return false
}
func avalancheTest1(k Key) {
	const REP = 100000
	r := rand.New(rand.NewSource(1234))
	n := k.bits()
	fmt.Printf("\t\tz=%d, n=%d\n", REP, n)

	// grid[i][j] is a count of whether flipping
	// input bit i affects output bit j.
	grid := make([][hashSize]int, n)

	for z := 0; z < REP; z++ {
		// pick a random key, hash it
		k.random(r)
		h := k.hash()

		// flip each bit, hash & compare the results
		for i := 0; i < n; i++ {
			k.flipBit(i)
			d := h ^ k.hash()
			k.flipBit(i)

			// record the effects of that bit flip
			g := &grid[i]
			for j := 0; j < hashSize; j++ {
				g[j] += int(d & 1)
				d >>= 1
			}
		}
	}

	// Each entry in the grid should be about REP/2.
	// More precisely, we did N = k.bits() * hashSize experiments where
	// each is the sum of REP coin flips.  We want to find bounds on the
	// sum of coin flips such that a truly random experiment would have
	// all sums inside those bounds with 99% probability.
	N := n * hashSize
	var c float64
	// find c such that Prob(mean-c*stddev < x < mean+c*stddev)^N > .9999
	for c = 0.0; math.Pow(math.Erf(c/math.Sqrt(2)), float64(N)) < .9999; c += .1 {
	}
	c *= 4.0 // allowed slack - we don't need to be perfectly random
	mean := .5 * REP
	stddev := .5 * math.Sqrt(REP)
	low := int(mean - c*stddev)
	high := int(mean + c*stddev)
	for i := 0; i < n; i++ {
		for j := 0; j < hashSize; j++ {
			x := grid[i][j]
			if x < low || x > high {
				//t.Errorf("bad bias for %s bit %d -> bit %d: %d/%d\n", k.name(), i, j, x, REP)
			}
		}
	}
}

// All bit rotations of a set of distinct keys
func TestSmhasherWindowed(t *TState) bool {
	windowed(&Int32Key{})
	windowed(&Int64Key{})
	windowed(&BytesKey{make([]byte, 128)})
	return false
}
func windowed(k Key) {
	if Short() {
		//t.Skip("Skipping in short mode")
	}
	const BITS = 16

	for r := 0; r < k.bits(); r++ {
		h := newHashSet()
		for i := 0; i < 1<<BITS; i++ {
			k.clear()
			for j := 0; j < BITS; j++ {
				if i>>uint(j)&1 != 0 {
					k.flipBit((j + r) % k.bits())
				}
			}
			h.add(k.hash())
		}
		h.check()
	}
}

// All keys of the form prefix + [A-Za-z0-9]*N + suffix.
func TestSmhasherText(t *TState) bool {
	if Short() {
		//t.Skip("Skipping in short mode")
	}
	text("Foo", "Bar")
	text("FooBar", "")
	text("", "FooBar")
	return false
}
func text(prefix, suffix string) {
	const N = 4
	const S = "ABCDEFGHIJKLMNOPQRSTabcdefghijklmnopqrst0123456789"
	const L = len(S)
	b := make([]byte, len(prefix)+N+len(suffix))
	copy(b, prefix)
	copy(b[len(prefix)+N:], suffix)
	h := newHashSet()
	c := b[len(prefix):]
	for i := 0; i < L; i++ {
		c[0] = S[i]
		for j := 0; j < L; j++ {
			c[1] = S[j]
			for k := 0; k < L; k++ {
				c[2] = S[k]
				for x := 0; x < L; x++ {
					c[3] = S[x]
					h.addB(b)
				}
			}
		}
	}
	h.check()
}

// Make sure different seed values generate different hashes.
func TestSmhasherSeed(t *TState) bool {
	h := newHashSet()
	const N = 100000
	s := "hello"
	for i := 0; i < N; i++ {
		h.addS_seed(s, uintptr(i))
	}
	return h.check()
}

// size of the hash output (32 or 64 bits)
const hashSize = 32 + int(^uintptr(0)>>63<<5)

func randBytes(r *rand.Rand, b []byte) {
	for i := range b {
		b[i] = byte(r.Uint32())
	}
}

func benchmarkHash(n, N int) {
	s := strings.Repeat("A", n)

	for i := 0; i < N; i++ {
		StringHash(s, 0)
	}
	SetBytes(int64(n))
}

var N = 0

func BenchmarkHash5()     { benchmarkHash(5, N) }
func BenchmarkHash16()    { benchmarkHash(16, N) }
func BenchmarkHash64()    { benchmarkHash(64, N) }
func BenchmarkHash1024()  { benchmarkHash(1024, N) }
func BenchmarkHash65536() { benchmarkHash(65536, N) }

type Test struct {
	pf   func(ts *TState) bool
	desc string
}

var Tests = []Test{
	Test{TestSmhasherSanity, "TestSmhasherSanity"},
	Test{TestSmhasherSeed, "TestSmhasherSeed"},
	Test{TestSmhasherText, "TestSmhasherText"},
	Test{TestSmhasherWindowed, "TestSmhasherWindowed"},
	Test{TestSmhasherAvalanche, "TestSmhasherAvalanche"},
	Test{TestSmhasherPermutation, "TestSmhasherPermutation"},
	Test{TestSmhasherSparse, "TestSmhasherSparse"},
	Test{TestSmhasherCyclic, "TestSmhasherCyclic"},
	Test{TestSmhasherTwoNonzero, "TestSmhasherSmallKeys"},
	Test{TestSmhasherSmallKeys, "TestSmhasherZeros"},
	Test{TestSmhasherAppendedZeros, "TestSmhasherAppendedZeros"},
}

type TState struct {
	hashf   string
	verbose bool
}

func tdiff(begin, end time.Time) time.Duration {
	d := end.Sub(begin)
	return d
}

func RunSMhasher(hashf string, v bool) {
	ts := TState{hashf: hashf, verbose: v}
	for _, test := range Tests {
		if v {
			fmt.Printf("\t%q: ", test.desc)
		}
		start := time.Now()
		_ = test.pf(&ts)
		stop := time.Now()
		if v {
			fmt.Printf("%v\n", tdiff(start, stop))
		}
	}
	return
}
