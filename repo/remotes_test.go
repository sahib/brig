package repo

import (
	"bytes"
	"strings"
	"testing"
)

var TestYml = `alice@wonderland.de /:
- read
alice@wonderland.de /public:
- read
- write
`

func TestRemoteList(t *testing.T) {
	rl, err := NewRemotes(strings.NewReader(TestYml))
	if err != nil {
		t.Fatalf("Failed to load remote list: %v", err)
	}

	buf := &bytes.Buffer{}
	if err := rl.Export(buf); err != nil {
		t.Fatalf("Failed to export remote list: %v", err)
	}

	out := string(buf.Bytes())
	if out != TestYml {
		t.Fatalf("Exported remote list differs from input: %v", out)
	}
}
