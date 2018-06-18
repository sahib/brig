package config

import "os"

// FromFile creates a new config from the YAML file located at `path`
func FromFile(path string, defaults DefaultMapping) (cfg *Config, err error) {
	fd, fErr := os.Open(path)
	if fErr != nil {
		return nil, fErr
	}

	defer func() {
		if clErr := fd.Close(); clErr != nil && err == nil {
			err = clErr
		}
	}()

	cfg, err = Open(fd, defaults)
	return
}

// ToFile saves `cfg` as YAML at a file located at `path`.
func ToFile(path string, cfg *Config) (err error) {
	fd, fErr := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if fErr != nil {
		return fErr
	}

	defer func() {
		if clErr := fd.Close(); clErr != nil && err == nil {
			err = clErr
		}
	}()

	err = cfg.Save(fd)
	return
}
