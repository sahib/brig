package config

import (
	"fmt"
	"testing"
)

const configPath = "/tmp/brig_test.cfg"

func TestConfig(t *testing.T) {
	fmt.Println("Creating default config.")
	cfg := CreateDefaultConfig()
	fmt.Println("Saving default config to ", configPath)
	if _, err := SaveConfig(configPath, cfg); err != nil {
		t.Errorf("Cannot save config: %v", err)
	}
	fmt.Println("Loading default config from ", configPath)
	c, err := LoadConfig(configPath)
	if err != nil {
		t.Errorf("Unable to load config: %v", err)
		return
	}

	inputValues := map[string]string{
		"repository.id":   "test@enterprise.fr/waffeln",
		"repository.uuid": "L@#K:JLKR:O#KJRLKQR",
		"ipfs.path":       "/tmp/katzenauge",
	}

	fmt.Println("\nSetting some test parameters...")
	for key, value := range inputValues {
		fmt.Printf("Setting %s to %s\n", key, value)
		if err := c.Set(key, value); err != nil {
			t.Errorf("Cannot set config value %s: %v", key, value)
			break
		}
	}

	fmt.Println("\nSaving config to ", configPath)
	if _, err = SaveConfig(configPath, c); err != nil {
		t.Errorf("Cannot save config: %v", err)
	}

	fmt.Println("Loading default config from ", configPath)
	c, err = LoadConfig(configPath)
	if err != nil {
		t.Errorf("Unable to load config: %v", err)
		return
	}

	fmt.Println("\nPrinting config after manipulating parameters...")
	expectedValues := map[string]interface{}{
		"repository.id":   "test@enterprise.fr/waffeln",
		"repository.uuid": "L@#K:JLKR:O#KJRLKQR",
		"repository.mid":  "",
		"ipfs.path":       "/tmp/katzenauge",
	}
	for key, expectedValue := range expectedValues {
		configValue, _ := c.String(key)
		fmt.Printf("Reading %s from config: %s\n", key, configValue)
		if configValue != expectedValue {
			t.Logf("%s read, but %s was expected.", configValue, expectedValue)
		}
	}
	configValue, _ := c.Int("ipfs.port")
	fmt.Printf("Reading %s from config: %d\n", "ipfs.port", configValue)
	if configValue != 5001 {
		t.Logf("%d read, but %d was expected.\n", configValue, 5001)
	}
}
