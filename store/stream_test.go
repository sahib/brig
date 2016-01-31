package store

import (
	"bytes"
	"io"
	"testing"
)

var TestKey = []byte("01234567890ABCDE01234567890ABCDE")

func TestWriteAndRead(t *testing.T) {
	raw := []byte("Hello World")
	rawBuf := &bytes.Buffer{}
	rawBuf.Write(raw)

	encStream, err := NewFileReader(TestKey, rawBuf)
	if err != nil {
		t.Errorf("Creating encryption stream failed: %v", err)
		return
	}

	encrypted := &bytes.Buffer{}
	if _, err := io.Copy(encrypted, encStream); err != nil {
		t.Errorf("Reading encrypted data failed: %v", err)
		return
	}

	decStream, err := NewIpfsReader(TestKey, bytes.NewReader(encrypted.Bytes()))
	if err != nil {
		t.Errorf("Creating decryption stream failed: %v", err)
		return
	}

	decrypted := &bytes.Buffer{}
	if _, err := io.Copy(decrypted, decStream); err != nil {
		t.Errorf("Reading decrypted data failed: %v", err)
		return
	}

	if !bytes.Equal(decrypted.Bytes(), raw) {
		t.Errorf("Raw and decrypted is not equal => BUG.")
		t.Errorf("RAW:\n  %v", raw)
		t.Errorf("DEC:\n  %v", decrypted.Bytes())
		return
	}
}
