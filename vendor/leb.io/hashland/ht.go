// Copyright © 2014,2015 Lawrence E. Bakst. All rights reserved.
package main

// based on http://amsoftware.narod.ru/algo.html
// However I can't replicate the result

import (
	"bufio"
	"crypto/sha1"
	"flag"
	"fmt"
	"hash"
	"io"
	. "leb.io/hashland/hashf"     // cleaved
	. "leb.io/hashland/hashtable" // cleaved
	"leb.io/hashland/nhash"
	"leb.io/hashland/smhasher"
	"leb.io/hrff"
	"math/rand"
	"os"
	"sort"
	"time"

	// remove these at some point
	"leb.io/aeshash"           // remove
	"leb.io/hashland/gomap"    // remove
	"leb.io/hashland/jenkins"  // remove
	"leb.io/hashland/keccak"   // remove
	"leb.io/hashland/keccakpg" // remove
	"leb.io/hashland/nullhash" // remove
	"leb.io/hashland/siphash"  // remove
)

func ReadFile(file string, cb func(line string)) int {
	var lines int
	//fmt.Printf("ReadFile: file=%q\n", file)
	f, err := os.Open(file)
	if err != nil {
		panic("ReadFile: opening file")
	}
	defer f.Close()

	rl := bufio.NewReader(f)
	//rs := csv.NewReader(f)
	// rs.Comma = '\t'      // Use tab-separated values

	for {
		//r, err := rs.Read()
		s, err := rl.ReadString(10) // 0x0A separator = newline
		if err == io.EOF {
			// fmt.Printf("ReadFile: EOF\n")
			return lines
		} else if err != nil {
			panic("reading file")
		}
		if s[len(s)-1] == '\n' {
			s = s[:len(s)-1]
		}
		if s[len(s)-1] == '\r' {
			s = s[:len(s)-1]
		}
		if s[len(s)-1] == ' ' {
			s = s[:len(s)-1]
		}
		//fmt.Printf("%q\n", s)
		if cb != nil {
			cb(s)
		}
		lines++
	}
}

func Test0(file string, lines int, hf2 string) (ht *HashTable) {
	var cnt int
	var countlines = func(line string) {
		cnt++
	}
	ht = NewHashTable(lines, *extra, *pd, *oa, *prime)
	start := time.Now()
	ReadFile(file, countlines)
	stop := time.Now()
	ht.Dur = tdiff(start, stop)
	return
}

func TestA(file string, lines int, hf2 string) (ht *HashTable) {
	//var lines int
	/*
		var countlines = func(line string) {
			lines++
		}
	*/
	var addLine = func(line string) {
		ht.Insert([]byte(line))
	}

	//fmt.Printf("\t%20q: ", hf2)
	//fmt.Printf("run: file=%q\n", file)
	//fmt.Printf("TestA: lines=%d, hf2=%q\n", lines, hf2)
	ht = NewHashTable(lines, *extra, *pd, *oa, *prime)
	//fmt.Printf("ht=%v\n", ht)
	start := time.Now()
	ReadFile(file, addLine)
	stop := time.Now()
	ht.Dur = tdiff(start, stop)
	return
}

func TestB(file string, lines int, hf2 string) (ht *HashTable) {
	var addLine = func(line string) {
		line += "\n"
		ht.Insert([]byte(line))
	}
	ht = NewHashTable(lines, *extra, *pd, *oa, *prime)
	start := time.Now()
	ReadFile(file, addLine)
	stop := time.Now()
	ht.Dur = tdiff(start, stop)
	return
}

func TestC(file string, lines int, hf2 string) (ht *HashTable) {
	var addLine = func(line string) {
		line += line + "\n\n\n\n"
		ht.Insert([]byte(line))
	}
	ht = NewHashTable(lines, *extra, *pd, *oa, *prime)
	start := time.Now()
	ReadFile(file, addLine)
	stop := time.Now()
	ht.Dur = tdiff(start, stop)
	return
}

func TestD(file string, lines int, hf2 string) (ht *HashTable) {
	var addLine = func(line string) {
		line = "ABCDE" + line
		ht.Insert([]byte(line))
	}
	ht = NewHashTable(lines, *extra, *pd, *oa, *prime)
	start := time.Now()
	ReadFile(file, addLine)
	stop := time.Now()
	ht.Dur = tdiff(start, stop)
	return
}

func TestE(file string, lines int, hf2 string) (ht *HashTable) {
	var addLine = func(line string) {
		line = line + line
		ht.Insert([]byte(line))
	}
	ht = NewHashTable(lines, *extra, *pd, *oa, *prime)
	start := time.Now()
	ReadFile(file, addLine)
	stop := time.Now()
	ht.Dur = tdiff(start, stop)
	return
}

func TestF(file string, lines int, hf2 string) (ht *HashTable) {
	var addLine = func(line string) {
		line = line + line + line + line
		ht.Insert([]byte(line))
	}
	ht = NewHashTable(lines, *extra, *pd, *oa, *prime)
	start := time.Now()
	ReadFile(file, addLine)
	stop := time.Now()
	ht.Dur = tdiff(start, stop)
	return
}

func reverse(s string) string {
	if len(s) == 0 {
		return ""
	}
	return reverse(s[1:]) + string(s[0])
}

func TestG(file string, lines int, hf2 string) (ht *HashTable) {
	var addLine = func(line string) {
		line2 := reverse(line)
		//fmt.Printf("line=%q, line2=%q", line, line2)
		ht.Insert([]byte(line2))
	}
	ht = NewHashTable(lines, *extra, *pd, *oa, *prime)
	start := time.Now()
	ReadFile(file, addLine)
	stop := time.Now()
	ht.Dur = tdiff(start, stop)
	return
}

func TestH(file string, lines int, hf2 string) (ht *HashTable) {
	var cnt int
	var counter = func(word string) {
		cnt++
	}
	var addWord = func(word string) {
		ht.Insert([]byte(word))
	}
	//test := []string{"abcdefgh", "efghijkl", "ijklmnop", "mnopqrst", "qrstuvwx", "uvwxyz01"} // 262144 words

	genWords(letters, counter)
	ht = NewHashTable(cnt, *extra, *pd, *oa, *prime)
	start := time.Now()
	genWords(letters, addWord)
	stop := time.Now()
	ht.Dur = tdiff(start, stop)
	return
}

// integers 0 to n
func TestI(file string, lines int, hf2 string) (ht *HashTable) {
	//fmt.Printf("ni=%d\n", *ni)
	bs := make([]byte, 4, 4)
	ht = NewHashTable(*ni, *extra, *pd, *oa, *prime)
	start := time.Now()
	for i := 0; i < *ni; i++ {
		bs[0], bs[1], bs[2], bs[3] = byte(i), byte(i>>8), byte(i>>16), byte(i>>24)
		ht.Insert(bs)
		//fmt.Printf("i=%d, 0x%08x, h=0x%08x\n", i, i, h)
	}
	stop := time.Now()
	ht.Dur = tdiff(start, stop)
	return
}

// marching 1
func TestJ(file string, lines int, hf2 string) (ht *HashTable) {
	length := 900
	keys := length * 8
	key := make([]byte, length, length)
	key = key[:]
	ht = NewHashTable(keys, *extra, *pd, *oa, *prime)
	start := time.Now()
	for k := range key {
		for i := uint(0); i < 8; i++ {
			key[k] = 1 << i
			//fmt.Printf("k=%d, i=%d, key=%#0x2\n", k, i, key)
			ht.Insert(key)
			key[k] = 0
		}
	}
	stop := time.Now()
	ht.Dur = tdiff(start, stop)
	return
}

func unhex(c byte) uint8 {
	switch {
	case '0' <= c && c <= '9':
		return c - '0'
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10
	}
	panic("unhex: bad input")
}

func hexToBytes(s string) []byte {
	var data = make([]byte, 1000, 1000)
	data = data[0 : len(s)/2]

	n := len(s)
	if (n & 1) == 1 {
		panic("gethex: string must be even")
	}
	for i := range data {
		data[i] = unhex(s[2*i])<<4 | unhex(s[2*i+1])
	}
	//fmt.Printf("hexToBytes: len(data)=%d, len(s)=%d\n", len(data), len(s))
	return data[0 : len(s)/2]
}

var r = rand.Float64

func rbetween(a uint64, b uint64) uint64 {
	rf := r()
	diff := float64(b - a + 1)
	r2 := rf * diff
	r3 := r2 + float64(a)
	//      fmt.Printf("rbetween: a=%d, b=%d, rf=%f, diff=%f, r2=%f, r3=%f\n", a, b, rf, diff, r2, r3)
	ret := uint64(r3)
	return ret
}

func TestK(file string, lines int, hf2 string) (ht *HashTable) {
	var seed uint64
	var rseed uint64
	var b = make([]byte, 1000, 1000)
	var hashLine = func(line string) {
		//fmt.Printf("line=%q\n", line)
		h := Hashf(b, rseed)
		b := hexToBytes(line)
		fmt.Printf("\t\tseed=%d, key=%x, hash=0x%x\n", rseed, b, h)
	}
	fmt.Printf("\n")
	ht = NewHashTable(lines, *extra, *pd, *oa, *prime)
	start := time.Now()
	for seed = 0; seed < uint64(*ns); seed++ {
		rseed = rbetween(0, uint64(1)<<63)
		ReadFile(file, hashLine)
	}
	stop := time.Now()
	ht.Dur = tdiff(start, stop)
	return
}

/*
func TestS(file string, lines int, hf2 string) (ht *HashTable) {
	var seed uint64
	var rseed uint64

	sha1160 := sha1.New()
	fp20 := make([]byte, 20, 20)

	sha1160.Reset()
	sha1160.Write(k)
	fp20 = fp20[0:0]
	fp20 = sha1160.Sum(fp20)

	return
}
*/

// [ABCDEFGH][EFGHIJKL][IJKLMNOP][MNOPQRST][QRSTUVWX][UVWXYZ01]
// given a slice of strings, generate all the combinations in order
func genWords(perms []string, f func(word string)) {
	var indices = make([]int, len(perms), len(perms))
	var idx int
	var inc = func() bool {
		// increment counter with carry
		for idx = 0; ; {
			indices[idx]++
			if indices[idx] >= len(perms[idx]) {
				indices[idx] = 0
				idx++
				if idx >= len(perms) {
					return true
				}
				continue
			} else {
				break
			}
		}
		return false
	}
	var letter = func(idx int, s string) string {
		return string(s[indices[idx]])
	}
	var word func(p []string) string
	word = func(p []string) string {
		if len(p) == 0 {
			return ""
		}
		l := len(p)
		idx := len(perms) - l
		tmp := letter(idx, p[0]) + word(p[1:])
		return tmp
	}
	// generate a word, hand it out, bump counter, repeat
	for {
		aword := word(perms)
		f(aword)
		if inc() {
			return
		}
	}
}

func tdiff(begin, end time.Time) time.Duration {
	d := end.Sub(begin)
	return d
}

// IntSlice attaches the methods of Interface to []int, sorting in increasing order.
type Uint32Slice []uint32

func (p Uint32Slice) Len() int           { return len(p) }
func (p Uint32Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Uint32Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// Sort is a convenience method.
func (p Uint32Slice) Sort() { sort.Sort(p) }

func checkForDups32(u Uint32Slice) (dups, mrun int) {
	i := 0
	run := 0
	for k, v := range u {
		if k == 0 || i == k {
			continue
		}
		if u[i] == v {
			run++
			dups++
			continue
		} else {
			if run > mrun {
				mrun = run
			}
			run = 0
			i = k
		}
	}
	return
}

var keySizes = []int{4, 8, 16, 32, 64, 512, 1024, 4096}

// WARNING: The benchmark run here is hardcoded (in two places)
// You must recompile to change the benchmark
func benchmark32s(n int) {
	//var hashes = make(Uint32Slice, n)
	const nbytes = 16384 // fix
	bs := make([]byte, nbytes, nbytes)
	bs = bs[:]
	//fp20 := make([]byte, 20, 20)
	Hf2 = "nullhash"
	for _, ksiz := range keySizes {
		if ksiz == 512 {
			n = n / 10
		}
		bs = bs[:ksiz]
		//hashes = make(Uint32Slice, n, n)
		//hashes = hashes[:]
		//fmt.Printf("ksiz=%d, len(bs)=%d\n", ksiz, len(bs))
		pn := hrff.Int64{int64(n), ""}
		ps := hrff.Int64{int64(n * ksiz), "B"}
		//fmt.Printf("benchmark32s: gen n=%d, n=%h, keySize=%d, size=%h\n", n, pn, ksiz, ps)
		if false {
			_ = gomap.Hash64(bs, 0)
			_ = aeshash.Hash(bs, 0)
		}
		start, stop := time.Now(), time.Now()
		switch ksiz {
		case 4:
			start = time.Now()
			for i := 0; i < n; i++ {
				_ = aeshash.Hash(bs, 0)
				// sha1160.Reset()
				// sha1160.Write(bs)
				// fp20 = fp20[0:0]
				// fp20 = sha1160.Sum(fp20)
				// _ = uint64(fp20[0])<<56 | uint64(fp20[1])<<48 | uint64(fp20[2])<<40 | uint64(fp20[3])<<32 |
				// 	uint64(fp20[4])<<24 | uint64(fp20[5])<<16 | uint64(fp20[6])<<8  | uint64(fp20[7])<<0
				//k224.Reset()
				//k224.Write(bs)
				//_ = k224.Sum(nil)
				//bs[0], bs[1], bs[2], bs[3] = byte(i), byte(i>>8), byte(i>>16), byte(i>>24)
				//_, _ = jenkins.Jenkins364(bs, 0, 0, 0)
				//_ = jenkins.Hash232(bs, 0)
				//_, _ = jenkins.Jenkins364(bs, 0, 0, 0)
				//_ = gomap.Hash64(bs, 0)
				//_ =  nhf64.Hash64S(bs, 0) // chnage below too
				//nh.Reset()
				//nh.Write(bs)
				//_ = nh.Sum64()
				//Hashf(bs, 0)
				//_ = siphash.Hash(0, 0, bs)
				//hashes[i] = h
				//fmt.Printf("i=%d, 0x%08x, h=0x%08x\n", i, i, h)
				//hashes[i] = h
				//fmt.Printf("i=%d, 0x%08x, h=0x%08x\n", i, i, h)
			}
			stop = time.Now()
		default:
			start = time.Now()
			for i := 0; i < n; i++ {
				// sha1160.Reset()
				// sha1160.Write(bs)
				// fp20 = fp20[0:0]
				// fp20 = sha1160.Sum(fp20)
				// _ = uint64(fp20[0])<<56 | uint64(fp20[1])<<48 | uint64(fp20[2])<<40 | uint64(fp20[3])<<32 |
				// 	uint64(fp20[4])<<24 | uint64(fp20[5])<<16 | uint64(fp20[6])<<8  | uint64(fp20[7])<<0
				_ = aeshash.Hash(bs, 0)
				//k224.Reset()
				//k224.Write(bs)
				//_ = k224.Sum(nil)
				//bs[0], bs[1], bs[2], bs[3], bs[4], bs[5], bs[6], bs[7] = byte(i), byte(i>>8), byte(i>>16), byte(i>>24), byte(i>>32), byte(i>>40), byte(i>>48), byte(i>>56)
				//_, _ = jenkins.Jenkins364(bs, 0, 0, 0)
				//_ = aeshash.Hash(bs, 0)
				//_ = gomap.Hash64(bs, 0)
				//_ = siphash.Hash(0, 0, bs)
				//_ =  nhf64.Hash64S(bs, 0)
				//nh.Reset()
				//nh.Write(bs)
				//_ = nh.Sum64()
				//Hashf(bs, 0)
			}
			stop = time.Now()
		}
		d := tdiff(start, stop)
		hsec := hrff.Float64{(float64(n) / d.Seconds()), "hashes/sec"}
		bsec := hrff.Float64{(float64(n) * float64(ksiz) / d.Seconds()), "B/sec"}
		//fmt.Printf("benchmark32s: %h\n", hsec)
		//fmt.Printf("benchmark32s: %.2h\n\n", bsec)
		fmt.Printf("\tksize=%d, n=%h, size=%h, %h, %.1h, time=%v\n", ksiz, pn, ps, hsec, bsec, d)
	}
	return

	fmt.Printf("benchmark32s: sort n=%d\n", n)
	//hashes.Sort()
	/*
		for i := 0; i < n; i++ {
			fmt.Printf("i=%d, 0x%08x, h=0x%08x\n", i, i, hashes[i])
		}
	*/
	fmt.Printf("benchmark32s: dup check n=%d\n", n)
	//dups, mrun := checkForDups32(hashes)
	//fmt.Printf("benchmark32: dups=%d, mrun=%d\n", dups, mrun)
}

func benchmark32g(h hash.Hash64, hf2 string, n int) {
	//var hashes Uint32Slice
	const nbytes = 16384 // fix

	Hf2 = hf2
	bs := make([]byte, nbytes, nbytes)
	bs = bs[:]
	for _, ksiz := range keySizes {
		if ksiz == 512 {
			n = n / 10
		}
		bs = bs[:ksiz]
		//hashes = make(Uint32Slice, n, n)
		//hashes = hashes[:]
		//fmt.Printf("ksiz=%d, len(bs)=%d\n", ksiz, len(bs))
		pn := hrff.Int64{int64(n), ""}
		ps := hrff.Int64{int64(n * ksiz), "B"}
		//fmt.Printf("benchmark32g: gen n=%d, n=%h, keySize=%d, size=%h\n", n, pn, ksiz, ps)
		start, stop := time.Now(), time.Now()
		if h != nil {
			switch ksiz {
			case 4:
				start = time.Now()
				for i := 0; i < n; i++ {
					bs[0], bs[1], bs[2], bs[3] = byte(i), byte(i>>8), byte(i>>16), byte(i>>24)
					h.Reset()
					h.Write(bs)
					_ = h.Sum64()
				}
				stop = time.Now()
			default:
				start = time.Now()
				for i := 0; i < n; i++ {
					bs[0], bs[1], bs[2], bs[3], bs[4], bs[5], bs[6], bs[7] = byte(i), byte(i>>8), byte(i>>16), byte(i>>24), byte(i>>32), byte(i>>40), byte(i>>48), byte(i>>56)
					h.Reset()
					h.Write(bs)
					_ = h.Sum64()
					//hashes[i] = h
					//fmt.Printf("i=%d, 0x%08x, h=0x%08x\n", i, i, h)
				}
				stop = time.Now()
			}
		} else {
			switch ksiz {
			case 4:
				start = time.Now()
				for i := 0; i < n; i++ {
					//bs[0], bs[1], bs[2], bs[3] = byte(i), byte(i>>8), byte(i>>16), byte(i>>24)
					Hashf(bs, 0) // the generic adapter is very inefficient, as much as 6X slower, however same for everyone
				}
				stop = time.Now()
			default:
				start = time.Now()
				for i := 0; i < n; i++ {
					//bs[0], bs[1], bs[2], bs[3], bs[4], bs[5], bs[6], bs[7] = byte(i), byte(i>>8), byte(i>>16), byte(i>>24), byte(i>>32), byte(i>>40), byte(i>>48), byte(i>>56)
					Hashf(bs, 0) // the generic adapter is very inefficient, as much as 6X slower, however same for everyone
					//hashes[i] = h
					//fmt.Printf("i=%d, 0x%08x, h=0x%08x\n", i, i, h)
				}
				stop = time.Now()
			}
		}
		d := tdiff(start, stop)
		hsec := hrff.Float64{(float64(n) / d.Seconds()), "hashes/sec"}
		bsec := hrff.Float64{(float64(n) * float64(ksiz) / d.Seconds()), "B/sec"}
		//fmt.Printf("benchmark32g: %h\n", hsec)
		//fmt.Printf("benchmark32g: %.2h\n\n", bsec)
		fmt.Printf("\tksize=%d, n=%h, size=%h, %h, %.1h, time=%v\n", ksiz, pn, ps, hsec, bsec, d)
	}

	if *cd {
		fmt.Printf("benchmark32g: sort n=%d\n", n)
		//hashes.Sort()
		/*
			for i := 0; i < n; i++ {
				fmt.Printf("i=%d, 0x%08x, h=0x%08x\n", i, i, hashes[i])
			}
		*/
		fmt.Printf("benchmark32g: dup check n=%d\n", n)
		//dups, mrun := checkForDups32(hashes)
		//fmt.Printf("benchmark32: dups=%d, mrun=%d\n", dups, mrun)
	}
}

var benchmarks = []string{"j332c", "j232", "sbox", "CrapWow"}

//var benchmarks = []string{"j332c"}

func benchmark(hashes []string, hf2 string, n int) {
	//for _, v := range hashes {
	//hf32 := Halloc(v)
	//fmt.Printf("benchmark32g: %q\n", v)
	benchmark32g(nil, hf2, n)
	fmt.Printf("\n")
	//}
}

type Test struct {
	name string
	flag **bool
	ptf  func(file string, lines int, hashf string) (ht *HashTable)
	desc string
}

var Tests = []Test{
	{"TestA", &A, TestA, "insert keys"},
	{"TestB", &B, TestB, "add newline to key"},
	{"TestC", &C, TestC, "add 4 newlines to key"},
	{"TestD", &D, TestD, "prepend ABCDE to key"},
	{"TestE", &E, TestE, "add 1 duplicate key"},
	{"TestF", &F, TestF, "add 3 duplicate keys"},
	{"TestG", &G, TestF, "reverse letter order in key"},
	{"TestH", &H, TestH, "words from letter combinations in wc"},
	{"TestI", &I, TestI, "integers from 0 to ni-1 (does not read file)"},
	{"TestJ", &J, TestJ, "one bit keys (does not read file)"},
	{"TestK", &K, TestK, "read file of keys and print hashes"},
}

func runTestsWithFileAndHashes(file string, lines int, hf []string) {
	var test Test
	if file != "" {
		if lines <= 0 {
			lines = ReadFile(file, nil)
		}
		//fmt.Printf("file=%q, lines=%d\n", file, lines)
		if *T0 {
			fmt.Printf("Test0 - ReadFile\n\t%20q: ", "ReadFile")
			ht := Test0(file, lines, "")
			ht.Print()
		}
	}
	for _, test = range Tests {
		if **test.flag {
			fmt.Printf("%s - %s\n", test.name, test.desc)
			for _, Hf2 = range hf {
				hi := HashFunctions[Hf2]
				if *c && !hi.Crypto {
					continue
				}
				if *h32 && hi.Size != 32 {
					continue
				}
				if *h64 && hi.Size != 64 {
					continue
				}
				fmt.Printf("\t%20q: ", Hf2)
				ht := test.ptf(file, lines, Hf2)
				ht.Print()
			}
		}
	}
}

var file = flag.String("file", "", "words to read")
var lines = flag.Int("lines", 0, "number of lines to read in file")
var hf = flag.String("hf", "all", "hash function")
var extra = flag.Int("e", 1, "extra bis in table size")
var prime = flag.Bool("p", false, "table size is primes and use mod")
var all = flag.Bool("a", false, "run all tests")
var pd = flag.Bool("pd", false, "print duplicate hashes")
var oa = flag.Bool("oa", false, "open addressing (no buckets)")

var c = flag.Bool("c", false, "only test crypto hash functions")
var h32 = flag.Bool("h32", false, "only test 32 bit has functions")
var h64 = flag.Bool("h64", false, "only test 64 bit has functions")

var b = flag.Bool("b", false, "run benchmarks")
var hcb = flag.Bool("hcb", false, "run hard coded benchmark")
var sm = flag.Bool("sm", false, "run SMHasher")
var v = flag.Bool("v", false, "verbose")
var cd = flag.Bool("cd", false, "check for duplicate hashs when running benchmarks")

//var wc = flags.String("wc", "abcdefgh, efghijkl, ijklmnop, mnopqrst, qrstuvwx, uvwxyz01", "letter combinations for word") // 262144 words)
var ni = flag.Int("ni", 200000, "number of integer keys")
var n = flag.Int("n", 10000000, "number of hashes for benchmark")
var ns = flag.Int("ns", 1, "number of seeds to test")

var T0 = flag.Bool("0", false, "test 0")
var A = flag.Bool("A", false, "test A")
var B = flag.Bool("B", false, "test B")
var C = flag.Bool("C", false, "test C")
var D = flag.Bool("D", false, "test D")
var E = flag.Bool("E", false, "test E")
var F = flag.Bool("F", false, "test F")
var G = flag.Bool("G", false, "test G")
var H = flag.Bool("H", false, "test H")
var I = flag.Bool("I", false, "test I")
var J = flag.Bool("J", false, "test J")
var K = flag.Bool("K", false, "test K")
var S = flag.Bool("S", false, "test S")

var letters = []string{"abcdefgh", "efghijkl", "ijklmnop", "mnopqrst", "qrstuvwx", "uvwxyz01"} // 262144 words
var TestPointers = []**bool{&A, &B, &C, &D, &E, &F, &G, &H, &I, &J, &K}

func allTestsOn() {
	*A, *B, *C, *D, *E, *F, *G, *H, *I, *J = true, true, true, true, true, true, true, true, true, true
}

func allTestsOff() {
	*A, *B, *C, *D, *E, *F, *G, *H, *I, *J = false, false, false, false, false, false, false, false, false, false
}

func main() {
	/*
		var cnt int
		var f = func(word string) {
			cnt++
			//fmt.Printf("%q\n", word)
		}

		test := []string{"ab", "cd"}
		test := []string{"abcdefgh", "efghijkl", "ijklmnop", "mnopqrst", "qrstuvwx", "uvwxyz01"} // 262144 words
		genWords(test, f)
		fmt.Printf("cnt=%d\n", cnt)
		return
	*/
	flag.Parse()
	if *all {
		allTestsOn()
	}
	if *hcb {
		*b = true
	}
	//fmt.Printf("%d lines read\n", lines)

	// read file and count lines
	// create table
	// read file and insert
	// stats

	switch {
	case *sm:
		if *hf == "all" {
			for _, Hf2 = range TestHashFunctions {
				hi := HashFunctions[Hf2]
				if *c && !hi.Crypto {
					continue
				}
				if *h32 && hi.Size != 32 {
					continue
				}
				if *h64 && hi.Size != 64 {
					continue
				}
				fmt.Printf("%q\n", Hf2)
				smhasher.RunSMhasher(Hf2, *v)
			}
		} else {
			Hf2 = *hf
			smhasher.RunSMhasher(Hf2, *v)
		}
		return
	case *b:
		fmt.Printf("\n")
		if *hf == "all" {
			for _, Hf2 = range TestHashFunctions {
				hi := HashFunctions[Hf2]
				if *c && !hi.Crypto {
					continue
				}
				if *h32 && hi.Size != 32 {
					continue
				}
				if *h64 && hi.Size != 64 {
					continue
				}
				fmt.Printf("%q\n", Hf2)
				benchmark(benchmarks, Hf2, *n)
			}
		} else {
			if *hcb { // this benchmark is hard coded
				fmt.Printf("%q\n", "hardcoded benchmark")
				benchmark32s(*n)
			} else {
				Hf2 = *hf
				fmt.Printf("%q\n", Hf2)
				benchmark32g(nil, *hf, *n)
			}
		}
		return
	case *file != "":
		if *hf == "all" {
			runTestsWithFileAndHashes(*file, *lines, TestHashFunctions)
		} else {
			Hf2 = *hf
			runTestsWithFileAndHashes(*file, *lines, []string{*hf})
		}
	case len(flag.Args()) != 0:
		for _, v := range flag.Args() {
			if *hf == "all" {
				runTestsWithFileAndHashes(v, *lines, TestHashFunctions)
			} else {
				Hf2 = *hf
				runTestsWithFileAndHashes(v, *lines, []string{*hf})
			}
		}
	case len(flag.Args()) == 0 && !*b:
		// no files specified run the only two tests we can with the specified hash functions
		allTestsOff()
		*I, *J = true, true
		if *hf == "all" {
			runTestsWithFileAndHashes("", *lines, TestHashFunctions)
		} else {
			Hf2 = *hf
			runTestsWithFileAndHashes("", *lines, []string{*hf})
		}
	}
}

var nhf64 nhash.HashF64
var nh hash.Hash64
var k224 hash.Hash
var k643 hash.Hash
var sha1160 hash.Hash

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "%s: [flags] [dictionary-files]\n", os.Args[0])
		flag.PrintDefaults()
	}

	bs := make([]byte, 4, 4)
	bs = bs[:]
	i := 0
	bs[0], bs[1], bs[2], bs[3] = byte(i), byte(i>>8), byte(i>>16), byte(i>>24)
	_, _ = jenkins.Jenkins364(bs, 0, 0, 0)
	_ = jenkins.Hash232(bs, 0)
	_, _ = jenkins.Jenkins364(bs, 0, 0, 0)
	_ = aeshash.Hash(bs, 0)
	_ = siphash.Hash(0, 0, bs)
	nh = nullhash.New()
	nhf64 = nullhash.NewF64()
	k224 = keccak.New224()
	k643 = keccakpg.NewCustom(64, 3)
	sha1160 = sha1.New()
}
