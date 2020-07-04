package httpipfs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

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

func (nd *Node) IsCached(hash h.Hash) (bool, error) {
	// This feature is only supported for ipfs >= 0.4.19.
	// Check this and issue a warning if that's not the case.
	if nd.version.LT(semver.MustParse("0.4.19")) {
		return false, fmt.Errorf("cache queries are not supported in ipfs < 0.4.19")
	}

	ctx := context.Background()
	req := nd.sh.Request("block/stat", hash.B58String())
	req.Option("offline", "true")
	resp, err := req.Send(ctx)
	if err != nil {
		return false, err
	}

	defer resp.Close()

	if resp.Error != nil {
		return false, nil
	}

	io.Copy(ioutil.Discard, resp.Output)
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
