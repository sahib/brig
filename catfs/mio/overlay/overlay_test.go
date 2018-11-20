package overlay

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/sahib/brig/catfs/mio/encrypt"
	"github.com/sahib/brig/util/testutil"
	"github.com/stretchr/testify/require"
)

var TestKey = []byte("01234567890ABCDE01234567890ABCDE")

func makeMod(off, size int64) *Modification {
	s := make([]byte, size)
	for i := int64(0); i < size; i++ {
		s[i] = byte(off + i)
	}

	return &Modification{off, s}
}

func TestMerge(t *testing.T) {
	i := &IntervalIndex{}
	i.Add(makeMod(0, 10))
	i.Add(makeMod(15, 5))
	i.Add(makeMod(10, 5))
	i.Add(makeMod(9, 8))
	i.Add(makeMod(90, 8))

	check := func(m *Modification, lo, hi int64) bool {
		if int64(m.data[0]) != m.offset {
			t.Errorf("Offset and first element do not match.")
			t.Errorf("Off: %v Data: %v", m.offset, m.data)
			return false
		}

		for i := lo; i < hi; i++ {
			if int64(m.data[i-lo]) != i {
				t.Errorf("Merge hickup: %v != %v", m.data[i-lo], i)
				return false
			}
		}

		return true
	}

	// First three intervals should be merged to one:
	if !check(i.r[0].(*Modification), 0, 20) {
		return
	}

	// Last one should be totally untouched:
	if !check(i.r[1].(*Modification), 90, 98) {
		return
	}
}

var SOURCE = []byte("0123456789")

func createLayer(t *testing.T, modifier func(l *Layer) error) *bytes.Buffer {
	src := bytes.NewReader(SOURCE)
	var dst *bytes.Buffer

	for _, size := range []int{ /*1, 2, 3, 4, 8, 16, 32,*/ 64} {
		tempDst := bytes.NewBuffer(nil)
		copyBuf := make([]byte, size)

		l := NewLayer(src)
		defer func() {
			if err := l.Close(); err != nil {
				t.Errorf("close(layer) failed: %v", err)
			}
		}()

		if modifier != nil {
			if err := modifier(l); err != nil {
				t.Errorf("overlay-modifier failed: %v", err)
				t.FailNow()
			}
		}

		if n, err := l.Seek(0, io.SeekStart); err != nil || n != 0 {
			t.Errorf("overlay-seek failed: %v (offset %v)", err, n)
			t.FailNow()
		}

		if _, err := io.CopyBuffer(tempDst, l, copyBuf); err != nil {
			t.Errorf("overlay: copy failed: %v", err)
			return nil
		}

		if dst != nil {
			// Changing the buffer size should not yield a different result:
			if !bytes.Equal(dst.Bytes(), tempDst.Bytes()) {
				t.Errorf("Different result with different buf size (size: %v)", size)
				t.Errorf("\tOLD: %x", dst.Bytes())
				t.Errorf("\tNEW: %x", tempDst.Bytes())
			}
		}

		dst = tempDst
	}

	return dst
}

func TestOverlayClean(t *testing.T) {
	buf := createLayer(t, nil)
	if !bytes.Equal(buf.Bytes(), SOURCE) {
		t.Errorf("overlay-simple: Expected enumerated values; got %v", buf.Bytes())
	}
}

var SingleWrites = map[string]struct {
	want     []byte
	modifier func(l *Layer) error
}{
	"no-modification": {
		SOURCE,
		nil,
	},
	"empty": {
		SOURCE,
		func(l *Layer) error {
			_, err := l.Write([]byte{})
			return err
		},
	},
	"onebyte": {
		SOURCE,
		func(l *Layer) error {
			_, err := l.Write([]byte{'0'})
			return err
		},
	},
	"onebytediff": {
		[]byte("1123456789"),
		func(l *Layer) error {
			_, err := l.Write([]byte{'1'})
			return err
		},
	},
	"extend": {
		[]byte("9876543210!!"),
		func(l *Layer) error {
			_, err := l.Write([]byte("9876543210!!"))
			return err
		},
	},
	"extend-gap": {
		[]byte("!!23456789??"),
		func(l *Layer) error {
			if _, err := l.Write([]byte("!!")); err != nil {
				return err
			}

			if _, err := l.Seek(10, io.SeekStart); err != nil {
				return err
			}

			if _, err := l.Write([]byte("??")); err != nil {
				return err
			}

			return nil
		},
	},
	"truncate": {
		[]byte("01234"),
		func(l *Layer) error {
			l.Truncate(5)

			if n := l.Limit(); n != 5 {
				return fmt.Errorf("Truncate() did not cut to 5, but to %v", n)
			}
			return nil
		},
	},
	"truncate-then-write": {
		[]byte("0123498765"),
		func(l *Layer) error {
			l.Truncate(0)
			if n, err := l.Seek(5, io.SeekStart); err != nil || n != 5 {
				return fmt.Errorf("Seek() did not work after Truncate(): %v (off: %v)", err, n)
			}

			if n, err := l.Write([]byte("98765")); err != nil || n != 5 {
				return fmt.Errorf("Write errored or short write after truncate: %v (bytes: %v)", err, n)
			}

			return nil
		},
	},
	"truncate-seek": {
		[]byte("01234"),
		func(l *Layer) error {
			l.Truncate(5)
			n, err := l.Seek(5, io.SeekStart)
			if err != nil {
				return fmt.Errorf("Seek to end failed: %v", err)
			}

			if n != 5 {
				return fmt.Errorf("Seek tells the wrong position: 5 != %d", n)
			}

			b := make([]byte, 10)
			if n, err := l.Read(b); n > 0 || err != io.EOF {
				return fmt.Errorf("Read delivers data over limit (%d bytes): %v", n, err)
			}

			return nil
		},
	},
}

func TestOverlaySimple(t *testing.T) {
	for name, test := range SingleWrites {
		buf := createLayer(t, test.modifier)

		if !bytes.Equal(test.want, buf.Bytes()) {
			t.Errorf("overlay-simple-write failed on `%s`.", name)
			t.Errorf("\tExpected: %v", test.want)
			t.Errorf("\tGot:      %v", buf.Bytes())
			return
		}
	}
}

func TestBigFile(t *testing.T) {
	src := testutil.CreateDummyBuf(147611)
	dst := &bytes.Buffer{}

	srcEnc := &bytes.Buffer{}
	wEnc, err := encrypt.NewWriter(srcEnc, TestKey)
	if err != nil {
		t.Errorf("Cannot create write-encryption layer: %v", err)
		return
	}

	if err := wEnc.Close(); err != nil {
		t.Errorf("Cannot close write-encryption layer: %v", err)
		return
	}

	wDec, err := encrypt.NewReader(bytes.NewReader(srcEnc.Bytes()), TestKey)
	if err != nil {
		t.Errorf("Cannot create read-encryption layer: %v", err)
		return
	}

	// Act a bit like the fuse layer:
	lay := NewLayer(wDec)
	lay.Truncate(0)

	bufSize := 128 * 1024
	if _, err := io.CopyBuffer(lay, bytes.NewReader(src), make([]byte, bufSize)); err != nil {
		t.Errorf("Could not encrypt data")
		return
	}

	lay.Truncate(int64(len(src)))

	if _, err := lay.Seek(0, io.SeekStart); err != nil {
		t.Errorf("Seeking to 0 in big file failed: %v", err)
		return
	}

	n, err := io.CopyBuffer(dst, lay, make([]byte, bufSize))
	if err != nil {
		t.Errorf("Could not copy big file data over overlay: %v", err)
		return
	}

	if n != int64(len(src)) {
		t.Errorf("Did not fully copy big file: got %d, should be %d bytes", n, len(src))
		return
	}

	if !bytes.Equal(dst.Bytes(), src) {
		t.Errorf("Source and destination buffers differ.")
		return
	}
}

func TestHasGaps(t *testing.T) {
	// gap between:
	require.True(t, hasGaps(
		[]Interval{
			&Modification{20, make([]byte, 20)},
			&Modification{41, make([]byte, 20)},
		},
		20,
		61,
	))

	// overlay *after* actual range:
	require.True(t, hasGaps(
		[]Interval{
			&Modification{20, make([]byte, 20)},
		},
		40,
		60,
	))

	// overlay *before* actual range:
	require.True(t, hasGaps(
		[]Interval{
			&Modification{20, make([]byte, 20)},
		},
		0,
		20,
	))

	// Fully occluded by single:
	require.False(t, hasGaps(
		[]Interval{
			&Modification{20, make([]byte, 20)},
		},
		20,
		40,
	))

	// Fully occluded by two:
	require.False(t, hasGaps(
		[]Interval{
			&Modification{20, make([]byte, 10)},
			&Modification{30, make([]byte, 10)},
		},
		20,
		40,
	))

	// Overlap at the beginning:
	require.False(t, hasGaps(
		[]Interval{
			&Modification{10, make([]byte, 20)},
			&Modification{30, make([]byte, 20)},
		},
		20,
		40,
	))

	// Small gap with overlaps before & after:
	require.True(t, hasGaps(
		[]Interval{
			&Modification{10, make([]byte, 15)},
			&Modification{35, make([]byte, 20)},
		},
		20,
		40,
	))
}
