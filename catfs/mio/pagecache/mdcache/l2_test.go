package mdcache

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/sahib/brig/catfs/mio/pagecache/page"
	"github.com/sahib/brig/util/testutil"
	"github.com/stretchr/testify/require"
)

func dummyPage(off, length uint32) *page.Page {
	buf := testutil.CreateDummyBuf(int64(length))
	return page.New(off, buf)
}

func withL2Cache(t *testing.T, fn func(l2 *l2cache)) {
	for _, compress := range []bool{false, true} {

		tmpDir, err := ioutil.TempDir("", "brig-page-l2")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		l2, err := newL2Cache(tmpDir, compress)
		require.NoError(t, err)

		tname := "no-compress"
		if compress {
			tname = "compress"
		}

		t.Run(tname, func(t *testing.T) {
			fn(l2)
		})

		// double check we do not waste any storage:
		require.NoError(t, l2.Close())
		_, err = os.Stat(tmpDir)
		require.True(t, os.IsNotExist(err))
	}
}

func TestL2GetSetDel(t *testing.T) {
	withL2Cache(t, func(l2 *l2cache) {
		pk := pageKey{1, 0}
		_, err := l2.Get(pk)
		require.Error(t, page.ErrCacheMiss)

		pset := dummyPage(0, 1024)
		require.NoError(t, l2.Set(pk, pset))

		pgot, err := l2.Get(pk)
		require.NoError(t, err)

		require.Equal(t, pset.Data, pgot.Data)
		require.Equal(t, pset.Extents, pgot.Extents)

		l2.Del([]pageKey{pk})
		_, err = l2.Get(pk)
		require.Error(t, page.ErrCacheMiss)
	})
}

func TestL2Nil(t *testing.T) {
	// l2 is optional, so a nil l2 cache should "work":
	l2, err := newL2Cache("", false)
	require.NoError(t, err)

	_, err = l2.Get(pageKey{0, 1})
	require.Error(t, page.ErrCacheMiss)

	require.NoError(t, l2.Set(pageKey{0, 1}, dummyPage(0, 1024)))
	l2.Del([]pageKey{{0, 1}})
}
