package compress

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var (
	plainFile   string = "WWarum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?Warum bin ich nur eine kleine Katzen aus dem polnischen Lande?arum bin ich nur eine kleine Katzen aus dem polnischen Lande?"
	ZipFilePath        = filepath.Join(os.TempDir(), "compressed.zip")
)

func openDest(src string) *os.File {
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		fmt.Printf("%s already exists.\n", src)
		os.Exit(-1)
	}

	fd, err := os.OpenFile(src, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	return fd
}

func openSrc(dest string) *os.File {
	fd, err := os.Open(ZipFilePath)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	return fd
}

func cleanTestdata(path string) {
	if err := os.Remove(path); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func TestCompressDecompress(t *testing.T) {
	// Fake data file is written to disk,
	// as compression reader has to be a ReadSeeker.
	dataToZip := bytes.NewBufferString(plainFile)
	zipFileDest := openDest(ZipFilePath)

	// Compress.
	w := NewWriter(zipFileDest, AlgoSnappy)
	io.Copy(w, dataToZip)
	w.Close()
	zipFileDest.Close()

	// Read compressed file into buffer.
	dataUncomp := bytes.NewBufferString("")
	dataFromZip := openSrc(ZipFilePath)
	defer dataFromZip.Close()

	// Uncompress.
	r := NewReader(dataFromZip)
	io.Copy(dataUncomp, r)

	// Compare.
	if strings.Compare(dataUncomp.String(), plainFile) != 0 {
		fmt.Println("Uncompressed data and plain file does not match.")
		os.Exit(-1)
	}
	cleanTestdata(ZipFilePath)
}
