// Package id implements the parsing of brig-ids.
//
// user[@domain[/resource]
package id

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/disorganizer/brig/util/ipfsutil"

	mh "github.com/jbenet/go-multihash"
	"golang.org/x/text/unicode/norm"
)

type ID string

type ErrBadID struct {
	reason string
}

func (e ErrBadID) Error() string {
	return e.reason
}

func valid(id string) error {
	if !utf8.ValidString(id) {
		return ErrBadID{fmt.Sprintf("Invalid utf-8: %v", id)}
	}

	for idx, rn := range id {
		if unicode.IsSpace(rn) {
			return ErrBadID{
				fmt.Sprintf("Space not allowed: %s (at %d)", id, idx),
			}
		}

		if !unicode.IsPrint(rn) {
			return ErrBadID{
				fmt.Sprintf("Only printable runes allowed: %s (at %d)", id, idx),
			}
		}
	}

	return nil
}

// Cast checks `id` to be correct and returns
// a wrapped ID.
func Cast(id string) (ID, error) {
	if err := valid(id); err != nil {
		return "", err
	}

	return ID(norm.NFKC.Bytes([]byte(id))), nil
}

func IsValid(id string) bool {
	return valid(id) == nil
}

func (id ID) Domain() string {
	a := strings.IndexRune(string(id), '@')
	if a < 0 {
		return ""
	}

	b := strings.LastIndexByte(string(id), '/')
	if b < 0 {
		return string(id)[a+1:]
	}

	return string(id)[a+1 : b]
}

func (id ID) Resource() string {
	idx := strings.LastIndexByte(string(id), '/')
	if idx < 0 {
		return ""
	}

	return string(id)[idx+1:]
}

func (id ID) User() string {
	idx := strings.Index(string(id), "@")
	if idx < 0 {
		return string(id)
	}

	return string(id)[:idx]
}

func (id ID) asBlockData() []byte {
	return []byte("brig:" + string(id))
}

var (
	ErrAlreadyRegistered = errors.New("Username already registered")
)

func (id ID) Register(node *ipfsutil.Node) error {
	blockData := id.asBlockData()
	hash, err := mh.Sum(blockData, mh.SHA2_256, -1)
	if err != nil {
		return err
	}

	peers, err := ipfsutil.Locate(node, hash, 1, 5*time.Second)
	if err != nil && err != ipfsutil.ErrTimeout {
		return err
	}

	// Check if some id is our own:
	if len(peers) > 0 {
		self, err := node.Identity()
		if err != nil {
			return err
		}

		wasSelf := false
		for _, peer := range peers {
			if peer.ID == self {
				wasSelf = true
			}
		}

		if wasSelf {
			return ErrAlreadyRegistered
		}
	}

	// If it was an timeout, it's probably not yet registered.
	otherHash, err := ipfsutil.AddBlock(node, blockData)
	fmt.Println(otherHash, hash) // TODO: should be same

	if err != nil {
		return err
	}

	return nil
}

func (id ID) Lookup(node *ipfsutil.Node) error {
	return nil
}
