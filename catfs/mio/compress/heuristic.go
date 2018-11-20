package compress

import (
	"mime"
	"path/filepath"
	"strings"

	"bitbucket.org/taruti/mimemagic"
)

var (
	// TextfileExtensions not covered by mime.TypeByExtension
	TextFileExtensions = map[string]bool{
		".go":   true,
		".json": true,
		".yaml": true,
		".xml":  true,
		".txt":  true,
	}
)

const (
	// HeaderSizeThreshold is the number of bytes needed to enable compression at all.
	HeaderSizeThreshold = 2048
)

func guessMime(path string, buf []byte) string {
	// try to guess it from the buffer we pass:
	match := mimemagic.Match("", buf)
	if match == "" {
		// try to guess it from the file path:
		match = mime.TypeByExtension(filepath.Ext(path))
	}

	// handle a few edge cases:
	if TextFileExtensions[filepath.Ext(path)] {
		return "text/plain"
	}

	return match
}

func isCompressible(mimetype string) bool {
	if strings.HasPrefix(mimetype, "text/") {
		return true
	}

	return CompressibleMapping[mimetype]
}

func GuessAlgorithm(path string, header []byte) (AlgorithmType, error) {
	if len(header) < HeaderSizeThreshold {
		return AlgoNone, nil
	}

	mime := guessMime(path, header)
	compressible := isCompressible(mime)

	if !compressible {
		return AlgoNone, nil
	}

	// text like files probably deserve some thorough compression:
	if strings.HasPrefix(mime, "text/") {
		return AlgoLZ4, nil
	}

	// fallback to snappy for generic files:
	return AlgoSnappy, nil
}
