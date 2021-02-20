package overlay

import (
	"bytes"
	"io"
	"testing"

	"github.com/sahib/brig/util"
	"github.com/sahib/brig/util/testutil"
	"github.com/stretchr/testify/require"
)

func TestZeroPaddedReader(t *testing.T) {
	tcs := []struct {
		name              string
		off, length, size int64
	}{
		{
			name:   "usual-case",
			off:    0,
			length: 1024,
			size:   512,
		}, {
			name:   "truncate-short",
			off:    0,
			length: 512,
			size:   1024,
		}, {
			name:   "equal",
			off:    0,
			length: 1024,
			size:   1024,
		}, {
			name:   "zero",
			off:    0,
			length: 0,
			size:   0,
		},
	}

	const maxSize = 4 * 1024
	data := testutil.CreateDummyBuf(maxSize)
	zero := make([]byte, maxSize)

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			zpr := &zeroPadReader{
				r:      bytes.NewReader(data),
				off:    tc.off,
				size:   tc.size,
				length: tc.length,
			}

			out := &bytes.Buffer{}
			n, err := io.Copy(out, zpr)
			require.NoError(t, err)
			require.Equal(t, tc.length, n)

			a := util.Min64(tc.size, tc.length)
			b := util.Max64(tc.size, tc.length)

			outData := out.Bytes()
			require.Equal(t, data[0:a], outData[0:a])
			require.Equal(t, zero[a:b], outData[a:b])
		})
	}
}

func TestIOBuf(t *testing.T) {
	tcs := []struct {
		name    string
		max     int
		srcSize int
		dstSize int
	}{
		{
			name:    "all-equal",
			max:     1024,
			srcSize: 1024,
			dstSize: 1024,
		}, {
			name:    "src<max<dst",
			max:     1024,
			srcSize: 512,
			dstSize: 2048,
		}, {
			name:    "max<src<dst",
			max:     512,
			srcSize: 1024,
			dstSize: 2048,
		}, {
			name:    "zero",
			max:     0,
			srcSize: 0,
			dstSize: 0,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			src := testutil.CreateDummyBuf(int64(tc.srcSize))
			ib := &iobuf{
				dst: make([]byte, tc.dstSize),
				max: tc.max,
			}

			n, err := ib.Write(src)
			require.NoError(t, err)
			require.Equal(t, n, ib.Len())

			if diff := tc.max - tc.srcSize; diff <= 0 {
				require.Equal(t, n, tc.max)
				require.Equal(t, 0, ib.Left(), "too many bytes left")
			} else {
				require.Equal(t, n, tc.srcSize)
				require.Equal(t, diff, ib.Left(), "wrong number of bytes left")
			}
		})
	}
}
