package server

import (
	"fmt"

	e "github.com/pkg/errors"
	"github.com/sahib/brig/catfs"
	"github.com/sahib/brig/gateway/remotesapi"
	"github.com/sahib/brig/net/peer"
	"github.com/sahib/brig/repo"
)

// RemotesAPI is an adapter of base for the gateway.
type RemotesAPI struct {
	base *base
}

// NewRemotesAPI returns a new RemotesAPI.
func NewRemotesAPI(base *base) *RemotesAPI {
	return &RemotesAPI{
		base: base,
	}
}

// List all existing remotes.
func (a *RemotesAPI) List() ([]*remotesapi.Remote, error) {
	// TODO: Do this in parallel.
	rmts, err := a.base.repo.Remotes.ListRemotes()
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

// Get a remote by its `name`.
func (a *RemotesAPI) Get(name string) (*remotesapi.Remote, error) {
	return a.get(name)
}

func (a *RemotesAPI) get(name string) (*remotesapi.Remote, error) {
	rmt, err := a.base.repo.Remotes.Remote(name)
	if err != nil {
		return nil, err
	}

	extRmt := &remotesapi.Remote{}
	extRmt.Name = rmt.Name
	extRmt.Fingerprint = string(rmt.Fingerprint)
	extRmt.AcceptAutoUpdates = rmt.AcceptAutoUpdates
	extRmt.AcceptPush = rmt.AcceptPush
	extRmt.ConflictStrategy = rmt.ConflictStrategy

	for _, folder := range rmt.Folders {
		extRmt.Folders = append(extRmt.Folders, remotesapi.Folder{
			Folder:   folder.Folder,
			ReadOnly: folder.ReadOnly,
		})
	}

	addr := rmt.Fingerprint.Addr()
	psrv := a.base.peerServer
	pinger, err := psrv.PingMap().For(addr)
	if err != nil {
		// early exit: peer is not online.
		return extRmt, nil
	}

	extRmt.IsOnline = pinger.Roundtrip() > 0
	extRmt.LastSeen = pinger.LastSeen()
	extRmt.IsAuthenticated = psrv.PingMap().IsAuthenticated(addr)
	return extRmt, nil
}

// Set (i.e. add or modify) a remote.
// IsAuthenticated, IsOnline and LastSeen will be ignored.
func (a *RemotesAPI) Set(rm remotesapi.Remote) error {
	fp, err := peer.CastFingerprint(rm.Fingerprint)
	if err != nil {
		return err
	}

	folders := []repo.Folder{}
	for _, folder := range rm.Folders {
		folders = append(folders, repo.Folder{
			Folder:   folder.Folder,
			ReadOnly: folder.ReadOnly,
		})
	}

	err = a.base.repo.Remotes.AddOrUpdateRemote(repo.Remote{
		Name:              rm.Name,
		Fingerprint:       fp,
		Folders:           folders,
		AcceptAutoUpdates: rm.AcceptAutoUpdates,
		AcceptPush:        rm.AcceptPush,
		ConflictStrategy:  rm.ConflictStrategy,
	})

	if err != nil {
		return err
	}

	return a.base.syncRemoteStates()
}

// Remove removes a remote by `name`.
func (a *RemotesAPI) Remove(name string) error {
	if err := a.base.repo.Remotes.RmRemote(name); err != nil {
		return err
	}

	return a.base.syncRemoteStates()
}

// Self returns the identity of this repository.
func (a *RemotesAPI) Self() (remotesapi.Identity, error) {
	kr, err := a.base.repo.Keyring()
	if err != nil {
		return remotesapi.Identity{}, err
	}

	ownPubKey, err := kr.OwnPubKey()
	if err != nil {
		return remotesapi.Identity{}, err
	}

	identity, err := a.base.peerServer.Identity()
	if err != nil {
		return remotesapi.Identity{}, err
	}

	owner := a.base.repo.Immutables.Owner()
	fp := peer.BuildFingerprint(identity.Addr, ownPubKey)
	return remotesapi.Identity{
		Name:        owner,
		Fingerprint: string(fp),
	}, nil
}

// Sync synchronizes the latest state of `name` with our latest state.
func (a *RemotesAPI) Sync(name string) error {
	msg := fmt.Sprintf("sync with »%s« from gateway", name)
	_, err := a.base.doSync(name, true, msg)
	return err
}

// MakeDiff produces a diff to the remote with `name`.
func (a *RemotesAPI) MakeDiff(name string) (*catfs.Diff, error) {
	if err := a.base.doFetch(name); err != nil {
		return nil, e.Wrapf(err, "fetch-remote")
	}

	var diff *catfs.Diff
	return diff, a.base.withCurrFs(func(localFs *catfs.FS) error {
		return a.base.withRemoteFs(name, func(remoteFs *catfs.FS) error {
			newDiff, err := localFs.MakeDiff(remoteFs, "CURR", "CURR")
			if err != nil {
				return err
			}

			diff = newDiff
			return nil
		})
	})
}

// OnChange register a callback to be called once the remote list changes.
func (a *RemotesAPI) OnChange(fn func()) {
	a.base.repo.Remotes.OnChange(fn)
}
