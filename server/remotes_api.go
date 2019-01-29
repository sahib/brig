package server

import (
	"fmt"

	e "github.com/pkg/errors"
	"github.com/sahib/brig/catfs"
	"github.com/sahib/brig/gateway/remotesapi"
	p2pnet "github.com/sahib/brig/net"
	"github.com/sahib/brig/net/peer"
	"github.com/sahib/brig/repo"
)

// Mock is for testing purposes whenever a normal RemotesAPI is needed.
// It stores remotes in memory and does not implement realy syncing or diffing.
type RemotesAPI struct {
	base *base
}

// NewMock creates a new Mock.
// `name` and `fingerprint` will be returned
func NewRemotesAPI(base *base) *RemotesAPI {
	return &RemotesAPI{
		base: base,
	}
}

func (a *RemotesAPI) List() ([]*remotesapi.Remote, error) {
	// TODO: Do this in parallel.
	rp, err := a.base.Repo()
	if err != nil {
		return nil, err
	}

	rmts, err := rp.Remotes.ListRemotes()
	if err != nil {
		return nil, err
	}

	extRmts := []*remotesapi.Remote{}
	for _, rmt := range rmts {
		extRmt, err := a.get(rmt.Name)
		if err != nil {
			return nil, err
		}

		extRmts = append(extRmts, extRmt)
	}

	return extRmts, nil
}

func (a *RemotesAPI) Get(name string) (*remotesapi.Remote, error) {
	return a.get(name)
}

func (a *RemotesAPI) get(name string) (*remotesapi.Remote, error) {
	rp, err := a.base.Repo()
	if err != nil {
		return nil, err
	}

	rmt, err := rp.Remotes.Remote(name)
	if err != nil {
		return nil, err
	}

	psrv, err := a.base.PeerServer()
	if err != nil {
		return nil, err
	}

	// TODO: Use some basic caching.
	// TODO: Use this also for server/net_handlers.go
	extRmt := &remotesapi.Remote{}
	extRmt.Name = rmt.Name
	extRmt.Fingerprint = string(rmt.Fingerprint)
	extRmt.AcceptAutoUpdates = rmt.AcceptAutoUpdates
	for _, folder := range rmt.Folders {
		extRmt.Folders = append(extRmt.Folders, folder.Folder)
	}

	pinger, err := psrv.PingMap().For(rmt.Fingerprint.Addr())
	if err != nil {
		// early exit: peer is not online.
		return extRmt, nil
	}

	extRmt.IsOnline = pinger.Roundtrip() > 0
	extRmt.LastSeen = pinger.LastSeen()

	// Try to ping the client:
	a.base.withNetClient(name, func(ctl *p2pnet.Client) error {
		if err := ctl.Ping(); err != nil {
			return err
		}

		extRmt.IsAuthenticated = true
		return nil
	})

	return extRmt, nil
}

func (a *RemotesAPI) Set(rm remotesapi.Remote) error {
	rp, err := a.base.Repo()
	if err != nil {
		return err
	}

	fp, err := peer.CastFingerprint(rm.Fingerprint)
	if err != nil {
		return err
	}

	folders := []repo.Folder{}
	for _, path := range rm.Folders {
		folders = append(folders, repo.Folder{
			Folder: path,
		})
	}

	err = rp.Remotes.AddOrUpdateRemote(repo.Remote{
		Name:        rm.Name,
		Fingerprint: fp,
		Folders:     folders,
	})

	if err != nil {
		return err
	}

	return a.base.syncRemoteStates()
}

func (a *RemotesAPI) Remove(name string) error {
	rp, err := a.base.Repo()
	if err != nil {
		return err
	}

	if err := rp.Remotes.RmRemote(name); err != nil {
		return err
	}

	return a.base.syncRemoteStates()
}

func (a *RemotesAPI) Self() (remotesapi.Identity, error) {
	rp, err := a.base.Repo()
	if err != nil {
		return remotesapi.Identity{}, err
	}

	ownPubKey, err := rp.Keyring().OwnPubKey()
	if err != nil {
		return remotesapi.Identity{}, err
	}

	psrv, err := a.base.PeerServer()
	if err != nil {
		return remotesapi.Identity{}, err
	}

	identity, err := psrv.Identity()
	if err != nil {
		return remotesapi.Identity{}, err
	}

	fp := peer.BuildFingerprint(identity.Addr, ownPubKey)
	return remotesapi.Identity{
		Name:        string(identity.Name),
		Fingerprint: string(fp),
	}, nil
}

func (a *RemotesAPI) Sync(name string) error {
	msg := fmt.Sprintf("sync with »%s« from gateway", name)
	_, err := a.base.doSync(name, true, msg)
	return err
}

func (a *RemotesAPI) MakeDiff(name string) (*catfs.Diff, error) {
	if err := a.base.doFetch(name); err != nil {
		return nil, e.Wrapf(err, "fetch-remote")
	}

	var diff *catfs.Diff
	return diff, a.base.withCurrFs(func(localFs *catfs.FS) error {
		return a.base.withRemoteFs(name, func(remoteFs *catfs.FS) error {
			newDiff, err := localFs.MakeDiff(remoteFs, "HEAD", "HEAD")
			if err != nil {
				return err
			}

			diff = newDiff
			return nil
		})
	})
}
