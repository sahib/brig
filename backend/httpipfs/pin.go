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
	"github.com/patrickmn/go-cache"
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

// Gets all locally available ipfs refs (hashes)
func (nd *Node) FillLocalRefs(m map[string]bool) (error) {
	ctx := context.Background()

	if m == nil {
		return errors.New("Need non nil reference to fill the map")
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
		m[ref.Ref] = true // reference with given reference hash is checked
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

	locCache := nd.cache.refsLinks
	links, found := locCache.Get(hash.B58String())
	if found && links != nil {
		return links.([]link), nil
	}

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
		return nil, err // nil indicates unknown status
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

	locCache.Set(hash.B58String(), linksResp.Links, cache.DefaultExpiration)
	return linksResp.Links, nil // returns empty array
}

// Checks if hash is cached. Does not check status of hash children
func (nd *Node) isThisHashOnlyCached(hash h.Hash) (bool, error) {
	locCache := nd.cache.localRefs
	localRefsMap, found := locCache.Get("all")
	if !found {
		// we need to get all the local refs
		var m = map[string]bool{}
		if err := nd.FillLocalRefs(m); err != nil {
			return false, err
		}
		localRefsMap = m
		locCache.Set("all", m, cache.DefaultExpiration)
	}
	var m = localRefsMap.(map[string]bool)
	if _, found := m[hash.B58String()]; !found {
		// this hash is not locally available and thus not cached
		return false, nil
	}
	return true, nil
	
}

// Checks if hash and all its children are cached
func (nd *Node) IsCached(hash h.Hash) (bool, error) {
	locallyCached := nd.cache.locallyCached;
	stat, found := locallyCached.Get(hash.B58String())
	if found {
		return stat.(bool), nil
	}
	// Nothing in the cache, we have to figure it out

	// Now, we are ready to check if the hash under the question is cached
	yes, err := nd.isThisHashOnlyCached(hash)
	if !yes {
		// no need to check for children if the parent is not cached
		if err == nil {
			locallyCached.Set(hash.B58String(), false, cache.DefaultExpiration)
		}
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
		var isChildCached bool = false
		if l.Size <= 262158 { // 256kB + 14B
			// Heuristic: at least up to the IPFS version v0.6.0
			// the child with size 262158 bytes is not going to have children.
			// Than we do not need to run the full recursive check
			isChildCached, err = nd.isThisHashOnlyCached(childHash)
		} else {
			// Size is large, we need to run the recursive check
			isChildCached, err = nd.IsCached(childHash)
		}
		if err != nil {
			return false, err
		}
		if !isChildCached {
			// If even one child/link is uncached, we call everything uncached
			// TODO: we can report how much of content is pre-cached
			locallyCached.Set(hash.B58String(), false, cache.DefaultExpiration)
			return false, nil
		}
	}
	// if we are here, the parent hash and all its children links/hashes are cached
	locallyCached.Set(hash.B58String(), true, cache.DefaultExpiration)
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
