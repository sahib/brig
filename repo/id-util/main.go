package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/cathalgarvey/go-minilock"
	"github.com/disorganizer/brig/repo"
)

func main() {
	genflag := flag.String("g", "", "Generate flag to generate a MID for a specific user.")
	encflag := flag.String("e", "", "Encrypt.")
	decflag := flag.String("d", "", "Decrypt.")
	flag.Parse()

	if *genflag != "" {
		jabberid, password := repo.GetUserlogin(*genflag)
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
		jid, pass := repo.GetUserlogin("")
		fmt.Println(jid, pass, mid, plaintext)
		encMsg, _ := repo.EncryptMSG(jid, pass, mid, plaintext, "MagicByte")
		fmt.Println(encMsg)
	}

	if *decflag != "" {
		encMsg := *decflag
		decMsg, _ := repo.DecryptMSG(flag.Arg(1), flag.Arg(2), encMsg)
		fmt.Println(decMsg)
	}
}
