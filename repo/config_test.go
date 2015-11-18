package repo

import (
	"fmt"
	"testing"
)

func TestConfig(t *testing.T) {
	CreateDefaultConfig("/tmp/nein")
	c, err := LoadConfig("config.yaml")

	if err != nil {
		t.Errorf("Unable to create config: %v", err)
		return
	}

	fmt.Println(c.String("animalzoo.overlord.name"))
	fmt.Println(c.Root)
	c.Set("animalzoo.overlord.name", "Gabriele")
	fmt.Println(c.Root)
	v, _ := c.Int("hs-augsburg.mensa.offen")
	fmt.Printf("%d\n", v)
	_, _ = SaveConfig("config.yaml", c)
}
