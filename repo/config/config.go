// Package config implements a very opinionated config utility.  It relies on a
// "default spec", i.e. a structure that defines all existing configuration
// keys, their types and their initial default values.  This is used as
// fallback and source of validation. The idea is similar to python's configobj
// (albeit much smaller). Surprisingly I didn't find any similar library in Go.
//
// Note that passing invalid keys to a few methods will cause a panic - on purpose.
// Using a wrong config key is seen as a bug and should be corrected immediately.
// This allows this package to skip error handling on Get() and Set() entirely.
// Also note that I'm not particularly proud of some parts of this code.
package config

import (
	"fmt"
	"io"
	"io/ioutil"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"sync"

	e "github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// DefaultEntry represents the metadata for a default value in the config.
// Every possible key has to have a DefaultEntry.
type DefaultEntry struct {
	// Default is the fallback value for this config key.
	// The confg type will be inferred from its literal type.
	Default interface{}

	// NeedsRestart indicates that we need to restart the daemon
	// to have an effect here.
	NeedsRestart bool

	// Docs describes the meaning of the configuration value.
	Docs string
}

// DefaultMapping is a container to hold all required DefaultEntries.
// It is a nested map with sections as string keys.
type DefaultMapping map[interface{}]interface{}

var (
	typeIntPattern   = regexp.MustCompile(`u{0,1}int(64|32|16|8|)`)
	typeFloatPattern = regexp.MustCompile(`float(32|64|)`)
)

func getDefaultByKeys(keys []string, defaults DefaultMapping) *DefaultEntry {
	if len(keys) == 0 {
		return nil
	}

	child, ok := defaults[keys[0]]
	if !ok {
		return nil
	}

	defaultEntry, ok := child.(DefaultEntry)
	if ok {
		if len(keys) > 1 {
			return nil
		}

		// scalar type, return immediately.
		return &defaultEntry
	}

	section, ok := child.(DefaultMapping)
	if !ok {
		panic(fmt.Errorf("got bad type in default table: %T", child))
	}

	return getDefaultByKeys(keys[1:], section)
}

func getDefaultByKey(key string, defaults DefaultMapping) *DefaultEntry {
	return getDefaultByKeys(strings.Split(key, "."), defaults)
}

func getTypeOf(val interface{}) string {
	typ := reflect.TypeOf(val)
	if typ == nil {
		return ""
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

	return typeA == typeB
}

func getTypeOfDefaultKey(key string, defaults DefaultMapping) string {
	defauttEntry := getDefaultByKey(key, defaults)
	if defauttEntry == nil {
		return ""
	}

	return getTypeOf(defauttEntry.Default)
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

func mergeDefaults(base map[interface{}]interface{}, overlay DefaultMapping) error {
	for keyVal := range overlay {
		key, ok := keyVal.(string)
		if !ok {
			return fmt.Errorf("config contains non string keys: %v", keyVal)
		}

		switch overlayChild := overlay[key].(type) {
		case DefaultMapping:
			baseSection, ok := base[key].(map[interface{}]interface{})
			if !ok {
				baseSection = make(map[interface{}]interface{})
				base[key] = baseSection
			}

			if err := mergeDefaults(baseSection, overlayChild); err != nil {
				return err
			}
		case DefaultEntry:
			if _, ok := base[key]; !ok {
				base[key] = overlayChild.Default
			}
		}
	}

	return nil
}

func validationChecker(root map[interface{}]interface{}, defaults DefaultMapping, prefix []string) error {
	err := keys(root, nil, func(section map[interface{}]interface{}, key []string) error {
		// It's a scalar key. Let's run some diagnostics.
		lastKey := key[len(key)-1]
		child := section[lastKey]

		fullKey := strings.Join(key, ".")
		defType := getTypeOfDefaultKey(fullKey, defaults)
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

		// Handle a few special cases here that come from go's type system.
		// Doing something like this will lead to a panic:
		//
		//     interface{}(int(42)).(int64)
		//
		// Since this is a config we do not care very much for extremely
		// big numbers and can therefore convert all numbers to int64.
		// The code below does that + something similar for float{32,64}.

		if typeIntPattern.MatchString(valType) {
			destType := reflect.TypeOf(int64(0))
			section[lastKey] = reflect.ValueOf(child).Convert(destType).Int()
		}

		if typeFloatPattern.MatchString(valType) {
			destType := reflect.TypeOf(float64(0))
			section[lastKey] = reflect.ValueOf(child).Convert(destType).Float()
		}

		// Valid key.
		return nil
	})

	if err != nil {
		return err
	}

	return mergeDefaults(root, defaults)
}

////////////

// Config s a helper that built is around a YAML file.
// It supports typed gets and sets, change notifications and
// basic validation with defaults.
type Config struct {
	mu sync.Mutex

	defaults        DefaultMapping
	memory          map[interface{}]interface{}
	callbackCount   int
	onChangeSignals map[string]map[int]func(string)
}

// Open creates a new config from the data in `r`.
// The mapping in `defaults ` tells the config which keys to expect
// and what type each of it should have.
func Open(r io.Reader, defaults DefaultMapping) (*Config, error) {
	if defaults == nil {
		return nil, fmt.Errorf("need a default mapping")
	}

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	memory := make(map[interface{}]interface{})
	if err := yaml.Unmarshal(data, memory); err != nil {
		return nil, err
	}

	if err := validationChecker(memory, defaults, []string{}); err != nil {
		return nil, e.Wrapf(err, "validate")
	}

	return &Config{
		defaults:        defaults,
		memory:          memory,
		onChangeSignals: make(map[string]map[int]func(string)),
	}, nil
}

// Save will write a YAML representation of the current config to `w`.
func (cfg *Config) Save(w io.Writer) error {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	data, err := yaml.Marshal(cfg.memory)
	if err != nil {
		return err
	}

	if _, err := w.Write(data); err != nil {
		return err
	}

	return nil
}

////////////

// splitKey splits `key` into it's parent container and base key
func (cfg *Config) splitKey(key string) (map[interface{}]interface{}, string) {
	return splitKeyRecursive(strings.Split(key, "."), cfg.memory)
}

// actual worker for splitKey
func splitKeyRecursive(keys []string, root map[interface{}]interface{}) (map[interface{}]interface{}, string) {
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

	return splitKeyRecursive(keys[1:], section)
}

// get is the worker for the higher level typed accessors
func (cfg *Config) get(key string) interface{} {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	parent, base := cfg.splitKey(key)
	if parent == nil {
		panic(fmt.Sprintf("bug: invalid config key: %v", key))
	}

	return parent[base]
}

// set is worker behind the Set*() methods.
func (cfg *Config) set(key string, val interface{}) {
	cfg.mu.Lock()

	fns := []func(string){}
	defer func() {
		// Call the callbacks without the lock:
		for _, fn := range fns {
			fn(key)
		}
	}()

	// Note that the unlock is called before the other defer.
	defer cfg.mu.Unlock()

	parent, base := cfg.splitKey(key)
	if parent == nil {
		panic(fmt.Sprintf("bug: invalid config key: %v", key))
	}

	defType := getTypeOf(parent[base])
	valType := getTypeOf(val)

	if !isCompatibleType(defType, valType) {
		cfg.mu.Unlock()
		panic(
			fmt.Sprintf(
				"bug: wrong type in set for key `%v`: want: %v but got %v",
				key, defType, valType,
			),
		)
	}

	parent[base] = val

	// Gather callbacks while still holding the lock:
	for _, ckey := range []string{key, ""} {
		if callbacks, ok := cfg.onChangeSignals[ckey]; ok {
			for _, callback := range callbacks {
				fns = append(fns, callback)
			}
		}
	}
}

////////////

// AddChangedKeyEvent registers a callback to be called when `key` is changed.
// Special case: if key is the empy string, the registered callback will get
// called for every change (with the respective key)
// This function supports registering several callbacks for the same `key`.
// The returned id can be used to unregister a callback with RemoveChangedKeyEvent()
// Note: This function will panic when using an invalid key.
func (cfg *Config) AddChangedKeyEvent(key string, fn func(key string)) int {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	if key != "" {
		defaultEntry := getDefaultByKey(key, cfg.defaults)
		if defaultEntry == nil {
			panic(fmt.Sprintf("bug: invalid config key: %v", key))
		}
	}

	callbacks, ok := cfg.onChangeSignals[key]
	if !ok {
		callbacks = make(map[int]func(string))
		cfg.onChangeSignals[key] = callbacks
	}

	oldCount := cfg.callbackCount
	callbacks[oldCount] = fn
	cfg.callbackCount++

	return oldCount
}

// RemoveChangedKeyEvent removes a previously registered callback.
// Note: This function will panic when using an invalid key.
func (cfg *Config) RemoveChangedKeyEvent(key string, id int) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	if key != "" {
		defaultEntry := getDefaultByKey(key, cfg.defaults)
		if defaultEntry == nil {
			panic(fmt.Sprintf("bug: invalid config key: %v", key))
		}
	}

	callbacks, ok := cfg.onChangeSignals[key]
	if !ok {
		return
	}

	delete(callbacks, id)
	if len(callbacks) == 0 {
		delete(cfg.onChangeSignals, key)
	}
}

////////////

// Bool returns the boolean value (or default) at `key`.
// Note: This function will panic if the key does not exist.
func (cfg *Config) Bool(key string) bool {
	return cfg.get(key).(bool)
}

// String returns the string value (or default) at `key`.
// Note: This function will panic if the key does not exist.
func (cfg *Config) String(key string) string {
	return cfg.get(key).(string)
}

// Int returns the int value (or default) at `key`.
// Note: This function will panic if the key does not exist.
func (cfg *Config) Int(key string) int64 {
	return cfg.get(key).(int64)
}

// Float returns the float value (or default) at `key`.
// Note: This function will panic if the key does not exist.
func (cfg *Config) Float(key string) float64 {
	return cfg.get(key).(float64)
}

////////////

// SetBool creates or sets the `val` at `key`.
// Note: This function will panic if the key does not exist.
func (cfg *Config) SetBool(key string, val bool) {
	cfg.set(key, val)
}

// SetString creates or sets the `val` at `key`.
// Note: This function will panic if the key does not exist.
func (cfg *Config) SetString(key string, val string) {
	cfg.set(key, val)
}

// SetInt creates or sets the `val` at `key`.
// Note: This function will panic if the key does not exist.
func (cfg *Config) SetInt(key string, val int64) {
	cfg.set(key, val)
}

// SetFloat creates or sets the `val` at `key`.
// Note: This function will panic if the key does not exist.
func (cfg *Config) SetFloat(key string, val float64) {
	cfg.set(key, val)
}

// Set creates or sets the `val` at `key`.
// Please only use this function only if you have an interface{}
// that you do not want to cast yourself.
// Note: This function will panic if the key does not exist.
func (cfg *Config) Set(key string, val interface{}) {
	cfg.set(key, val)
}

////////////

// GetDefault retrieves the default for a certain key.
// Note: This function will panic if the key does not exist.
func (cfg *Config) GetDefault(key string) DefaultEntry {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	// The lock here is probably not necessary,
	// since we wont't modify defaults.
	entry := getDefaultByKey(key, cfg.defaults)
	if entry == nil {
		panic(fmt.Sprintf("bug: invalid config key: %v", key))
	}

	return *entry
}

// Keys returns all keys that are currently set (including the default keys)
func (cfg *Config) Keys() ([]string, error) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	allKeys := []string{}
	err := keys(cfg.memory, nil, func(section map[interface{}]interface{}, key []string) error {
		allKeys = append(allKeys, strings.Join(key, "."))
		return nil
	})

	if err != nil {
		return nil, err
	}

	sort.Strings(allKeys)
	return allKeys, nil
}
