package util

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sahib/brig/util/testutil"
	"github.com/stretchr/testify/require"
)

func fileIsSame(t *testing.T, a, b string) {
	data, err := ioutil.ReadFile(a)
	require.NoError(t, err)

	datb, err := ioutil.ReadFile(b)
	require.NoError(t, err)

	require.Equal(t, data, datb)
}

func TestTarUntar(t *testing.T) {
	tmpDirPack, err := ioutil.TempDir("", "brig-taruntar-pack-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDirPack)

	tmpDirUnpack, err := ioutil.TempDir("", "brig-taruntar-unpack-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDirUnpack)

	require.NoError(
		t,
		os.MkdirAll(filepath.Join(tmpDirPack, "sub"), 0700),
	)

	expectedPaths := []string{"a", "b", "sub/c", "sub/d"}
	for idx, path := range expectedPaths {
		data := testutil.CreateRandomDummyBuf(16*1024, int64(idx%2))
		require.NoError(
			t,
			ioutil.WriteFile(filepath.Join(tmpDirPack, path), data, 0600),
		)
	}

	buf := &bytes.Buffer{}
	require.NoError(t, Tar(tmpDirPack, "test-archive", buf))

	gotPaths := []string{}
	require.NoError(t, Untar(buf, tmpDirUnpack))
	require.NoError(
		t,
		filepath.Walk(tmpDirUnpack, func(path string, info os.FileInfo, err error) error {
			require.NoError(t, err)
			if !info.Mode().IsRegular() {
				return nil
			}
			path = path[len(tmpDirUnpack):]
			path = strings.TrimLeftFunc(path, func(r rune) bool {
				return r == filepath.Separator
			})

			gotPaths = append(gotPaths, path)
			return nil
		}),
	)

	require.Equal(t, expectedPaths, gotPaths)
	for idx, expectedPath := range expectedPaths {
		fileIsSame(
			t,
			filepath.Join(tmpDirPack, expectedPath),
			filepath.Join(tmpDirUnpack, gotPaths[idx]),
		)
	}
}
