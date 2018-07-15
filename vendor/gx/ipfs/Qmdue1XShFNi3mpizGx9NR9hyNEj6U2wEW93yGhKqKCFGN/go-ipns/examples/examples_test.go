package examples_test

import (
	"testing"

	"gx/ipfs/Qmdue1XShFNi3mpizGx9NR9hyNEj6U2wEW93yGhKqKCFGN/go-ipns/examples"
	crypto "gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
)

var testPath = "/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5"

func TestKeyGeneration(t *testing.T) {
	_, err := generateRSAKey()
	if err != nil {
		t.Error(err)
	}

	_, err = generateEDKey()
	if err != nil {
		t.Error(err)
	}
}

func TestEmbeddedEntryCreation(t *testing.T) {
	rk, err := generateRSAKey()
	if err != nil {
		t.Fatal(err)
	}

	ek, err := generateEDKey()
	if err != nil {
		t.Fatal(err)
	}
	_, err = examples.CreateEntryWithEmbed(testPath, rk.GetPublic(), rk)
	if err != nil {
		t.Error(err)
	}

	_, err = examples.CreateEntryWithEmbed(testPath, ek.GetPublic(), ek)
	if err != nil {
		t.Error(err)
	}

}
func generateRSAKey() (crypto.PrivKey, error) {
	// DO NOT USE 1024 BITS IN PRODUCTION
	// THIS IS ONLY FOR TESTING PURPOSES
	k, err := examples.GenerateRSAKeyPair(1024)
	if err != nil {
		return nil, err
	}
	return k, nil
}

func generateEDKey() (crypto.PrivKey, error) {
	// ED25519 uses 256bit keys, and ignore the bit param
	k, err := examples.GenerateEDKeyPair()
	if err != nil {
		return nil, err
	}
	return k, nil
}
