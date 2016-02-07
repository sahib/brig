package main

import (
	"flag"
	"log"
	"os"

	"github.com/disorganizer/brig/store/encrypt"
)

func main() {
	decryptMode := flag.Bool("d", false, "Decrypt.")
	flag.Parse()

	key := []byte("01234567890ABCDE01234567890ABCDE")

	var err error
	if *decryptMode == false {
		_, err = encrypt.Encrypt(key, os.Stdin, os.Stdout, 0)
	} else {
		_, err = encrypt.Decrypt(key, os.Stdin, os.Stdout)
	}

	if err != nil {
		log.Fatal(err)
		return
	}
}
