package store

import (
	"bytes"
	"io"
	"io/ioutil"
	"testing"

	"github.com/disorganizer/brig/util/testutil"
)

var TestKey = []byte("01234567890ABCDE01234567890ABCDE")

type wrapReader struct {
	io.Reader
	io.Seeker
	io.Closer
	io.WriterTo
}

func testWriteAndRead(size int64, t *testing.T) {
	raw := testutil.CreateDummyBuf(size)
	rawBuf := &bytes.Buffer{}
	if _, err := rawBuf.Write(raw); err != nil {
		t.Errorf("Huh, buf-write failed?")
		return
	}

	encStream, err := NewFileReader(TestKey, rawBuf)
	if err != nil {
		t.Errorf("Creating encryption stream failed: %v", err)
		return
	}

	encrypted := &bytes.Buffer{}
	if _, err = io.Copy(encrypted, encStream); err != nil {
		t.Errorf("Reading encrypted data failed: %v", err)
		return
	}

	// Fake a close method:
	br := bytes.NewReader(encrypted.Bytes())

	r := wrapReader{
		Reader:   br,
		Seeker:   br,
		WriterTo: br,
		Closer:   ioutil.NopCloser(nil),
	}

	decStream, err := NewIpfsReader(TestKey, r)
	if err != nil {
		t.Errorf("Creating decryption stream failed: %v", err)
		return
	}

	decrypted := &bytes.Buffer{}
	if _, err = io.Copy(decrypted, decStream); err != nil {
		t.Errorf("Reading decrypted data failed: %v", err)
		return
	}

	if !bytes.Equal(decrypted.Bytes(), raw) {
		t.Errorf("Raw and decrypted is not equal => BUG.")
		t.Errorf("RAW:\n  %v", raw)
		t.Errorf("DEC:\n  %v", decrypted.Bytes())
		return
	}
}

func TestWriteAndRead(t *testing.T) {
	s64k := int64(64 * 1024)
	for _, size := range []int64{0, 1, 10, s64k, s64k - 1, s64k + 1, s64k * 2} {
		t.Logf("Testing stream at size %d", size)
		testWriteAndRead(size, t)
	}
}
