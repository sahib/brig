package encrypt

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/disorganizer/brig/util/testutil"
)

var TestKey = []byte("01234567890ABCDE01234567890ABCDE")

func openFiles(from, to string) (*os.File, *os.File, error) {
	fdFrom, err := os.Open(from)
	if err != nil {
		return nil, nil, err
	}

	fdTo, err := os.OpenFile(to, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		fdFrom.Close()
		return nil, nil, err
	}

	return fdFrom, fdTo, nil
}

func encryptFile(key []byte, from, to string) (n int64, outErr error) {
	fdFrom, fdTo, err := openFiles(from, to)
	if err != nil {
		return 0, err
	}

	defer func() {
		// Only fdTo needs to be closed, Decrypt closes fdFrom.
		if err := fdFrom.Close(); err != nil {
			outErr = err
		}
		if err := fdTo.Close(); err != nil {
			outErr = err
		}
	}()

	info, err := fdFrom.Stat()
	if err != nil {
		return 0, err
	}

	return Encrypt(key, fdFrom, fdTo, info.Size())
}

func decryptFile(key []byte, from, to string) (n int64, outErr error) {
	fdFrom, fdTo, err := openFiles(from, to)
	if err != nil {
		return 0, err
	}

	defer func() {
		// Only fdTo needs to be closed, Decrypt closes fdFrom.
		if err := fdTo.Close(); err != nil {
			outErr = err
		}
	}()

	return Decrypt(key, fdFrom, fdTo)
}

func remover(t *testing.T, path string) {
	if err := os.Remove(path); err != nil {
		t.Errorf("Could not remove temp file: %v", err)
	}
}

func testSimpleEncDec(t *testing.T, size int) {
	path := testutil.CreateFile(int64(size))
	defer remover(t, path)

	encPath := path + "_enc"
	decPath := path + "_dec"

	_, err := encryptFile(TestKey, path, encPath)
	defer remover(t, encPath)

	if err != nil {
		t.Errorf("Encrypt failed: %v", err)
	}

	_, err = decryptFile(TestKey, encPath, decPath)
	defer remover(t, decPath)

	if err != nil {
		t.Errorf("Decrypt failed: %v", err)
	}

	a, _ := ioutil.ReadFile(path)
	b, _ := ioutil.ReadFile(decPath)
	c, _ := ioutil.ReadFile(encPath)

	if !bytes.Equal(a, b) {
		t.Errorf("Source and decrypted not equal")
	}

	if bytes.Equal(a, c) && size != 0 {
		t.Errorf("Source was not encrypted (same as source)")
		t.Errorf("%v|||%v|||%v", a, b, c)
	}
}

func TestSimpleEncDec(t *testing.T) {
	t.Parallel()

	sizes := []int{
		0,
		1,
		MaxBlockSize - 1,
		MaxBlockSize,
		MaxBlockSize + 1,
		GoodDecBufferSize - 1,
		GoodDecBufferSize,
		GoodDecBufferSize + 1,
		GoodEncBufferSize - 1,
		GoodEncBufferSize,
		GoodEncBufferSize + 1,
	}

	for size := range sizes {
		t.Logf("Testing SimpleEncDec for size %d", size)
		testSimpleEncDec(t, size)
	}
}

func TestSeek(t *testing.T) {
	N := int64(2 * MaxBlockSize)
	a := testutil.CreateDummyBuf(N)
	b := make([]byte, 0, N)

	source := bytes.NewBuffer(a)
	shared := &bytes.Buffer{}
	dest := bytes.NewBuffer(b)

	enc, err := NewWriter(shared, TestKey, N)
	if err != nil {
		panic(err)
	}

	buf := make([]byte, GoodEncBufferSize)

	// Encrypt:
	_, err = io.CopyBuffer(enc, source, buf)
	if err != nil {
		panic(err)
	}

	// This needs to be here, since close writes
	// left over data to the write stream
	if err := enc.Close(); err != nil {
		t.Errorf("close(enc): %v", err)
		return
	}

	sharedReader := bytes.NewReader(shared.Bytes())
	decLayer, err := NewReader(sharedReader, TestKey)
	if err != nil {
		t.Errorf("cannot create new reader: %v", err)
		return
	}

	seekTest := int64(MaxBlockSize)
	pos, err := decLayer.Seek(seekTest, os.SEEK_SET)
	if err != nil {
		t.Errorf("Seek error'd: %v", err)
		return
	}

	if pos != seekTest {
		t.Errorf("Seek is a bad jumper: %d (should %d)", pos, MaxBlockSize)
		return
	}

	pos, _ = decLayer.Seek(0, os.SEEK_CUR)
	if pos != seekTest {
		t.Errorf("SEEK_CUR(0) deliverd wrong status")
		return
	}

	pos, _ = decLayer.Seek(seekTest/2, os.SEEK_CUR)
	if pos != seekTest+seekTest/2 {
		t.Errorf("SEEK_CUR jumped to the wrong pos: %d", pos)
		return
	}

	pos, _ = decLayer.Seek(-seekTest, os.SEEK_CUR)
	if pos != seekTest/2 {
		t.Errorf("SEEK_CUR does not like negative indices: %d", pos)
		return
	}

	pos, _ = decLayer.Seek(seekTest/2, os.SEEK_CUR)
	if pos != seekTest {
		t.Errorf("SEEK_CUR has problems after negative indices: %d", pos)
		return
	}

	if N != int64(decLayer.info.Length) {
		t.Errorf(
			"Input length (%d) does not match header length (%d).",
			N,
			decLayer.info.Length,
		)
		return
	}

	// Check if SEEK_END appears to work
	// (jump to same position, but from end of file)
	endPos, _ := decLayer.Seek(seekTest, os.SEEK_END)
	if endPos != pos {
		t.Errorf("SEEK_END failed; should be %d, was %d", pos, endPos)
		return
	}

	// Decrypt:
	_, err = io.CopyBuffer(dest, decLayer, buf)
	if err != nil {
		t.Errorf("Decrypt failed: %v", err)
		return
	}

	if err := decLayer.Close(); err != nil {
		t.Errorf("close(dec): %v", err)
		return
	}

	if !bytes.Equal(a[seekTest:], dest.Bytes()) {
		b := dest.Bytes()
		t.Errorf("Buffers are not equal:")
		t.Errorf("\tAAA %d %x %x\n", len(a), a[:10], a[len(a)-10:])
		t.Errorf("\tBBB %d %x %x\n", len(b), b[:10], b[len(b)-10:])
		return
	}
}

func BenchmarkEncDec(b *testing.B) {
	for n := 0; n < b.N; n++ {
		testSimpleEncDec(nil, MaxBlockSize*100)
	}
}

// Regression test:
// check that reader does not read first block first,
// even if jumping right into the middle of the file.
func TestSeekThenRead(t *testing.T) {
	N := int64(2 * MaxBlockSize)
	a := testutil.CreateDummyBuf(N)
	b := make([]byte, 0, N)

	source := bytes.NewBuffer(a)
	shared := &bytes.Buffer{}
	dest := bytes.NewBuffer(b)

	enc, err := NewWriter(shared, TestKey, N)
	if err != nil {
		panic(err)
	}

	// Use a different buf size for a change:
	buf := make([]byte, 4096)

	// Encrypt:
	_, err = io.CopyBuffer(enc, source, buf)
	if err != nil {
		t.Errorf("copy(enc, source) failed %v", err)
		return
	}

	// This needs to be here, since close writes
	// left over data to the write stream
	if err = enc.Close(); err != nil {
		t.Errorf("close(enc): %v", err)
		return
	}

	sharedReader := bytes.NewReader(shared.Bytes())
	decLayer, err := NewReader(sharedReader, TestKey)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := decLayer.Close(); err != nil {
			t.Errorf("close(dec) failed: %v", err)
		}
	}()

	// Jump somewhere inside the large file:
	jumpPos := N/2 + N/4 + 1
	newPos, err := decLayer.Seek(jumpPos, os.SEEK_SET)
	if err != nil {
		t.Errorf("Seek failed in SeekThenRead: %v", err)
		return
	}

	if newPos != jumpPos {
		t.Errorf("Seek jumped to %v (should be %v)", newPos, N/2+N/4)
		return
	}

	// Decrypt:
	copiedBytes, err := io.CopyBuffer(dest, decLayer, buf)
	if err != nil {
		t.Errorf("Decrypt failed: %v", err)
		return
	}

	if copiedBytes != N-jumpPos {
		t.Errorf("Copied different amount of decrypted data than expected.")
		t.Errorf("Should be %v, was %v bytes.", copiedBytes, N-jumpPos)
		return
	}

	// Check the data actually matches the source data.
	if !bytes.Equal(a[newPos:], dest.Bytes()) {
		t.Errorf("Seeked data does not match expectations.")
		t.Errorf("\tEXPECTED: %v...", a[newPos:newPos:10])
		t.Errorf("\tGOT:      %v...", dest.Bytes()[:10])
	}
}

func TestEmptyFile(t *testing.T) {
	srcBuf := []byte{}
	dstBuf := []byte{}
	tmpBuf := &bytes.Buffer{}

	src := bytes.NewReader(srcBuf)
	dst := bytes.NewBuffer(dstBuf)

	enc, err := NewWriter(tmpBuf, TestKey, 0)
	if err != nil {
		t.Errorf("TestEmpyFile: creating writer failed: %v", err)
		return
	}

	if _, err = io.Copy(enc, src); err != nil {
		t.Errorf("TestEmpyFile: copy(enc, src) failed: %v", err)
		return
	}

	if err = enc.Close(); err != nil {
		t.Errorf("TestEmpyFile: close(enc) failed: %v", err)
		return
	}

	dec, err := NewReader(bytes.NewReader(tmpBuf.Bytes()), TestKey)
	if err != nil {
		t.Errorf("TestEmpyFile: creating reader failed: %v", err)
		return
	}

	if err = dec.Close(); err != nil {
		t.Errorf("TestEmpyFile: close(dec) failed: %v", err)
		return
	}

	if _, err = dec.Seek(10, os.SEEK_SET); err != nil {
		t.Errorf("Seek failed: %v", err)
		return
	}

	if _, err = io.Copy(dst, dec); err != nil {
		t.Errorf("TestEmpyFile: copy(dst, dec) failed: %v", err)
		return
	}

	if !bytes.Equal(srcBuf, dstBuf) {
		t.Errorf("TestEmpyFile: Not empty: src=%v dst=%v", srcBuf, dstBuf)
		return
	}
}

// Test if encrypting the same plaintext twice yields
// the same ciphertext. This is a crucial property for brig, although it
// has some security implications (i.e. no real random etc.)
func TestEncryptedTheSame(t *testing.T) {
	sourceData := testutil.CreateDummyBuf(3 * MaxBlockSize)
	encOne := &bytes.Buffer{}
	encTwo := &bytes.Buffer{}

	n1, err := Encrypt(TestKey, bytes.NewReader(sourceData), encOne, int64(len(sourceData)))
	if err != nil {
		t.Errorf("TestEncryptedTheSame: Encrypting first failed: %v", err)
		return
	}

	n2, err := Encrypt(TestKey, bytes.NewReader(sourceData), encTwo, int64(len(sourceData)))
	if err != nil {
		t.Errorf("TestEncryptedTheSame: Encrypting second failed: %v", err)
		return
	}

	if n1 != n2 {
		t.Errorf("TestEncryptedTheSame: Ciphertexts differ in length.")
		return
	}

	if !bytes.Equal(encOne.Bytes(), encTwo.Bytes()) {
		t.Errorf("TestEncryptedTheSame: Ciphertext differ, you failed.")
		t.Errorf("\tOne: %v", encOne.Bytes())
		t.Errorf("\tTwo: %v", encTwo.Bytes())
		return
	}
}
