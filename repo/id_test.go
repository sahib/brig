package repo

import (
	"fmt"
	"testing"
)

type User struct {
	Name     string
	Jid      string
	Password string
	Mid      string
}

func encrypt(jid, password, secretMsg, filename string, selfenc bool, mid ...string) (string, error) {
	ciphertext, err := EncryptMSG(jid, password, secretMsg, filename, selfenc, mid...)
	return ciphertext, err
}

func maliciousUserHasDecrypted(decryptedText, originalText string, user *User, maliciousUsers ...*User) bool {
	for _, maliciousUser := range maliciousUsers {
		// message successfully decrypted by a malicious user?
		if decryptedText == originalText && user == maliciousUser {
			return true
		}
	}
	return false
}

func (u *User) String() string {
	return fmt.Sprintf("%s", u.Name)
}

func TestID(t *testing.T) {

	// sender
	alice := &User{
		"Alice",
		"alice@jabber.de/laptop",
		"3lrj;2lq3rj;lkqjwflkjwf",
		"Jw7xyd3jrG4d4TkQmUzDKLwbH9RPcEV47SAFRJtCEFY6c",
	}

	// receivers
	bob := &User{
		"Bob,",
		"bob@jabber.de/work",
		"lk23j4lk2jlk3j4l2k3j12333",
		"2JHpZWEypyBNxN1pe6mptBa4uFsNwj54r3DXegdLGuKanh",
	}
	bruce := &User{
		"Bruce",
		"bruce@jabber.de/rsa",
		"l3kjr;l33;)JLJK90092",
		"j9VD7e2vgrxbxJX4i3ut4AGg47S8yoyJN5793ti1NNdWc",
	}

	// indruder
	micrathene := &User{
		"Micrathene",
		"micrathene@jabber.de/forest",
		"lijk3lk*(3l#KJ8#:Lk#",
		"cewNAcGCRoqbB95JfgAyHFpXv4ka7hroUUkqQEx6vpdVE",
	}

	originalText := "This is a very secret message."
	receivers := []*User{bob, bruce}
	receiverMids := []string{}
	for _, receiver := range receivers {
		receiverMids = append(receiverMids, receiver.Mid)
	}
	fmt.Printf("%s encrypts for %s\n", alice, receivers)
	ciphertext, err := encrypt(alice.Jid, alice.Password, originalText, "0xDEADBEEF", false, receiverMids...)
	if err != nil {
		t.Log("Error enctypting plaintext.", err)
	}
	for _, user := range []*User{alice, bob, bruce, micrathene} {
		decryptedtext, _ := DecryptMSG(user.Jid, user.Password, ciphertext)
		if maliciousUserHasDecrypted(decryptedtext, originalText, user, micrathene, alice /* malicious users*/) {
			t.Errorf("%s souldn't be able to decrypt the ciphertext.\n", user.Jid)
		}
		fmt.Printf("User %s tries to encrypt: %t\n", user, decryptedtext == originalText)
	}
}
