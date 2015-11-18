package util

import (
	"fmt"
	"testing"
)

func TestDerive(t *testing.T) {
	key, salt, _ := DeriveAESKey("elch@jabber.nullcat.de", "Katznwald", 32)
	fmt.Printf("Key: % x\nSalt: % x\n", key, salt)
}
