package mdcache

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/sahib/brig/catfs/mio/pagecache/page"
	"github.com/sahib/brig/util/testutil"
	"github.com/stretchr/testify/require"
)

func withMDCache(t *testing.T, fn func(mdc *MDCache)) {
	tmpDir, err := ioutil.TempDir("", "brig-page-l2")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	md, err := New(Options{
		MaxMemoryUsage:    4 * page.Size,
		SwapDirectory:     tmpDir,
		L1CacheMissRefill: true,
	})

	require.NoError(t, err)

	fn(md)

	require.NoError(t, md.Close())
}

func TestMDBasic(t *testing.T) {
	withMDCache(t, func(mdc *MDCache) {
		for idx := 0; idx < 8; idx++ {
			err := mdc.Merge(1, uint32(idx), 0, testutil.CreateDummyBuf(page.Size))
			require.NoError(t, err)
		}

		for idx := 0; idx < 8; idx++ {
			p, err := mdc.Lookup(1, uint32(idx))
			require.NoError(t, err)

			require.Equal(t, testutil.CreateDummyBuf(page.Size), p.Data)
			require.Equal(t, []page.Extent{{
				OffLo: 0,
				OffHi: page.Size,
			}}, p.Extents)
		}

		require.NoError(t, mdc.Evict(1, 8*page.Size))
	})
}
