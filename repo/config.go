package repo

import (
	"os"

	"github.com/olebedev/config"
)

// LoadConfig loads a yaml configuration file.
func LoadConfig(path string) (*config.Config, error) {
	cfg, err := config.ParseYamlFile(path)

	if err != nil {
		return nil, err
	}
	return cfg, nil
}

// SaveConfig saves a given config as yaml encoded configuration file.
func SaveConfig(path string, cfg *config.Config) (int, error) {
	yamlString, err := config.RenderYaml(cfg.Root)
	if err != nil {
		return 0, err
	}
	file, err := os.Create(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	written, err := file.WriteString(yamlString)
	if err != nil {
		return 0, err
	}

	return written, nil
}

// Make default config template a bit prettier:
type section map[string]interface{}

// CreateDefaultConfig creates a configfile with default values.
func CreateDefaultConfig(path string) (int, error) {
	cfg := section{
		"development": section{
			"database": section{
				"host": "localhost",
			},
			"users": []interface{}{
				section{
					"name":     "calvin",
					"password": "yukon",
				},
				section{
					"name":     "hobbes",
					"password": "tuna",
				},
			},
		},
		"production": section{
			"database": section{
				"host": "192.168.1.1",
			},
		},
	}

	defaultCfg := config.Config{Root: cfg}
	return SaveConfig(path, &defaultCfg)
}
