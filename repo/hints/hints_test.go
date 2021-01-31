package hints

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHintManager(t *testing.T) {
	mgr, err := NewManager(nil)
	require.NoError(t, err)

	expect := Hint{
		CompressionAlgo: CompressionLZ4,
		EncryptionAlgo:  EncryptionNone,
	}

	mgr.Set("/a/b/c", expect)
	hint := mgr.Lookup("/a/b/c/d")
	require.Equal(t, expect, hint)

	require.Equal(t, map[string]Hint{
		"/":      Default(),
		"/a/b/c": expect,
	}, mgr.List())

	yamlBuf := bytes.NewBuffer(nil)
	require.NoError(t, mgr.Save(yamlBuf))
	oldYaml := yamlBuf.String()

	// Check if a freshly loaded one behaves exactly same:
	newMgr, err := NewManager(yamlBuf)
	require.NoError(t, err)

	newHint := newMgr.Lookup("/a/b/c/d")
	require.Equal(t, expect, newHint)

	newYamlBuf := bytes.NewBuffer(nil)
	require.NoError(t, newMgr.Save(newYamlBuf))
	require.Equal(t, oldYaml, newYamlBuf.String())

	require.Equal(t, ErrNoSuchHint, newMgr.Remove("/a/b/c/d"))
	require.NoError(t, newMgr.Remove("/a/b/c"))
	require.Equal(t, Default(), newMgr.Lookup("/a/b/c/d"))
}
