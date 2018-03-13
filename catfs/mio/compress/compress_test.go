package compress

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/sahib/brig/util"
	"github.com/sahib/brig/util/testutil"
	"github.com/stretchr/testify/require"
)

var (
	TestOffsets      = []int64{-1, -500, 0, 1, -C64K, -C32K, C64K - 1, C64K, C64K + 1, C32K - 1, C32K, C32K + 1, C64K - 5, C64K + 5, C32K - 5, C32K + 5}
	TestSizes        = []int64{0, 1, C64K - 1, C64K, C64K + 1, C32K - 1, C32K, C32K + 1, C64K - 5, C64K + 5, C32K - 5, C32K + 5}
	CompressionAlgos = []AlgorithmType{AlgoLZ4}
)

func openDest(t *testing.T, dest string) *os.File {
	if _, err := os.Stat(dest); !os.IsNotExist(err) && err != nil {
		t.Fatalf("Opening destination %v failed: %v\n", dest, err)
	}
	fd, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		t.Fatalf("Opening source %v failed: %v\n", dest, err)
	}
	return fd
}

func openSrc(t *testing.T, src string) *os.File {
	fd, err := os.Open(src)
	if err != nil {
		t.Fatalf("Opening source %v failed: %v\n", src, err)
	}
	return fd
}

func createTempFile(t *testing.T) string {
	fd, err := ioutil.TempFile("", "brig-mio-compress")
	path := fd.Name()
	require.Nil(t, err)
	require.Nil(t, fd.Close())
	return path
}

const (
	C64K = 64 * 1024
	C32K = 32 * 1024
)

func TestCompressDecompress(t *testing.T) {
	t.Parallel()

	sizes := TestSizes
	algos := CompressionAlgos

	for _, algo := range algos {
		for _, size := range sizes {
			name := fmt.Sprintf("%s-size%d", algo, size)
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				testCompressDecompress(t, size, algo, true, true)
				testCompressDecompress(t, size, algo, false, false)
				testCompressDecompress(t, size, algo, true, false)
				testCompressDecompress(t, size, algo, false, true)
			})
		}
	}
}

func testCompressDecompress(t *testing.T, size int64, algo AlgorithmType, useReadFrom, useWriteTo bool) {
	// Fake data file is written to disk,
	// as compression reader has to be a ReadSeeker.
	zipPath := createTempFile(t)
	defer testutil.Remover(t, zipPath)

	data := testutil.CreateDummyBuf(size)
	zipFileDest := openDest(t, zipPath)

	// Compress.
	w, err := NewWriter(zipFileDest, algo)
	if err != nil {
		t.Errorf("Writer init failed %v", err)
		return

	}

	if _, err := testutil.DumbCopy(w, bytes.NewReader(data), useReadFrom, useWriteTo); err != nil {
		t.Errorf("Compress failed %v", err)
		return
	}

	if err := w.Close(); err != nil {
		t.Errorf("Compression writer close failed: %v", err)
		return
	}

	if err := zipFileDest.Close(); err != nil {
		t.Errorf("close(zipFileDest) failed: %v", err)
		return
	}

	// Read compressed file into buffer.
	dataUncomp := bytes.NewBuffer(nil)
	dataFromZip := openSrc(t, zipPath)

	// Uncompress.
	r := NewReader(dataFromZip)
	if _, err := testutil.DumbCopy(dataUncomp, r, useReadFrom, useWriteTo); err != nil {
		t.Errorf("Decompression failed: %v", err)
		return
	}
	if err := dataFromZip.Close(); err != nil {
		t.Errorf("Zip close failed: %v", err)
		return
	}

	// Compare.
	got, want := dataUncomp.Bytes(), data
	if !bytes.Equal(got, want) {
		t.Error("Uncompressed data and input data does not match.")
		t.Errorf("\tGOT:   %v", util.OmitBytes(got, 10))
		t.Errorf("\tWANT:  %v", util.OmitBytes(want, 10))
		return
	}
}

func TestSeek(t *testing.T) {
	t.Parallel()

	sizes := TestSizes
	offsets := TestOffsets
	algos := CompressionAlgos
	for _, algo := range algos {
		for _, size := range sizes {
			for _, off := range offsets {
				name := fmt.Sprintf("%s-size%d-off%d", algo, size, off)
				t.Run(name, func(t *testing.T) {
					t.Parallel()

					testSeek(t, size, off, algo, false, false)
					testSeek(t, size, off, algo, true, true)
					testSeek(t, size, off, algo, false, true)
					testSeek(t, size, off, algo, true, false)
				})
			}
		}
	}
}

func testSeek(t *testing.T, size, offset int64, algo AlgorithmType, useReadFrom, useWriteTo bool) {
	// Fake data file is written to disk,
	// as compression reader has to be a ReadSeeker.
	zipPath := createTempFile(t)
	defer testutil.Remover(t, zipPath)

	data := testutil.CreateDummyBuf(size)
	zipFileDest := openDest(t, zipPath)

	// Compress.
	w, err := NewWriter(zipFileDest, algo)
	if err != nil {
		t.Errorf("Writer init failed %v", err)
		return
	}
	if _, err := testutil.DumbCopy(w, bytes.NewReader(data), useReadFrom, useWriteTo); err != nil {
		t.Errorf("Compress failed %v", err)
		return
	}
	defer testutil.Remover(t, zipPath)

	if err := w.Close(); err != nil {
		t.Errorf("Compression writer close failed: %v", err)
		return
	}

	if err := zipFileDest.Close(); err != nil {
		t.Errorf("close(zipFileDest) failed: %v", err)
		return
	}

	// Read compressed file into buffer.
	dataUncomp := bytes.NewBuffer(nil)
	dataFromZip := openSrc(t, zipPath)
	zr := NewReader(dataFromZip)

	// Set specific offset before read.
	_, err = zr.Seek(offset, io.SeekStart)
	if err == io.EOF && offset < size && offset > -1 {
		t.Errorf("Seek failed even with EOF: %d <= %d", offset, size)
		return
	}
	if err != io.EOF && err != nil {
		t.Errorf("Seek failed: %v", err)
		return
	}

	// Read starting at a specific offset.
	if _, err := testutil.DumbCopy(dataUncomp, zr, useReadFrom, useWriteTo); err != nil {
		t.Errorf("Decompression failed: %v", err)
		return
	}
	if err := dataFromZip.Close(); err != nil {
		t.Errorf("Zip close failed: %v", err)
		return
	}

	// Compare.
	maxOffset := offset
	if offset > size {
		maxOffset = size
	}

	if offset < 0 {
		maxOffset = 0
	}

	got, want := dataUncomp.Bytes(), data[maxOffset:]
	if !bytes.Equal(got, want) {
		t.Error("Uncompressed data and input data does not match.")
		t.Errorf("\tGOT:   %v", util.OmitBytes(got, 10))
		t.Errorf("\tWANT:  %v", util.OmitBytes(want, 10))
		return
	}
}
