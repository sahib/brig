package config

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func openFromData(data []byte, defaults DefaultMapping) (*Config, error) {
	return Open(NewYamlDecoder(bytes.NewReader([]byte(data))), defaults)
}

func openFromString(data string, defaults DefaultMapping) (*Config, error) {
	return Open(NewYamlDecoder(bytes.NewReader([]byte(data))), defaults)
}

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
	cfg, err := openFromString(testConfig, TestDefaults)
	require.Nil(t, err)

	require.Equal(t, int64(6667), cfg.Int("daemon.port"))
}

func TestGetNonExisting(t *testing.T) {
	defer func() { require.NotNil(t, recover()) }()

	cfg, err := openFromString(testConfig, TestDefaults)
	require.Nil(t, err)

	// should panic.
	require.Nil(t, cfg.SetFloat("not.existing", 23))
}

func TestSet(t *testing.T) {
	cfg, err := openFromString(testConfig, TestDefaults)
	require.Nil(t, err)

	require.Nil(t, cfg.SetInt("daemon.port", 6666))
	require.Equal(t, int64(6666), cfg.Int("daemon.port"))
}

func TestSetBucketKey(t *testing.T) {
	defer func() { require.NotNil(t, recover()) }()

	cfg, err := openFromString(testConfig, TestDefaults)
	require.Nil(t, err)

	// should panic.
	require.Nil(t, cfg.SetString("daemon", "oh oh"))
}

func TestAddChangeSignal(t *testing.T) {
	cfg, err := openFromString(testConfig, TestDefaults)
	require.Nil(t, err)

	callCount := 0
	cbID := cfg.AddEvent("data.ipfs.path", func(key string) {
		require.Equal(t, "new-value", cfg.String("data.ipfs.path"))
		callCount++
	})

	require.Equal(t, 0, callCount)
	require.Nil(t, cfg.SetInt("daemon.port", 42))
	require.Equal(t, 0, callCount)
	require.Nil(t, cfg.SetString("data.ipfs.path", "new-value"))
	require.Equal(t, 1, callCount)

	// Setting it twice should not trigger the callbacks.
	// The value did not change after all.
	require.Nil(t, cfg.SetString("data.ipfs.path", "new-value"))
	require.Equal(t, 1, callCount)

	cfg.RemoveEvent(cbID)
	require.Nil(t, cfg.SetString("data.ipfs.path", "newer-value"))
	require.Equal(t, 1, callCount)
}

func TestAddChangeSignalAll(t *testing.T) {
	cfg, err := openFromString(testConfig, TestDefaults)
	require.Nil(t, err)

	callCount := 0
	cbID := cfg.AddEvent("", func(key string) {
		callCount++
	})

	require.Equal(t, 0, callCount)
	require.Nil(t, cfg.SetInt("daemon.port", 42))
	require.Equal(t, 1, callCount)
	require.Nil(t, cfg.SetString("data.ipfs.path", "new-value"))
	require.Equal(t, 2, callCount)
	require.Nil(t, cfg.SetInt("daemon.port", 42))
	require.Equal(t, 2, callCount)

	cfg.RemoveEvent(cbID)
	require.Nil(t, cfg.SetString("data.ipfs.path", "newer-value"))
	require.Equal(t, 2, callCount)
}

func TestOpenSave(t *testing.T) {
	cfg, err := openFromString(testConfig, TestDefaults)
	require.Nil(t, err)

	require.Nil(t, cfg.SetInt("daemon.port", 6666))
	require.Nil(t, cfg.SetString("data.ipfs.path", "y"))

	buf := &bytes.Buffer{}
	require.Nil(t, cfg.Save(NewYamlEncoder(buf)))

	newCfg, err := Open(NewYamlDecoder(buf), TestDefaults)
	require.Nil(t, err)

	require.Equal(t, int64(6666), newCfg.Int("daemon.port"))
	require.Equal(t, "y", newCfg.String("data.ipfs.path"))
}

func TestKeys(t *testing.T) {
	cfg, err := openFromString(testConfig, TestDefaults)
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
	_, err := openFromString("a: 1", TestDefaults)
	require.NotNil(t, err)
}

func TestSection(t *testing.T) {
	cfg, err := openFromString(testConfig, TestDefaults)
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
	cfg, err := openFromString(testConfig, TestDefaults)
	require.Nil(t, err)

	parentCallCount := 0
	parentID := cfg.AddEvent("fs.compress.default_algo", func(key string) {
		require.Equal(t, "fs.compress.default_algo", key)
		parentCallCount++
	})

	fsSec := cfg.Section("fs")

	childCallCount := 0
	childID := fsSec.AddEvent("compress.default_algo", func(key string) {
		require.Equal(t, "compress.default_algo", key)
		childCallCount++
	})

	require.Nil(t, cfg.SetString("fs.compress.default_algo", "none"))
	require.Nil(t, fsSec.SetString("compress.default_algo", "lz4"))

	require.Equal(t, 2, parentCallCount)
	require.Equal(t, 1, childCallCount)

	fsSec.RemoveEvent(childID)
	cfg.RemoveEvent(parentID)
}

func TestIsValidKey(t *testing.T) {
	cfg, err := openFromString(testConfig, TestDefaults)
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

	cfg, err := Open(nil, defaults)
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

func TestToFileFromYamlFile(t *testing.T) {
	cfg, err := Open(nil, TestDefaults)
	require.Nil(t, err)

	path := "/tmp/brig-test-config.yml"
	require.Nil(t, ToYamlFile(path, cfg))

	defer os.Remove(path)

	loadCfg, err := FromYamlFile(path, TestDefaults)
	require.Nil(t, err)

	configMustEquals(t, cfg, loadCfg)
}

func TestSetIncompatibleType(t *testing.T) {
	cfg, err := Open(nil, TestDefaults)
	require.Nil(t, err)

	require.NotNil(t, cfg.SetString("daemon.port", "xxx"))
}

func TestVersionPersisting(t *testing.T) {
	cfg, err := Open(nil, TestDefaults)
	require.Nil(t, err)

	require.Equal(t, Version(0), cfg.Version())
	cfg.version = Version(1)

	buf := &bytes.Buffer{}
	require.Nil(t, cfg.Save(NewYamlEncoder(buf)))

	cfg, err = openFromData(buf.Bytes(), TestDefaults)
	require.Nil(t, err)

	require.Equal(t, Version(1), cfg.Version())
}

func TestOpenMalformed(t *testing.T) {
	malformed := make([]byte, 1024)
	for idx := 0; idx < 1024; idx++ {
		malformed[idx] = byte(idx % 256)
	}

	// Not panicking here is okay for now as test.
	// Later one might want to add something like a fuzzer for this.
	_, err := openFromData(malformed, TestDefaults)
	require.NotNil(t, err)
}

func TestReload(t *testing.T) {
	cfg, err := Open(nil, TestDefaultsV0)
	require.Nil(t, err)

	text := `# version: 666
a:
  b: 70
  child:
    c: "world"
`

	require.Nil(t, cfg.Reload(NewYamlDecoder(strings.NewReader(text))))
	require.Equal(t, int64(70), cfg.Int("a.b"))
	require.Equal(t, "world", cfg.String("a.child.c"))
	require.Equal(t, Version(666), cfg.Version())
}

func TestReloadSignal(t *testing.T) {
	cfg, err := Open(nil, TestDefaultsV0)
	require.Nil(t, err)

	text := `# version: 666
a:
  b: 70
  child:
    c: "hello"
`
	globalCallCount := 0
	localCallCount := 0

	cfg.AddEvent("", func(key string) {
		globalCallCount++
	})

	cfg.AddEvent("a.b", func(key string) {
		localCallCount++
	})

	require.Nil(t, cfg.Reload(NewYamlDecoder(strings.NewReader(text))))

	require.Equal(t, 1, globalCallCount)
	require.Equal(t, 1, localCallCount)
}

func TestMerge(t *testing.T) {
	baseYml := `# version: 666
a:
  b: 70
`

	overYml := `# version: 666
a:
  child:
    c: "world"
`
	baseCfg, err := Open(NewYamlDecoder(strings.NewReader(baseYml)), TestDefaultsV0)
	require.Nil(t, err)

	overCfg, err := Open(NewYamlDecoder(strings.NewReader(overYml)), TestDefaultsV0)
	require.Nil(t, err)

	require.Equal(t, int64(70), baseCfg.Get("a.b"))
	require.Equal(t, "hello", baseCfg.Get("a.child.c"))

	require.Equal(t, int64(15), overCfg.Get("a.b"))
	require.Equal(t, "world", overCfg.Get("a.child.c"))

	err = baseCfg.Merge(overCfg)
	require.Nil(t, err)

	// We shouldn't take over the default value of overCfg.
	require.Equal(t, int64(70), baseCfg.Get("a.b"))
	require.Equal(t, "world", baseCfg.Get("a.child.c"))

	// Check that the old values were not touched:
	require.Equal(t, int64(15), overCfg.Get("a.b"))
	require.Equal(t, "world", overCfg.Get("a.child.c"))
}

func TestMergeDifferentDefaults(t *testing.T) {
	baseYml := `# version: 666
a:
  b: 70
`

	overYml := `# version: 666
a:
  child:
    c: "x"
`

	baseCfg, err := Open(NewYamlDecoder(strings.NewReader(baseYml)), TestDefaultsV0)
	require.Nil(t, err)

	overCfg, err := Open(NewYamlDecoder(strings.NewReader(overYml)), TestDefaultsV1)
	require.Nil(t, err)

	err = baseCfg.Merge(overCfg)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "refusing")
}

func TestDuration(t *testing.T) {
	baseYml := `# version: 666
duration: 5m20s
`

	defaults := DefaultMapping{
		"duration": DefaultEntry{
			Default:   "10m",
			Validator: DurationValidator(),
		},
	}

	cfg, err := Open(NewYamlDecoder(strings.NewReader(baseYml)), defaults)
	require.Nil(t, err)

	require.Equal(t, 5*time.Minute+20*time.Second, cfg.Duration("duration"))
	cfg.SetDuration("duration", 20*time.Minute)
	require.Equal(t, 20*time.Minute, cfg.Duration("duration"))

	cfg.SetDuration("duration", time.Duration(0))
	require.Equal(t, 0*time.Minute, cfg.Duration("duration"))
}
