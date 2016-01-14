package encrypt

import (
	"bytes"
	"fmt"
	"github.com/disorganizer/brig/util/testutil"
	"io"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

var TestKey = []byte("01234567890ABCDE01234567890ABCDE")

func encryptFile(key []byte, from, to string) (int64, error) {
	fdFrom, _ := os.Open(from)
	defer fdFrom.Close()

	fdTo, _ := os.OpenFile(to, os.O_CREATE|os.O_WRONLY, 0755)
	defer fdTo.Close()

	return Encrypt(key, fdFrom, fdTo)
}

func decryptFile(key []byte, from, to string) (int64, error) {
	fdFrom, _ := os.Open(from)
	defer fdFrom.Close()

	fdTo, _ := os.OpenFile(to, os.O_CREATE|os.O_WRONLY, 0755)
	defer fdTo.Close()

	return Decrypt(key, fdFrom, fdTo)
}

func testSimpleEncDec(t *testing.T, size int) {
	path := testutil.CreateFile(int64(size))
	defer os.Remove(path)

	encPath := path + "_enc"
	decPath := path + "_dec"

	_, err := encryptFile(TestKey, path, encPath)
	defer os.Remove(encPath)

	if err != nil {
		log.Println(err)
		t.Errorf("Encrypt failed: %v", err)
	}

	_, err = decryptFile(TestKey, encPath, decPath)
	defer os.Remove(decPath)

	if err != nil {
		log.Println(err)
		t.Errorf("Decrypt failed: %v", err)
	}

	a, _ := ioutil.ReadFile(path)
	b, _ := ioutil.ReadFile(decPath)
	c, _ := ioutil.ReadFile(encPath)

	if !bytes.Equal(a, b) {
		t.Errorf("Source and decrypted not equal")
	}

	if bytes.Equal(a, c) {
		t.Errorf("Source was not encrypted (same as source)")
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

	encLayer, err := NewWriter(shared, TestKey)
	if err != nil {
		panic(err)
	}

	buf := make([]byte, GoodEncBufferSize)

	// Encrypt:
	_, err = io.CopyBuffer(encLayer, source, buf)
	if err != nil {
		panic(err)
	}

	// This needs to be here, since close writes
	// left over data to the write stream
	encLayer.Close()

	sharedReader := bytes.NewReader(shared.Bytes())
	decLayer, err := NewReader(sharedReader, TestKey)
	if err != nil {
		panic(err)
	}
	defer decLayer.Close()

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
	}

	pos, _ = decLayer.Seek(-seekTest, os.SEEK_CUR)
	if pos != seekTest/2 {
		t.Errorf("SEEK_CUR does not like negative indices: %d", pos)
	}

	pos, _ = decLayer.Seek(seekTest/2, os.SEEK_CUR)
	if pos != seekTest {
		t.Errorf("SEEK_CUR has problems after negative indices: %d", pos)
	}

	// Decrypt:
	_, err = io.CopyBuffer(dest, decLayer, buf)
	if err != nil {
		t.Errorf("Decrypt failed: %v", err)
		return
	}

	if !bytes.Equal(a[seekTest:], dest.Bytes()) {
		b := dest.Bytes()
		fmt.Printf("AAA %d %x %x\n", len(a), a[:10], a[len(a)-10:])
		fmt.Printf("BBB %d %x %x\n", len(b), b[:10], b[len(b)-10:])
		t.Errorf("Buffers are not equal")
		return
	}
}

func BenchmarkEncDec(b *testing.B) {
	for n := 0; n < b.N; n++ {
		testSimpleEncDec(nil, MaxBlockSize*100)
	}
}
