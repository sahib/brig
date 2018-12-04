// Package config implements a very opinionated config utility.  It relies on a
// "default spec", i.e. a structure that defines all existing configuration
// keys, their types and their initial default values.  This is used as
// fallback and source of validation. The idea is similar to python's configobj
// (albeit much smaller). Surprisingly I didn't find any similar library in Go.
//
// Note that passing invalid keys to a few methods will cause a panic - on purpose.
// Using a wrong config key is seen as a bug and should be corrected immediately.
// This allows this package to skip error handling on Get() and Set() entirely.
//
// In short: This config  does a few things different than the ones I saw for
// Go.  Instead of providing numerous possible sources and formats to save your
// config it simply relies on YAML out of the box. The focus is not on ultimate
// convinience but on:
//
// - Providing meaningful validation and default values.
//
// - Providing built-in documentation for all config values.
//
// - Making it able to react on changed config values.
//
// - Being usable from several go routines.
//
// - In future: Provide an easy way to migrate configs.
package config

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	e "github.com/pkg/errors"
)

// Strictness defines how the API of the config reacts when it thinks
// that the programmer made a mistake. These kind of mistakes usually
// fall into one of the following categories:
//
// - A wrong key was used for a Get(), Set() etc.
// - The wrong type getter was used for a certain key (bool for a string e.g.)
// - The defaults are faulty (wrong types, __many__ in the wrong place etc.)
//
// All of those errors should be catched early in the development and should
// not make it to the user. Therefore StrictnessPanic is recommended for
// developing. When building the software in release mode, one might want to
// switch to StrictnessWarn, which just logs when it found something awful.
type Strictness int

const (
	// StrictnessIgnore silently ignores any programmer error
	StrictnessIgnore = Strictness(iota)
	// StrictnessWarn will log a complaint via log.Println()
	StrictnessWarn
	// StrictnessPanic will panic when a programmer error was made.
	StrictnessPanic
)

// DefaultEntry represents the metadata for a default value in the config.
// Every possible key has to have a DefaultEntry, otherwise Get() and Set()
// will panic at you since this is considered a programmer error.
type DefaultEntry struct {
	// Default is the fallback value for this config key.
	// The confg type will be inferred from its literal type.
	Default interface{}

	// NeedsRestart indicates that we need to restart the daemon
	// to have an effect here.
	NeedsRestart bool

	// Docs describes the meaning of the configuration value.
	Docs string

	// Function that can be used to check
	Validator func(val interface{}) error
}

// DefaultMapping is a container to hold all required DefaultEntries.
// It is a nested map with sections as string keys.
type DefaultMapping map[interface{}]interface{}

var (
	// all types that we will cast into int64
	typeIntPattern = regexp.MustCompile(`u{0,1}int(64|32|16|8|)`)
	// all types that we will cast into float64
	typeFloatPattern = regexp.MustCompile(`float(32|64|)`)
	// all types that are inside of a slice
	typeSlicePattern = regexp.MustCompile(`^\[.*\]$`)
	// pattern for the version tag
	versionTag = regexp.MustCompile(`^# version:\s*(\d+).*`)
	// manyMarker is a special key in the default mapping
	manyMarker = "__many__"
)

func getDefaultSectionByKeys(keys []string, defaults DefaultMapping, strictness Strictness) DefaultMapping {
	if len(keys) == 0 {
		return defaults
	}

	child, hasChild := defaults[keys[0]]
	if !hasChild {
		// The key might still be used if we have a __many__ entry.
		// If not, it has to be a wrong key.
		placeHolderFound := false
		for defaultKeyVal := range defaults {
			defaultKey, ok := defaultKeyVal.(string)
			if !ok {
				complain(
					fmt.Sprintf("bug: default key is not a string: %v", defaultKeyVal),
					strictness,
				)

				return nil
			}

			// We found a __many__ entry. Use it as validation base.
			if defaultKey == manyMarker {
				child = defaults[manyMarker]
				placeHolderFound = true
			}
		}

		if !placeHolderFound {
			return nil
		}
	}

	if child == nil {
		return nil
	}

	section, ok := child.(DefaultMapping)
	if !ok {
		return nil
	}

	return getDefaultSectionByKeys(keys[1:], section, strictness)
}

// recursive implementation of getDefaultByKey
func getDefaultByKeys(keys []string, defaults DefaultMapping, strictness Strictness) *DefaultEntry {
	if len(keys) == 0 {
		return nil
	}

	section := getDefaultSectionByKeys(keys[:len(keys)-1], defaults, strictness)
	if section == nil {
		return nil
	}

	lastKey := keys[len(keys)-1]
	if lastKey == manyMarker {
		complain("__many__ used for default entries", strictness)
		return nil
	}

	child, ok := section[lastKey]
	if !ok {
		return nil
	}

	defaultEntry, ok := child.(DefaultEntry)
	if !ok {
		return nil
	}

	// scalar type, return immediately.
	return &defaultEntry
}

func getDefaultByKey(key string, defaults DefaultMapping, strictness Strictness) *DefaultEntry {
	return getDefaultByKeys(strings.Split(key, "."), defaults, strictness)
}

func getTypeOf(val interface{}) string {
	typ := reflect.TypeOf(val)
	if typ == nil {
		return ""
	}

	if typ.Kind() == reflect.Slice {
		return fmt.Sprintf("[%s]", typ.Elem().Name())
	}

	return typ.Name()
}

func isCompatibleType(typeA, typeB string) bool {
	// Be a bit more tolerant regarding integer values.
	if typeIntPattern.MatchString(typeA) {
		return typeIntPattern.MatchString(typeB)
	}

	if typeFloatPattern.MatchString(typeA) {
		return typeFloatPattern.MatchString(typeB)
	}

	if typeSlicePattern.MatchString(typeA) {
		return typeSlicePattern.MatchString(typeB)
	}

	return typeA == typeB
}

func keys(root map[interface{}]interface{}, prefix []string, fn func(section map[interface{}]interface{}, key []string) error) error {
	for keyVal := range root {
		key, ok := keyVal.(string)
		if !ok {
			return fmt.Errorf("config contains non string keys: %v", keyVal)
		}

		// Create the next prefix for the next call or the validation check.
		nextPrefix := make([]string, len(prefix), len(prefix)+1)
		copy(nextPrefix, prefix)
		nextPrefix = append(nextPrefix, key)

		child := root[key]
		section, ok := child.(map[interface{}]interface{})
		if ok {
			// It's another sub section we have to visit.
			if err := keys(section, nextPrefix, fn); err != nil {
				return err
			}

			continue
		}

		if err := fn(root, nextPrefix); err != nil {
			return err
		}
	}

	return nil
}

func generalizeScalarType(val interface{}) interface{} {
	// Handle a few special cases here that come from go's type system.
	// Doing something like this will lead to a panic:
	//
	//     interface{}(int(42)).(int64)
	//
	// Since this is a config we do not care very much for extremely
	// big numbers and can therefore convert all numbers to int64.
	// The code below does that + something similar for float{32,64}.
	if typeIntPattern.MatchString(getTypeOf(val)) {
		destType := reflect.TypeOf(int64(0))
		val = reflect.ValueOf(val).Convert(destType).Int()
	}

	if typeFloatPattern.MatchString(getTypeOf(val)) {
		destType := reflect.TypeOf(float64(0))
		val = reflect.ValueOf(val).Convert(destType).Float()
	}

	return val
}

func generalizeType(val interface{}, defType string) (interface{}, error) {
	if typ := reflect.TypeOf(val); typ.Kind() == reflect.Slice {
		interfaces := val.([]interface{})
		switch defType {
		case "[string]":
			results := []string{}
			for _, inter := range interfaces {
				val, ok := inter.(string)
				if !ok {
					return nil, fmt.Errorf("string list contanins non-strings: %v (%T)", inter, inter)
				}

				results = append(results, val)
			}

			return results, nil
		case "[int]", "[int64]", "[int32]", "[int16]", "[int8]",
			"[uint]", "[uint64]", "[uint32]", "[uint16]", "[uint8]":
			results := []int64{}
			for _, inter := range interfaces {
				// always cast to int64:
				inter = generalizeScalarType(inter)
				val, ok := inter.(int64)
				if !ok {
					return nil, fmt.Errorf("int list contanins non-int64: %v (%T)", inter, inter)
				}

				results = append(results, val)
			}

			return results, nil
		case "[float32]", "[float64]":
			results := []float64{}
			for _, inter := range interfaces {
				inter = generalizeScalarType(inter)
				val, ok := inter.(float64)
				if !ok {
					return nil, fmt.Errorf("float list contanins non-float: %v (%T)", inter, inter)
				}

				results = append(results, val)
			}

			return results, nil
		case "[bool]":
			results := []bool{}
			for _, inter := range interfaces {
				val, ok := inter.(bool)
				if !ok {
					return nil, fmt.Errorf("bool list contanins non-bool: %v (%T)", inter, inter)
				}

				results = append(results, val)
			}

			return results, nil
		default:
			return nil, fmt.Errorf("unsupported list type: %v", typ)
		}
	}

	return generalizeScalarType(val), nil
}

func maybeMakeInterfaceList(val interface{}) interface{} {
	typ := reflect.TypeOf(val)
	if typ == nil {
		return nil
	}

	if typ.Kind() == reflect.Slice {
		rval := reflect.ValueOf(val)
		results := []interface{}{}
		for idx := 0; idx < rval.Len(); idx++ {
			results = append(results, rval.Index(idx).Interface())
		}

		return results
	}

	return val
}

// fill up any not explicitly set key with default values
func mergeDefaults(base map[interface{}]interface{}, overlay DefaultMapping, defaultKeys map[string]struct{}, prefix string) error {
	for overlayKeyVal := range overlay {
		overlayKey, ok := overlayKeyVal.(string)
		if !ok {
			return fmt.Errorf("config contains non string keys: %v", overlayKeyVal)
		}

		baseKeys := []string{}
		if overlayKey == manyMarker {
			for baseKeyVal := range base {
				baseKey, ok := baseKeyVal.(string)
				if !ok {
					return fmt.Errorf("key in config is not a string: %v", baseKeyVal)
				}

				if _, ok := overlay[baseKey]; !ok {
					baseKeys = append(baseKeys, baseKey)
				}
			}
		} else {
			baseKeys = append(baseKeys, overlayKey)
		}

		for _, baseKey := range baseKeys {
			switch overlayChild := overlay[overlayKey].(type) {
			case DefaultMapping:
				baseSection, ok := base[baseKey].(map[interface{}]interface{})
				if !ok {
					baseSection = make(map[interface{}]interface{})
					base[baseKey] = baseSection
				}

				newPrefix := prefixKey(prefix, baseKey)
				if err := mergeDefaults(baseSection, overlayChild, defaultKeys, newPrefix); err != nil {
					return err
				}
			case DefaultEntry:
				if _, ok := base[baseKey]; !ok {
					defType := getTypeOf(overlayChild.Default)
					defaultVal := maybeMakeInterfaceList(overlayChild.Default)

					fixedDefault, err := generalizeType(defaultVal, defType)
					if err != nil {
						return err
					}

					base[baseKey] = fixedDefault
					defaultKeys[prefixKey(prefix, baseKey)] = struct{}{}
				}
			}
		}
	}

	return nil
}

// validationChecker validates the incoming config
func validationChecker(
	root map[interface{}]interface{},
	defaults DefaultMapping,
	defaultKeys map[string]struct{},
	strictness Strictness,
) error {
	err := keys(root, nil, func(section map[interface{}]interface{}, key []string) error {
		// It's a scalar key. Let's run some diagnostics.
		lastKey := key[len(key)-1]
		child := section[lastKey]

		fullKey := strings.Join(key, ".")
		defaultEntry := getDefaultByKey(fullKey, defaults, strictness)
		if defaultEntry == nil {
			return fmt.Errorf("no default for key: %v", fullKey)
		}

		defType := getTypeOf(defaultEntry.Default)
		if defType == "" {
			return fmt.Errorf("no default found for key `%v`", fullKey)
		}

		valType := getTypeOf(child)
		if !isCompatibleType(valType, defType) {
			return fmt.Errorf(
				"type mismatch: want `%v`, got `%v` for key `%v`",
				defType,
				valType,
				fullKey,
			)
		}

		generalizedChild, err := generalizeType(child, defType)
		if err != nil {
			return err
		}

		// Do user defined validation:
		if defaultEntry.Validator != nil {
			if err := defaultEntry.Validator(generalizedChild); err != nil {
				return err
			}
		}

		// Valid key. Set the value:
		section[lastKey] = generalizedChild
		return nil
	})

	if err != nil {
		return err
	}

	// Fill in keys that are not present in the passed config:
	return mergeDefaults(root, defaults, defaultKeys, "")
}

////////////

// keyChangedEvent is a single entry added by AddChangedKeyEvent
type keyChangedEvent struct {
	fn  func(key string)
	key string
}

// Config is a helper that is built around a representation defined by a Encoder/Decoder.
// It supports typed gets and sets, change notifications and
// basic validation with defaults.
type Config struct {
	mu *sync.Mutex

	section         string
	defaults        DefaultMapping
	memory          map[interface{}]interface{}
	callbackCount   int
	changeCallbacks map[string]map[int]keyChangedEvent
	defaultKeys     map[string]struct{}
	version         Version
	strictness      Strictness
}

func prefixKey(section, key string) string {
	if section == "" {
		return key
	}

	return strings.Trim(section, ".") + "." + strings.Trim(key, ".")
}

// Open creates a new config from the data in `r`. The mapping in `defaults `
// tells the config which keys to expect and what type each of it should have.
// It is allowed to pass `nil` as decoder. In this case a config purely with
// default values will be created.
func Open(dec Decoder, defaults DefaultMapping, strictness Strictness) (*Config, error) {
	if defaults == nil {
		return nil, fmt.Errorf("need a default mapping")
	}

	var memory map[interface{}]interface{}
	var version Version
	var err error

	if dec != nil {
		version, memory, err = dec.Decode()
		if err != nil {
			return nil, err
		}
	} else {
		memory = make(map[interface{}]interface{})
		version = Version(0)
	}

	return open(version, memory, defaults, strictness)
}

// open does the actual struct creation. It is also used by the migrater.
func open(
	version Version,
	memory map[interface{}]interface{},
	defaults DefaultMapping,
	strictness Strictness,
) (*Config, error) {
	defaultKeys := make(map[string]struct{})
	if err := validationChecker(memory, defaults, defaultKeys, strictness); err != nil {
		return nil, e.Wrapf(err, "validate")
	}

	return &Config{
		mu:              &sync.Mutex{},
		defaults:        defaults,
		memory:          memory,
		version:         version,
		changeCallbacks: make(map[string]map[int]keyChangedEvent),
		defaultKeys:     defaultKeys,
		strictness:      strictness,
	}, nil
}

func complain(msg string, strictness Strictness) {
	switch strictness {
	case StrictnessWarn:
		log.Println(msg)
	case StrictnessPanic:
		panic(msg)
	default:
		return
	}
}

// Reload re-sets all values in the config to the data in `dec`.
// If `dec` is nil, all default values will be returned.
// All keys that changed will trigger a signal, if registered.
//
// Note that you cannot pass different defaults on Reload,
// since this might alter the structure of the config,
// potentially causing incompatibillies. Use the migration
// interface if you really need to change the layout.
func (cfg *Config) Reload(dec Decoder) error {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	var memory map[interface{}]interface{}
	var version Version
	var err error

	if dec != nil {
		version, memory, err = dec.Decode()
		if err != nil {
			return err
		}
	} else {
		memory = make(map[interface{}]interface{})
		version = Version(0)
	}

	// old config used to check for old values:
	oldCfg := &Config{
		memory:        cfg.memory,
		version:       cfg.version,
		defaults:      cfg.defaults,
		callbackCount: cfg.callbackCount,
		defaultKeys:   cfg.defaultKeys,
		section:       cfg.section,
		strictness:    cfg.strictness,
	}

	if err := validationChecker(memory, cfg.defaults, cfg.defaultKeys, cfg.strictness); err != nil {
		return e.Wrapf(err, "validate")
	}

	cfg.memory = memory
	cfg.version = version

	// Keys did not change, since it's the same defaults:
	callbacks := []keyChangedEvent{}
	for _, key := range cfg.keys() {
		if !reflect.DeepEqual(oldCfg.get(key), cfg.get(key)) {
			callbacks = append(callbacks, cfg.gatherCallbacks(key)...)
		}
	}

	for _, callback := range callbacks {
		callback.fn(callback.key)
	}

	return nil
}

// Save will write a representation defined by `enc` of the current config to `w`.
func (cfg *Config) Save(enc Encoder) error {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	return enc.Encode(cfg.version, cfg.memory)
}

////////////

// splitKey splits `key` into it's parent container and base key
func (cfg *Config) splitKey(key string, sectionAllowed bool) (map[interface{}]interface{}, string) {
	return splitKeyRecursive(strings.Split(key, "."), cfg.memory, sectionAllowed)
}

// actual worker for splitKey
func splitKeyRecursive(keys []string, root map[interface{}]interface{}, sectionAllowed bool) (map[interface{}]interface{}, string) {
	if len(keys) == 0 {
		return nil, ""
	}

	child, ok := root[keys[0]]
	if !ok {
		return nil, ""
	}

	section, ok := child.(map[interface{}]interface{})
	if !ok {
		if len(keys) > 1 {
			return nil, ""
		}

		// scalar type, return immediately.
		return root, keys[0]
	}

	if sectionAllowed && len(keys) == 1 {
		return root, keys[0]
	}

	return splitKeyRecursive(keys[1:], section, sectionAllowed)
}

// get is the worker for the higher level typed accessors
func (cfg *Config) get(key string) interface{} {
	key = prefixKey(cfg.section, key)
	parent, base := cfg.splitKey(key, false)
	if parent == nil {
		// It is not present in cfg.memory.
		// Maybe it's an entry below __many__?
		defEntry := getDefaultByKey(key, cfg.defaults, cfg.strictness)
		if defEntry != nil {
			return defEntry.Default
		}

		complain(
			fmt.Sprintf("bug: invalid config key: %v", key),
			cfg.strictness,
		)
		return nil
	}

	return parent[base]
}

// call this with cfg.mu locked!
func (cfg *Config) gatherCallbacks(key string) []keyChangedEvent {
	callbacks := []keyChangedEvent{}
	for _, ckey := range []string{key, ""} {
		if ckey == "" || strings.HasPrefix(ckey, cfg.section) {
			if bucket, ok := cfg.changeCallbacks[ckey]; ok {
				for _, callback := range bucket {
					callbacks = append(callbacks, callback)
				}
			}
		}
	}

	return callbacks
}

func (cfg *Config) punchHole(key []string, root map[interface{}]interface{}) (map[interface{}]interface{}, string, error) {
	if len(key) == 0 {
		return nil, "", nil
	}

	if len(key) == 1 {
		return root, key[0], nil
	}

	child, ok := root[key[0]]
	if !ok {
		child = make(map[interface{}]interface{})
		root[key[0]] = child
	}

	section, ok := child.(map[interface{}]interface{})
	if !ok {
		return nil, "", fmt.Errorf("trying to override value with section: %v", key)
	}

	return cfg.punchHole(key[1:], section)
}

// setLocked is worker behind the Set*() methods.
func (cfg *Config) setLocked(key string, val interface{}) error {
	cfg.mu.Lock()

	key = prefixKey(cfg.section, key)
	callbacks := []keyChangedEvent{}
	defer func() {
		// Call the callbacks without the lock:
		for _, callback := range callbacks {
			callback.fn(callback.key)
		}
	}()

	// NOTE: the unlock is called before the other defer!
	defer cfg.mu.Unlock()

	parent, base := cfg.splitKey(key, false)
	if parent == nil {
		// section, sectionBase := cfg.splitKey(key, true)
		// mergeDefaults(, , cfg.defaultKeys, cfg.section)
		def := getDefaultByKey(key, cfg.defaults, cfg.strictness)
		if def != nil {
			splitKey := strings.Split(key, ".")

			var err error
			parent, base, err = cfg.punchHole(splitKey, cfg.memory)
			if err != nil {
				return err
			}

			parent[base] = def.Default
		} else {
			msg := fmt.Sprintf("bug: invalid config key: %v", key)
			complain(msg, cfg.strictness)
			return errors.New(msg)
		}
	}

	defType := getTypeOf(parent[base])
	valType := getTypeOf(val)

	if !isCompatibleType(defType, valType) {
		return fmt.Errorf(
			"wrong type in set for key `%v`: want: `%v` but got `%v`",
			key, defType, valType,
		)
	}

	// Remember that we've overwritten this key:
	delete(cfg.defaultKeys, key)

	// Check if something was changed. If not we do not need to notify anyone.
	if reflect.DeepEqual(val, parent[base]) {
		return nil
	}

	// If there is an validator defined, we should check now.
	defEntry := getDefaultByKey(key, cfg.defaults, cfg.strictness)
	if defEntry.Validator != nil {
		if err := defEntry.Validator(val); err != nil {
			return err
		}
	}

	parent[base] = val
	callbacks = cfg.gatherCallbacks(key)

	return nil
}

////////////

// AddEvent registers a callback to be called when `key` is changed.
// Special case: if key is the empy string, the registered callback will get
// called for every change (with the respective key)
// This function supports registering several callbacks for the same `key`.
// The returned id can be used to unregister a callback with RemoveEvent()
// Note: This function will panic when using an invalid key.
func (cfg *Config) AddEvent(key string, fn func(key string)) int {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	event := keyChangedEvent{
		fn:  fn,
		key: key,
	}

	if key != "" {
		key = prefixKey(cfg.section, key)
		defaultEntry := getDefaultByKey(key, cfg.defaults, cfg.strictness)
		if defaultEntry == nil {
			complain(
				fmt.Sprintf("bug: invalid config key: %v", key),
				cfg.strictness,
			)
			return 0
		}
	}

	callbacks, ok := cfg.changeCallbacks[key]
	if !ok {
		callbacks = make(map[int]keyChangedEvent)
		cfg.changeCallbacks[key] = callbacks
	}

	oldCount := cfg.callbackCount
	callbacks[oldCount] = event

	cfg.callbackCount++

	return oldCount
}

// RemoveEvent removes a previously registered callback.
func (cfg *Config) RemoveEvent(id int) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	toDelete := []string{}
	for key, bucket := range cfg.changeCallbacks {
		delete(bucket, id)
		if len(bucket) == 0 {
			toDelete = append(toDelete, key)
		}
	}

	for _, key := range toDelete {
		delete(cfg.changeCallbacks, key)
	}
}

// ClearEvents removes all registered events.
func (cfg *Config) ClearEvents() {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	cfg.changeCallbacks = make(map[string]map[int]keyChangedEvent)
}

////////////

func (cfg *Config) checkZeroType(key string, val, zero interface{}) interface{} {
	valTyp := reflect.TypeOf(val)
	zeroTyp := reflect.TypeOf(zero)

	if valTyp.ConvertibleTo(zeroTyp) {
		// We can convert the types, all good.
		return reflect.ValueOf(val).Convert(zeroTyp).Interface()
	}

	// Let's complain about the wrong type.
	// The user of the api probably used he wrong type for this key.
	// (i.e. Bool() for a string)
	complain(
		fmt.Sprintf(
			"bug: wrong type in get for key `%s`. Want `%s`, but got `%s`. Wrong getter used?",
			key,
			zeroTyp.Name(),
			valTyp.Name(),
		),
		cfg.strictness,
	)
	return zero
}

// Get returns the raw value at `key`.
// Do not use this method when possible, use the typeed convinience methods.
// Note: This function might panic when they key does not exist and StrictnessPanic is used.
// If an error happens it will return the zero value.
func (cfg *Config) Get(key string) interface{} {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	return cfg.get(key)
}

// Bool returns the boolean value (or default) at `key`.
// Note: This function might panic when they key does not exist and StrictnessPanic is used.
// If an error happens it will return the zero value.
func (cfg *Config) Bool(key string) bool {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	val := cfg.get(key)
	if val == nil {
		return false
	}

	return cfg.checkZeroType(key, val, false).(bool)
}

// String returns the string value (or default) at `key`.
// Note: This function might panic when they key does not exist and StrictnessPanic is used.
// If an error happens it will return the zero value.
func (cfg *Config) String(key string) string {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	val := cfg.get(key)
	if val == nil {
		return ""
	}

	return cfg.checkZeroType(key, val, "").(string)
}

// Int returns the int value (or default) at `key`.
// Note: This function might panic when they key does not exist and StrictnessPanic is used.
// If an error happens it will return the zero value.
func (cfg *Config) Int(key string) int64 {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	val := cfg.get(key)
	if val == nil {
		return 0
	}

	return cfg.checkZeroType(key, val, int64(0)).(int64)
}

// Float returns the float value (or default) at `key`.
// Note: This function might panic when they key does not exist and StrictnessPanic is used.
// If an error happens it will return the zero value.
func (cfg *Config) Float(key string) float64 {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	val := cfg.get(key)
	if val == nil {
		return 0
	}

	return cfg.checkZeroType(key, val, float64(0)).(float64)
}

// Duration returns the duration value (or default) at `key`.
// Note: This function might panic when they key does not exist and StrictnessPanic is used.
// If an error happens it will return the zero value.
func (cfg *Config) Duration(key string) time.Duration {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	val := cfg.get(key)
	if val == nil {
		return time.Duration(0)
	}

	s := cfg.checkZeroType(key, val, "").(string)
	d, err := time.ParseDuration(s)
	if err != nil {
		complain(
			fmt.Sprintf("invalid duration: %v; use the duration validator!", s),
			cfg.strictness,
		)
		return time.Duration(0)
	}

	return d
}

// Strings returns the string list value (or default) at `key`.
// Note: This function might panic when they key does not exist and StrictnessPanic is used.
// If an error happens it will return the zero value.
func (cfg *Config) Strings(key string) []string {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	val := cfg.get(key)
	if val == nil {
		return nil
	}

	return cfg.checkZeroType(key, val, []string{}).([]string)
}

// Ints returns the int list value (or default) at `key`.
// Note: This function might panic when they key does not exist and StrictnessPanic is used.
// If an error happens it will return the zero value.
func (cfg *Config) Ints(key string) []int64 {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	val := cfg.get(key)
	if val == nil {
		return nil
	}

	return cfg.checkZeroType(key, val, []int64{}).([]int64)
}

// Floats returns the float list value (or default) at `key`.
// Note: This function might panic when they key does not exist and StrictnessPanic is used.
// If an error happens it will return the zero value.
func (cfg *Config) Floats(key string) []float64 {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	val := cfg.get(key)
	if val == nil {
		return nil
	}

	return cfg.checkZeroType(key, val, []float64{}).([]float64)
}

// Bools returns the boolean list value (or default) at `key`.
// Note: This function might panic when they key does not exist and StrictnessPanic is used.
// If an error happens it will return the zero value.
func (cfg *Config) Bools(key string) []bool {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	val := cfg.get(key)
	if val == nil {
		return nil
	}

	return cfg.checkZeroType(key, val, []bool{}).([]bool)
}

// Durations returns the duration value (or default) at `key`.
// Note: This function might panic when they key does not exist and StrictnessPanic is used.
// If an error happens it will return the zero value.
func (cfg *Config) Durations(key string) []time.Duration {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	val := cfg.get(key)
	if val == nil {
		return nil
	}

	strings := cfg.checkZeroType(key, val, []string{}).([]string)
	durations := []time.Duration{}

	for _, s := range strings {
		d, err := time.ParseDuration(s)
		if err != nil {
			complain(
				fmt.Sprintf("invalid duration: %v; use the durations validator!", s),
				cfg.strictness,
			)
			return nil
		}

		durations = append(durations, d)
	}

	return durations
}

////////////

// IsDefault will return true if this key was not explicitly set,
// but taken over from the defaults.
// Note: This function might panic when they key does not exist and StrictnessPanic is used.
func (cfg *Config) IsDefault(key string) bool {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	_, ok := cfg.defaultKeys[key]
	return ok
}

// Merge takes all values from `other` that were set explicitly
// and sets them in `cfg`. If any key changes, the respective
// event callback will be called.
func (cfg *Config) Merge(other *Config) error {
	cfg.mu.Lock()

	callbacks := []keyChangedEvent{}
	defer func() {
		// Call the callbacks without the lock:
		for _, callback := range callbacks {
			callback.fn(callback.key)
		}
	}()

	if !reflect.DeepEqual(cfg.defaults, other.defaults) {
		return fmt.Errorf("refusing to merge configs with different defaults")
	}

	// NOTE: the unlock is called before the other defer!
	defer cfg.mu.Unlock()

	for _, key := range cfg.keys() {
		if _, ok := other.defaultKeys[key]; ok {
			// It is a default key on the other side.
			// No need to set it, since we might have
			// overwritten this key.
			continue
		}

		oldVal := cfg.get(key)
		newVal := other.get(key)

		// Only use callbacks if the key really changed:
		if !reflect.DeepEqual(newVal, oldVal) {
			callbacks = append(callbacks, cfg.gatherCallbacks(key)...)
			parent, base := cfg.splitKey(key, false)
			parent[base] = newVal
		}
	}

	return nil
}

////////////

// SetBool creates or sets the `val` at `key`.
// Note: This function might panic when they key does not exist and StrictnessPanic is used.
func (cfg *Config) SetBool(key string, val bool) error {
	return cfg.setLocked(key, val)
}

// SetString creates or sets the `val` at `key`.
// Note: This function might panic when they key does not exist and StrictnessPanic is used.
func (cfg *Config) SetString(key string, val string) error {
	return cfg.setLocked(key, val)
}

// SetInt creates or sets the `val` at `key`.
// Note: This function might panic when they key does not exist and StrictnessPanic is used.
func (cfg *Config) SetInt(key string, val int64) error {
	return cfg.setLocked(key, val)
}

// SetFloat creates or sets the `val` at `key`.
// Note: This function might panic when they key does not exist and StrictnessPanic is used.
func (cfg *Config) SetFloat(key string, val float64) error {
	return cfg.setLocked(key, val)
}

// SetDuration creates or sets the `val` at `key`.
// Note: This function might panic when they key does not exist and StrictnessPanic is used.
func (cfg *Config) SetDuration(key string, val time.Duration) error {
	return cfg.setLocked(key, val.String())
}

// SetBools creates or sets the `val` at `key`.
// Note: This function might panic when they key does not exist and StrictnessPanic is used.
func (cfg *Config) SetBools(key string, val []bool) error {
	return cfg.setLocked(key, val)
}

// SetStrings creates or sets the `val` at `key`.
// Note: This function might panic when they key does not exist and StrictnessPanic is used.
func (cfg *Config) SetStrings(key string, val []string) error {
	return cfg.setLocked(key, val)
}

// SetInts creates or sets the `val` at `key`.
// Note: This function might panic when they key does not exist and StrictnessPanic is used.
func (cfg *Config) SetInts(key string, val []int64) error {
	return cfg.setLocked(key, val)
}

// SetFloats creates or sets the `val` at `key`.
// Note: This function might panic when they key does not exist and StrictnessPanic is used.
func (cfg *Config) SetFloats(key string, val []float64) error {
	return cfg.setLocked(key, val)
}

// SetDurations creates or sets the `val` at `key`.
// Note: This function might panic when they key does not exist and StrictnessPanic is used.
func (cfg *Config) SetDurations(key string, val []time.Duration) error {
	strings := []string{}
	for _, d := range val {
		strings = append(strings, d.String())
	}

	return cfg.setLocked(key, strings)
}

// Set creates or sets the `val` at `key`.
// Please only use this function only if you have an interface{}
// that you do not want to cast yourself.
// Note: This function might panic when they key does not exist and StrictnessPanic is used.
func (cfg *Config) Set(key string, val interface{}) error {
	return cfg.setLocked(key, val)
}

////////////

// GetDefault retrieves the default for a certain key.
// Note: This function might panic when they key does not exist and StrictnessPanic is used.
func (cfg *Config) GetDefault(key string) DefaultEntry {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	// The lock here is probably not necessary,
	// since we wont't modify defaults.
	key = prefixKey(cfg.section, key)
	entry := getDefaultByKey(key, cfg.defaults, cfg.strictness)
	if entry == nil {
		complain(fmt.Sprintf("bug: invalid config key: %v", key), cfg.strictness)
		return DefaultEntry{}
	}

	return *entry
}

// Keys returns all possible keys (including the default keys)
func (cfg *Config) Keys() []string {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	return cfg.keys()
}

func (cfg *Config) keys() []string {
	allKeys := []string{}
	err := keys(cfg.memory, nil, func(section map[interface{}]interface{}, key []string) error {
		fullKey := strings.Join(key, ".")
		if strings.HasPrefix(fullKey, cfg.section) {
			if len(cfg.section) != 0 {
				fullKey = fullKey[len(cfg.section)+1:]
			}
			allKeys = append(allKeys, fullKey)
		}

		return nil
	})

	if err != nil {
		// keys() should only return an error if the function passed to it
		// error in some way. Since we don't do that it should not produce
		// any non-nil error return.
		complain(fmt.Sprintf("Keys() failed internally: %v", err), cfg.strictness)
		return nil
	}

	sort.Strings(allKeys)
	return allKeys
}

// Section returns a new config that
func (cfg *Config) Section(section string) *Config {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	return &Config{
		// mutex is shared with parent, since they protect the same memory.
		mu:            cfg.mu,
		section:       section,
		callbackCount: cfg.callbackCount,
		// The data is shared, any set to a section will cause a set in the parent.
		defaults: cfg.defaults,
		memory:   cfg.memory,
		// Sections may have own callbacks.
		// The parent callbacks are still called though.
		changeCallbacks: cfg.changeCallbacks,
		strictness:      cfg.strictness,
	}
}

// IsValidKey can be checked to see if untrusted keys actually are valid.
// It should not be used to check keys from string literals.
func (cfg *Config) IsValidKey(key string) bool {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	key = prefixKey(cfg.section, key)
	return getDefaultByKey(key, cfg.defaults, cfg.strictness) != nil
}

const sliceSeparator = " ;; "

func castStringSlice(val string) ([]string, error) {
	res := []string{}
	for _, val := range strings.Split(val, sliceSeparator) {
		res = append(res, val)
	}

	return res, nil
}

func castIntSlice(val string) ([]int64, error) {
	res := []int64{}
	for _, val := range strings.Split(val, sliceSeparator) {
		conv, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return nil, err
		}

		res = append(res, conv)
	}

	return res, nil
}

func castFloatSlice(val string) ([]float64, error) {
	res := []float64{}
	for _, val := range strings.Split(val, sliceSeparator) {
		conv, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return nil, err
		}

		res = append(res, conv)
	}

	return res, nil
}

func castBoolSlice(val string) ([]bool, error) {
	res := []bool{}
	for _, val := range strings.Split(val, sliceSeparator) {
		conv, err := strconv.ParseBool(val)
		if err != nil {
			return nil, err
		}

		res = append(res, conv)
	}

	return res, nil
}

// Cast takes `val` and reads the type of `key`.  It then tries to convert it
// to one of the supported types (and possibly fails due to that)
//
// This cast assumes that `val` is always a string, which is useful for data
// coming fom the client.  Note: This function will panic if the key does not
// exist.
func (cfg *Config) Cast(key, val string) (interface{}, error) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	key = prefixKey(cfg.section, key)
	entry := getDefaultByKey(key, cfg.defaults, cfg.strictness)
	if entry == nil {
		msg := fmt.Sprintf("bug: invalid config key: %v", key)
		complain(msg, cfg.strictness)
		return nil, errors.New(msg)
	}

	switch entry.Default.(type) {
	case int, int16, int32, int64, uint, uint16, uint32, uint64:
		return strconv.ParseInt(val, 10, 64)
	case float32, float64:
		return strconv.ParseFloat(val, 64)
	case bool:
		return strconv.ParseBool(val)
	case string:
		return val, nil
	case []int, []int16, []int32, []int64, []uint, []uint16, []uint32, []uint64:
		return castIntSlice(val)
	case []float32, []float64:
		return castFloatSlice(val)
	case []bool:
		return castBoolSlice(val)
	case []string:
		return castStringSlice(val)
	}

	return nil, nil
}

// Uncast is the opposite of Cast. It works like Get(), but always returns
// a stringified version of a value, even in case of slices. The representation
// returned by Uncast() can be later read in again by using Cast().
//
// Cast() and Uncast() are mainly useful in systems where you can only use strings,
// e.g. when building an API between different programming languages.
// Note: Slice items are separated by " ;; ".
func (cfg *Config) Uncast(key string) string {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	fullKey := prefixKey(cfg.section, key)
	entry := getDefaultByKey(fullKey, cfg.defaults, cfg.strictness)
	if entry == nil {
		msg := fmt.Sprintf("bug: invalid config key: %v", fullKey)
		complain(msg, cfg.strictness)
		return ""
	}

	switch entry.Default.(type) {
	case []int, []int16, []int32, []int64, []uint, []uint16, []uint32, []uint64:
		res := []string{}
		for _, val := range cfg.get(key).([]int64) {
			res = append(res, strconv.FormatInt(val, 10))
		}

		return strings.Join(res, sliceSeparator)
	case []float32, []float64:
		res := []string{}
		for _, val := range cfg.get(key).([]float64) {
			res = append(res, strconv.FormatFloat(val, 'f', -1, 64))
		}

		return strings.Join(res, sliceSeparator)
	case []bool:
		res := []string{}
		for _, val := range cfg.get(key).([]bool) {
			res = append(res, strconv.FormatBool(val))
		}

		return strings.Join(res, sliceSeparator)
	case []string:
		return strings.Join(cfg.get(key).([]string), sliceSeparator)
	}

	return fmt.Sprintf("%v", cfg.get(key))
}

// Version returns the version of the config The initial version is always 0.
// It will be only updated by migrating to newer versions.
func (cfg *Config) Version() Version {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	return cfg.version
}

// Reset reverts a key or a section to its defaults.
// If key points to a value, only this value is reset.
// If key points to a section, all keys in it are reset.
// If key is an empty string, the whole config is reset to defaults.
//
// When calling reset on parts of the config that includes __many__ sections,
// those will be totally cleared and won't be serialized when calling Save().
// Retrieving values from those will yield default values as if they were not set.
func (cfg *Config) Reset(key string) error {
	cfg.mu.Lock()

	key = prefixKey(cfg.section, key)
	entry := getDefaultByKey(key, cfg.defaults, cfg.strictness)
	if entry != nil {
		// Key points to a value.
		cfg.defaults[key] = struct{}{}
		cfg.mu.Unlock()
		return cfg.setLocked(key, entry.Default)
	}

	defer cfg.mu.Unlock()

	if key == "" {
		// The whole config needs to be reset:
		cfg.memory = make(map[interface{}]interface{})
		return mergeDefaults(cfg.memory, cfg.defaults, cfg.defaultKeys, cfg.section)
	}

	// We need to clear a section:
	splitKey := strings.Split(key, ".")
	defaultSection := getDefaultSectionByKeys(
		splitKey[:len(splitKey)-1],
		cfg.defaults,
		cfg.strictness,
	)

	if defaultSection == nil {
		msg := fmt.Sprintf("no such section: %v", key)
		complain(msg, cfg.strictness)
		return errors.New(msg)
	}

	parent, base := cfg.splitKey(key, true)
	if parent == nil {
		msg := fmt.Sprintf("bug: invalid config key: %v", key)
		complain(msg, cfg.strictness)
		return errors.New(msg)
	}

	delete(parent, base)
	return mergeDefaults(parent, defaultSection, cfg.defaultKeys, cfg.section)
}
