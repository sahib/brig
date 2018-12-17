package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/sahib/brig/catfs/mio/compress"
	"github.com/sahib/brig/catfs/mio/encrypt"
)

func process(r io.ReadSeeker, w io.Writer, readEncrypted, readCompress, writeEncrypted, writeCompress bool) error {
	key := make([]byte, 32)

	if readEncrypted {
		rEnc, err := encrypt.NewReader(r, key)
		if err != nil {
			return err
		}

		r = rEnc
	}

	if readCompress {
		r = compress.NewReader(r)
	}

	if writeEncrypted {
		wEnc, encErr := encrypt.NewWriter(w, key)
		if encErr != nil {
			return encErr
		}

		defer wEnc.Close()
		w = wEnc
	}

	if writeCompress {
		wZip, zipErr := compress.NewWriter(w, compress.AlgoLZ4)
		if zipErr != nil {
			return zipErr
		}

		defer wZip.Close()
		w = wZip
	}

	n, err := io.Copy(w, r)
	if err != nil {
		return err
	}

	fmt.Printf("Wrote %d bytes.\n", n)
	return nil
}

func main() {
	inputFlag := flag.String("input", "", "input path")
	outputFlag := flag.String("output", "", "output path")

	encryptFlag := flag.Bool("encrypt", false, "Do encryption?")
	decryptFlag := flag.Bool("decrypt", false, "Do decryption?")

	compressFlag := flag.Bool("compress", false, "Do compression?")
	decompressFlag := flag.Bool("decompress", false, "Do decompression?")

	flag.Parse()

	if *inputFlag == "" {
		fmt.Println("Please specify an input path.")
		os.Exit(1)
	}

	if *outputFlag == "" {
		fmt.Println("Please specify an output path.")
		os.Exit(1)
	}

	if *encryptFlag && *decryptFlag {
		fmt.Println("Cannot encrypt and decrypt at the same time.")
		os.Exit(1)
	}

	if *compressFlag && *decompressFlag {
		fmt.Println("Cannot compress and decompress at the same time.")
		os.Exit(1)
	}

	inFd, err := os.Open(*inputFlag)
	if err != nil {
		fmt.Printf("Failed to open input file: %v\n", err)
		os.Exit(1)
	}

	defer inFd.Close()

	outFd, err := os.OpenFile(*outputFlag, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		fmt.Printf("Failed to open output file: %v\n", err)
		os.Exit(1)
	}

	defer outFd.Close()

	if err := process(inFd, outFd, *decryptFlag, *decompressFlag, *encryptFlag, *compressFlag); err != nil {
		fmt.Printf("Processing failed: %v\n", err)
		os.Exit(2)
	}
}
