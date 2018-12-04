package config

import "os"

// FromYamlFile creates a new config from the YAML file located at `path`
func FromYamlFile(path string, defaults DefaultMapping, strictness Strictness) (*Config, error) {
	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	cfg, err := Open(NewYamlDecoder(fd), defaults, strictness)
	if err != nil {
		return nil, err
	}

	return cfg, fd.Close()
}

// ToYamlFile saves `cfg` as YAML at a file located at `path`.
func ToYamlFile(path string, cfg *Config) error {
	fd, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}

	if err := cfg.Save(NewYamlEncoder(fd)); err != nil {
		return err
	}

	return fd.Close()
}
