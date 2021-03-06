package peer

import "testing"

var validityTests = []struct {
	id       string
	ok       bool
	user     string
	domain   string
	resource string
}{
	{"", false, "", "", ""},
	{"\n", false, "", "", ""},
	{"a", true, "a", "", ""},
	{"ä", true, "ä", "", ""},
	{string([]byte{0xC2}), false, "", "", ""}, // Invalid utf8 start byte
	{"alice", true, "alice", "", ""},
	{"alice bobsen", false, "", "", ""},
	{"alice@nullcat.de", true, "alice", "nullcat.de", ""},
	{"alice@nullcat.de/laptop", true, "alice", "nullcat.de", "laptop"},
	{"alice @nullcat.de/laptop", false, "", "", ""},
}

func TestValidity(t *testing.T) {
	for _, test := range validityTests {
		valid := IsValid(test.id)
		if valid != test.ok {
			t.Errorf("valid(`%s`) was `%t`, should be `%t`", test.id, valid, test.ok)
			continue
		}

		if !valid {
			continue
		}

		id, err := CastName(test.id)
		if err != nil {
			t.Errorf("Casting `%s` failed: %v", test.id, err)
			continue
		}

		if id.User() != test.user {
			t.Errorf(
				"User differs; wanted `%s`; got `%s`",
				test.user, id.User(),
			)
			continue
		}

		if id.Domain() != test.domain {
			t.Errorf(
				"Domain differs; wanted `%s`; got `%s`",
				test.domain, id.Domain(),
			)
			continue
		}

		if id.Resource() != test.resource {
			t.Errorf(
				"Resource differs; wanted `%s`; got `%s`",
				test.resource, id.Resource(),
			)
			continue
		}
	}
}
