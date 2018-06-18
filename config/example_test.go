package config

import (
	"bytes"
	"fmt"
	"log"
	"strings"
)

/////////////////////////
// BASIC USAGE EXAMPLE //
/////////////////////////

// Everything starts with writing down the defaults.
// It's a nested data structure that you define right in your program:
var ExampleDefaultsV0 = DefaultMapping{
	"ui": DefaultMapping{
		"show_tooltips": DefaultEntry{
			Default: true,
			// Make sure to note down what this key does.
			Docs: "Show tooltips when you least expect them",
			// This key can be set without restarting the service.
			NeedsRestart: false,
		},
	},
	"backend": DefaultMapping{
		"name": DefaultEntry{
			Default:      "the_good_one",
			Docs:         "Choose what backend you want to use.",
			NeedsRestart: true,
			// Limit what config options are available.
			// You can always write more validators.
			Validator: EnumValidator(
				"the_good_one",
				"the_bad_one",
				"the_ugly_one",
			),
		},
		"workers": DefaultEntry{
			Default:      10,
			Docs:         "How many workers to start.",
			NeedsRestart: false,
			// Make sure the user can't start more than 100
			// and not less than 1.
			Validator: IntRangeValidator(1, 100),
		},
	},
}

func ExampleConfig() {
	// You either open it via existing yaml data - or in case of the initial
	// write, you just let it take over the defaults. This is also the step
	// where the first validation happens.
	cfg, err := Open(bytes.NewReader(nil), ExampleDefaultsV0)
	if err != nil {
		log.Fatalf("Failed to open config: %v", err)
	}

	// Fetching keys is easy now and requires no error handling:
	cfg.String("backend.name")   // -> the_good_one
	cfg.Int("backend.workers")   // -> 10
	cfg.Bool("ui.show_tooltips") // -> true

	// You can set also set keys:
	// This one will return an error though because
	cfg.SetString("backend.name", "the_great_one")

	// If you'd like to print an overview over all config keys,
	// you can just get a list of all default entries:
	for _, key := range cfg.Keys() {
		entry := cfg.GetDefault(key)
		fmt.Printf("%s: %s (restart: %t)\n", key, entry.Docs, entry.NeedsRestart)
	}

	// If you have only a string (e.g. from a cmdline config set),
	// you can ask the config to convert it to the right type:
	alienKey, alienVal := "backend.workers", "15"
	if cfg.IsValidKey(alienKey) {
		safeVal, err := cfg.Cast(alienKey, alienVal)
		if err != nil {
			log.Fatalf("Uh, oh, could not cast key to the right type: %v", err)
		}

		// safeVal is an integer now:
		cfg.Set(alienKey, safeVal)
	}

	// Want to know if something changed?
	// Just register a callback for it. If you pass an empty string,
	// you'll get callbacks for every set.
	cid := cfg.AddChangedKeyEvent("backend.workers", func(key string) {
		fmt.Println("Key was changed:", key)
	})

	// You can get rid of callbacks too of course:
	cfg.RemoveChangedKeyEvent(cid)

	// One nifty feature is to pass only a sub section of the config
	// to specific parts of the program - Which saves you from typing
	// the full keys and disallowing them to remove other parts.
	backendCfg := cfg.Section("backend")
	backendCfg.Get("name") // -> the_good_one

	// When you're done you can always serialize the config:
	buf := &bytes.Buffer{}
	if err := cfg.Save(buf); err != nil {
		log.Fatalf("Failed to save config: %v", err)
	}

	fmt.Println(buf.String())

	// Output: backend.name: Choose what backend you want to use. (restart: true)
	// backend.workers: How many workers to start. (restart: false)
	// ui.show_tooltips: Show tooltips when you least expect them (restart: false)
	// # version: 0 (DO NOT MODIFY THIS LINE)
	// backend:
	//   name: the_good_one
	//   workers: 15
	// ui:
	//   show_tooltips: true
	//
}

/////////////////////////////
// BASIC MIGRATION EXAMPLE //
/////////////////////////////

// The first version of a config is always "0".
// If you jump in a version you simply increase the number
// that you pass to NewMigrater()
const CurrentVersion = 1

// For every migration you should have your own defaults.
// This might seem wasteful at first, but it will document
// in your code what you changed over time and how it converts.
// Also, it's needed to open old configs (shortened for brevity)
var ExampleDefaultsV1 = DefaultMapping{
	"ui": DefaultMapping{
		"show_tooltips": DefaultEntry{
			Default: true,
		},
	},
	"backend": DefaultMapping{
		"name": DefaultEntry{
			Default: "the_good_one",
			Validator: EnumValidator(
				"the_good_one",
				"the_bad_one",
				"the_ugly_one",
			),
		},
		// workers is gone, we only have accuracy here now.
		// (and the type also changed to a float)
		"accuracy": DefaultEntry{
			Default: 10.5,
		},
	},
}

func ExampleMigration() {
	// This config package optionally supports versioned configs.
	// Whenever you decide to change the layout of the config,
	// you can bump the version and register a new migration func
	// that will be run over older config upon opening them.
	mgr := NewMigrater(CurrentVersion)

	// Add a migration - the first one for version "0" has no func attached.
	mgr.Add(0, nil, ExampleDefaultsV0)

	// For version "1" we gonna need a function that transforms the config:
	migrateToV1 := func(oldCfg, newCfg *Config) error {
		// Use the helpful MigrateKeys method to migrate most of the keys.
		// It will call you back on every missing key or any errors.
		return MigrateKeys(oldCfg, newCfg, func(key string, err error) error {
			switch key {
			case "backend.accuracy":
				// Do something based on the old config key:
				return newCfg.SetFloat(
					"backend.accuracy",
					float64(oldCfg.Int("backend.workers"))+0.5,
				)
			default:
				return fmt.Errorf("Incomplete migration for key: %v", key)
			}
		})
	}

	// Add it with the new respective results:
	mgr.Add(1, migrateToV1, ExampleDefaultsV1)

	rawConfig := `# version: 0
ui:
  show_tooltips: true
backend:
  name: the_good_one
  workers: 10
`

	// The Migrate call works like a factory method.
	// It creates the config in a versioned way:
	cfg, err := mgr.Migrate(strings.NewReader(rawConfig))
	if err != nil {
		// Handle errors...
	}

	cfg.Version() // -> 1 now.

	// If you print it, you will notice a changed version tag:
	buf := &bytes.Buffer{}
	cfg.Save(buf)
	fmt.Println(buf.String())

	// Output: # version: 1 (DO NOT MODIFY THIS LINE)
	// backend:
	//   accuracy: 10.5
	//   name: the_good_one
	// ui:
	//   show_tooltips: true
}
