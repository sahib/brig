package pagecache

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"testing"

	"github.com/sahib/brig/catfs/mio/pagecache/mdcache"
	"github.com/sahib/brig/catfs/mio/pagecache/page"
	"github.com/sahib/brig/util"
	"github.com/sahib/brig/util/testutil"
	"github.com/stretchr/testify/require"
)

func withLayer(t *testing.T, size int64, fn func(expected []byte, p *Layer)) {
	md, err := mdcache.New(mdcache.Options{
		MaxMemoryUsage: 4 * page.Size,
		SwapDirectory:  "",
	})

	require.NoError(t, err)

	data := testutil.CreateDummyBuf(size)
	p, err := NewLayer(bytes.NewReader(data), md, 42, size)
	require.NoError(t, err)

	expected := make([]byte, size)
	copy(expected, data)
	fn(expected, p)

	require.NoError(t, md.Close())
}

var (
	testSizes = []int64{
		16*page.Size + 0,
		16*page.Size - 1,
		16*page.Size + 1,
		page.Size + 0,
		page.Size - 1,
		page.Size + 1,
		0,
		1,
	}
)

func TestReadOnly(t *testing.T) {
	for _, testSize := range testSizes {
		t.Run(fmt.Sprintf("%d", testSize), func(t *testing.T) {
			withLayer(t, testSize, func(expected []byte, p *Layer) {
				got := bytes.NewBuffer([]byte{})
				n, err := p.WriteTo(got)
				require.NoError(t, err)
				require.Equal(t, testSize, n)
				require.Equal(t, expected, got.Bytes())
			})
		})
	}
}

func padOrCutToLength(buf []byte, length int64) []byte {
	if int64(len(buf)) >= length {
		return buf[:length]
	}

	c := make([]byte, length)
	copy(c, buf)
	return c
}

func TestReadOnlyTruncate(t *testing.T) {
	truncOffsets := []int64{
		-2*page.Size + 0,
		-2*page.Size - 1,
		-2*page.Size + 1,
		+2*page.Size + 0,
		+2*page.Size - 1,
		+2*page.Size + 1,
		+page.Size + 0,
		+page.Size - 1,
		+page.Size + 1,
		-page.Size + 0,
		-page.Size - 1,
		-page.Size + 1,
		+0,
		+1,
		-1,
	}

	for _, testSize := range testSizes {
		t.Run(fmt.Sprintf("%d", testSize), func(t *testing.T) {
			for _, truncOff := range truncOffsets {
				length := util.Max64(0, testSize+truncOff)
				if length == testSize {
					// no need to run test with no truncation.
					// already covered by TestReadOnly()
					continue
				}

				t.Run(fmt.Sprintf("trunc-to-%d", length), func(t *testing.T) {
					withLayer(t, testSize, func(expected []byte, p *Layer) {
						got := bytes.NewBuffer([]byte{})
						p.Truncate(length)

						n, err := p.WriteTo(got)
						require.NoError(t, err)
						require.Equal(t, length, n)

						res := padOrCutToLength(got.Bytes(), length)
						require.Equal(
							t,
							padOrCutToLength(expected, length),
							res,
						)
					})
				})
			}
		})
	}
}

func TestWriteSingle(t *testing.T) {
	for _, testReadSize := range testSizes {
		t.Run(fmt.Sprintf("read-%d", testReadSize), func(t *testing.T) {
			for _, testWriteSize := range testSizes {
				t.Run(fmt.Sprintf("write-%d", testWriteSize), func(t *testing.T) {
					withLayer(t, testReadSize, func(expected []byte, p *Layer) {
						expected = testutil.CreateRandomDummyBuf(testWriteSize, 23)
						wn, err := p.WriteAt(expected, 0)
						require.NoError(t, err)
						require.Equal(t, int64(wn), testWriteSize)

						got := make([]byte, testWriteSize)
						rn, err := p.Read(got)
						if testReadSize == 0 {
							// special case: that will immediately return EOF.
							require.Error(t, io.EOF, err)
							return
						}

						require.NoError(t, err)
						require.Equal(t, wn, rn)
					})
				})
			}
		})
	}
}

func TestWriteRandomOffset(t *testing.T) {
	// Randomly generate writes and write them to the layer.
	// The randomness is controlled by seed to be reproducible.
	// The generated data is also copy()'d to a slice which
	// serves as way to check the overlay on the final read.

	for seed := 0; seed < 40; seed++ {
		t.Run(fmt.Sprintf("seed-%d", seed), func(t *testing.T) {
			for _, testReadSize := range testSizes {
				if testReadSize == 0 {
					continue
				}

				t.Run(fmt.Sprintf("size-%d", testReadSize*2), func(t *testing.T) {
					withLayer(t, testReadSize, func(expected []byte, p *Layer) {

						// NOTE: We do not write beyond p.Length()
						// to make this test easier to check.
						p.Truncate(testReadSize * 2)

						expected = padOrCutToLength(expected, p.Length())
						require.Equal(t, testReadSize*2, p.Length())

						rand.Seed(int64(seed))
						for nwrites := 0; nwrites < seed; nwrites++ {
							writeOff := rand.Int63n(p.Length())
							writeLen := rand.Int63n(p.Length() - writeOff + 1)

							// stream contains 0-254 data, overwrite with random:
							buf := testutil.CreateRandomDummyBuf(writeLen, int64(seed))
							copy(expected[writeOff:writeOff+writeLen], buf)
							wn, err := p.WriteAt(buf, writeOff)
							require.NoError(t, err)
							require.Equal(t, int(writeLen), wn)
						}

						got := &bytes.Buffer{}
						rn, err := io.Copy(got, p)
						require.NoError(t, err)
						require.Equal(t, p.Length(), int64(rn))
						require.Equal(t, p.Length(), int64(len(expected)))
						require.Equal(t, p.Length(), int64(got.Len()))

						// This for loop here is just for easier digest
						// debug output. require.Equal() outputs huge
						// diffs that are seldom helpful.
						for idx := 0; idx < got.Len(); idx++ {
							if expected[idx] != got.Bytes()[idx] {
								require.Equal(
									t,
									expected[idx:idx+256],
									got.Bytes()[idx:idx+256],
								)
								return
							}
						}

						// This assert is just here in case the for loop
						// above has a bug or gets lost somehow.
						require.Equal(t, expected, got.Bytes())
					})
				})
			}
		})
	}
}

func TestReadRandomOffset(t *testing.T) {
	tcs := []struct {
		size   int64
		length int64
		nops   int
	}{
		{page.Size, page.Size, 10},
		{2 * page.Size, page.Size, 20},
		{page.Size, 2 * page.Size, 30},

		{16 * page.Size, 16 * page.Size, 30},
		{16 * page.Size, 16*page.Size + 1, 20},
		{16*page.Size + 1, 16 * page.Size, 10},
	}
	//		16*page.Size + 0,
	//		16*page.Size - 1,
	//		16*page.Size + 1,
	//		page.Size + 0,
	//		page.Size - 1,
	//		page.Size + 1,
	//		0,
	//		1,

	for _, tc := range tcs {
		name := fmt.Sprintf("s%d-l%d-o%d", tc.size, tc.length, tc.nops)
		t.Run(name, func(t *testing.T) {
			// always use the same writing distribution,
			// just increasingly more writes:

			withLayer(t, tc.size, func(expected []byte, p *Layer) {
				p.Truncate(tc.length)
				expected = padOrCutToLength(expected, p.Length())
				require.Equal(t, tc.length, p.Length())
				rand.Seed(42)

				overlayed := make([]byte, len(expected))
				copy(overlayed, expected)

				for idx := 0; idx < tc.nops; idx++ {
					writeOff := rand.Int63n(p.Length())
					writeLen := rand.Int63n(p.Length() - writeOff + 1)

					// stream contains 0-254 data, overwrite with random:
					buf := testutil.CreateRandomDummyBuf(writeLen, int64(42))
					copy(overlayed[writeOff:writeOff+writeLen], buf)
					wn, err := p.WriteAt(buf, writeOff)
					require.NoError(t, err)
					require.Equal(t, int(writeLen), wn)
				}

				for seed := 0; seed < 40; seed++ {
					rand.Seed(int64(seed))
					t.Run(fmt.Sprintf("seed-%d", seed), func(t *testing.T) {
						for idx := 0; idx < tc.nops; idx++ {
							readOff := rand.Int63n(p.Length())
							readLen := rand.Int63n(p.Length() - readOff + 1)
							readBuf := make([]byte, readLen)

							rn, err := p.ReadAt(readBuf, readOff)
							require.NoError(t, err)

							require.Equal(t, readLen, int64(rn))
							for idx := int64(0); idx < readLen; idx++ {
								if overlayed[readOff+idx] != readBuf[idx] {
									require.Failf(t, "bad data read", "first wrong offset: %d", readOff+idx)
									return
								}
							}

							require.Equal(t, overlayed[readOff:readOff+readLen], readBuf)
						}
					})
				}
			})
		})
	}
}
