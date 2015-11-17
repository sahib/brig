package main

import (
	"flag"
	"fmt"

	"github.com/cathalgarvey/go-minilock"
	zxcvbn "github.com/nbutton23/zxcvbn-go"
	"os"
)

func encryptMSG(jid, pass, mid, plaintext, filename string) (string, error) {
	ciphertext, err := minilock.EncryptFileContentsWithStrings(filename, []byte(plaintext), jid, pass, true, mid)
	if err != nil {
		return "", nil
	}
	return string(ciphertext), nil
}

func decryptMSG(jid, pass, msg string) (string, error) {
	userKey, err := minilock.GenerateKey(jid, pass)
	if err != nil {
		return "", nil
	}
	_, _, plaintext, _ := minilock.DecryptFileContents([]byte(msg), userKey)
	return string(plaintext), nil
}

func getUserlogin(jabberid string) (string, string) {
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

func main() {

	genflag := flag.String("g", "", "Generate flag to generate a MID for a specific user.")
	encflag := flag.String("e", "", "Encrypt.")
	decflag := flag.String("d", "", "Decrypt.")
	flag.Parse()

	if *genflag != "" {
		jabberid, password := getUserlogin(*genflag)
		keys, _ := minilock.GenerateKey(jabberid, password)
		mid, err := keys.EncodeID()
		if err != nil {
			fmt.Println(err)
			os.Exit(-3)
		}
		fmt.Println("JabberID: ", jabberid, " MID: ", mid)
	}

	if *encflag != "" {
		mid := *encflag
		plaintext := flag.Arg(1)
		jid, pass := getUserlogin("")
		fmt.Println(jid, pass, mid, plaintext)
		encMsg, _ := encryptMSG(jid, pass, mid, plaintext, "MagicByte")
		fmt.Println(encMsg)
	}

	if *decflag != "" {
		encMsg := *decflag
		decMsg, _ := decryptMSG(flag.Arg(1), flag.Arg(2), encMsg)
		fmt.Println(decMsg)
	}
}
