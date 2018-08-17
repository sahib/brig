package config

import (
	"bytes"
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

var TestDefaultsV0 = DefaultMapping{
	"a": DefaultMapping{
		"b": DefaultEntry{
			Default: 15,
		},
		"child": DefaultMapping{
			"c": DefaultEntry{
				Default: "hello",
			},
		},
	},
}

var TestDefaultsV1 = DefaultMapping{
	"a": DefaultMapping{
		"b": DefaultEntry{
			Default: 10,
		},
		"child": DefaultMapping{
			"c": DefaultEntry{
				Default:   "z",
				Validator: EnumValidator("x", "y", "z"),
			},
		},
		"new_key": DefaultEntry{
			Default: 2.0,
		},
	},
}

func migrateToV1(oldCfg, newCfg *Config) error {
	return MigrateKeys(oldCfg, newCfg, func(key string, err error) error {
		switch key {
		case "a.child.c":
			return newCfg.SetString(key, "z")
		case "a.new_key":
			return newCfg.SetFloat(key, float64(oldCfg.Int("a.b"))*3.0)
		default:
			return fmt.Errorf("Incomplete migration for key: %v", key)
		}
	})
}

func createInitialConfigData(t *testing.T) []byte {
	cfg, err := Open(nil, TestDefaultsV0, StrictnessPanic)
	require.Nil(t, err)
	require.Equal(t, Version(0), cfg.Version())

	buf := &bytes.Buffer{}
	require.Nil(t, cfg.Save(NewYamlEncoder(buf)))
	return buf.Bytes()
}

func checkIfSorted(t *testing.T, mgr *Migrater) {
	versions := []int{}
	for _, mig := range mgr.migrations {
		versions = append(versions, int(mig.version))
	}

	require.True(t, sort.IntsAreSorted(versions), "migrations are not sorted")
}

func TestAddMigration(t *testing.T) {
	mgr := NewMigrater(0, StrictnessPanic)
	for idx := 0; idx < 10; idx++ {
		mgr.Add(Version(idx), nil, nil)
	}

	checkIfSorted(t, mgr)

	mgr = NewMigrater(0, StrictnessPanic)
	for idx := 9; idx >= 0; idx-- {
		mgr.Add(Version(idx), nil, nil)
	}

	checkIfSorted(t, mgr)

	mgr = NewMigrater(0, StrictnessPanic)
	for idx := 1; idx < 11; idx++ {
		// Pseudo random order:
		mgr.Add(Version(int64(idx)*114007148193231984%61), nil, nil)
	}
}

func TestBasicMigration(t *testing.T) {
	initialData := createInitialConfigData(t)

	mgr := NewMigrater(1, StrictnessPanic)
	mgr.Add(0, nil, TestDefaultsV0)
	mgr.Add(1, migrateToV1, TestDefaultsV1)
	cfg, err := mgr.Migrate(NewYamlDecoder(bytes.NewReader(initialData)))
	require.Nil(t, err)

	buf := &bytes.Buffer{}
	require.Nil(t, cfg.Save(NewYamlEncoder(buf)))
	require.Equal(t, Version(1), cfg.Version())

	// This should be taken from the old config (old default is 10)
	require.Equal(t, int64(15), cfg.Get("a.b"))

	// The old config value was a free string, now it's am emum.
	// We should take over the migrated value this time.
	require.Equal(t, "z", cfg.Get("a.child.c"))

	// The new key should take the migrated value.
	require.Equal(t, float64(45), cfg.Get("a.new_key"))
}

func TestBiggerThan(t *testing.T) {
	mgr := NewMigrater(0, StrictnessPanic)
	for idx := 0; idx < 10; idx++ {
		mgr.migrations = append(mgr.migrations, migrationEntry{
			version: Version(idx),
		})
	}

	require.Empty(t, mgr.biggerThan(10))
	require.Empty(t, mgr.biggerThan(11))
	require.Equal(t, mgr.migrations, mgr.biggerThan(-1))

	for idx := 0; idx < 10; idx++ {
		require.Equal(t, mgr.migrations[idx+1:], mgr.biggerThan(Version(idx)))
	}
}

func TestMigrationFor(t *testing.T) {
	mgr := NewMigrater(0, StrictnessPanic)
	for idx := 0; idx < 10; idx++ {
		mgr.migrations = append(mgr.migrations, migrationEntry{
			version: Version(idx),
		})
	}

	for idx := 0; idx < 10; idx++ {
		require.Equal(t, mgr.migrationFor(Version(idx)), &migrationEntry{
			version: Version(idx),
		})
	}
}
