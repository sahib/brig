package httpipfs

import (
	"context"
	"encoding/json"

	h "github.com/sahib/brig/util/hashlib"
)

func (nd *Node) GC() ([]h.Hash, error) {
	ctx := context.Background()
	resp, err := nd.sh.Request("repo/gc").Send(ctx)

	if err != nil {
		return nil, err
	}

	defer resp.Close()

	if resp.Error != nil {
		return nil, resp.Error
	}

	raw := struct {
		Key map[string]string
	}{}

	if err := json.NewDecoder(resp.Output).Decode(&raw); err != nil {
		return nil, err
	}

	hs := []h.Hash{}
	for _, cid := range raw.Key {
		h, err := h.FromB58String(cid)
		if err != nil {
			return nil, err
		}

		hs = append(hs, h)
	}

	return hs, nil
}
