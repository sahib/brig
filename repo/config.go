package repo

import "github.com/spf13/viper"

var Defaults = []struct {
	Name  string
	Value interface{}
}{
	{"daemon.port", 6666},
	{"sync.ignore_removed", false},
	{"sync.conflict_strategy", "marker"},
	{"data.backend", "memory"},
	{"data.ipfs.swarmport", 4001},
	{"data.ipfs.path", ""},
	{"data.compress.algo", "snappy"},
}

func setConfigDefaults(config *viper.Viper) error {
	for _, fallback := range Defaults {
		config.SetDefault(fallback.Name, fallback.Value)
	}

	return nil
}
