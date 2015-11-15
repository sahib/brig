package main

import (
	"fmt"
	"github.com/olebedev/config"
	"os"
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
	written, err := file.WriteString(yamlString)
	if err != nil {
		return 0, err
	}

	defer file.Close()
	return written, nil
}

// CreateDefaultConfig creates a configfile with default values.
func CreateDefaultConfig(path string) (int, error) {
	cfg := map[string]interface{}{
		"development": map[string]interface{}{
			"database": map[string]interface{}{
				"host": "localhost",
			},
			"users": []interface{}{
				map[string]interface{}{
					"name":     "calvin",
					"password": "yukon",
				},
				map[string]interface{}{
					"name":     "hobbes",
					"password": "tuna",
				},
			},
		},
		"production": map[string]interface{}{
			"database": map[string]interface{}{
				"host": "192.168.1.1",
			},
		},
	}

	defaultCfg := config.Config{Root: cfg}
	return SaveConfig(path, &defaultCfg)
}

func main() {

	CreateDefaultConfig("nein")
	c, err := LoadConfig("config.yaml")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(c.String("animalzoo.overlord.name"))
	fmt.Println(c.Root)
	c.Set("animalzoo.overlord.name", "Gabriele")
	fmt.Println(c.Root)
	v, _ := c.Int("hs-augsburg.mensa.offen")
	fmt.Printf("%d\n", v)
	_, _ = SaveConfig("config.yaml", c)
}
