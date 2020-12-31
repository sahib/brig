package httpipfs

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"

	ipfsutil "github.com/ipfs/go-ipfs-util"
	mh "github.com/multiformats/go-multihash"
	"github.com/sahib/brig/net/peer"
	h "github.com/sahib/brig/util/hashlib"
	shell "github.com/ipfs/go-ipfs-api"
	log "github.com/sirupsen/logrus"
)

// PublishName will announce `name` to the network
// and make us discoverable.
func (nd *Node) PublishName(name string) error {
	if !nd.isOnline() {
		return ErrOffline
	}

	fullName := "brig:" + string(name)
	key, err := nd.sh.BlockPut([]byte(fullName), "v0", "sha2-256", -1)
	log.Debugf("published name: »%s« (key %s)", name, key)
	return err
}

// Identity returns our own identity.
// It will cache the identity after the first request.
func (nd *Node) Identity() (peer.Info, error) {
	nd.mu.Lock()
	if nd.cachedIdentity != "" {
		defer nd.mu.Unlock()
		return peer.Info{
			Name: "httpipfs",
			Addr: nd.cachedIdentity,
		}, nil
	}

	// Do not hold the lock during net ops:
	nd.mu.Unlock()

	id, err := nd.sh.ID()
	if err != nil {
		return peer.Info{}, err
	}

	nd.mu.Lock()
	nd.cachedIdentity = id.ID
	nd.mu.Unlock()

	return peer.Info{
		Name: "httpipfs",
		Addr: id.ID,
	}, nil
}

func findProvider(ctx context.Context, sh *shell.Shell, hash h.Hash) ([]string, error) {
	resp, err := sh.Request("dht/findprovs", hash.B58String()).Send(ctx)
	if err != nil {
		return nil, err
	}

	defer resp.Output.Close()

	if resp.Error != nil {
		return nil, resp.Error
	}

	ids := make(map[string]bool)
	br := bufio.NewReader(resp.Output)
	interrupted := false

	for len(ids) < 20 && !interrupted {
		line, err := br.ReadBytes('\n')
		if err != nil {
			break
		}

		raw := struct {
			Responses []struct {
				ID string
			}
		}{}

		lr := bytes.NewReader(line)
		if err := json.NewDecoder(lr).Decode(&raw); err != nil {
			return nil, err
		}

		for _, resp := range raw.Responses {
			ids[resp.ID] = true
		}

		select {
		case <-ctx.Done():
			interrupted = true
			break
		}
	}

	linearIDs := []string{}
	for id := range ids {
		linearIDs = append(linearIDs, id)
	}

	return linearIDs, nil
}

// ResolveName will return all peers that identify themselves as `name`.
// If ctx is canceled it will return early, but return no error.
func (nd *Node) ResolveName(ctx context.Context, name string) ([]peer.Info, error) {
	if !nd.isOnline() {
		return nil, ErrOffline
	}

	name = "brig:" + name
	mhash, err := mh.Sum([]byte(name), ipfsutil.DefaultIpfsHash, -1)
	if err != nil {
		return nil, err
	}

	log.Debugf("backend: resolve »%s« (%s)", name, mhash.B58String())

	ids, err := findProvider(ctx, nd.sh, h.Hash(mhash))
	if err != nil {
		return nil, err
	}

	infos := []peer.Info{}
	for _, id := range ids {
		infos = append(infos, peer.Info{
			Addr: id,
			Name: peer.Name(name),
		})
	}

	return infos, nil
}
