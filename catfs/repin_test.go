package catfs

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRepin(t *testing.T) {
	withDummyFS(t, func(fs *FS) {
		fs.cfg.SetBool("repin.enabled", true)
		fs.cfg.SetString("repin.quota", "10B")
		fs.cfg.SetInt("repin.min_depth", 1)
		fs.cfg.SetInt("repin.max_depth", 10)

		for idx := 0; idx < 20; idx++ {
			require.Nil(t, fs.Stage("/dir/a", bytes.NewReader([]byte{byte(idx)})))
			require.Nil(t, fs.Stage("/dir/b", bytes.NewReader([]byte{byte(255 - idx)})))
			require.Nil(t, fs.MakeCommit(fmt.Sprintf("state: %d", idx)))
		}

		for idx := 0; idx < 20; idx++ {
			require.Nil(t, fs.Pin("/dir/a", "HEAD"+strings.Repeat("^", idx)))
			require.Nil(t, fs.Pin("/dir/b", "HEAD"+strings.Repeat("^", idx)))
		}

		require.Nil(t, fs.repin())
	})
}
