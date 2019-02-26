package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	p2pnet "github.com/sahib/brig/net"
	"github.com/sahib/brig/net/peer"
	"github.com/sahib/brig/repo"
	"github.com/sahib/brig/server/capnp"
	"github.com/sahib/brig/util/conductor"
	log "github.com/sirupsen/logrus"
	capnplib "zombiezen.com/go/capnproto2"
	"zombiezen.com/go/capnproto2/server"
)

type netHandler struct {
	base *base
}

func (nh *netHandler) Whoami(call capnp.Net_whoami) error {
	server.Ack(call.Options)

	capID, err := capnp.NewIdentity(call.Results.Segment())
	if err != nil {
		return err
	}

	psrv := nh.base.peerServer
	self, err := psrv.Identity()
	if err != nil {
		return err
	}

	// Compute our own fingerprint:
	rp := nh.base.repo
	ownPubKey, err := rp.Keyring().OwnPubKey()
	if err != nil {
		return err
	}

	finger := peer.BuildFingerprint(self.Addr, ownPubKey)

	if err := capID.SetOwner(rp.Owner); err != nil {
		return err
	}

	if err := capID.SetFingerprint(string(finger)); err != nil {
		return err
	}

	if err := capID.SetCurrentUser(rp.CurrentUser()); err != nil {
		return err
	}

	capID.SetIsOnline(psrv.IsOnline())
	return call.Results.SetWhoami(capID)
}

func (nh *netHandler) Connect(call capnp.Net_connect) error {
	server.Ack(call.Options)
	log.Infof("backend is going online...")
	return nh.base.peerServer.Connect()
}

func (nh *netHandler) Disconnect(call capnp.Net_disconnect) error {
	server.Ack(call.Options)
	log.Infof("backend is going offline...")
	return nh.base.peerServer.Disconnect()
}

func (nh *netHandler) RemoteOnlineList(call capnp.Net_remoteOnlineList) error {
	server.Ack(call.Options)

	rp := nh.base.repo
	psrv := nh.base.peerServer

	remotes, err := rp.Remotes.ListRemotes()
	if err != nil {
		return err
	}

	seg := call.Results.Segment()
	statuses, err := capnp.NewRemoteStatus_List(seg, int32(len(remotes)))
	if err != nil {
		return err
	}

	for idx, remote := range remotes {
		status, err := capnp.NewRemoteStatus(call.Results.Segment())
		if err != nil {
			return err
		}

		capRemote, err := remoteToCapRemote(remote, seg)
		if err != nil {
			return err
		}

		if err := status.SetRemote(*capRemote); err != nil {
			return err
		}

		pinger, err := psrv.PingMap().For(remote.Fingerprint.Addr())
		if err != nil {
			status.SetError(err.Error())
		}

		authenticated := false
		if pinger != nil {
			roundtrip := int32(pinger.Roundtrip() / time.Millisecond)
			status.SetRoundtripMs(roundtrip)

			lastSeen := pinger.LastSeen().Format(time.RFC3339)
			if err := status.SetLastSeen(lastSeen); err != nil {
				return err
			}

			err = nh.base.withNetClient(remote.Name, func(ctl *p2pnet.Client) error {
				authenticated = ctl.Ping() == nil
				return nil
			})

			status.SetAuthenticated(authenticated)

			if err != nil {
				return err
			}
		} else {
			errMsg := fmt.Sprintf("no route")
			if err := status.SetError(errMsg); err != nil {
				return err
			}
		}

		if err := statuses.Set(idx, status); err != nil {
			return err
		}
	}

	return call.Results.SetInfos(statuses)
}

func (nh *netHandler) RemotePing(call capnp.Net_remotePing) error {
	who, err := call.Params.Who()
	if err != nil {
		return err
	}

	return nh.base.withNetClient(who, func(ctl *p2pnet.Client) error {
		start := time.Now()
		if err := ctl.Ping(); err != nil {
			return err
		}

		roundtrip := time.Since(start).Seconds()
		call.Results.SetRoundtrip(roundtrip)
		return nil
	})
}

func capRemoteToRemote(remote capnp.Remote) (*repo.Remote, error) {
	remoteName, err := remote.Name()
	if err != nil {
		return nil, err
	}

	capFingerprint, err := remote.Fingerprint()
	if err != nil {
		return nil, err
	}

	// Check the fingerprint to be valid:
	fingerprint, err := peer.CastFingerprint(capFingerprint)
	if err != nil {
		return nil, err
	}

	remoteFolders, err := remote.Folders()
	if err != nil {
		return nil, err
	}

	folders := []repo.Folder{}
	for idx := 0; idx < remoteFolders.Len(); idx++ {
		folder := remoteFolders.At(idx)
		folderName, err := folder.Folder()
		if err != nil {
			return nil, err
		}

		folders = append(folders, repo.Folder{
			Folder: folderName,
		})
	}

	return &repo.Remote{
		Name:              remoteName,
		Fingerprint:       peer.Fingerprint(fingerprint),
		Folders:           folders,
		AcceptAutoUpdates: remote.AcceptAutoUpdates(),
	}, nil
}

func remoteToCapRemote(remote repo.Remote, seg *capnplib.Segment) (*capnp.Remote, error) {
	capRemote, err := capnp.NewRemote(seg)
	if err != nil {
		return nil, err
	}

	if err := capRemote.SetName(remote.Name); err != nil {
		return nil, err
	}

	if err := capRemote.SetFingerprint(string(remote.Fingerprint)); err != nil {
		return nil, err
	}

	capFolders, err := capnp.NewRemoteFolder_List(seg, int32(len(remote.Folders)))
	if err != nil {
		return nil, err
	}

	for idx, folder := range remote.Folders {
		capFolder, err := capnp.NewRemoteFolder(seg)
		if err != nil {
			return nil, err
		}

		if err := capFolder.SetFolder(folder.Folder); err != nil {
			return nil, err
		}

		if err := capFolders.Set(idx, capFolder); err != nil {
			return nil, err
		}
	}

	if err := capRemote.SetFolders(capFolders); err != nil {
		return nil, err
	}

	capRemote.SetAcceptAutoUpdates(remote.AcceptAutoUpdates)
	return &capRemote, nil
}

func (nh *netHandler) RemoteByName(call capnp.Net_remoteByName) error {
	server.Ack(call.Options)

	name, err := call.Params.Name()
	if err != nil {
		return err
	}

	rp := nh.base.repo
	rmt, err := rp.Remotes.Remote(name)
	if err != nil {
		return err
	}

	capRemote, err := remoteToCapRemote(rmt, call.Results.Segment())
	if err != nil {
		return err
	}

	return call.Results.SetRemote(*capRemote)
}

func (nh *netHandler) RemoteAddOrUpdate(call capnp.Net_remoteAddOrUpdate) error {
	server.Ack(call.Options)

	rp := nh.base.repo
	capRemote, err := call.Params.Remote()
	if err != nil {
		return err
	}

	remote, err := capRemoteToRemote(capRemote)
	if err != nil {
		return err
	}

	if err := rp.Remotes.AddOrUpdateRemote(*remote); err != nil {
		return err
	}

	return nh.base.syncRemoteStates()
}

func (nh *netHandler) RemoteClear(call capnp.Net_remoteClear) error {
	server.Ack(call.Options)
	return nh.base.repo.Remotes.Clear()
}

func (nh *netHandler) RemoteRm(call capnp.Net_remoteRm) error {
	server.Ack(call.Options)

	name, err := call.Params.Name()
	if err != nil {
		return err
	}

	rp := nh.base.repo
	if err := rp.Remotes.RmRemote(name); err != nil {
		return err
	}

	return nh.base.syncRemoteStates()
}

func (nh *netHandler) RemoteLs(call capnp.Net_remoteLs) error {
	server.Ack(call.Options)

	rp := nh.base.repo
	remotes, err := rp.Remotes.ListRemotes()
	if err != nil {
		return err
	}

	seg := call.Results.Segment()
	capRemotes, err := capnp.NewRemote_List(seg, int32(len(remotes)))
	if err != nil {
		return err
	}

	for idx, remote := range remotes {
		capRemote, err := remoteToCapRemote(remote, seg)
		if err != nil {
			return err
		}

		if err := capRemotes.Set(idx, *capRemote); err != nil {
			return err
		}
	}

	return call.Results.SetRemotes(capRemotes)
}

func (nh *netHandler) RemoteUpdate(call capnp.Net_remoteUpdate) error {
	server.Ack(call.Options)

	rp := nh.base.repo
	capRemote, err := call.Params.Remote()
	if err != nil {
		return err
	}

	remote, err := capRemoteToRemote(capRemote)
	if err != nil {
		return err
	}

	return rp.Remotes.AddOrUpdateRemote(*remote)
}

func (nh *netHandler) RemoteSave(call capnp.Net_remoteSave) error {
	server.Ack(call.Options)

	remotes := []repo.Remote{}
	capRemotes, err := call.Params.Remotes()
	if err != nil {
		return err
	}

	for idx := 0; idx < capRemotes.Len(); idx++ {
		capRemote := capRemotes.At(idx)
		remote, err := capRemoteToRemote(capRemote)
		if err != nil {
			return err
		}

		remotes = append(remotes, *remote)
	}

	rp := nh.base.repo
	if err := rp.Remotes.SaveList(remotes); err != nil {
		return err
	}

	return nh.base.syncRemoteStates()
}

// LocateResult is one entry in the result of the "net locate" command.
type LocateResult struct {
	Name        string
	Mask        string
	Addr        string
	Fingerprint string
}

func (nh *netHandler) peekAndCachePeer(peer peer.Info, mask p2pnet.LocateMask, ticket uint64) error {
	peekCtx, cancel := context.WithTimeout(nh.base.ctx, 2*time.Second)
	defer cancel()

	psrv := nh.base.peerServer
	fingerprint, remoteName, err := psrv.PeekFingerprint(peekCtx, peer.Addr)
	if err != nil {
		log.Warningf(
			"No fingerprint for %v %v",
			peer.Addr,
			err,
		)
		return err
	}

	if string(fingerprint) == "" {
		return nil
	}

	result := &LocateResult{
		Name:        string(remoteName),
		Addr:        string(peer.Addr),
		Mask:        mask.String(),
		Fingerprint: string(fingerprint),
	}

	log.Debugf("Pushing partial result: %v", result)
	if err := nh.base.conductor.Push(ticket, result); err != nil {
		log.Debugf("Failed to push result: %v", err)
		return err
	}

	return nil
}

func (nh *netHandler) NetLocate(call capnp.Net_netLocate) error {
	timeoutSec := call.Params.TimeoutSec()

	who, err := call.Params.Who()
	if err != nil {
		return err
	}

	locateMaskSpec, err := call.Params.LocateMask()
	if err != nil {
		return err
	}

	locateMask := p2pnet.LocateMask(p2pnet.LocateAll)
	if locateMaskSpec != "" {
		locateMask, err = p2pnet.LocateMaskFromString(locateMaskSpec)
		if err != nil {
			return err
		}
	}

	psrv := nh.base.peerServer
	ident, err := psrv.Identity()
	if err != nil {
		return err
	}

	addrCache := sync.Map{}
	addrCache.Store(ident.Addr, true)

	ticket := nh.base.conductor.Exec(func(ticket uint64) error {
		timeoutDur := time.Duration(timeoutSec * float64(time.Second))
		ctx, cancel := context.WithTimeout(nh.base.ctx, timeoutDur)
		defer cancel()

		log.WithFields(log.Fields{
			"who":     who,
			"timeout": timeoutSec,
			"mask":    locateMask,
		}).Debug("Starting locate...")

		locateCh := psrv.Locate(ctx, peer.Name(who), locateMask)

		wg := sync.WaitGroup{}
		for located := range locateCh {
			if located.Err != nil {
				log.Debugf("Locate failed for %s: %v", located.Name, located.Err)
				continue
			}

			// Every result might have more than one peer.
			// We should quickly check all of them to see if they are valid.
			// If some addrs are duplicated, we'll retrieve them from addrCache.
			for _, locatedPeer := range located.Peers {
				if _, ok := addrCache.Load(locatedPeer.Addr); ok {
					continue
				}

				log.Debugf("Fetching fingerprint for %v", locatedPeer.Addr)

				// Do the actual lookup in the background and start
				// the other lookups in parallel.
				wg.Add(1)
				go func(peer peer.Info, located p2pnet.LocateResult) {
					defer wg.Done()
					addrCache.Store(peer.Addr, true)
					nh.peekAndCachePeer(peer, located.Mask, ticket)
				}(locatedPeer, located)
			}
		}

		// Let background worker run until all go routines are finished.
		// Otherwise the result-killing timeout would set in too early.
		wg.Wait()
		return nil
	})

	// Tell the client under what ticket he can query for results.
	call.Results.SetTicket(ticket)
	return nil
}

func (nh *netHandler) NetLocateNext(call capnp.Net_netLocateNext) error {
	ticket := call.Params.Ticket()
	data, err := nh.base.conductor.Pop(ticket)
	if err != nil && !conductor.IsNoDataLeft(err) {
		return err
	}

	if conductor.IsNoDataLeft(err) {
		return nil
	}

	result, ok := data.(*LocateResult)
	if !ok {
		return fmt.Errorf("internal error: wrong type for LocateResult")
	}

	capResult, err := capnp.NewLocateResult(call.Results.Segment())
	if err != nil {
		return err
	}

	if err := capResult.SetName(result.Name); err != nil {
		return err
	}

	if err := capResult.SetAddr(result.Addr); err != nil {
		return err
	}

	if err := capResult.SetMask(result.Mask); err != nil {
		return err
	}

	if err := capResult.SetFingerprint(result.Fingerprint); err != nil {
		return err
	}

	return call.Results.SetResult(capResult)
}
