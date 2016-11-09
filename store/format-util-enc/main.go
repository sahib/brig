package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/disorganizer/brig/store/compress"
	"github.com/disorganizer/brig/store/encrypt"
	//"github.com/disorganizer/brig/util/pwd"
	"golang.org/x/crypto/scrypt"
)

const (
	aeadCipherChaCha = iota
	aeadCipherAES
)

type options struct {
	algo              string
	encalgo           string
	args              []string
	compress          bool
	encrypt           bool
	maxblocksize      int64
	decompress        bool
	useDevNull        bool
	forceDstOverwrite bool
}

func withTime(fn func()) time.Duration {
	now := time.Now()
	fn()
	return time.Since(now)
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

func dstFilename(compressor bool, src, algo string) string {
	if compressor {
		return fmt.Sprintf("%s.%s", src, algo)
	}
	return fmt.Sprintf("%s.%s", src, "uncompressed")
}

func dieWithUsage() {
	fmt.Printf("Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(-1)

}

func die(err error) {
	log.Fatal(err)
	os.Exit(-1)
}

func parseFlags() options {
	decompress := flag.Bool("d", false, "Decompress.")
	compress := flag.Bool("c", false, "Compress.")
	encrypt := flag.Bool("e", false, "Enable encryption/decryption.")
	maxblocksize := flag.Int64("b", 1, "BlockSize.")
	algo := flag.String("a", "none", "Possible compression algorithms: none, snappy, lz4.")
	encalgo := flag.String("n", "aes", "Possible encryption algorithms: aes, chacha.")
	forceDstOverwrite := flag.Bool("f", false, "Force overwriting destination file.")
	useDevNull := flag.Bool("D", false, "Write to /dev/null.")
	flag.Parse()
	return options{
		decompress:        *decompress,
		compress:          *compress,
		encrypt:           *encrypt,
		algo:              *algo,
		encalgo:           *encalgo,
		maxblocksize:      *maxblocksize,
		forceDstOverwrite: *forceDstOverwrite,
		useDevNull:        *useDevNull,
		args:              flag.Args(),
	}
}

func derivateAesKey(pwd, salt []byte, keyLen int) []byte {
	key, err := scrypt.Key(pwd, salt, 16384, 8, 1, keyLen)
	if err != nil {
		panic("Bad scrypt parameters: " + err.Error())
	}
	return key
}

func main() {
	opts := parseFlags()

	if len(opts.args) != 1 {
		dieWithUsage()
	}
	if opts.compress && opts.decompress {
		dieWithUsage()
	}
	if !opts.compress && !opts.decompress {
		dieWithUsage()
	}

	srcPath := opts.args[0]
	algo, err := compress.FromString(opts.algo)
	if err != nil {
		die(err)
	}

	src := openSrc(srcPath)
	defer src.Close()

	dstPath := dstFilename(opts.compress, srcPath, opts.algo)
	if opts.useDevNull {
		dstPath = os.DevNull
	}

	dst := openDst(dstPath, opts.forceDstOverwrite)
	defer dst.Close()

	key := derivateAesKey([]byte("defaultpassword"), nil, 32)
	if key == nil {
		die(err)
	}
	var chiper uint16 = aeadCipherAES
	if opts.encalgo == "chacha" {
		chiper = aeadCipherChaCha
	}

	if opts.encalgo == "aes" {
		chiper = aeadCipherAES
	}

	if opts.encalgo == "none" {
		opts.encrypt = false
	}

	if opts.compress {
		ew := io.WriteCloser(dst)
		if opts.encrypt {
			ew, err = encrypt.NewWriterWithTypeAndBlockSize(dst, key, chiper, opts.maxblocksize)
			if err != nil {
				die(err)
			}
		}
		zw, err := compress.NewWriter(ew, algo)
		if err != nil {
			die(err)
		}
		_, err = io.Copy(zw, src)
		if err != nil {
			die(err)
		}
		if err := zw.Close(); err != nil {
			die(err)
		}
		if err := ew.Close(); err != nil {
			die(err)
		}
	}
	if opts.decompress {
		var reader io.ReadSeeker = src
		if opts.encrypt {
			er, err := encrypt.NewReader(src, key)
			if err != nil {
				die(err)
			}
			reader = er
		}
		zr := compress.NewReader(reader)
		_, err = io.Copy(dst, zr)
		if err != nil {
			die(err)
		}
	}
}
