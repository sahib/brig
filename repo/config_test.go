package repo

import "testing"

func TestDefaultConfig(t *testing.T) {
	defaults := string(buildConfigDefault())

	// TODO: Do more sensitive check here.
	if len(defaults) < 100 {
		t.Errorf("Default config looks bogus: %v", defaults)
	}
}
