package config

import "os"

// FromYamlFile creates a new config from the YAML file located at `path`
func FromYamlFile(path string, defaults DefaultMapping) (cfg *Config, err error) {
	fd, fErr := os.Open(path)
	if fErr != nil {
		return nil, fErr
	}

	defer func() {
		if clErr := fd.Close(); clErr != nil && err == nil {
			err = clErr
		}
	}()

	cfg, err = Open(NewYamlDecoder(fd), defaults)
	return cfg, err
}

// ToYamlFile saves `cfg` as YAML at a file located at `path`.
func ToYamlFile(path string, cfg *Config) (err error) {
	fd, fErr := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if fErr != nil {
		return fErr
	}

	defer func() {
		if clErr := fd.Close(); clErr != nil && err == nil {
			err = clErr
		}
	}()

	return cfg.Save(NewYamlEncoder(fd))
}
