package config

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

// CreateDefaultConfig creates a configfile with default values.
func CreateDefaultConfig() *config.Config {
	cfg := map[string]interface{}{
		"repository": map[string]interface{}{
			"jid":  "",
			"mid":  "",
			"uuid": "",
		},
		"ipfs": map[string]interface{}{
			"port":     5001,
			"hostname": "localhost",
			"path":     "",
		},
	}

	return &config.Config{Root: cfg}
}
