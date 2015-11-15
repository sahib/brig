package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func createFile(size int64) string {
	fd, err := ioutil.TempFile("", "brig_test")
	if err != nil {
		panic("Cannot create temp file")
	}

	defer fd.Close()

	blockSize := int64(1 * 1024 * 1024)
	buf := make([]byte, blockSize)

	for i := int64(0); i < blockSize; i++ {
		buf[i] = byte(i)
	}

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

	key := []byte("01234567890ABCDE01234567890ABCDE")

	encPath := path + "_enc"
	decPath := path + "_dec"

	var err error
	_, err = encryptFile(key, path, encPath)
	defer os.Remove(encPath)

	if err != nil {
		log.Println(err)
		t.Errorf("Encrypt failed: %v", err)
	}

	_, err = decryptFile(key, encPath, decPath)
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
	// rEnc, wEnc := io.Pipe()
	// enc := NewEncryptedWriter()
}

func BenchmarkEncDec(b *testing.B) {
	for n := 0; n < b.N; n++ {
		testSimpleEncDec(nil, MaxBlockSize*100)
	}
}
