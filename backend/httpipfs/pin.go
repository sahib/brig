package httpipfs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"errors"

	"github.com/blang/semver"
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

type ipfsStateCache struct {
	LocalRefs map[string]bool
}

func NewIpfsStateCache() *ipfsStateCache {
	cache := ipfsStateCache{}
	cache.LocalRefs = map[string]bool{}
	return &cache
}

// Gets all locally available ipfs refs (hashes)
func (nd *Node) FillLocalRefs(cache *ipfsStateCache) (error) {
	ctx := context.Background()

	if cache == nil {
		return errors.New("Need non nil reference to fill the cache")
	}
	req := nd.sh.Request("refs/local")
	resp, err := req.Send(ctx)
	if err != nil {
		return err
	}
	defer resp.Close()
	if resp.Error != nil {
		return resp.Error
	}

	ref := objectRef{}
	jsonDecoder := json.NewDecoder(resp.Output)

	for {
		if err := jsonDecoder.Decode(&ref); err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		cache.LocalRefs[ref.Ref] = true // reference with given reference hash is checked
	}

	return nil
}

type link  struct {
	Name string
	Hash string
	Size uint64
}

// Get children (links in the ipfs lingo) of the hash
func (nd *Node) GetLinks(hash h.Hash) ([]link, error) {
	// Empty return array is not the same as nil!
	// Nil means we were not able to check for links
	// i.e. parent hash is not available or there were an error.
	// Empty means that everything worked but there are no children/links

	// The "Option: offline" feature is only supported for ipfs >= 0.4.19.
	// Check this and issue a warning if that's not the case.
	if nd.version.LT(semver.MustParse("0.4.19")) {
		return nil, fmt.Errorf("offline cache queries are not supported in ipfs < 0.4.19")
	}

	ctx := context.Background()
	req := nd.sh.Request("object/links", hash.B58String())
	req.Option("offline", "true")
	resp, err := req.Send(ctx)
	if err != nil {
		return nil, err // nil indicates unknown statur
	}
	defer resp.Close()
	if resp.Error != nil {
		return nil, resp.Error
	}

	type objectLinksResp struct {
		Hash string
		Links []link
	}
	linksResp := objectLinksResp{}
	if err := json.NewDecoder(resp.Output).Decode(&linksResp); err != nil {
		return nil, err
	}
	if linksResp.Links == nil {
		// If we are here everything work, and no children case is communicated
		// with an empty return array.
		linksResp.Links = []link{}
	}

	return linksResp.Links, nil // returns empty array
}

// Checks if hash is cached. Does not check status of hash children
func (nd *Node) isCached(hash h.Hash, cache *ipfsStateCache) (bool, error) {
	if cache == nil {
		return nd.IsCached(hash)
	}
	if _, ok := cache.LocalRefs[hash.B58String()]; !ok {
		// this hash is not locally available and thus not cached
		return false, nil
	}
	return true, nil
	
}

// Checks if hash and all its children are cached
func (nd *Node) IsCached(hash h.Hash) (bool, error) {
	// Let's get all locally available ipfs refs or hashes
	cache := NewIpfsStateCache()
	err := nd.FillLocalRefs(cache) // backend status cache
	if err != nil {
		return false, err
	}

	// Now, we are ready to check if the hash under the question is cached
	yes, err := nd.isCached(hash, cache)
	if !yes {
		return false, err
	}

	// By now we know that parent object/block is cached by what about linked ones?
	// Lets get the list of linked (children) objects
	links, err := nd.GetLinks(hash)
	if err != nil {
		return false, err
	}
	for _, l := range(links) {
		childHash, err := h.FromB58String(l.Hash)
		if err != nil {
			return false, err
		}
		// WARNING: isCached does not check if hash has children!
		isChildCached, err := nd.isCached(childHash, cache)
		if err != nil {
			return false, err
		}
		if !isChildCached {
			// If even one child/link is uncached, we call everything uncached
			// TODO: we can report how much of content is pre-cached
			return false, nil
		}
	}
	// if we are here, the parent hash and all its children links/hashes are cached
	return true, nil
}

func (nd *Node) CachedSize(hash h.Hash) (uint64, error) {
	// MaxUint64 indicates that cachedSize is unknown
	MaxUint64 := uint64(1<<64 - 1)
	ctx := context.Background()
	req := nd.sh.Request("object/stat", hash.B58String())
	// provides backend size only for cached objects
	req.Option("offline", "true")
	resp, err := req.Send(ctx)
	if err != nil {
		return MaxUint64, err
	}

	defer resp.Close()

	if resp.Error != nil {
		return MaxUint64, resp.Error
	}

	raw := struct {
		CumulativeSize uint64
		Key string
	}{}

	if err := json.NewDecoder(resp.Output).Decode(&raw); err != nil {
		return MaxUint64, err
	}

	return raw.CumulativeSize, nil
}
