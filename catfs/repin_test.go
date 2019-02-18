package catfs

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRepinDepthOnly(t *testing.T) {
	withDummyFS(t, func(fs *FS) {
		fs.cfg.SetBool("repin.enabled", true)
		fs.cfg.SetString("repin.quota", "10G")
		fs.cfg.SetInt("repin.min_depth", 1)
		fs.cfg.SetInt("repin.max_depth", 10)

		testRun(t, fs, 10, 20)
	})
}

func TestRepinNoMaxDepth(t *testing.T) {
	withDummyFS(t, func(fs *FS) {
		fs.cfg.SetBool("repin.enabled", true)
		fs.cfg.SetString("repin.quota", "10G")
		fs.cfg.SetInt("repin.min_depth", 1)
		fs.cfg.SetInt("repin.max_depth", 100)

		testRun(t, fs, 20, 20)
	})
}

func TestRepinDisabled(t *testing.T) {
	withDummyFS(t, func(fs *FS) {
		fs.cfg.SetBool("repin.enabled", false)
		testRun(t, fs, 20, 20)
	})
}

func TestRepinQuota(t *testing.T) {
	withDummyFS(t, func(fs *FS) {
		fs.cfg.SetBool("repin.enabled", true)
		fs.cfg.SetString("repin.quota", "11B")
		fs.cfg.SetInt("repin.min_depth", 1)
		fs.cfg.SetInt("repin.max_depth", 100)

		testRun(t, fs, 10, 20)
	})
}

func TestRepinKillAll(t *testing.T) {
	withDummyFS(t, func(fs *FS) {
		fs.cfg.SetBool("repin.enabled", true)
		fs.cfg.SetString("repin.quota", "0B")
		fs.cfg.SetInt("repin.min_depth", 0)
		fs.cfg.SetInt("repin.max_depth", 0)

		testRun(t, fs, -1, 20)
	})
}

func TestRepinOldBehaviour(t *testing.T) {
	withDummyFS(t, func(fs *FS) {
		fs.cfg.SetBool("repin.enabled", true)
		fs.cfg.SetString("repin.quota", "100G")
		fs.cfg.SetInt("repin.min_depth", 1)
		fs.cfg.SetInt("repin.max_depth", 1)

		testRun(t, fs, 1, 20)
	})
}

func testRun(t *testing.T, fs *FS, split, n int) {
	for idx := 0; idx < n; idx++ {
		require.Nil(t, fs.Stage("/dir/a", bytes.NewReader([]byte{byte(idx)})))
		require.Nil(t, fs.MakeCommit(fmt.Sprintf("state: %d", idx)))
	}

	for idx := 0; idx < n; idx++ {
		require.Nil(t, fs.Pin("/dir/a", "HEAD"+strings.Repeat("^", idx), false))
	}

	require.Nil(t, fs.repin("/"))

	histA, err := fs.History("/dir/a")
	require.Nil(t, err)

	for idx := 0; idx <= split; idx++ {
		require.True(t, histA[idx].IsPinned, fmt.Sprintf("%d", idx))
		require.False(t, histA[idx].IsExplicit, fmt.Sprintf("%d", idx))
	}

	for idx := split + 1; idx < n; idx++ {
		require.False(t, histA[idx].IsPinned, fmt.Sprintf("%d", idx))
		require.False(t, histA[idx].IsExplicit, fmt.Sprintf("%d", idx))
	}
}
