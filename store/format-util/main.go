package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/disorganizer/brig/store/compress"
)

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
		os.Exit(-1)
	}
}

func openDst(dest string, overwrite bool) *os.File {
	if !overwrite {
		if _, err := os.Stat(dest); !os.IsNotExist(err) {
			log.Fatalf("Opening destination failed, %v exists.\n", dest)
		}
	}
	fd, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		log.Fatalf("Opening destination %v failed: %v\n", dest, err)
	}
	return fd
}

func openSrc(src string) *os.File {
	fd, err := os.Open(src)
	if err != nil {
		log.Fatalf("Opening source %v failed: %v\n", src, err)
	}
	return fd
}

func getDstFilename(compressor bool, src, algo string) string {
	if compressor {
		return fmt.Sprintf("%s.%s", src, algo)
	}
	return fmt.Sprintf("%s.%s", src, "uncompressed")
}

func main() {
	algorithms := map[string]int{
		"none":   0,
		"snappy": 1,
		"lz4":    2,
	}
	decompressMode := flag.Bool("d", false, "Decompress.")
	compressMode := flag.Bool("c", false, "Compress.")
	useAlgo := flag.String("s", "none", "Possible compression algorithms: none, snappy, lz4.")
	forceOverwrite := flag.Bool("f", false, "Force overwriting destination file.")
	flag.Parse()
	Args := flag.Args()

	if len(Args) != 1 || !((*compressMode || *decompressMode) && !(*compressMode && *decompressMode)) {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(-1)
	}

	srcPath := Args[0]
	algo, ok := algorithms[*useAlgo]
	if !ok {
		log.Fatalf("Invalid algorithm type: %s", *useAlgo)
		os.Exit(-1)
	}

	src := openSrc(srcPath)
	dstFileName := getDstFilename(*compressMode, srcPath, *useAlgo)
	dst := openDst(dstFileName, *forceOverwrite)
	defer dst.Close()
	defer src.Close()

	nBytes, err := int64(0), errors.New("huh, this should never happen")
	if *compressMode {
		zw, err := compress.NewWriter(dst, compress.AlgorithmType(algo))
		checkError(err)
		nBytes, err = io.Copy(zw, src)
		checkError(err)
		zw.Close()
	}
	if *decompressMode {
		zr := compress.NewReader(src)
		nBytes, err = io.Copy(dst, zr)
		checkError(err)
	}
	fmt.Printf("%s created, %d bytes processed.\n", dstFileName, nBytes)
}
