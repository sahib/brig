package repo

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"

	yml "gopkg.in/yaml.v2"
)

var Defaults = []struct {
	Name  string
	Value interface{}
}{
	{"daemon.port", 6666},
	{"sync.ignore_removed", false},
	{"sync.conflict_strategy", "marker"},
	{"data.ipfs.path", ""},
	{"data.compress.algo", "snappy"},
}

func setConfigDefaults(config *viper.Viper) error {
	for _, fallback := range Defaults {
		config.SetDefault(fallback.Name, fallback.Value)
	}

	return nil
}

func recursiveSet(defaults map[string]interface{}, key []string, val interface{}) {
	if len(key) > 1 {
		sub, ok := defaults[key[0]]
		if !ok {
			sub = make(map[string]interface{})
			defaults[key[0]] = sub
		}

		subMap, ok := sub.(map[string]interface{})
		if !ok {
			return
		}

		recursiveSet(subMap, key[1:], val)
		return
	}

	defaults[key[0]] = val
}

func buildConfigDefault() []byte {
	defaults := make(map[string]interface{})
	for _, entry := range Defaults {
		key := strings.Split(entry.Name, ".")
		recursiveSet(defaults, key, entry.Value)
	}

	data, err := yml.Marshal(defaults)
	if err != nil {
		panic(fmt.Sprintf("Failed to convert default config to yml: %v", err))
	}

	return data
}

func buildMetaDefault(backendName, owner string) []byte {
	data, err := yml.Marshal(map[string]interface{}{
		"data": map[string]string{
			"backend": backendName,
		},
		"repo": map[string]string{
			"owner": owner,
		},
	})

	if err != nil {
		panic(fmt.Sprintf("Failed to convert default meta to yml: %v", err))
	}

	return data
}
