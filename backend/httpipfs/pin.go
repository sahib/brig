package httpipfs

import (
	"context"
	"encoding/json"
	"io"
	"strings"

	"github.com/patrickmn/go-cache"
	h "github.com/sahib/brig/util/hashlib"
)

// IsPinned returns true when `hash` is pinned in some way.
func (nd *Node) IsPinned(hash h.Hash) (bool, error) {
	ctx := context.Background()
	resp, err := nd.sh.Request("pin/ls", hash.B58String()).Send(ctx)
	if err != nil {
		return false, err
	}

	defer resp.Close()

	if resp.Error != nil {
		if strings.HasSuffix(resp.Error.Message, "is not pinned") {
			return false, nil
		}

		return false, resp.Error
	}

	raw := struct {
		Keys map[string]struct {
			Type string
		}
	}{}

	if err := json.NewDecoder(resp.Output).Decode(&raw); err != nil {
		return false, err
	}

	if len(raw.Keys) == 0 {
		return false, nil
	}

	return true, nil
}

// Pin will pin `hash`.
func (nd *Node) Pin(hash h.Hash) error {
	return nd.sh.Pin(hash.B58String())
}

// Unpin will unpin `hash`.
func (nd *Node) Unpin(hash h.Hash) error {
	err := nd.sh.Unpin(hash.B58String())
	if err == nil || err.Error() == "pin/rm: not pinned or pinned indirectly" {
		return nil
	}
	return err
}

type objectRef struct {
	Ref string // hash of the ref
	Err string
}

// Link is a child of a hash.
// Used by IPFS when files get bigger.
type Link struct {
	Name string
	Hash string
	Size uint64
}

// IsCached checks if hash and all its children are cached
func (nd *Node) IsCached(hash h.Hash) (bool, error) {
	locallyCached := nd.cache.locallyCached
	stat, found := locallyCached.Get(hash.B58String())
	if found {
		return stat.(bool), nil
	}

	// Nothing in the cache, we have to figure it out.
	// We will execute equivalent of
	//   ipfs refs --offline --recursive hash
	// note the `--recursive` switch, we need to check all children links
	// if command fails at least one child link/hash is missing
	ctx := context.Background()
	req := nd.sh.Request("refs", hash.B58String())
	req.Option("offline", "true")
	req.Option("recursive", "true")
	resp, err := req.Send(ctx)
	if err != nil {
		return false, err
	}
	defer resp.Close()
	if resp.Error != nil {
		return false, resp.Error
	}

	ref := objectRef{}
	jsonDecoder := json.NewDecoder(resp.Output)
	for {
		if err := jsonDecoder.Decode(&ref); err == io.EOF {
			break
		} else if err != nil {
			return false, err
		}
		if ref.Err != "" {
			// Either main hash or one of its refs/links is not available locally
			// consequently the whole hash is not cached
			locallyCached.Set(hash.B58String(), false, cache.DefaultExpiration)
			return false, nil
		}
	}
	// if we are here, the parent hash and all its children links/hashes are cached
	locallyCached.Set(hash.B58String(), true, cache.DefaultExpiration)
	return true, nil
}

// CachedSize returns the cached size of the node.
// Negative indicates unknow eithe due to error or hash not stored locally
func (nd *Node) CachedSize(hash h.Hash) (int64, error) {
	ctx := context.Background()
	req := nd.sh.Request("object/stat", hash.B58String())
	// provides backend size only for cached objects
	req.Option("offline", "true")
	resp, err := req.Send(ctx)
	if err != nil {
		return -1, err
	}

	defer resp.Close()

	if resp.Error != nil {
		return -1, resp.Error
	}

	raw := struct {
		CumulativeSize int64
		Key            string
	}{}

	if err := json.NewDecoder(resp.Output).Decode(&raw); err != nil {
		return -1, err
	}

	return raw.CumulativeSize, nil
}
