// Copyright © 2014 Lawrence E. Bakst. All rights reserved.
package jenkins_test

//import "flag"
import "fmt"
import "time"
import "unsafe"
//import "math"
//import "math/rand"
//import "runtime"
import "github.com/tildeleb/hashland/jenkins"
import "github.com/tildeleb/hrff"
import "testing"

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

func tdiff(begin, end time.Time) time.Duration {
    d := end.Sub(begin)
    return d
}

func TestCheck(t *testing.T) {
	jenkins.Check()
}

func TestBasic(t *testing.T) {

	q := "This is the time for all good men to come to the aid of their country..."
	//qq := []byte{"xThis is the time for all good men to come to the aid of their country..."}
	//qqq := []byte{"xxThis is the time for all good men to come to the aid of their country..."}
	//qqqq[] := []byte{"xxxThis is the time for all good men to come to the aid of their country..."}

	u := stu(q)
	fmt.Printf("len(q)=%d, len(u)=%d\n", len(q), len(u))
	h1 := jenkins.HashWordsLen(u, 13, 0)
	fmt.Printf("%08x, %0x8, %08x\n", h1, h1, h1)

	b, c := uint32(0), uint32(0)
	c, b = jenkins.HashString("", c, b)
	//fmt.Printf("%08x, %08x\n", c, b)
	if c != 0xdeadbeef || b != 0xdeadbeef {
		t.Logf("c=0x%x != 0xdeadbeef || b=0x%x != 0xdeadbeef\n", c, b)
		t.FailNow()
	}

	b, c = 0xdeadbeef, 0
	c, b = jenkins.HashString("", c, b)
	//fmt.Printf("%08x, %08x\n", c, b)	// bd5b7dde deadbeef
	if c != 0xbd5b7dde || b != 0xdeadbeef {
		t.Logf("c=0x%x != 0xbd5b7dde || b=0x%x != 0xdeadbeef\n", c, b)
		t.FailNow()
	}

  	b, c = 0xdeadbeef, 0xdeadbeef
	c, b = jenkins.HashString("", c, b)
	//fmt.Printf("%08x, %08x\n", c, b)	// 9c093ccd bd5b7dde
	if c != 0x9c093ccd || b != 0xbd5b7dde {
		t.Logf("c=0x%x != 0x9c093ccd || b=0x%x != 0xbd5b7dde\n", c, b)
		t.FailNow()
	}

	b, c = 0, 0
	c, b = jenkins.HashString("Four score and seven years ago", c, b)
	//fmt.Printf("%08x, %08x\n", c, b)	// 17770551 ce7226e6
	if c != 0x17770551 || b != 0xce7226e6 {
		t.Logf("c=0x%x != 0x17770551 || b=0x%x != 0xce7226e6\n", c, b)
		t.FailNow()
	}

	b, c = 1, 0
	c, b = jenkins.HashString("Four score and seven years ago", c, b)
	//fmt.Printf("%08x, %08x\n", c, b)	// e3607cae bd371de4
	if c != 0xe3607cae || b != 0xbd371de4 {
		t.Logf("c=0x%x != 0xe3607cae || b=0x%x != 0xbd371de4\n", c, b)
		t.FailNow()
	}

	b, c = 0, 1
	c, b = jenkins.HashString("Four score and seven years ago", c, b)
	//fmt.Printf("%08x, %08x\n", c, b)	// cd628161 6cbea4b3
	if c != 0xcd628161 || b != 0x6cbea4b3 {
		t.Logf("c=0x%x != 0xcd628161 || b=0x%x != 0x6cbea4b3\n", c, b)
		t.FailNow()
	}

}

func TestHash32v(t *testing.T) {
	k := make([]byte, 30, 30)
	seed := uint32(0)
	for i := 0; i < len(k); i++ {
		h := jenkins.Hash232(k[:i], seed)
		fmt.Printf("i=%03d, h=0x%08x, k=%v\n", i, h, k[:i]) // 0x8965bbe9
	}
}


func TestHash32(t *testing.T) {
	k := make([]byte, 4, 4)
	k = k[:]
	seed := uint32(0)
	h := jenkins.Hash232(k, seed)
	fmt.Printf("k=%v, h=0x%08x\n", k, h) // 0x8965bbe9
}

func TestHash64(t *testing.T) {
	k := make([]byte, 24, 24)
	k = k[:]
	seed := uint64(0)
	h := jenkins.Hash264(k, seed)
	fmt.Printf("k=%v, h=0x%016x\n", k, h) // 0x74882dd69da95fae
}

func BenchmarkJenkins332(b *testing.B) {
	//tmp := make([]byte, 4, 4)
	us := make([]uint32, 1)
	start := time.Now()
	b.SetBytes(int64(b.N * 4))
	for i := 1; i <= b.N; i++ {
		us[0] = uint32(i)
		//tmp[0], tmp[1], tmp[2], tmp[3] = byte(key&0xFF), byte((key>>8)&0xFF), byte((key>>16)&0xFF), byte((key>>24)&0xFF)
		jenkins.HashWords332(us, 0)
	}
	stop := time.Now()
	dur := tdiff(start, stop)
	hsec := hrff.Float64{(float64(b.N) / dur.Seconds()), "hashes/sec"}
	fmt.Printf("bench: %h\n", hsec)
	bsec := hrff.Float64{(float64(b.N) * 4 / dur.Seconds()), "B/sec"}
	fmt.Printf("bench: %h\n", bsec)
}

func BenchmarkJenkins232(b *testing.B) {
	bs := make([]byte, 4, 4)
	start := time.Now()
	b.SetBytes(int64(b.N * 4))
	for i := 1; i <= b.N; i++ {
		bs[0], bs[1], bs[2], bs[3] = byte(i)&0xFF, (byte(i)>>8)&0xFF, (byte(i)>>16)&0xFF, (byte(i)>>24)&0xFF
		//tmp[0], tmp[1], tmp[2], tmp[3] = byte(key&0xFF), byte((key>>8)&0xFF), byte((key>>16)&0xFF), byte((key>>24)&0xFF)
		jenkins.Hash232(bs, 0)
	}
	stop := time.Now()
	dur := tdiff(start, stop)
	hsec := hrff.Float64{(float64(b.N) / dur.Seconds()), "hashes/sec"}
	fmt.Printf("bench: %h\n", hsec)
	bsec := hrff.Float64{(float64(b.N) * 4 / dur.Seconds()), "B/sec"}
	fmt.Printf("bench: %h\n", bsec)
}

func BenchmarkJenkins264Bytes4(b *testing.B) {
	bs := make([]byte, 4, 4)
	start := time.Now()
	b.SetBytes(int64(b.N * 4))
	for i := 1; i <= b.N; i++ {
		bs[0], bs[1], bs[2], bs[3] = byte(i)&0xFF, (byte(i)>>8)&0xFF, (byte(i)>>16)&0xFF, (byte(i)>>24)&0xFF
		//tmp[0], tmp[1], tmp[2], tmp[3] = byte(key&0xFF), byte((key>>8)&0xFF), byte((key>>16)&0xFF), byte((key>>24)&0xFF)
		jenkins.Hash264(bs, 0)
	}
	stop := time.Now()
	dur := tdiff(start, stop)
	hsec := hrff.Float64{(float64(b.N) / dur.Seconds()), "hashes/sec"}
	fmt.Printf("bench: %h\n", hsec)
	bsec := hrff.Float64{(float64(b.N) * 4 / dur.Seconds()), "B/sec"}
	fmt.Printf("bench: %h\n", bsec)
}

func BenchmarkJenkins264Bytes24(b *testing.B) {
	bs := make([]byte, 24, 24)
	start := time.Now()
	b.SetBytes(int64(b.N * 4))
	for i := 1; i <= b.N; i++ {
		bs[0], bs[1], bs[2], bs[3] = byte(i)&0xFF, (byte(i)>>8)&0xFF, (byte(i)>>16)&0xFF, (byte(i)>>24)&0xFF
		bs[4], bs[5], bs[6], bs[7] = bs[0], bs[1], bs[2], bs[3]
		bs[8], bs[9], bs[10], bs[11], bs[12], bs[13], bs[14], bs[15] = bs[0], bs[1], bs[2], bs[3], bs[4], bs[5], bs[6], bs[7]
		bs[16], bs[17], bs[18], bs[19], bs[20], bs[21], bs[22], bs[23] = bs[0], bs[1], bs[2], bs[3], bs[4], bs[5], bs[6], bs[7]
		jenkins.Hash264(bs, 0)
	}
	stop := time.Now()
	dur := tdiff(start, stop)
	hsec := hrff.Float64{(float64(b.N) / dur.Seconds()), "hashes/sec"}
	fmt.Printf("bench: %h\n", hsec)
	bsec := hrff.Float64{(float64(b.N) * 4 / dur.Seconds()), "B/sec"}
	fmt.Printf("bench: %h\n", bsec)
}

/*
func main() {
	q := "This is the time for all good men to come to the aid of their country..."
	//qq := []byte{"xThis is the time for all good men to come to the aid of their country..."}
	//qqq := []byte{"xxThis is the time for all good men to come to the aid of their country..."}
	//qqqq[] := []byte{"xxxThis is the time for all good men to come to the aid of their country..."}

	u := stu(q)
	h1 := hashword(u, (len(q)-1)/4, 13)
	h2 := hashword(u, (len(q)-5)/4, 13)
	h3 := hashword(u, (len(q)-9)/4, 13)
	fmt.Printf("%08x, %0x8, %08x\n", h1, h2, h3)


}
*/