package main

import (
	"flag"
	"log"
	"os"
	"time"

	multihash "github.com/jbenet/go-multihash"
)

type File interface {
	// Path relative to the repo root
	Path() string

	// File size of the file in bytes
	Size() int

	// Modification timestamp (with timezone)
	Mtime() time.Time

	// Hash of the unencrypted file
	Hash() multihash.Multihash

	// Hash of the encrypted file from IPFS
	IpfsHash() multihash.Multihash
}

func NewFile(path string) (*File, error) {
	// TODO:
	return nil, nil
}

func main() {
	decryptMode := flag.Bool("d", false, "Decrypt.")
	flag.Parse()

	key := []byte("01234567890ABCDE01234567890ABCDE")

	var err error
	if *decryptMode == false {
		_, err = Encrypt(key, os.Stdin, os.Stdout)
	} else {
		source, _ := os.Open("/tmp/dump")

		_, err = Decrypt(key, source, os.Stdout)
	}

	if err != nil {
		log.Fatal(err)
		return
	}
}
