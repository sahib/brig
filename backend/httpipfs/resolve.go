package httpipfs

import (
	"context"
	"encoding/json"

	shell "github.com/ipfs/go-ipfs-api"
	ipfsutil "github.com/ipfs/go-ipfs-util"
	mh "github.com/multiformats/go-multihash"
	"github.com/sahib/brig/net/peer"
	h "github.com/sahib/brig/util/hashlib"
)

func (nd *Node) PublishName(name string) error {
	fullName := "brig:" + string(name)
	_, err := nd.sh.BlockPut([]byte(fullName), "v0", "sha2-256", -1)
	return err
}

func (nd *Node) Identity() (peer.Info, error) {
	id, err := nd.sh.ID()
	if err != nil {
		return peer.Info{}, err
	}

	return peer.Info{
		Name: "ipfs",
		Addr: id.ID,
	}, nil
}

func findProvider(sh *shell.Shell, hash h.Hash) ([]string, error) {
	ctx := context.Background()
	resp, err := sh.Request("dht/findprovs", hash.B58String()).Send(ctx)

	if err != nil {
		return nil, err
	}

	defer resp.Close()

	if resp.Error != nil {
		return nil, resp.Error
	}

	raw := struct {
		Responses []struct {
			ID string
		}
	}{}

	if err := json.NewDecoder(resp.Output).Decode(&raw); err != nil {
		return nil, err
	}

	ids := []string{}
	for _, entry := range raw.Responses {
		ids = append(ids, entry.ID)
	}

	return ids, nil
}

func (nd *Node) ResolveName(ctx context.Context, name string) ([]peer.Info, error) {
	name = "brig:" + name
	mhash, err := mh.Sum([]byte(name), ipfsutil.DefaultIpfsHash, -1)
	if err != nil {
		return nil, err
	}

	// TODO: Use ctx somehow.
	ids, err := findProvider(nd.sh, h.Hash(mhash))
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
