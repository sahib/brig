package main

import (
	"flag"
	"log"
	"os"

	"github.com/disorganizer/brig/bit/format"
)

func main() {
	decryptMode := flag.Bool("d", false, "Decrypt.")
	flag.Parse()

	key := []byte("01234567890ABCDE01234567890ABCDE")

	var err error
	if *decryptMode == false {
		_, err = format.Encrypt(key, os.Stdin, os.Stdout)
	} else {
		_, err = format.Decrypt(key, os.Stdin, os.Stdout)
	}

	if err != nil {
		log.Fatal(err)
		return
	}
}
