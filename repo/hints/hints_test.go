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

	mgr.Remember("/a/b/c", expect)
	hint := mgr.Lookup("/a/b/c/d")
	require.Equal(t, expect, hint)

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
}
