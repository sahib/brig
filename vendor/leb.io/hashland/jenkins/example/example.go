// Copyright © 2014 Lawrence E. Bakst. All rights reserved.
package main

import "flag"
import "fmt"
import "unsafe"
import "time"
import "sort"
import "os"
import "log"
import "runtime/pprof"
import "leb.io/hashland/jenkins"
import "leb.io/hrff"

//import "math/rand"
//import "runtime"

func stu(s string) []uint32 {
	//fmt.Printf("stu: s=%q\n", s)
	l := (len(s) + 3) / 4
	d := make([]uint32, l, l)
	d = d[0:0]
	b := ([]byte)(s)
	//fmt.Printf("b=%x\n", b)
	for i := 0; i < l; i++ {
		t := *(*uint32)(unsafe.Pointer(&b[i*4]))
		//fmt.Printf("t=%x \n", t)
		d = append(d, t)
	}
	//fmt.Printf("stu: len(s)=%d, len(d)=%d, d=%x\n", len(s), len(d), d)
	return d
}

/* check that every input bit changes every output bit half the time */

const (
	HASHSTATE = 1
	HASHLEN   = 1
	MAXPAIR   = 6022
	MAXLEN    = 70
)

func driver2() {
	var qa [MAXLEN + 1]byte
	var qb [MAXLEN + 2]byte
	// *a = &qa[0], *b = &qb[1];   uint8_t
	var c [HASHSTATE]uint32
	var d [HASHSTATE]uint32
	var e [HASHSTATE]uint32
	var f [HASHSTATE]uint32
	var g [HASHSTATE]uint32
	var h [HASHSTATE]uint32
	var x [HASHSTATE]uint32
	var y [HASHSTATE]uint32

	a := qa[0:MAXLEN]
	b := qb[1 : MAXLEN+1]

	//i=0, j=0, k, l, m=0, z;
	// uint32_t hlen;

	fmt.Printf("No more than %d trials should ever be needed \n", MAXPAIR/2)
	for hlen := 0; hlen < MAXLEN; hlen++ {
		z := uint32(0)
		i, m := z, z
		for i = uint32(0); i < uint32(hlen); i++ { // for each input byte,
			for j := uint32(0); j < 8; j++ { // for each input bit,
				for m = uint32(1); m < 8; m++ { // for serveral possible initvals
					for l := 0; l < HASHSTATE; l++ {
						e[l] = ^(uint32(0))
						f[l] = e[l]
						g[l] = e[l]
						h[l] = e[l]
						x[l] = e[l]
						y[l] = e[l]
					}

					// check that every output bit is affected by that input bit
					k := uint32(0)
					for k = 0; k < MAXPAIR; k += 2 {
						finished := true
						/* keys have one bit different */
						for l := 0; l < hlen+1; l++ {
							a[l], b[l] = byte(0), byte(0)
						}
						/* have a and b be two keys differing in only one bit */
						a[i] ^= byte(k << j)
						a[i] ^= byte(k >> (8 - j))
						c[0] = jenkins.HashBytesLength(a, hlen, m)

						b[i] ^= byte((k + 1) << j)
						b[i] ^= byte((k + 1) >> (8 - j))
						d[0] = jenkins.HashBytesLength(b, hlen, m)
						// check every bit is 1, 0, set, and not set at least once
						for l := 0; l < HASHSTATE; l++ {
							e[l] &= (c[l] ^ d[l])
							f[l] &= ^(c[l] ^ d[l])
							g[l] &= c[l]
							h[l] &= ^c[l]
							x[l] &= d[l]
							y[l] &= ^d[l]
							if e[l]|f[l]|g[l]|h[l]|x[l]|y[l] != 0 {
								finished = false
							}
						}
						if finished {
							break
						}
					}
					if k > z {
						z = k
					}
					if k == MAXPAIR {
						fmt.Printf("Some bit didn't change: ")
						fmt.Printf("%.8x %.8x %.8x %.8x %.8x %.8x  ", e[0], f[0], g[0], h[0], x[0], y[0])
						fmt.Printf("i %d j %d m %d len %d\n", i, j, m, hlen)
					}
					if z == MAXPAIR {
						goto done
					}
				}
			}
		}
	done:
		if z < MAXPAIR {
			fmt.Printf("Mix success  %2d bytes  %2d initvals  ", i, m)
			fmt.Printf("required  %d  trials\n", z/2)
		}
	}
	fmt.Printf("\n")
}

// IntSlice attaches the methods of Interface to []int, sorting in increasing order.
type Uint32Slice []uint32

func (p Uint32Slice) Len() int           { return len(p) }
func (p Uint32Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Uint32Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// Sort is a convenience method.
func (p Uint32Slice) Sort() { sort.Sort(p) }

// IntSlice attaches the methods of Interface to []int, sorting in increasing order.
type Uint64Slice []uint64

func (p Uint64Slice) Len() int           { return len(p) }
func (p Uint64Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Uint64Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// Sort is a convenience method.
func (p Uint64Slice) Sort() { sort.Sort(p) }

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

func checkForDups64(u Uint64Slice) (dups int) {
	i := 0
	for k, v := range u {
		if k == 0 || i == k {
			continue
		}
		if u[i] == v {
			dups++
			continue
		} else {
			i = k
		}
	}
	return
}

func tdiff(begin, end time.Time) time.Duration {
	d := end.Sub(begin)
	return d
}

func benchmark64(n int64) {
	var hashes = make(Uint64Slice, n)
	bs := make([]byte, 24, 24)

	fmt.Printf("benchmark64: gen n=%d, n=%h, size=%H\n", n, hrff.Int64{n, ""}, hrff.Int64{n * 8, "B"})
	start := time.Now()
	for i := int64(0); i < n; i++ {
		bs[0], bs[1], bs[2], bs[3] = byte(i&0xFF), byte((i>>8)&0xFF), byte((i>>16)&0xFF), byte((i>>24)&0xFF)
		bs[4], bs[5], bs[6], bs[7] = bs[0], bs[1], bs[2], bs[3]
		bs[8], bs[9], bs[10], bs[11], bs[12], bs[13], bs[14], bs[15] = bs[0], bs[1], bs[2], bs[3], bs[4], bs[5], bs[6], bs[7]
		bs[16], bs[17], bs[18], bs[19], bs[20], bs[21], bs[22], bs[23] = bs[0], bs[1], bs[2], bs[3], bs[4], bs[5], bs[6], bs[7]
		h := jenkins.Hash264(bs, 0)
		hashes[i] = h
		//fmt.Printf("i=%d, 0x%08x, h=0x%016x\n", i, i, h)
	}
	stop := time.Now()
	d := tdiff(start, stop)
	hsec := hrff.Float64{(float64(n) / d.Seconds()), "hashes/sec"}
	bsec := hrff.Float64{(float64(n) * float64(24) / d.Seconds()), "B/sec"}
	fmt.Printf("benchmark64: %h\n", hsec)
	fmt.Printf("benchmark64: %h\n", bsec)
	fmt.Printf("benchmark64: sort n=%d\n", n)
	hashes.Sort()

	if false {
		for i := int64(0); i < n; i++ {
			fmt.Printf("i=%d, 0x%08x, h=0x%08x\n", i, i, hashes[i])
		}
	}

	fmt.Printf("benchmark64: dup check n=%d\n", n)
	dups := checkForDups64(hashes)
	fmt.Printf("benchmark64: dups=%d\n", dups)
}

//dr := hrff.Float64{float64(tlen) / d.Seconds(), "B/sec"}
//ns := float64(psum.Recvt.Sub(p.Packets[0][i].Sendt))
//ms := ns / 1e6

func benchmark32(n int) {
	//var hashes = make(Uint32Slice, n)
	//var u = make([]uint32, 1, 1)
	bs := make([]byte, 4, 4)
	var pn = hrff.Int64{int64(n), ""}
	var ps = hrff.Int64{int64(n * 4), "B"}
	fmt.Printf("benchmark32: gen n=%d, n=%h, size=%h\n", n, pn, ps)
	start := time.Now()
	for i := 0; i < n; i++ {
		bs[0], bs[1], bs[2], bs[3] = byte(i)&0xFF, (byte(i)>>8)&0xFF, (byte(i)>>16)&0xFF, (byte(i)>>24)&0xFF
		_ = jenkins.Hash232(bs, 0)
		//hashes[i] = h
		//fmt.Printf("i=%d, 0x%08x, h=0x%08x\n", i, i, h)
	}
	stop := time.Now()
	d := tdiff(start, stop)
	hsec := hrff.Float64{(float64(n) / d.Seconds()), "hashes/sec"}
	bsec := hrff.Float64{(float64(n) * 4 / d.Seconds()), "B/sec"}
	fmt.Printf("benchmark32: %h\n", hsec)
	fmt.Printf("benchmark32: %h\n", bsec)
	return

	fmt.Printf("benchmark32: sort n=%d\n", n)
	//hashes.Sort()
	/*
		for i := 0; i < n; i++ {
			fmt.Printf("i=%d, 0x%08x, h=0x%08x\n", i, i, hashes[i])
		}
	*/
	fmt.Printf("benchmark32: dup check n=%d\n", n)
	//dups, mrun := checkForDups32(hashes)
	//fmt.Printf("benchmark32: dups=%d, mrun=%d\n", dups, mrun)
}

var n = flag.Int("n", 5, "number of hashes")
var p = flag.String("p", "", "write cpu profile to file")

/*
func ShortTest(n int) {
	var u = make([]uint32, 1, 1)

	for i := 0; i < n; i++ {
		u[0] = uint32(i)
		h := jenkins3.HashWords32(u, 0)
		fmt.Printf("i=%d, 0x%08x, h=0x%08x\n", i, i, h)
	}
}
*/

func main() {
	flag.Parse()
	if *p != "" {
		f, err := os.Create(*p)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	//ShortTest(*n)
	//return
	benchmark32(*n)
	benchmark64(int64(*n))

	q := "This is the time for all good men to come to the aid of their country..."
	//qq := []byte{"xThis is the time for all good men to come to the aid of their country..."}
	//qqq := []byte{"xxThis is the time for all good men to come to the aid of their country..."}
	//qqqq[] := []byte{"xxxThis is the time for all good men to come to the aid of their country..."}

	u := stu(q)
	h1 := jenkins.HashWordsLen(u, (len(q)-1)/4, 13)
	h2 := jenkins.HashWordsLen(u, (len(q)-5)/4, 13)
	h3 := jenkins.HashWordsLen(u, (len(q)-9)/4, 13)
	fmt.Printf("%08x, %08x, %08x\n", h1, h2, h3)

	b, c := uint32(0), uint32(0)
	c, b = jenkins.HashString("", c, b)
	fmt.Printf("%08x, %08x\n", c, b) // deadbeef deadbeef

	b, c = 0xdeadbeef, 0
	c, b = jenkins.HashString("", c, b)
	fmt.Printf("%08x, %08x\n", c, b) // bd5b7dde deadbeef

	b, c = 0xdeadbeef, 0xdeadbeef
	c, b = jenkins.HashString("", c, b)
	fmt.Printf("%08x, %08x\n", c, b) // 9c093ccd bd5b7dde

	b, c = 0, 0
	c, b = jenkins.HashString("Four score and seven years ago", c, b)
	fmt.Printf("%08x, %08x\n", c, b) // 17770551 ce7226e6

	b, c = 1, 0
	c, b = jenkins.HashString("Four score and seven years ago", c, b)
	fmt.Printf("%08x, %08x\n", c, b) // e3607cae bd371de4

	b, c = 0, 1
	c, b = jenkins.HashString("Four score and seven years ago", c, b)
	fmt.Printf("%08x, %08x\n", c, b) // cd628161 6cbea4b3

	driver2()
}
