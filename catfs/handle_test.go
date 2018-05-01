package catfs

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"testing"

	"github.com/sahib/brig/catfs/mio/compress"
	"github.com/sahib/brig/util/testutil"
	"github.com/stretchr/testify/require"
)

// TODO: write more tests for the handle interface.

func TestOpenRead(t *testing.T) {
	withDummyFS(t, func(fs *FS) {
		rawData := []byte{1, 2, 3}
		require.Nil(t, fs.Stage("/x", bytes.NewReader(rawData)))

		fd, err := fs.Open("/x")
		require.Nil(t, err)

		data, err := ioutil.ReadAll(fd)
		require.Nil(t, err)

		require.Equal(t, data, rawData)
		require.Nil(t, fd.Close())
	})
}

func TestOpenWrite(t *testing.T) {
	withDummyFS(t, func(fs *FS) {
		rawData := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		require.Nil(t, fs.Stage("/x", bytes.NewReader(rawData)))

		fd, err := fs.Open("/x")
		require.Nil(t, err)

		n, err := fd.Write([]byte{3, 2, 1})
		require.Nil(t, err)
		require.Equal(t, n, 3)

		data, err := ioutil.ReadAll(fd)
		require.Nil(t, err)
		require.Equal(t, rawData[3:], data)

		// Check that we can also seek back to start after reading to the end.
		// (and also check if the write overlay actually did work)
		pos, err := fd.Seek(0, io.SeekStart)
		require.Nil(t, err)
		require.Equal(t, pos, int64(0))

		data, err = ioutil.ReadAll(fd)
		require.Nil(t, err)
		require.Equal(t, data, []byte{3, 2, 1, 4, 5, 6, 7, 8, 9, 10})
		require.Nil(t, fd.Close())
	})
}

func TestOpenTruncate(t *testing.T) {
	rawData := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	for idx := 0; idx < len(rawData)+5; idx++ {
		t.Run(fmt.Sprintf("truncate_%d", idx), func(t *testing.T) {
			withDummyFS(t, func(fs *FS) {
				require.Nil(t, fs.Stage("/x", bytes.NewReader(rawData)))

				fd, err := fs.Open("/x")
				require.Nil(t, err)

				require.Nil(t, fd.Truncate(uint64(idx)))

				data, err := ioutil.ReadAll(fd)
				require.Nil(t, err)

				// cap rawData index:
				rawIdx := idx
				if idx >= len(rawData) {
					rawIdx = len(rawData)
				}

				require.Equal(t, rawData[:rawIdx], data)
				require.Nil(t, fd.Close())

				// Check if the result was really written:
				stream, err := fs.Cat("/x")
				require.Nil(t, err)

				persistentData, err := ioutil.ReadAll(stream)
				require.Nil(t, err)
				require.Equal(t, rawData[:rawIdx], persistentData)
			})
		})
	}
}

func TestOpenOpAfterClose(t *testing.T) {
	withDummyFS(t, func(fs *FS) {
		rawData := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		require.Nil(t, fs.Stage("/x", bytes.NewReader(rawData)))

		fd, err := fs.Open("/x")
		require.Nil(t, err)

		require.Nil(t, fd.Close())

		_, err = ioutil.ReadAll(fd)
		require.Equal(t, err, ErrIsClosed)
	})
}

// TODO: More tests. This is still very buggy.
//       Cases needed for:
//       - 0, SEEK_END
//       - 9, SEEK_SET
//       - ...
func TestOpenExtend(t *testing.T) {
	withDummyFS(t, func(fs *FS) {
		rawData := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		require.Nil(t, fs.Stage("/x", bytes.NewReader(rawData)))

		fd, err := fs.Open("/x")
		require.Nil(t, err)

		pos, err := fd.Seek(10, io.SeekStart)
		require.Nil(t, err)
		require.Equal(t, pos, int64(10))

		n, err := fd.Write([]byte{11, 12, 13})
		require.Nil(t, err)
		require.Equal(t, n, 3)

		require.Nil(t, fd.Close())

		stream, err := fs.Cat("/x")
		require.Nil(t, err)

		postData, err := ioutil.ReadAll(stream)
		require.Nil(t, err)

		expected := make([]byte, 13)
		copy(expected, rawData)
		copy(expected[10:], []byte{11, 12, 13})
		require.Equal(t, expected, postData)
	})
}

// Read data from the handle like fuse would:
// Seek to an offset, read a chunk and then advance to next block.
// block size and file size may var heavily here.
func TestHandleFuseLikeRead(t *testing.T) {
	tcs := []struct {
		fileSize  int
		blockSize int
	}{
		{2048, 400},
	}

	for _, tc := range tcs {

		testHandleFuseLikeRead(t, tc.fileSize, tc.blockSize)
	}
}

func testHandleFuseLikeRead(t *testing.T, fileSize, blockSize int) {
	// fuse reads data always with a prior seek.
	// try to emulate this behaviour here.
	withDummyFS(t, func(fs *FS) {
		rawData := testutil.CreateDummyBuf(int64(fileSize))
		require.Nil(t, fs.Stage("/x", bytes.NewReader(rawData)))

		fd, err := fs.Open("/x")
		require.Nil(t, err)

		left := len(rawData)
		for left > 0 {
			toRead := blockSize
			if left < blockSize {
				toRead = left
			}

			offset := len(rawData) - left
			buf := make([]byte, toRead)
			if _, err = fd.Seek(int64(offset), io.SeekStart); err != nil {
				t.Fatalf("Seek to %d failed", offset)
			}

			n, err := fd.Read(buf)
			if err != nil {
				t.Fatalf("Read failed: %v", err)
			}

			if n != toRead {
				t.Fatalf("Handle read less than expected (wanted %d, got %d)", toRead, n)
			}

			if !bytes.Equal(buf, rawData[offset:offset+toRead]) {
				t.Fatalf("Block [%d:%d] differs from raw data", offset, offset+toRead)
			}

			left -= blockSize
		}

		require.Nil(t, fd.Close())
	})
}

func TestHandleChangeCompression(t *testing.T) {
	withDummyFS(t, func(fs *FS) {
		// Create a file which will not be compressed.
		size := int64(compress.HeaderSizeThreshold - 1)
		oldData := testutil.CreateDummyBuf(size)
		require.Nil(t, fs.Mkdir("/sub", false))
		require.Nil(t, fs.Stage("/sub/a-text-file.go", bytes.NewReader(oldData)))

		// Second run will use another compress algorithm, since we're
		// over the header size limit in the compression guesser.
		fd, err := fs.Open("/sub/a-text-file.go")
		require.Nil(t, err)

		// "echo(1)" does a flush after open (for whatever reason)
		require.Nil(t, fd.Flush())

		offset, err := fd.Seek(size, io.SeekStart)
		require.Nil(t, err)
		require.Equal(t, offset, size)

		expectedData := []byte("xxxxx")
		n, err := fd.Write(expectedData)
		require.Nil(t, err)
		require.Equal(t, n, len(expectedData))
		require.Nil(t, fd.Flush())
		require.Nil(t, fd.Close())

		stream, err := fs.Cat("/sub/a-text-file.go")
		require.Nil(t, err)

		gotData, err := ioutil.ReadAll(stream)
		require.Nil(t, err)

		expectData := append(oldData, expectedData...)
		require.Equal(t, expectData, gotData)
	})
}
