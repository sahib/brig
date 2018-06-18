package config

import (
	"bytes"
	"os"
	"testing"

	"github.com/sahib/brig/util/testutil"
	"github.com/stretchr/testify/require"
)

var TestDefaults = DefaultMapping{
	"daemon": DefaultMapping{
		"port": DefaultEntry{
			Default:      6666,
			NeedsRestart: true,
			Docs:         "Port of the daemon process",
			Validator:    IntRangeValidator(1, 655356),
		},
	},
	"fs": DefaultMapping{
		"sync": DefaultMapping{
			"ignore_removed": DefaultEntry{
				Default:      false,
				NeedsRestart: false,
				Docs:         "Do not remove what the remote removed",
			},
			"ignore_moved": DefaultEntry{
				Default:      false,
				NeedsRestart: false,
				Docs:         "Do not move what the remote moved",
			},
			"conflict_strategy": DefaultEntry{
				Default:      "marker",
				NeedsRestart: false,
				Validator: EnumValidator(
					"marker", "ignore",
				),
			},
		},
		"compress": DefaultMapping{
			"default_algo": DefaultEntry{
				Default:      "snappy",
				NeedsRestart: false,
				Docs:         "What compression algorithm to use by default",
				Validator: EnumValidator(
					"snappy", "lz4", "none",
				),
			},
		},
	},
	"repo": DefaultMapping{
		"current_user": DefaultEntry{
			Default:      "",
			NeedsRestart: true,
			Docs:         "The repository owner that is published to the outside",
		},
	},
	"data": DefaultMapping{
		"ipfs": DefaultMapping{
			"path": DefaultEntry{
				Default:      "",
				NeedsRestart: true,
				Docs:         "Root directory of the ipfs repository",
			},
		},
	},
}

func getTypeOfDefaultKey(key string, defaults DefaultMapping) string {
	defaultEntry := getDefaultByKey(key, defaults)
	if defaultEntry == nil {
		return ""
	}

	return getTypeOf(defaultEntry.Default)
}

func TestGetDefaults(t *testing.T) {
	require.Equal(t, getDefaultByKey("daemon.port", TestDefaults).Default, 6666)
	require.Nil(t, getDefaultByKey("daemon.port.sub", TestDefaults))
	require.Nil(t, getDefaultByKey("daemon.xxx", TestDefaults))
	require.Nil(t, getDefaultByKey("daemon", TestDefaults))
}

func TestDefaultsType(t *testing.T) {
	require.Equal(t, "int", getTypeOfDefaultKey("daemon.port", TestDefaults))
	require.Equal(t, "string", getTypeOfDefaultKey("data.ipfs.path", TestDefaults))
	require.Equal(t, "", getTypeOfDefaultKey("not.yet.there", TestDefaults))
}

var testConfig = `daemon:
  port: 6667
data:
  ipfs:
    path: x
`

func TestGetDefault(t *testing.T) {
	cfg, err := Open(bytes.NewReader([]byte(testConfig)), TestDefaults)
	require.Nil(t, err)

	require.Equal(t, int64(6667), cfg.Int("daemon.port"))
}

func TestGetNonExisting(t *testing.T) {
	defer func() { require.NotNil(t, recover()) }()

	cfg, err := Open(bytes.NewReader([]byte(testConfig)), TestDefaults)
	require.Nil(t, err)

	// should panic.
	require.Nil(t, cfg.SetFloat("not.existing", 23))
}

func TestSet(t *testing.T) {
	cfg, err := Open(bytes.NewReader([]byte(testConfig)), TestDefaults)
	require.Nil(t, err)

	require.Nil(t, cfg.SetInt("daemon.port", 6666))
	require.Equal(t, int64(6666), cfg.Int("daemon.port"))
}

func TestSetBucketKey(t *testing.T) {
	defer func() { require.NotNil(t, recover()) }()

	cfg, err := Open(bytes.NewReader([]byte(testConfig)), TestDefaults)
	require.Nil(t, err)

	// should panic.
	require.Nil(t, cfg.SetString("daemon", "oh oh"))
}

func TestAddChangeSignal(t *testing.T) {
	cfg, err := Open(bytes.NewReader([]byte(testConfig)), TestDefaults)
	require.Nil(t, err)

	callCount := 0
	cbID := cfg.AddChangedKeyEvent("data.ipfs.path", func(key string) {
		require.Equal(t, "new-value", cfg.String("data.ipfs.path"))
		callCount++
	})

	require.Equal(t, 0, callCount)
	require.Nil(t, cfg.SetInt("daemon.port", 42))
	require.Equal(t, 0, callCount)
	require.Nil(t, cfg.SetString("data.ipfs.path", "new-value"))
	require.Equal(t, 1, callCount)
	require.Nil(t, cfg.SetString("data.ipfs.path", "new-value"))
	require.Equal(t, 1, callCount)

	require.Nil(t, cfg.RemoveChangedKeyEvent(cbID))
	require.Nil(t, cfg.SetString("data.ipfs.path", "newer-value"))
	require.Equal(t, 1, callCount)
}

func TestAddChangeSignalAll(t *testing.T) {
	cfg, err := Open(bytes.NewReader([]byte(testConfig)), TestDefaults)
	require.Nil(t, err)

	callCount := 0
	cbID := cfg.AddChangedKeyEvent("", func(key string) {
		callCount++
	})

	require.Equal(t, 0, callCount)
	require.Nil(t, cfg.SetInt("daemon.port", 42))
	require.Equal(t, 1, callCount)
	require.Nil(t, cfg.SetString("data.ipfs.path", "new-value"))
	require.Equal(t, 2, callCount)

	require.Nil(t, cfg.RemoveChangedKeyEvent(cbID))
	require.Nil(t, cfg.SetString("data.ipfs.path", "newer-value"))
	require.Equal(t, 2, callCount)
}

func TestOpenSave(t *testing.T) {
	cfg, err := Open(bytes.NewReader([]byte(testConfig)), TestDefaults)
	require.Nil(t, err)

	require.Nil(t, cfg.SetInt("daemon.port", 6666))
	require.Nil(t, cfg.SetString("data.ipfs.path", "y"))

	buf := &bytes.Buffer{}
	require.Nil(t, cfg.Save(buf))

	newCfg, err := Open(buf, TestDefaults)
	require.Nil(t, err)

	require.Equal(t, int64(6666), newCfg.Int("daemon.port"))
	require.Equal(t, "y", newCfg.String("data.ipfs.path"))
}

func TestKeys(t *testing.T) {
	cfg, err := Open(bytes.NewReader([]byte(testConfig)), TestDefaults)
	require.Nil(t, err)

	keys := cfg.Keys()
	require.Equal(t, []string{
		"daemon.port",
		"data.ipfs.path",
		"fs.compress.default_algo",
		"fs.sync.conflict_strategy",
		"fs.sync.ignore_moved",
		"fs.sync.ignore_removed",
		"repo.current_user",
	}, keys)
}

func TestAddExtraKeys(t *testing.T) {
	// There is no default for "a: 1" -> fail.
	_, err := Open(bytes.NewReader([]byte(`a: 1`)), TestDefaults)
	require.NotNil(t, err)
}

func TestSection(t *testing.T) {
	cfg, err := Open(bytes.NewReader([]byte(testConfig)), TestDefaults)
	require.Nil(t, err)

	fsSec := cfg.Section("fs")
	require.Equal(t, "snappy", fsSec.String("compress.default_algo"))
	require.Equal(t, "snappy", cfg.String("fs.compress.default_algo"))

	require.Nil(t, fsSec.SetString("compress.default_algo", "lz4"))
	require.Equal(t, "lz4", fsSec.String("compress.default_algo"))
	require.Equal(t, "lz4", cfg.String("fs.compress.default_algo"))

	require.Nil(t, cfg.SetString("fs.compress.default_algo", "none"))
	require.Equal(t, "none", fsSec.String("compress.default_algo"))
	require.Equal(t, "none", cfg.String("fs.compress.default_algo"))

	childKeys := fsSec.Keys()
	require.Equal(t, []string{
		"fs.compress.default_algo",
		"fs.sync.conflict_strategy",
		"fs.sync.ignore_moved",
		"fs.sync.ignore_removed",
	}, childKeys)
}

func TestSectionSignals(t *testing.T) {
	cfg, err := Open(bytes.NewReader([]byte(testConfig)), TestDefaults)
	require.Nil(t, err)

	parentCallCount := 0
	parentID := cfg.AddChangedKeyEvent("fs.compress.default_algo", func(key string) {
		require.Equal(t, "fs.compress.default_algo", key)
		parentCallCount++
	})

	fsSec := cfg.Section("fs")

	childCallCount := 0
	childID := fsSec.AddChangedKeyEvent("compress.default_algo", func(key string) {
		require.Equal(t, "compress.default_algo", key)
		childCallCount++
	})

	require.Nil(t, cfg.SetString("fs.compress.default_algo", "none"))
	require.Nil(t, fsSec.SetString("compress.default_algo", "lz4"))

	require.Equal(t, 2, parentCallCount)
	require.Equal(t, 1, childCallCount)

	require.Nil(t, fsSec.RemoveChangedKeyEvent(childID))
	require.Nil(t, cfg.RemoveChangedKeyEvent(parentID))

}

func TestIsValidKey(t *testing.T) {
	cfg, err := Open(bytes.NewReader([]byte(testConfig)), TestDefaults)
	require.Nil(t, err)

	require.True(t, cfg.IsValidKey("daemon.port"))
	require.False(t, cfg.IsValidKey("data.port"))
}

func TestCast(t *testing.T) {
	defaults := DefaultMapping{
		"string": DefaultEntry{
			Default: "a",
		},
		"int": DefaultEntry{
			Default: 2,
		},
		"float": DefaultEntry{
			Default: 3.0,
		},
		"bool": DefaultEntry{
			Default: false,
		},
	}

	cfg, err := Open(bytes.NewReader(nil), defaults)
	require.Nil(t, err)

	// Same string cast:
	strCast, err := cfg.Cast("string", "test")
	require.Nil(t, err)
	require.Equal(t, "test", strCast)

	// Int cast:
	intCast, err := cfg.Cast("int", "123")
	require.Nil(t, err)
	require.Equal(t, int64(123), intCast)

	// Float cast:
	floatCast, err := cfg.Cast("float", "5.0")
	require.Nil(t, err)
	require.Equal(t, float64(5.0), floatCast)

	// Bool cast:
	boolCast, err := cfg.Cast("bool", "true")
	require.Nil(t, err)
	require.Equal(t, true, boolCast)

	// Wrong cast types:
	_, err = cfg.Cast("int", "im a string")
	require.NotNil(t, err)

	_, err = cfg.Cast("int", "2.0")
	require.NotNil(t, err)
}

func configMustEquals(t *testing.T, aCfg, bCfg *Config) {
	require.Equal(t, aCfg.Keys(), bCfg.Keys())
	for _, key := range aCfg.Keys() {
		require.Equal(t, aCfg.Get(key), bCfg.Get(key), key)
	}
}

func TestToFileFromFile(t *testing.T) {
	cfg, err := Open(bytes.NewReader(nil), TestDefaults)
	require.Nil(t, err)

	path := "/tmp/brig-test-config.yml"
	require.Nil(t, ToFile(path, cfg))

	defer os.Remove(path)

	loadCfg, err := FromFile(path, TestDefaults)
	require.Nil(t, err)

	configMustEquals(t, cfg, loadCfg)
}

func TestSetIncompatibleType(t *testing.T) {
	cfg, err := Open(bytes.NewReader(nil), TestDefaults)
	require.Nil(t, err)

	require.NotNil(t, cfg.SetString("daemon.port", "xxx"))
}

func TestVersionPersisting(t *testing.T) {
	cfg, err := Open(bytes.NewReader(nil), TestDefaults)
	require.Nil(t, err)

	require.Equal(t, Version(0), cfg.Version())
	cfg.version = Version(1)

	buf := &bytes.Buffer{}
	require.Nil(t, cfg.Save(buf))

	cfg, err = Open(bytes.NewReader(buf.Bytes()), TestDefaults)
	require.Nil(t, err)

	require.Equal(t, Version(1), cfg.Version())
}

func TestOpenMalformed(t *testing.T) {
	malformed := testutil.CreateDummyBuf(1024)

	// Not panicking here is okay for now as test.
	// Later one might want to add something like a fuzzer for this.
	_, err := Open(bytes.NewReader(malformed), TestDefaults)
	require.NotNil(t, err)
}
