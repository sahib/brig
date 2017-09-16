package mio

import (
	"bytes"
	"io"
	"io/ioutil"
	"testing"

	"github.com/disorganizer/brig/store/compress"
	"github.com/disorganizer/brig/util/testutil"
)

var TestKey = []byte("01234567890ABCDE01234567890ABCDE")

type wrapReader struct {
	io.Reader
	io.Seeker
	io.Closer
	io.WriterTo
}

func testWriteAndRead(t *testing.T, raw []byte, algoType compress.AlgorithmType) {
	rawBuf := &bytes.Buffer{}
	if _, err := rawBuf.Write(raw); err != nil {
		t.Errorf("Huh, buf-write failed?")
		return
	}

	encStream, err := NewFileReader(TestKey, rawBuf, algoType)
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
	sizes := []int64{
		0, 1, 10, s64k, s64k - 1, s64k + 1,
		s64k * 2, s64k * 1024,
	}

	for _, size := range sizes {
		t.Logf("Testing stream at size %d", size)
		regularData := testutil.CreateDummyBuf(size)
		randomData := testutil.CreateRandomDummyBuf(size, 42)

		for algoType, _ := range compress.AlgoMap {
			testWriteAndRead(t, regularData, algoType)
			testWriteAndRead(t, randomData, algoType)
		}
	}
}
