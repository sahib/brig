package store

import (
	"bytes"
	"io"
	"os"
	"testing"
)

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

	for _, size := range []int{1, 2, 3, 4, 8, 16, 32, 64} {
		tempDst := bytes.NewBuffer(nil)
		copyBuf := make([]byte, size)

		l := NewLayer(src)
		defer l.Close()

		if modifier != nil {
			if err := modifier(l); err != nil {
				t.Errorf("overlay-modifier failed: %v", err)
				t.FailNow()
			}
		}

		if n, err := l.Seek(0, os.SEEK_SET); err != nil || n != 0 {
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
				t.Errorf("")
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

var SINGLE_WRITES = map[string]struct {
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

			if _, err := l.Seek(10, os.SEEK_SET); err != nil {
				return err
			}

			if _, err := l.Write([]byte("??")); err != nil {
				return err
			}

			return nil
		},
	},
	"truncate": {
		[]byte("012345"),
		func(l *Layer) error {
			return nil
		},
	},
}

func TestOverlaySimple(t *testing.T) {
	for name, test := range SINGLE_WRITES {
		buf := createLayer(t, test.modifier)
		t.Log(buf.Bytes())

		if !bytes.Equal(test.want, buf.Bytes()) {
			t.Errorf("overlay-simple-write failed on `%s`.", name)
			t.Errorf("\tExpected: %v", test.want)
			t.Errorf("\tGot:      %v", buf.Bytes())
			return
		}
	}
}
