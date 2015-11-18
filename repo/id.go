package repo

import (
	"fmt"

	"github.com/cathalgarvey/go-minilock"
	zxcvbn "github.com/nbutton23/zxcvbn-go"
)

func EncryptMSG(jid, pass, plaintext, filename string, selfenc bool, mid ...string) (string, error) {
	ciphertext, err := minilock.EncryptFileContentsWithStrings(filename, []byte(plaintext), jid, pass, selfenc, mid...)
	if err != nil {
		return "", nil
	}
	return string(ciphertext), nil
}

func DecryptMSG(jid, pass, msg string) (string, error) {
	userKey, err := minilock.GenerateKey(jid, pass)
	if err != nil {
		return "", nil
	}
	_, _, plaintext, _ := minilock.DecryptFileContents([]byte(msg), userKey)
	return string(plaintext), nil
}

// TODO(elk): bad name?
func GetUserlogin(jabberid string) (string, string) {
	var username string
	if jabberid == "" {
		fmt.Print("JabberID: ")
		fmt.Scanln(&username)
	} else {
		username = jabberid
	}

	var password string
	for {
		fmt.Print("Password:")
		fmt.Scanln(&password)
		passStrength := zxcvbn.PasswordStrength(password, []string{})
		if passStrength.Entropy < 60 {
			fmt.Printf("Password is to weak (%f bits).\n", passStrength.Entropy)
			continue
		}
		break
	}
	return username, password
}
