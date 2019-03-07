package httpipfs

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"

	e "github.com/pkg/errors"
	h "github.com/sahib/brig/util/hashlib"
	log "github.com/sirupsen/logrus"
)

// GC will trigger the garbage collector of IPFS.
// Cleaned up hashes will be returned as a list
// (note that those hashes are not always ours)
func (nd *Node) GC() ([]h.Hash, error) {
	ctx := context.Background()
	resp, err := nd.sh.Request("repo/gc").Send(ctx)

	if err != nil {
		return nil, e.Wrapf(resp.Error, "gc request")
	}

	defer resp.Close()

	if resp.Error != nil {
		return nil, e.Wrapf(resp.Error, "gc resp")
	}

	hs := []h.Hash{}
	br := bufio.NewReader(resp.Output)
	for {
		line, err := br.ReadBytes('\n')
		if err != nil {
			break
		}

		raw := struct {
			Key map[string]string
		}{}

		lr := bytes.NewReader(line)
		if err := json.NewDecoder(lr).Decode(&raw); err != nil {
			return nil, e.Wrapf(err, "json decode")
		}

		for _, cid := range raw.Key {
			h, err := h.FromB58String(cid)
			if err != nil {
				return nil, e.Wrapf(err, "gc: hash decode")
			}

			hs = append(hs, h)
		}
	}

	log.Debugf("GC returned %d hashes", len(hs))
	return hs, nil
}
