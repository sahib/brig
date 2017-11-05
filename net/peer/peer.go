// Package peer implements the basic data types needed to communicate
// with other brig instances.
//
// user[@domain[/resource]
package peer

import (
	"fmt"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

type Name string

type ErrBadName struct {
	reason string
}

func (e ErrBadName) Error() string {
	return e.reason
}

func valid(name string) error {
	if utf8.RuneCountInString(name) == 0 {
		return ErrBadName{"Empty name is not allowed"}
	}

	if !utf8.ValidString(name) {
		return ErrBadName{fmt.Sprintf("Invalid utf-8: %v", name)}
	}

	for idx, rn := range name {
		if unicode.IsSpace(rn) {
			return ErrBadName{
				fmt.Sprintf("Space not allowed: %s (at %d)", name, idx),
			}
		}

		if !unicode.IsPrint(rn) {
			return ErrBadName{
				fmt.Sprintf("Only printable runes allowed: %s (at %d)", name, idx),
			}
		}
	}

	return nil
}

// Cast checks `name` to be correct and returns
// a wrapped name.
func Cast(name string) (Name, error) {
	if err := valid(name); err != nil {
		return "", err
	}

	return Name(norm.NFKC.Bytes([]byte(name))), nil
}

func IsValid(name string) bool {
	return valid(name) == nil
}

func (name Name) Domain() string {
	a := strings.IndexRune(string(name), '@')
	if a < 0 {
		return ""
	}

	b := strings.LastIndexByte(string(name), '/')
	if b < 0 {
		return string(name)[a+1:]
	}

	return string(name)[a+1 : b]
}

func (name Name) Resource() string {
	idx := strings.LastIndexByte(string(name), '/')
	if idx < 0 {
		return ""
	}

	return string(name)[idx+1:]
}

func (name Name) WithoutResource() string {
	domain := name.Domain()
	if len(domain) > 0 {
		return name.User() + "@" + name.Domain()
	}

	return name.User()
}

func (name Name) AsPath() string {
	path := name.User()
	rsrc := name.Resource()
	if rsrc != "" {
		path += "-" + rsrc
	}

	return strings.Replace(path, string(os.PathSeparator), "|", -1)
}

func (name Name) User() string {
	idx := strings.Index(string(name), "@")
	if idx < 0 {
		return string(name)
	}

	return string(name)[:idx]
}

type Fingerprint string

func (fp Fingerprint) String() string {
	return ""
}

type Info struct {
	Addr string
}
