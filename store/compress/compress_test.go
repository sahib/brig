package compress

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/disorganizer/brig/util/testutil"
)

var (
	ZipFilePath = filepath.Join(os.TempDir(), "compressed.zip")
)

func openDest(t *testing.T, dest string) *os.File {
	if _, err := os.Stat(dest); !os.IsNotExist(err) {
		t.Fatalf("Opening destination %v failed: %v\n", dest, err)
	}

	fd, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		t.Fatalf("Opening srouce %v failed: %v\n", dest, err)
	}
	return fd
}

func openSrc(t *testing.T, src string) *os.File {
	fd, err := os.Open(src)
	if err != nil {
		t.Fatalf("Opening srouce %v failed: %v\n", src, err)
	}
	return fd
}

const (
	C64K = 64 * 1024
	C32K = 32 * 1024
)

func TestCompressDecompress(t *testing.T) {
	testValues := []int64{0, C64K - 1, C64K, C64K + 1, C32K - 1, C32K, C32K + 1}
	for _, size := range testValues {
		testCompressDecompress(t, size)
	}
}

func testCompressDecompress(t *testing.T, size int64) {
	// Fake data file is written to disk,
	// as compression reader has to be a ReadSeeker.
	data := testutil.CreateDummyBuf(size)
	zipFileDest := openDest(t, ZipFilePath)

	// Compress.
	w := NewWriter(zipFileDest, AlgoSnappy)
	if _, err := io.Copy(w, bytes.NewReader(data)); err != nil {
		t.Errorf("Compress failed %v", err)
	}

	if err := w.Close(); err != nil {
		t.Errorf("Compression writer close failed: %v", err)
	}

	// Read compressed file into buffer.
	dataUncomp := bytes.NewBuffer(nil)
	dataFromZip := openSrc(t, ZipFilePath)

	// Uncompress.
	r := NewReader(dataFromZip)
	if _, err := io.Copy(dataUncomp, r); err != nil {
		t.Errorf("Decompression failed: %v", err)
	}
	if err := dataFromZip.Close(); err != nil {
		t.Errorf("Zip close failed: %v", err)
	}

	// Compare.
	if !bytes.Equal(dataUncomp.Bytes(), data) {
		t.Error("Uncompressed data and input data does not match.")
	}
	testutil.Remover(t, ZipFilePath)
}
