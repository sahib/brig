package terminaldimensions

import "testing"

func TestParser(t *testing.T) {
	x, _, _ := parse("123 456\n")
	if x != uint(123) {
		t.Errorf("Expected: %d Got: %d", uint(123), x)
	}
}
