package main

import (
	"log"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"bitbucket.org/taruti/mimemagic"
	"github.com/fatih/color"
)

var (
	Threshold       = int64(1024)
	NotCompressable = []string{"video", "audio", "image", "zip", "rar"}
)

func isLargeEnough(f *os.File) bool {
	st, err := f.Stat()
	if err != nil {
		log.Fatal(err)
	}
	return st.Size() >= Threshold
}

func guessMime(buf []byte, path string) string {
	s := mimemagic.Match("", buf)
	if s == "" {
		s = mime.TypeByExtension(filepath.Ext(path))
	}
	if s == "" {
		s = "unknown filetype"
	}
	return s
}

func isCompressable(mimetype string) bool {
	for _, substr := range NotCompressable {
		if strings.Contains(mimetype, substr) {
			return false
		}
	}
	return true
}

func openFile(path string) *os.File {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	return file
}

func printFormated(path, mime string, compressable, largeEnough bool) {
	if compressable && largeEnough {
		color.Green("✔: %s,[%s]\n", path, mime)
	}

	if !largeEnough || !compressable {
		color.Red("✗: %s, [%s]\n", path, mime)
	}
}

func main() {
	buf := make([]byte, 1024)
	for _, path := range os.Args {
		if fd := openFile(path); fd != nil {
			fd.Read(buf)
			m := guessMime(buf, path)
			c := isCompressable(m)
			e := isLargeEnough(fd)
			printFormated(path, m, c, e)
			fd.Close()
		}
	}
}
