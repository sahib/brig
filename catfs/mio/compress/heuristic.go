package compress

import (
	"mime"
	"path/filepath"
	"strings"

	"bitbucket.org/taruti/mimemagic"
	log "github.com/Sirupsen/logrus"
)

var (
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

func ChooseCompressAlgo(path string, header []byte) (AlgorithmType, error) {
	mime := guessMime(path, header)
	compressAble := isCompressable(mime)

	log.Debugf("Guessed `%s` mime for `%s`", mime, path)
	if !compressAble {
		return AlgoNone, nil
	}

	if strings.HasPrefix(mime, "text/") {
		return AlgoLZ4, nil
	}

	return AlgoSnappy, nil
}
