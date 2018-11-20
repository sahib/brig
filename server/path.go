package server

import (
	"path"
	"strings"
)

// URL is a path to a file or directory,
// that might optionally include a user name.
type URL struct {
	User string
	Path string
}

func prefixSlash(p string) string {
	if !strings.HasPrefix(p, "/") {
		return "/" + p
	}

	return p
}

func clean(p string) string {
	return prefixSlash(path.Clean(p))
}

func parsePath(p string) (*URL, error) {
	if strings.HasPrefix(p, "/") {
		// no user part in there.
		return &URL{Path: clean(p), User: ""}, nil
	}

	if idx := strings.IndexRune(p, ':'); idx <= 0 || idx >= len(p)-1 {
		return &URL{Path: clean(p), User: ""}, nil
	}

	split := strings.SplitN(p, ":", 2)
	return &URL{
		Path: clean(split[1]),
		User: split[0],
	}, nil
}
