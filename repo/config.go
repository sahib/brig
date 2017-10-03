package repo

import "github.com/spf13/viper"

func setConfigDefaults(config *viper.Viper) error {
	config.SetDefault("daemon.port", 6666)
	config.SetDefault("sync.ignore_removed", false)
	config.SetDefault("sync.conflict_strategy", "marker")
	config.SetDefault("data.ipfs.swarmport", 4001)
	config.SetDefault("data.ipfs.path", "")
	config.SetDefault("data.compress.algo", "snappy")
	return nil
}
