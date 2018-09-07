package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	p2pnet "github.com/sahib/brig/net"
	"github.com/sahib/brig/net/peer"
	"github.com/sahib/brig/repo"
	"github.com/sahib/brig/server/capnp"
	"github.com/sahib/brig/util/conductor"
	capnplib "zombiezen.com/go/capnproto2"
)

type netHandler struct {
	base *base
}

func (nh *netHandler) Whoami(call capnp.Net_whoami) error {
	capId, err := capnp.NewIdentity(call.Results.Segment())
	if err != nil {
		return err
	}

	psrv, err := nh.base.PeerServer()
	if err != nil {
		return err
	}

	self, err := psrv.Identity()
	if err != nil {
		return err
	}

	rp, err := nh.base.Repo()
	if err != nil {
		return err
	}

	// Compute our own fingerprint:
	ownPubKey, err := rp.Keyring().OwnPubKey()
	if err != nil {
		return err
	}

	finger := peer.BuildFingerprint(self.Addr, ownPubKey)

	if err := capId.SetOwner(rp.Owner); err != nil {
		return err
	}

	if err := capId.SetFingerprint(string(finger)); err != nil {
		return err
	}

	if err := capId.SetCurrentUser(rp.CurrentUser()); err != nil {
		return err
	}

	// TODO: Asking for IsOnline() can cause an initial Connect() currently.
	capId.SetIsOnline(psrv.IsOnline())
	return call.Results.SetWhoami(capId)
}

func (nh *netHandler) Connect(call capnp.Net_connect) error {
	psrv, err := nh.base.PeerServer()
	if err != nil {
		return err
	}

	log.Infof("backend is going online...")
	return psrv.Connect()
}

func (nh *netHandler) Disconnect(call capnp.Net_disconnect) error {
	psrv, err := nh.base.PeerServer()
	if err != nil {
		return err
	}

	log.Infof("backend is going offline...")
	return psrv.Disconnect()
}

func (nh *netHandler) OnlinePeers(call capnp.Net_onlinePeers) error {
	rp, err := nh.base.Repo()
	if err != nil {
		return err
	}

	psrv, err := nh.base.PeerServer()
	if err != nil {
		return err
	}

	remotes, err := rp.Remotes.ListRemotes()
	if err != nil {
		return err
	}

	seg := call.Results.Segment()
	statuses, err := capnp.NewPeerStatus_List(seg, int32(len(remotes)))
	if err != nil {
		return err
	}

	for idx, remote := range remotes {
		status, err := capnp.NewPeerStatus(call.Results.Segment())
		if err != nil {
			return err
		}

		fp := string(remote.Fingerprint)
		if err := status.SetFingerprint(fp); err != nil {
			return err
		}

		if err := status.SetName(remote.Name); err != nil {
			return err
		}

		pinger, err := psrv.PingMap().For(remote.Fingerprint.Addr())
		if err != nil {
			status.SetError(err.Error())
		}

		if pinger != nil {
			roundtrip := int32(pinger.Roundtrip() / time.Millisecond)
			status.SetRoundtripMs(roundtrip)

			lastSeen := pinger.LastSeen().Format(time.RFC3339)
			if err := status.SetLastSeen(lastSeen); err != nil {
				return err
			}
		} else {
			errMsg := fmt.Sprintf("no route (yet)")
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
		Name:        remoteName,
		Fingerprint: peer.Fingerprint(fingerprint),
		Folders:     folders,
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

	return &capRemote, nil
}

func (nh *netHandler) syncPingMap() error {
	psrv, err := nh.base.PeerServer()
	if err != nil {
		return err
	}

	rp, err := nh.base.Repo()
	if err != nil {
		return err
	}

	addrs := []string{}
	remotes, err := rp.Remotes.ListRemotes()
	if err != nil {
		return err
	}

	for _, remote := range remotes {
		addrs = append(addrs, remote.Fingerprint.Addr())
	}

	return psrv.PingMap().Sync(addrs)
}

func (nh *netHandler) RemoteAdd(call capnp.Net_remoteAdd) error {
	rp, err := nh.base.Repo()
	if err != nil {
		return err
	}

	capRemote, err := call.Params.Remote()
	if err != nil {
		return err
	}

	remote, err := capRemoteToRemote(capRemote)
	if err != nil {
		return err
	}

	if err := rp.Remotes.AddRemote(*remote); err != nil {
		return err
	}

	return nh.syncPingMap()
}

func (nh *netHandler) RemoteClear(call capnp.Net_remoteClear) error {
	rp, err := nh.base.Repo()
	if err != nil {
		return err
	}

	return rp.Remotes.Clear()
}

func (nh *netHandler) RemoteRm(call capnp.Net_remoteRm) error {
	repo, err := nh.base.Repo()
	if err != nil {
		return err
	}

	name, err := call.Params.Name()
	if err != nil {
		return err
	}

	if err := repo.Remotes.RmRemote(name); err != nil {
		return err
	}

	return nh.syncPingMap()
}

func (nh *netHandler) RemoteLs(call capnp.Net_remoteLs) error {
	repo, err := nh.base.Repo()
	if err != nil {
		return err
	}

	remotes, err := repo.Remotes.ListRemotes()
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

func (nh *netHandler) RemoteSave(call capnp.Net_remoteSave) error {
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

	rp, err := nh.base.Repo()
	if err != nil {
		return err
	}

	if err := rp.Remotes.SaveList(remotes); err != nil {
		return err
	}

	return nh.syncPingMap()
}

type LocateResult struct {
	Name        string
	Mask        string
	Addr        string
	Fingerprint string
}

// TODO: This method is too complex, clean it up a bit
//       and move parts of it to net.Server
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

	psrv, err := nh.base.PeerServer()
	if err != nil {
		return err
	}

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

			for _, locatedPeer := range located.Peers {
				if _, ok := addrCache.Load(locatedPeer.Addr); ok {
					continue
				}

				log.Debugf("Fetching fingerprint for %v", locatedPeer.Addr)

				wg.Add(1)
				go func(peer peer.Info) {
					defer wg.Done()

					peekCtx, cancel := context.WithTimeout(nh.base.ctx, 2*time.Second)
					defer cancel()

					// Remember that we already resolved this addr.
					addrCache.Store(peer.Addr, true)

					fingerprint, remoteName, err := psrv.PeekFingerprint(peekCtx, peer.Addr)
					if err != nil {
						log.Warningf(
							"No fingerprint for %v (query: %s): %v",
							peer.Addr,
							who,
							err,
						)
						return
					}

					if string(fingerprint) == "" {
						return
					}

					result := &LocateResult{
						Name:        string(remoteName),
						Addr:        string(peer.Addr),
						Mask:        located.Mask.String(),
						Fingerprint: string(fingerprint),
					}

					log.Debugf("Pushing partial result: %v", result)
					if err := nh.base.conductor.Push(ticket, result); err != nil {
						log.Debugf("Failed to push result: %v", err)
					}
				}(locatedPeer)
			}
		}

		// Let background worker run until all go routines are finished.
		// Otherwise the result-killing timeout would set in too early.
		wg.Wait()
		return nil
	})

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
