package tunnel

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestTunnel(t *testing.T) {
	m := &bytes.Buffer{}

	ta, err := NewEllipticTunnel(m)
	if err != nil {
		t.Errorf("Could not create tunnel: %v", err)
		return
	}

	n, err := ta.Write([]byte("Hello"))
	if n != 5 {
		t.Errorf("Short write on tunnel: %v", err)
		return
	}

	if m.Len() > 0 && string(m.Bytes()) == "Hello" {
		t.Errorf("Tunnel failed to encrypt: %v", m)
		return
	}

	n, err = ta.Write([]byte("World"))
	if n != 5 {
		t.Errorf("Short write on tunnel: %v", err)
		return
	}

	data, _ := ioutil.ReadAll(ta)

	if string(data) != "HelloWorld" {
		t.Errorf("decrypted differs from source.")
		t.Errorf("\tWant: %v", "HelloWorld")
		t.Errorf("\tGot:  %v", string(data))
		return
	}
}
