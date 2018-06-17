package config

import (
	"bytes"
	"os"
	"testing"

	"github.com/sahib/brig/util/testutil"
	"github.com/stretchr/testify/require"
)

func getTypeOfDefaultKey(key string, defaults DefaultMapping) string {
	defaultEntry := getDefaultByKey(key, defaults)
	if defaultEntry == nil {
		return ""
	}

	return getTypeOf(defaultEntry.Default)
}

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
	require.Nil(t, cfg.SetFloat("not.existing", 23))
}

func TestSet(t *testing.T) {
	cfg, err := Open(bytes.NewReader([]byte(testConfig)), Defaults)
	require.Nil(t, err)

	require.Nil(t, cfg.SetInt("daemon.port", 6666))
	require.Equal(t, int64(6666), cfg.Int("daemon.port"))
}

func TestSetBucketKey(t *testing.T) {
	defer func() { require.NotNil(t, recover()) }()

	cfg, err := Open(bytes.NewReader([]byte(testConfig)), Defaults)
	require.Nil(t, err)

	// should panic.
	require.Nil(t, cfg.SetString("daemon", "oh oh"))
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
	cfg, err := Open(bytes.NewReader([]byte(testConfig)), Defaults)
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
	cfg, err := Open(bytes.NewReader([]byte(testConfig)), Defaults)
	require.Nil(t, err)

	require.Nil(t, cfg.SetInt("daemon.port", 6666))
	require.Nil(t, cfg.SetString("data.ipfs.path", "y"))

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
	_, err := Open(bytes.NewReader([]byte(`a: 1`)), Defaults)
	require.NotNil(t, err)
}

func TestSection(t *testing.T) {
	cfg, err := Open(bytes.NewReader([]byte(testConfig)), Defaults)
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
	cfg, err := Open(bytes.NewReader([]byte(testConfig)), Defaults)
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
	cfg, err := Open(bytes.NewReader([]byte(testConfig)), Defaults)
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

func TestValidation(t *testing.T) {
	defaults := DefaultMapping{
		"enum-val": DefaultEntry{
			Default:      "a",
			NeedsRestart: false,
			Validator:    EnumValidator("a", "b", "c"),
		},
	}

	// Check initial validation:
	_, err := Open(bytes.NewReader([]byte("enum-val: d")), defaults)
	require.NotNil(t, err)

	cfg, err := Open(bytes.NewReader([]byte("enum-val: c")), defaults)
	require.Nil(t, err)
	require.Equal(t, cfg.String("enum-val"), "c")

	// Set an invalid enum value:
	require.NotNil(t, cfg.SetString("enum-val", "C"))
	require.Nil(t, cfg.SetString("enum-val", "a"))
	require.Equal(t, cfg.String("enum-val"), "a")
}

func TestIntvalidator(t *testing.T) {
	vdt := IntRangeValidator(10, 100)
	require.Contains(t, vdt("x").Error(), "is not an integer")
	require.Contains(t, vdt(int64(9)).Error(), "may not be less than 10")
	require.Contains(t, vdt(int64(101)).Error(), "may not be more than 100")

	require.Nil(t, vdt(int64(10)))
	require.Nil(t, vdt(int64(100)))
	require.Nil(t, vdt(int64(50)))
}

func configMustEquals(t *testing.T, aCfg, bCfg *Config) {
	require.Equal(t, aCfg.Keys(), bCfg.Keys())
	for _, key := range aCfg.Keys() {
		require.Equal(t, aCfg.Get(key), bCfg.Get(key), key)
	}
}

func TestToFileFromFile(t *testing.T) {
	cfg, err := Open(bytes.NewReader(nil), Defaults)
	require.Nil(t, err)

	path := "/tmp/brig-test-config.yml"
	require.Nil(t, ToFile(path, cfg))

	defer os.Remove(path)

	loadCfg, err := FromFile(path, Defaults)
	require.Nil(t, err)

	configMustEquals(t, cfg, loadCfg)
}

func TestSetIncompatibleType(t *testing.T) {
	cfg, err := Open(bytes.NewReader(nil), Defaults)
	require.Nil(t, err)

	require.NotNil(t, cfg.SetString("daemon.port", "xxx"))
}

func TestVersionPersisting(t *testing.T) {
	cfg, err := Open(bytes.NewReader(nil), Defaults)
	require.Nil(t, err)

	require.Equal(t, Version(0), cfg.Version())
	cfg.version = Version(1)

	buf := &bytes.Buffer{}
	require.Nil(t, cfg.Save(buf))

	cfg, err = Open(bytes.NewReader(buf.Bytes()), Defaults)
	require.Nil(t, err)

	require.Equal(t, Version(1), cfg.Version())
}

func TestOpenMalformed(t *testing.T) {
	malformed := testutil.CreateDummyBuf(1024)

	// Not panicking here is okay for now as test.
	// Later one might want to add something like a fuzzer for this.
	_, err := Open(bytes.NewReader(malformed), Defaults)
	require.NotNil(t, err)
}
