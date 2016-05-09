// Package id implements the parsing of brig-ids.
//
// user[@domain[/resource]
package id

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/disorganizer/brig/util"
	"github.com/disorganizer/brig/util/ipfsutil"

	log "github.com/Sirupsen/logrus"
	mh "github.com/jbenet/go-multihash"
	"golang.org/x/text/unicode/norm"
)

var (
	ErrNoAddrs = errors.New("No addrs found for id (online?)")
)

type ID string

type ErrBadID struct {
	reason string
}

func (e ErrBadID) Error() string {
	return e.reason
}

func valid(id string) error {
	if utf8.RuneCountInString(id) == 0 {
		return ErrBadID{"Empty ID is not allowed"}
	}

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

func (id ID) Hash() mh.Multihash {
	// TODO: Use go-ipfs-util.DefaultIpfsHash
	//		 https://github.com/ipfs/go-ipfs-util/pull/1
	hash, err := mh.Sum(id.asBlockData(), mh.SHA2_256, -1)

	// Mulithash should only fail if an invalid len or code was passed.
	if err != nil {
		panic(err)
	}

	return hash
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

func (id ID) AsPath() string {
	path := id.User()
	rsrc := id.Resource()
	if rsrc != "" {
		path += "-" + rsrc
	}

	return strings.Replace(path, string(os.PathSeparator), "|", -1)
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

// TODO: bad name? Does not really register; just publishes the block(s)
func (id ID) Register(node *ipfsutil.Node) error {
	hash := id.Hash()

	peers, err := ipfsutil.Locate(node, hash, 1, 5*time.Second)
	if err != nil && err != util.ErrTimeout {
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
	otherHash, err := ipfsutil.AddBlock(node, id.asBlockData())
	if !bytes.Equal(otherHash, hash) {
		log.Warningf("Hash differ during register; did the hash func changed?")
	}

	if err != nil {
		return err
	}

	return nil
}

// TODO: Not sure if the next functions are actually useful...
// DelBlock just deletes the block *locally*
// Lookup returns the brig:$id value

func (id ID) Unregister(node *ipfsutil.Node) error {
	hash := id.Hash()

	if err := ipfsutil.DelBlock(node, hash); err != nil {
		return err
	}

	return nil
}

func (id ID) Taken(node *ipfsutil.Node) (bool, error) {
	data, err := ipfsutil.CatBlock(node, id.Hash(), 5*time.Second)
	if err != nil {
		return false, err
	}

	// This is kinda paranoid...
	// (Disclaimer: I doubt hash collisions, but bugs are everywhere)
	return bytes.Equal(data, id.asBlockData()), nil
}
