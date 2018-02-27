package server

import (
	"path"
	"strings"
)

// Path parsing utilities.

type Url struct {
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

func parsePath(p string) (*Url, error) {
	if strings.HasPrefix(p, "/") {
		// no user part in there.
		return &Url{Path: clean(p), User: ""}, nil
	}

	if idx := strings.IndexRune(p, ':'); idx <= 0 || idx >= len(p)-1 {
		return &Url{Path: clean(p), User: ""}, nil
	}

	split := strings.SplitN(p, ":", 2)
	return &Url{
		Path: clean(split[1]),
		User: split[0],
	}, nil
}
