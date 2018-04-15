package nullhash

import (
	"hash"
	"leb.io/hashland/nhash"
)

func Nullhash(b []byte, seed uint64) uint64 {
	return 0
}

type nullhashstate struct {
	ctr uint64
}

func (n *nullhashstate) Size() int {
	return 8
}

func (n *nullhashstate) BlockSize() int {
	return 8
}

func (n *nullhashstate) NumSeedBytes() int {
	return 8
}

func (n *nullhashstate) Write(b []byte) (int, error) {
	return len(b), nil
}

func (n *nullhashstate) Sum(b []byte) []byte {

	buf := make([]byte, 8, 8)
	buf = buf[:]
	return append(b, buf...)
}

func (n *nullhashstate) Sum64() uint64 {
	n.ctr++
	return n.ctr
}

func (n *nullhashstate) Reset() {
}

func (n *nullhashstate) Hash64(b []byte, seeds ...uint64) uint64 {
	n.ctr++
	return n.ctr
}

func (n *nullhashstate) Hash64S(b []byte, seed uint64) uint64 {
	n.ctr++
	return n.ctr
}

func New() hash.Hash64 {
	var n nullhashstate
	return &n
}

func NewF64() nhash.HashF64 {
	var n nullhashstate
	return &n
}

/*
// initial results are even worse than I thought
// nullhash: 1.135146477s
// Hash64 with seed: 59.690928864s
// Hash64 no seed: 3.684813561s

const n = 1000000000
var b = make([]byte, 100, 100)
var seed uint64
var intf = New()

func b1() time.Duration {
	start := time.Now()
	for i:= 0; i < n; i++ {
		nullhash(b, seed)
	}
	stop := time.Now()
	return tdiff(start, stop)
}

func b2() time.Duration {
	start := time.Now()
	for i:= 0; i < n; i++ {
		intf.Hash64(b, seed)
	}
	stop := time.Now()
	return tdiff(start, stop)
}

func b3() time.Duration {
	start := time.Now()
	for i := 0; i < n; i++ {
		intf.Hash64(b)
	}
	stop := time.Now()
	return tdiff(start, stop)
}

func main() {
	fmt.Printf("nullhash: ")
	t1 := b1()
	fmt.Printf("%v\n", t1)
	fmt.Printf("Hash64 with seed: ")
	t2 := b2()
	fmt.Printf("%v\n", t2)
	fmt.Printf("Hash64 no seed: ")
	t3 := b3()
	fmt.Printf("%v\n", t3)
}
*/
