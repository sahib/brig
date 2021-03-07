package mdcache

import (
	"testing"

	"github.com/sahib/brig/catfs/mio/pagecache/page"
	"github.com/stretchr/testify/require"
)

func withL1Cache(t *testing.T, fn func(l1, backing *l1cache)) {
	// some fake in-mem cache that stores everything that got removed
	// out of l1 due to size restrictions.
	backing, err := newL1Cache(nil, int64(^uint64(0)>>1))
	require.NoError(t, err)

	l1, err := newL1Cache(backing, 4*(page.Size+page.Meta))
	require.NoError(t, err)

	fn(l1, backing)

	require.NoError(t, l1.Close())
}

func TestL1GetSetDel(t *testing.T) {
	// NOTE: Only covers the very basic usage.
	withL1Cache(t, func(l1, _ *l1cache) {
		pk := pageKey{1, 0}
		_, err := l1.Get(pk)
		require.Error(t, page.ErrCacheMiss)

		pset := dummyPage(0, 1024)
		require.NoError(t, l1.Set(pk, pset))

		pgot, err := l1.Get(pk)
		require.NoError(t, err)

		require.Equal(t, pset.Data, pgot.Data)
		require.Equal(t, pset.Extents, pgot.Extents)

		l1.Del([]pageKey{pk})
		_, err = l1.Get(pk)
		require.Error(t, page.ErrCacheMiss)
	})
}

func TestL1SwapPriority(t *testing.T) {
	withL1Cache(t, func(l1, backing *l1cache) {
		// Insert 8 pages, only 4 can stay in l1.
		for idx := 0; idx < 8; idx++ {
			pk := pageKey{1, uint32(idx)}
			require.NoError(t, l1.Set(pk, dummyPage(0, uint32((idx+1)*100))))
		}

		for idx := 0; idx < 4; idx++ {
			pk := pageKey{1, uint32(idx)}
			_, err := l1.Get(pk)
			require.Error(t, err, page.ErrCacheMiss)

			// should be in backing store, check:
			p, err := backing.Get(pk)
			require.NoError(t, err)

			expected := dummyPage(0, uint32((idx+1)*100))
			require.Equal(t, expected, p)
		}

		for idx := 4; idx < 8; idx++ {
			pk := pageKey{1, uint32(idx)}
			p, err := l1.Get(pk)
			require.NoError(t, err)

			expected := dummyPage(0, uint32((idx+1)*100))
			require.Equal(t, expected, p)
		}
	})
}
