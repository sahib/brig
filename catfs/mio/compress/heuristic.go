package compress

import (
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/sdemontfort/go-mimemagic"
)

var (
	// TextFileExtensions not covered by mime.TypeByExtension
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
	httpMatch := http.DetectContentType(buf)
	if httpMatch != "application/octet-stream" {
		return httpMatch
	}

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

// GuessAlgorithm takes the path name and the header data of it
// and tries to guess a suitable compression algorithm.
func GuessAlgorithm(path string, header []byte) (AlgorithmType, error) {
	if len(header) < HeaderSizeThreshold {
		return AlgoNone, nil
	}

	mime := guessMime(path, header)
	if mime == "" {
		// the guesses below work only when mime is known
		return AlgoSnappy, nil
	}

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
