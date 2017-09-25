package catfs

type Config struct {
	CompressAlgo         string
	SyncIgnoreRemoved    bool
	SyncConflictStrategy string
}

var DefaultConfig = &Config{
	CompressAlgo:         "snappy",
	SyncIgnoreRemoved:    false,
	SyncConflictStrategy: "add",
}

type config struct {
}

func (cfg *Config) parseConfig() (*config, error) {
	return nil, nil
}
