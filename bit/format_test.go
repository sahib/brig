package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

var TestKey = []byte("01234567890ABCDE01234567890ABCDE")

func createDummyBuf(size int64) []byte {
	buf := make([]byte, size)

	for i := int64(0); i < size; i++ {
		// Be evil and stripe the data:
		buf[i] = byte(i % 255)
	}

	return buf
}

func createFile(size int64) string {
	fd, err := ioutil.TempFile("", "brig_test")
	if err != nil {
		panic("Cannot create temp file")
	}

	defer fd.Close()

	blockSize := int64(1 * 1024 * 1024)
	buf := createDummyBuf(blockSize)

	for size > 0 {
		take := size
		if size > int64(len(buf)) {
			take = int64(len(buf))
		}

		_, err := fd.Write(buf[:take])
		if err != nil {
			panic(err)
		}

		size -= blockSize
	}

	return fd.Name()
}

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
	path := createFile(int64(size))
	defer os.Remove(path)

	encPath := path + "_enc"
	decPath := path + "_dec"

	var err error
	_, err = encryptFile(TestKey, path, encPath)
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

	if !bytes.Equal(a, b) {
		t.Errorf("Source and decrypted not equal")
	}
}

func TestSimpleEncDec(t *testing.T) {
	t.Parallel()

	sizes := []int{MaxBlockSize - 1, MaxBlockSize, MaxBlockSize + 1}
	for size := range sizes {
		testSimpleEncDec(t, size)
	}
}

func TestSeek(t *testing.T) {
	N := int64(2 * MaxBlockSize)
	a := createDummyBuf(N)
	b := make([]byte, 0, N)

	source := bytes.NewBuffer(a)
	shared := &bytes.Buffer{}
	dest := bytes.NewBuffer(b)

	encLayer, err := NewEncryptedWriter(shared, TestKey)
	if err != nil {
		panic(err)
	}

	buf := make([]byte, GoodBufferSize)

	// Encrypt:
	_, err = io.CopyBuffer(encLayer, source, buf)
	if err != nil {
		panic(err)
	}

	// This needs to be here, since close writes
	// left over data to the write stream
	encLayer.Close()

	sharedReader := bytes.NewReader(shared.Bytes())
	decLayer, err := NewEncryptedReader(sharedReader, TestKey)
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
