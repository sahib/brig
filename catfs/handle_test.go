package catfs

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// TODO: write tests for the handle interface.

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

		pos, err := fd.Seek(0, os.SEEK_CUR)
		require.Nil(t, err)
		require.Equal(t, pos, int64(3))

		data, err := ioutil.ReadAll(fd)
		require.Nil(t, err)
		require.Equal(t, data, rawData[3:])

		// Check that we can also seek back to start after reading to the end.
		// (and also check if the write overlay actually did work)
		pos, err = fd.Seek(0, os.SEEK_SET)
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
		t.Run(fmt.Sprintf("truncate(%d)", idx), func(t *testing.T) {
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
				require.Equal(t, data, rawData[:rawIdx])
				require.Nil(t, fd.Close())

				// Check if the result was really written:
				stream, err := fs.Cat("/x")
				require.Nil(t, err)

				persistentData, err := ioutil.ReadAll(stream)
				require.Nil(t, err)
				require.Equal(t, persistentData, rawData[:rawIdx])
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

		pos, err := fd.Seek(10, os.SEEK_SET)
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
