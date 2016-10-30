package compress

import (
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"bitbucket.org/taruti/mimemagic"
)

var (
	// TODO Is it useful to make threshold dependend on mime type?
	Threshold = int64(1024)
	// Whitelist of all known uncompressed formats, that would be filtered out
	// by the following blacklist
	Compressable = []string{
		"image/bmp",
		"audio/x-wav",
	}
	// Blacklist
	NotCompressable = []string{
		"application/ogg",
		"video",
		"audio",
		"image",
		"zip",
		"rar",
		"7z",
	}
	// Textfile extensions not covered by mime.TypeByExtension
	TextFileExtensions = []string{
		".go",
		".json",
		".yaml",
		".xml",
		".txt",
	}
)

func guessMime(path string, buf []byte) string {
	s := mimemagic.Match("", buf)
	if s == "" {
		s = mime.TypeByExtension(filepath.Ext(path))
	}
	for _, extension := range TextFileExtensions {
		if extension == filepath.Ext(path) {
			return "text/generic"
		}
	}
	return s
}

func isCompressable(mimetype string) bool {
	for _, substr := range Compressable {
		if strings.Contains(mimetype, substr) {
			return true
		}
	}
	for _, substr := range NotCompressable {
		if strings.Contains(mimetype, substr) {
			return false
		}
	}
	return true
}

func ChooseCompressAlgo(repoPath string, rs io.ReadSeeker) (AlgorithmType, error) {
	buf := make([]byte, Threshold)
	bytesRead, err := rs.Read(buf)
	if err != nil {
		return AlgoNone, err
	}

	mime := guessMime(repoPath, buf)
	compressAble := isCompressable(mime)

	if _, err := rs.Seek(0, os.SEEK_SET); err != nil {
		return AlgoNone, err
	}

	if !compressAble || int64(bytesRead) != Threshold {
		return AlgoNone, nil
	}

	if strings.HasPrefix(mime, "text/") {
		return AlgoLZ4, nil
	} else {
		return AlgoSnappy, nil
	}
}
