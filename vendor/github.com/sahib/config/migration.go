package config

import (
	"errors"
	"fmt"
	"sort"

	e "github.com/pkg/errors"
)

var (
	// ErrNotVersioned is returned by Migrate() when it can't find the version tag.
	// If that happens you can still try to Open() the config normally.
	ErrNotVersioned = errors.New("config has no valid version tag")
)

// Version specifies the version of the config.  The exact number does not
// matter; the migration only cares about less or more. It's recommended to
// use successive numbers though. The first version is 0.
type Version int

// Migration is a function that is executed to replicate the changes between to
// versions. It should modify `newCfg` in a way so all keys from `oldCfg` that
// were portable are transferred.
type Migration func(oldCfg, newConfig *Config) error

type migrationEntry struct {
	fn       Migration
	version  Version
	defaults DefaultMapping
}

// Migrater is a factory for creating version'd configs.
// See NewMigrater() for more details.
type Migrater struct {
	currentVersion Version
	migrations     []migrationEntry
	strictness     Strictness
}

// Add a new migration entry.
// Adding a migration without a migration func is allowed
func (mm *Migrater) Add(version Version, migration Migration, defaults DefaultMapping) {
	entry := migrationEntry{
		fn:       migration,
		version:  version,
		defaults: defaults,
	}

	biggerIdx := sort.Search(len(mm.migrations), func(idx int) bool {
		return mm.migrations[idx].version > version
	})

	if biggerIdx == len(mm.migrations) {
		mm.migrations = append(mm.migrations, entry)
		return
	}

	// Insert somewhere in the middle:
	left := append(mm.migrations[:biggerIdx], entry)
	mm.migrations = append(left, mm.migrations[biggerIdx:]...)
}

// migrationFor returns the migration at `version` or nil if there is None.
func (mm *Migrater) migrationFor(version Version) *migrationEntry {
	migIdx := sort.Search(len(mm.migrations), func(idx int) bool {
		return mm.migrations[idx].version >= version
	})

	if migIdx == len(mm.migrations) {
		return nil
	}

	mig := mm.migrations[migIdx]
	if mig.version != version {
		return nil
	}

	return &mig
}

// biggerThan returns all migrations bigger than a certain version.
func (mm *Migrater) biggerThan(version Version) []migrationEntry {
	biggerIdx := sort.Search(len(mm.migrations), func(idx int) bool {
		return mm.migrations[idx].version > version
	})

	if biggerIdx == len(mm.migrations) {
		return []migrationEntry{}
	}

	return mm.migrations[biggerIdx:]
}

// Migrate reads the config from `dec` and converts it to the newest version if required.
func (mm *Migrater) Migrate(dec Decoder) (*Config, error) {
	if dec == nil {
		if len(mm.migrations) == 0 {
			return nil, fmt.Errorf("no migration given and nothing to decode from")
		}

		defaults := mm.migrations[len(mm.migrations)-1].defaults
		return Open(nil, defaults, mm.strictness)
	}

	currVersion, memory, err := dec.Decode()
	if err != nil {
		return nil, err
	}

	// Attempt to open the (potentially) old config and read it
	// with the respective & compatible defaults.
	currMig := mm.migrationFor(currVersion)
	if currMig == nil {
		return nil, fmt.Errorf("There are no defaults for `%d`", currVersion)
	}

	// TODO
	cfg, err := open(currVersion, memory, currMig.defaults, mm.strictness)
	if err != nil {
		return nil, err
	}

	// Find all migrations we have to do, to get to the most recent
	// versions. If already recent, migrations will be empty.
	for _, migration := range mm.biggerThan(currVersion) {
		if migration.fn == nil {
			continue
		}

		// Create an empty default config:
		newCfg, err := Open(nil, migration.defaults, mm.strictness)
		if err != nil {
			return nil, e.Wrapf(err, "failed creating default config for v%d", migration.version)
		}

		// Do the migration:
		if err := migration.fn(cfg, newCfg); err != nil {
			return nil, err
		}

		// Try again with current cfg in next round:
		cfg = newCfg
		cfg.version = migration.version
	}

	return cfg, nil
}

// NewMigrater returns a new Migrater.
//
// It can be seen as an migration capable version of config.Open().  Instead of
// directly passing the defaults you register a number of migrations (each with
// their own migration func, defaults and version).  The actual work is done by
// the migration functions which are written by the caller of this API.
// The caller defined migration method will likely call MigrateKeys() though.
//
// Call Migrate() on the migrater will read the current version and
// try to migrate to the most recent one.
func NewMigrater(currentVersion Version, strictness Strictness) *Migrater {
	return &Migrater{
		currentVersion: currentVersion,
		strictness:     strictness,
	}
}

// MigrateKeys is a helper function to write migrations easily. It takes the
// old and new config and copies all keys that are compatible (i.e. same key,
// same type). If calls fn() on any key that exists in the new config and not
// in the old config (i.e. new keys). If any error occurs during set (e.g.
// wrong type) fn is also called.  If fn returns a non-nil error this method
// stops and returns the error.
func MigrateKeys(oldCfg, newCfg *Config, fn func(key string, err error) error) error {
	for _, newKey := range newCfg.Keys() {
		var fnErr error
		isValid := oldCfg.IsValidKey(newKey)
		if isValid {
			if err := newCfg.Set(newKey, oldCfg.Get(newKey)); err != nil {
				fnErr = err
			}
		}

		if (!isValid || fnErr != nil) && fn != nil {
			if err := fn(newKey, fnErr); err != nil {
				return err
			}
		}
	}

	return nil
}
