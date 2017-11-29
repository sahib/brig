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

	h "github.com/disorganizer/brig/util/hashlib"
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

// CastName checks `name` to be correct and returns
// a wrapped name.
func CastName(name string) (Name, error) {
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

/////////////////////////
// FINGERPRINT HELPERS //
/////////////////////////

// Fingerprint encodes the addr of a remote and an ID (i.e. hash)
// of the remote's public key. It is later used to verify if a
// remote's addr or pubkey has changed and is presented to the user
// as initial identification token for another user.
type Fingerprint string

func CastFingerprint(s string) (Fingerprint, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return Fingerprint(""), fmt.Errorf(
			"bad fingerprint: invalid num of colons: %d",
			len(parts),
		)
	}

	fp := Fingerprint(s)
	if fp.Addr() == "" {
		return Fingerprint(""), fmt.Errorf(
			"bad fingerprint: addr could not be read",
		)
	}

	if fp.PubKeyID() == "" {
		return Fingerprint(""), fmt.Errorf(
			"bad fingerprint: bad pub key id",
		)
	}

	return fp, nil
}

func BuildFingerprint(addr string, pubKeyData []byte) Fingerprint {
	s := fmt.Sprintf("%s:%s", addr, h.Sum(pubKeyData).B58String())
	return Fingerprint(s)
}

func (fp Fingerprint) Addr() string {
	// We assume that a fingerprint was always safely casted with Cast(),
	// so errors should not happen. They can of course still happen if the API
	// was not used correctly. Simply return the zero string in this case.
	parts := strings.SplitN(string(fp), ":", 2)
	if len(parts) < 2 {
		return ""
	}

	return parts[0]
}

func (fp Fingerprint) PubKeyID() string {
	parts := strings.SplitN(string(fp), ":", 2)
	if len(parts) < 2 {
		return ""
	}

	return parts[1]
}

func (fp Fingerprint) PubKeyMatches(pubKeyData []byte) bool {
	own := fp.PubKeyID()
	if own == "" {
		return false
	}

	remote := h.Sum(pubKeyData).B58String()
	return own == remote
}

///////////////////////

// TODO: Make Info -> Addr one day?
//       Having a separate name is not that useful to justify complexity.

type Info struct {
	Name Name
	Addr string
}
