package config

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaults(t *testing.T) {
	require.Equal(t, getDefaultByKey("daemon.port", Defaults).Default, 6666)
	require.Nil(t, getDefaultByKey("daemon.port.sub", Defaults))
	require.Nil(t, getDefaultByKey("daemon.xxx", Defaults))
	require.Nil(t, getDefaultByKey("daemon", Defaults))
}

func TestDefaultsType(t *testing.T) {
	require.Equal(t, "int", getTypeOfDefaultKey("daemon.port", Defaults))
	require.Equal(t, "string", getTypeOfDefaultKey("data.ipfs.path", Defaults))
	require.Equal(t, "", getTypeOfDefaultKey("not.yet.there", Defaults))
}

var testConfig = `daemon:
  port: 6667
data:
  ipfs:
    path: x
`

func TestGetDefault(t *testing.T) {
	cfg, err := Open(bytes.NewReader([]byte(testConfig)), Defaults)
	require.Nil(t, err)

	require.Equal(t, int64(6667), cfg.Int("daemon.port"))
}

func TestGetNonExisting(t *testing.T) {
	defer func() { require.NotNil(t, recover()) }()

	cfg, err := Open(bytes.NewReader([]byte(testConfig)), Defaults)
	require.Nil(t, err)

	// should panic.
	cfg.SetFloat("not.existing", 23)
}

func TestSet(t *testing.T) {
	cfg, err := Open(bytes.NewReader([]byte(testConfig)), Defaults)
	require.Nil(t, err)

	cfg.SetInt("daemon.port", 6666)
	require.Equal(t, int64(6666), cfg.Int("daemon.port"))
}

func TestSetBucketKey(t *testing.T) {
	defer func() { require.NotNil(t, recover()) }()

	cfg, err := Open(bytes.NewReader([]byte(testConfig)), Defaults)
	require.Nil(t, err)

	// should panic.
	cfg.SetString("daemon", "oh oh")
}

func TestAddChangeSignal(t *testing.T) {
	cfg, err := Open(bytes.NewReader([]byte(testConfig)), Defaults)
	require.Nil(t, err)

	callCount := 0
	cbID := cfg.AddChangedKeyEvent("data.ipfs.path", func(key string) {
		require.Equal(t, "new-value", cfg.String("data.ipfs.path"))
		callCount++
	})

	require.Equal(t, 0, callCount)
	cfg.SetInt("daemon.port", 42)
	require.Equal(t, 0, callCount)
	cfg.SetString("data.ipfs.path", "new-value")
	require.Equal(t, 1, callCount)
	cfg.SetString("data.ipfs.path", "new-value")
	require.Equal(t, 1, callCount)

	require.Nil(t, cfg.RemoveChangedKeyEvent(cbID))
	cfg.SetString("data.ipfs.path", "newer-value")
	require.Equal(t, 1, callCount)
}

func TestAddChangeSignalAll(t *testing.T) {
	cfg, err := Open(bytes.NewReader([]byte(testConfig)), Defaults)
	require.Nil(t, err)

	callCount := 0
	cbID := cfg.AddChangedKeyEvent("", func(key string) {
		callCount++
	})

	require.Equal(t, 0, callCount)
	cfg.SetInt("daemon.port", 42)
	require.Equal(t, 1, callCount)
	cfg.SetString("data.ipfs.path", "new-value")
	require.Equal(t, 2, callCount)

	require.Nil(t, cfg.RemoveChangedKeyEvent(cbID))
	cfg.SetString("data.ipfs.path", "newer-value")
	require.Equal(t, 2, callCount)
}

func TestOpenSave(t *testing.T) {
	cfg, err := Open(bytes.NewReader([]byte(testConfig)), Defaults)
	require.Nil(t, err)

	cfg.SetInt("daemon.port", 6666)
	cfg.SetString("data.ipfs.path", "y")

	buf := &bytes.Buffer{}
	require.Nil(t, cfg.Save(buf))

	newCfg, err := Open(buf, Defaults)
	require.Nil(t, err)

	require.Equal(t, int64(6666), newCfg.Int("daemon.port"))
	require.Equal(t, "y", newCfg.String("data.ipfs.path"))
}

func TestKeys(t *testing.T) {
	cfg, err := Open(bytes.NewReader([]byte(testConfig)), Defaults)
	require.Nil(t, err)

	keys, err := cfg.Keys()
	require.Nil(t, err)
	require.Equal(t, []string{
		"daemon.port",
		"data.compress.default_algo",
		"data.ipfs.path",
		"sync.conflict_strategy",
		"sync.ignored_removed",
	}, keys)
}

func TestAddExtraKeys(t *testing.T) {
	// There is no default for "a: 1" -> fail.
	_, err := Open(bytes.NewReader([]byte(`a: 1`)), Defaults)
	require.NotNil(t, err)
}

func TestSection(t *testing.T) {
	cfg, err := Open(bytes.NewReader([]byte(testConfig)), Defaults)
	require.Nil(t, err)

	dataSec := cfg.Section("data")
	require.Equal(t, "snappy", dataSec.String("compress.default_algo"))
	require.Equal(t, "snappy", cfg.String("data.compress.default_algo"))

	dataSec.SetString("compress.default_algo", "lz4")
	require.Equal(t, "lz4", dataSec.String("compress.default_algo"))
	require.Equal(t, "lz4", cfg.String("data.compress.default_algo"))

	cfg.SetString("data.compress.default_algo", "none")
	require.Equal(t, "none", dataSec.String("compress.default_algo"))
	require.Equal(t, "none", cfg.String("data.compress.default_algo"))

	childKeys, err := dataSec.Keys()
	require.Nil(t, err)
	require.Equal(t, []string{
		"data.compress.default_algo",
		"data.ipfs.path",
	}, childKeys)
}

func TestSectionSignals(t *testing.T) {
	cfg, err := Open(bytes.NewReader([]byte(testConfig)), Defaults)
	require.Nil(t, err)

	parentCallCount := 0
	parentID := cfg.AddChangedKeyEvent("data.compress.default_algo", func(key string) {
		require.Equal(t, "data.compress.default_algo", key)
		parentCallCount++
	})

	dataSec := cfg.Section("data")

	childCallCount := 0
	childID := dataSec.AddChangedKeyEvent("compress.default_algo", func(key string) {
		require.Equal(t, "compress.default_algo", key)
		childCallCount++
	})

	cfg.SetString("data.compress.default_algo", "none")
	dataSec.SetString("compress.default_algo", "lz4")

	require.Equal(t, 2, parentCallCount)
	require.Equal(t, 1, childCallCount)

	require.Nil(t, dataSec.RemoveChangedKeyEvent(childID))
	require.Nil(t, cfg.RemoveChangedKeyEvent(parentID))

}
