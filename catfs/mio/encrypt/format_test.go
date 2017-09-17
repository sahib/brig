package encrypt

import (
	"bytes"
	"io"
	"io/ioutil"
	"math"
	"os"
	"testing"

	"github.com/disorganizer/brig/util"
	"github.com/disorganizer/brig/util/testutil"
)

var TestKey = []byte("01234567890ABCDE01234567890ABCDE")

const ExtraDebug = false

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

	return Encrypt(key, fdFrom, fdTo)
}

func decryptFile(key []byte, from, to string) (n int64, outErr error) {
	fdFrom, fdTo, err := openFiles(from, to)
	if err != nil {
		return 0, err
	}

	defer func() {
		if err := fdTo.Close(); err != nil {
			outErr = err
			return
		}

		if err := fdFrom.Close(); err != nil {
			outErr = err
			return
		}
	}()

	return Decrypt(key, fdFrom, fdTo)
}

func remover(t *testing.T, path string) {
	if err := os.Remove(path); err != nil {
		t.Errorf("Could not remove temp file: %v", err)
	}
}

func testSimpleEncDec(t *testing.T, size int64) {
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

	for _, size := range SizeTests {
		if ExtraDebug {
			t.Logf("Testing SimpleEncDec for size %d", size)
		}
		testSimpleEncDec(t, size)
	}
}

var SizeTests = []int64{
	0,
	1,
	defaultMaxBlockSize - 1,
	defaultMaxBlockSize,
	defaultMaxBlockSize + 10,
	defaultDecBufferSize - 1,
	defaultDecBufferSize,
	defaultDecBufferSize + 1,
	defaultEncBufferSize - 1,
	defaultEncBufferSize,
	7 * defaultEncBufferSize,
	7*defaultEncBufferSize - 1,
	defaultEncBufferSize + 1,
}

var SeekTests = []struct {
	Whence int
	Offset float64
	Error  error
}{
	// Jump to the mid:
	{os.SEEK_SET, 0.5, nil},
	// Should stay the same:
	{os.SEEK_CUR, 0, nil},
	// Jump a quarter forth:
	{os.SEEK_CUR, 0.25, nil},
	// Jump a half back:
	{os.SEEK_CUR, -0.5, nil},
	// Jump back to the half:
	{os.SEEK_CUR, 0.25, nil},
	// See if SEEK_END works:
	{os.SEEK_END, -0.5, nil},
	// This triggered a crash earlier:
	{os.SEEK_END, -2, io.EOF},
	// Im guessing now:
	{os.SEEK_END, -1.0 / 4096, nil},
}

func BenchmarkEncDec(b *testing.B) {
	for n := 0; n < b.N; n++ {
		testSimpleEncDec(nil, defaultMaxBlockSize*100)
	}
}

func TestSeek(t *testing.T) {
	for _, size := range SizeTests {
		testSeek(t, int64(size), false, false)
		testSeek(t, int64(size), false, true)
		testSeek(t, int64(size), true, false)
		testSeek(t, int64(size), true, true)

		if t.Failed() {
			break
		}
	}
}

func testSeek(t *testing.T, N int64, readFrom, writeTo bool) {
	sourceData := testutil.CreateDummyBuf(N)
	source := bytes.NewBuffer(sourceData)
	shared := &bytes.Buffer{}

	if ExtraDebug {
		t.Logf("Testing seek for size %d", N)
	}

	enc, err := NewWriter(shared, TestKey)
	if err != nil {
		t.Errorf("Creating an encrypted writer failed: %v", err)
		return
	}

	// Encrypt:
	if _, err = testutil.DumbCopy(enc, source, readFrom, writeTo); err != nil {
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
		t.Errorf("creating new reader failed: %v", err)
		return
	}

	lastJump := int64(0)

	for _, test := range SeekTests {
		realOffset := int64(math.Floor(.5 + test.Offset*float64(N)))

		whence := map[int]string{
			0: "SEEK_SET",
			1: "SEEK_CUR",
			2: "SEEK_END",
		}[test.Whence]

		exptOffset := int64(0)
		switch test.Whence {
		case os.SEEK_SET:
			exptOffset = realOffset
		case os.SEEK_CUR:
			exptOffset = lastJump + realOffset
		case os.SEEK_END:
			exptOffset = N + realOffset
		default:
			panic("Bad whence")
		}

		if ExtraDebug {
			t.Logf(
				" => Seek(%v, %v) -> %v (size: %v)",
				realOffset,
				whence,
				exptOffset,
				N,
			)
		}

		jumpedTo, err := decLayer.Seek(realOffset, test.Whence)
		if err != test.Error {
			if err != io.EOF && N != 0 {
				t.Errorf(
					"Seek(%v, %v) produced an error: %v (should be %v)",
					realOffset,
					whence,
					err,
					test.Error,
				)
			}
			return
		}

		if test.Error != nil {
			continue
		}

		if jumpedTo != exptOffset {
			t.Errorf(
				"Seek(%v, %v) jumped badly. Should be %v, was %v",
				realOffset,
				whence,
				exptOffset,
				jumpedTo,
			)
			return
		}

		lastJump = jumpedTo

		// Decrypt and check if the contents are okay:
		dest := bytes.NewBuffer(nil)

		copiedBytes, err := testutil.DumbCopy(dest, decLayer, readFrom, writeTo)
		if err != nil {
			t.Errorf("Decrypt failed: %v", err)
			return
		}

		if copiedBytes != N-jumpedTo {
			t.Errorf("Copied different amount of decrypted data than expected.")
			t.Errorf("Should be %v, was %v bytes.", N-jumpedTo, copiedBytes)
		}

		// Check the data actually matches the source data.
		if !bytes.Equal(sourceData[jumpedTo:], dest.Bytes()) {
			t.Errorf("Seeked data does not match expectations.")
			t.Errorf("\tEXPECTED: %v", util.OmitBytes(sourceData[jumpedTo:], 10))
			t.Errorf("\tGOT:      %v", util.OmitBytes(dest.Bytes(), 10))
			return
		}

		// Jump back, so the other tests continue to work:
		jumpedAgain, err := decLayer.Seek(jumpedTo, os.SEEK_SET)
		if err != nil {
			t.Errorf("Seeking not possible after reading: %v", err)
			return
		}

		if jumpedTo != jumpedAgain {
			t.Errorf("Jumping back to original pos failed.")
			t.Errorf("Should be %v, was %v.", jumpedTo, jumpedAgain)
			return
		}
	}
}

func TestEmptyFile(t *testing.T) {
	srcBuf := []byte{}
	dstBuf := []byte{}
	tmpBuf := &bytes.Buffer{}

	src := bytes.NewReader(srcBuf)
	dst := bytes.NewBuffer(dstBuf)

	enc, err := NewWriter(tmpBuf, TestKey)
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

	if _, err = dec.Seek(10, os.SEEK_SET); err != io.EOF {
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
	sourceData := testutil.CreateDummyBuf(3 * defaultMaxBlockSize)
	encOne := &bytes.Buffer{}
	encTwo := &bytes.Buffer{}

	n1, err := Encrypt(TestKey, bytes.NewReader(sourceData), encOne)
	if err != nil {
		t.Errorf("TestEncryptedTheSame: Encrypting first failed: %v", err)
		return
	}

	n2, err := Encrypt(TestKey, bytes.NewReader(sourceData), encTwo)
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
