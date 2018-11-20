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

	h "github.com/sahib/brig/util/hashlib"
	"golang.org/x/text/unicode/norm"
)

// Name is the display name of a peer.
// (i.e. how another repo calls itself)
type Name string

// ErrBadName is returned for invalidly formatted peer names.
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

// CastName checks `name` to be correct and returns a wrapped name.
func CastName(name string) (Name, error) {
	if err := valid(name); err != nil {
		return "", err
	}

	return Name(norm.NFKC.Bytes([]byte(name))), nil
}

// IsValid will return true if a peer name is formatted validly.
func IsValid(name string) bool {
	return valid(name) == nil
}

// Domain will return the domain part of a peer name, if present.
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

// Resource will return the resource part of a peer name, if present.
func (name Name) Resource() string {
	idx := strings.LastIndexByte(string(name), '/')
	if idx < 0 {
		return ""
	}

	return string(name)[idx+1:]
}

// WithoutResource returns the same peer name without its resource part.
func (name Name) WithoutResource() string {
	domain := name.Domain()
	if len(domain) > 0 {
		return name.User() + "@" + name.Domain()
	}

	return name.User()
}

// AsPath converts a peer name to a path that can be used for storage.
func (name Name) AsPath() string {
	path := name.User()
	rsrc := name.Resource()
	if rsrc != "" {
		path += "-" + rsrc
	}

	return strings.Replace(path, string(os.PathSeparator), "|", -1)
}

// User returns the user part of the peer name.
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

// CastFingerprint converts and checks `s` to be a valid Fingerprint.
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

// BuildFingerprint builds a fingerprint from `addr` and a public key.
func BuildFingerprint(addr string, pubKeyData []byte) Fingerprint {
	s := fmt.Sprintf("%s:%s", addr, h.Sum(pubKeyData).B58String())
	return Fingerprint(s)
}

// Addr returns the addr part of a fingerprint.
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

// PubKeyID returns the public key hash in the fingerprint.
func (fp Fingerprint) PubKeyID() string {
	parts := strings.SplitN(string(fp), ":", 2)
	if len(parts) < 2 {
		return ""
	}

	return parts[1]
}

// PubKeyMatches checks if the supplied public key matches with the
// hashed version in the fingerprint.
func (fp Fingerprint) PubKeyMatches(pubKeyData []byte) bool {
	own := fp.PubKeyID()
	if own == "" {
		return false
	}

	remote := h.Sum(pubKeyData).B58String()
	return own == remote
}

///////////////////////

// Info is a pair of addr and a peer name.
type Info struct {
	Name Name
	Addr string
}
