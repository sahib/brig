package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	e "github.com/pkg/errors"
	"github.com/sahib/brig/backend"
	p2pnet "github.com/sahib/brig/net"
	"github.com/sahib/brig/net/peer"
	"github.com/sahib/brig/repo"
	"github.com/sahib/brig/server/capnp"
	"github.com/sahib/brig/util/conductor"
	"github.com/sahib/brig/version"
	capnplib "zombiezen.com/go/capnproto2"
	"zombiezen.com/go/capnproto2/server"
)

type metaHandler struct {
	base *base
}

func (mh *metaHandler) Quit(call capnp.Meta_quit) error {
	mh.base.quitCh <- struct{}{}
	return nil
}

func (mh *metaHandler) Ping(call capnp.Meta_ping) error {
	server.Ack(call.Options)
	return call.Results.SetReply("PONG")
}

func (mh *metaHandler) Init(call capnp.Meta_init) error {
	server.Ack(call.Options)

	backendName, err := call.Params.Backend()
	if err != nil {
		return err
	}

	initFolder, err := call.Params.BasePath()
	if err != nil {
		return err
	}

	password, err := call.Params.Password()
	if err != nil {
		return err
	}

	owner, err := call.Params.Owner()
	if err != nil {
		return err
	}

	if !backend.IsValidName(backendName) {
		return fmt.Errorf("Invalid backend name: %v", backendName)
	}

	// Update the in-memory password.
	mh.base.password = password
	mh.base.basePath = initFolder

	err = repo.Init(initFolder, owner, password, backendName)
	if err != nil {
		return e.Wrapf(err, "repo-init")
	}

	rp, err := mh.base.Repo()
	if err != nil {
		return err
	}

	backendPath := rp.BackendPath(backendName)
	err = backend.InitByName(backendName, backendPath)
	return e.Wrapf(err, "backend-init")
}

func (mh *metaHandler) Mount(call capnp.Meta_mount) error {
	server.Ack(call.Options)

	mountPath, err := call.Params.MountPath()
	if err != nil {
		return err
	}

	mounts, err := mh.base.Mounts()
	if err != nil {
		return err
	}

	_, err = mounts.AddMount(mountPath)
	return err
}

func (mh *metaHandler) Unmount(call capnp.Meta_unmount) error {
	server.Ack(call.Options)

	mountPath, err := call.Params.MountPath()
	if err != nil {
		return err
	}

	mounts, err := mh.base.Mounts()
	if err != nil {
		return err
	}

	return mounts.Unmount(mountPath)
}

func (mh *metaHandler) ConfigGet(call capnp.Meta_configGet) error {
	repo, err := mh.base.Repo()
	if err != nil {
		return err
	}

	key, err := call.Params.Key()
	if err != nil {
		return err
	}

	if !repo.Config.IsValidKey(key) {
		return fmt.Errorf("invalid key: %v", key)
	}

	value := repo.Config.String(key)
	return call.Results.SetValue(value)
}

func (mh *metaHandler) ConfigDoc(call capnp.Meta_configDoc) error {
	repo, err := mh.base.Repo()
	if err != nil {
		return err
	}

	key, err := call.Params.Key()
	if err != nil {
		return err
	}

	if !repo.Config.IsValidKey(key) {
		return fmt.Errorf("invalid key: %v", key)
	}

	seg := call.Results.Segment()
	capPair, err := mh.configDefaultEntryToCapnp(seg, key)
	if err != nil {
		return err
	}

	return call.Results.SetDesc(*capPair)
}

func (mh *metaHandler) ConfigSet(call capnp.Meta_configSet) error {
	rp, err := mh.base.Repo()
	if err != nil {
		return err
	}

	key, err := call.Params.Key()
	if err != nil {
		return err
	}

	if !rp.Config.IsValidKey(key) {
		return fmt.Errorf("invalid key: %v", key)
	}

	rawVal, err := call.Params.Value()
	if err != nil {
		return err
	}

	val, err := rp.Config.Cast(key, rawVal)
	if err != nil {
		return err
	}

	log.Debugf("config: set `%s` to `%v`", key, val)
	if err := rp.Config.Set(key, val); err != nil {
		return err
	}

	return rp.SaveConfig()
}

func (mh *metaHandler) configDefaultEntryToCapnp(seg *capnplib.Segment, key string) (*capnp.ConfigEntry, error) {
	pair, err := capnp.NewConfigEntry(seg)
	if err != nil {
		return nil, err
	}

	if err := pair.SetKey(key); err != nil {
		return nil, err
	}

	repo, err := mh.base.Repo()
	if err != nil {
		return nil, err
	}

	clientVal := fmt.Sprintf("%v", repo.Config.Get(key))
	if err := pair.SetVal(clientVal); err != nil {
		return nil, err
	}

	entry := repo.Config.GetDefault(key)
	if err := pair.SetDoc(entry.Docs); err != nil {
		return nil, err
	}

	defVal := fmt.Sprintf("%v", entry.Default)
	if err := pair.SetDefault(defVal); err != nil {
		return nil, err
	}

	pair.SetNeedsRestart(entry.NeedsRestart)
	return &pair, nil
}

func (mh *metaHandler) ConfigAll(call capnp.Meta_configAll) error {
	repo, err := mh.base.Repo()
	if err != nil {
		return err
	}

	all := repo.Config.Keys()
	seg := call.Results.Segment()

	capLst, err := capnp.NewConfigEntry_List(seg, int32(len(all)))
	if err != nil {
		return err
	}

	for idx, key := range all {
		capPair, err := mh.configDefaultEntryToCapnp(seg, key)
		if err != nil {
			return err
		}

		if err := capLst.Set(idx, *capPair); err != nil {
			return err
		}
	}

	return call.Results.SetAll(capLst)
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

func (mh *metaHandler) syncPingMap() error {
	psrv, err := mh.base.PeerServer()
	if err != nil {
		return err
	}

	rp, err := mh.base.Repo()
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

func (mh *metaHandler) RemoteAdd(call capnp.Meta_remoteAdd) error {
	rp, err := mh.base.Repo()
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

	return mh.syncPingMap()
}

func (mh *metaHandler) RemoteClear(call capnp.Meta_remoteClear) error {
	rp, err := mh.base.Repo()
	if err != nil {
		return err
	}

	return rp.Remotes.Clear()
}

func (mh *metaHandler) RemoteRm(call capnp.Meta_remoteRm) error {
	repo, err := mh.base.Repo()
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

	return mh.syncPingMap()
}

func (mh *metaHandler) RemoteLs(call capnp.Meta_remoteLs) error {
	repo, err := mh.base.Repo()
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

func (mh *metaHandler) RemoteSave(call capnp.Meta_remoteSave) error {
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

	rp, err := mh.base.Repo()
	if err != nil {
		return err
	}

	if err := rp.Remotes.SaveList(remotes); err != nil {
		return err
	}

	return mh.syncPingMap()
}

type LocateResult struct {
	Name        string
	Mask        string
	Addr        string
	Fingerprint string
}

// TODO: This method is too complex, clean it up a bit
//       and move parts of it to net.Server
func (mh *metaHandler) NetLocate(call capnp.Meta_netLocate) error {
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

	psrv, err := mh.base.PeerServer()
	if err != nil {
		return err
	}

	ident, err := psrv.Identity()
	if err != nil {
		return err
	}

	addrCache := sync.Map{}
	addrCache.Store(ident.Addr, true)

	ticket := mh.base.conductor.Exec(func(ticket uint64) error {
		timeoutDur := time.Duration(timeoutSec * float64(time.Second))
		ctx, cancel := context.WithTimeout(mh.base.ctx, timeoutDur)
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

					peekCtx, cancel := context.WithTimeout(mh.base.ctx, 2*time.Second)
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
					if err := mh.base.conductor.Push(ticket, result); err != nil {
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

func (mh *metaHandler) NetLocateNext(call capnp.Meta_netLocateNext) error {
	ticket := call.Params.Ticket()
	data, err := mh.base.conductor.Pop(ticket)
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

func (mh *metaHandler) RemotePing(call capnp.Meta_remotePing) error {
	who, err := call.Params.Who()
	if err != nil {
		return err
	}

	return mh.base.withNetClient(who, func(ctl *p2pnet.Client) error {
		start := time.Now()
		if err := ctl.Ping(); err != nil {
			return err
		}

		roundtrip := time.Since(start).Seconds()
		call.Results.SetRoundtrip(roundtrip)
		return nil
	})
}

func (mh *metaHandler) Become(call capnp.Meta_become) error {
	who, err := call.Params.Who()
	if err != nil {
		return err
	}

	rp, err := mh.base.Repo()
	if err != nil {
		return err
	}

	// We can only be users that are present in the remote list (and owner)
	// (This is not a technical limitation)
	if who != rp.Owner {
		_, err = rp.Remotes.Remote(who)
		if err != nil {
			return err
		}
	}

	log.Infof("Becoming: %v", who)
	rp.SetCurrentUser(who)
	return nil
}

func (mh *metaHandler) Whoami(call capnp.Meta_whoami) error {
	capId, err := capnp.NewIdentity(call.Results.Segment())
	if err != nil {
		return err
	}

	psrv, err := mh.base.PeerServer()
	if err != nil {
		return err
	}

	self, err := psrv.Identity()
	if err != nil {
		return err
	}

	rp, err := mh.base.Repo()
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

func (mh *metaHandler) Connect(call capnp.Meta_connect) error {
	psrv, err := mh.base.PeerServer()
	if err != nil {
		return err
	}

	log.Infof("backend is going online...")
	return psrv.Connect()
}

func (mh *metaHandler) Disconnect(call capnp.Meta_disconnect) error {
	psrv, err := mh.base.PeerServer()
	if err != nil {
		return err
	}

	log.Infof("backend is going offline...")
	return psrv.Disconnect()
}

func (mh *metaHandler) OnlinePeers(call capnp.Meta_onlinePeers) error {
	rp, err := mh.base.Repo()
	if err != nil {
		return err
	}

	psrv, err := mh.base.PeerServer()
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

func (mh *metaHandler) Version(call capnp.Meta_version) error {
	rp, err := mh.base.Repo()
	if err != nil {
		return err
	}

	name := rp.BackendName()
	bkVersion := backend.Version(name)
	if bkVersion == nil {
		return fmt.Errorf("bug: invalid backend name: %v", name)
	}

	capVersion, err := capnp.NewVersion(call.Results.Segment())
	if err != nil {
		return err
	}

	if err := capVersion.SetServerVersion(version.String()); err != nil {
		return err
	}

	if err := capVersion.SetServerRev(version.GitRev); err != nil {
		return err
	}

	if err := capVersion.SetBackendVersion(bkVersion.SemVer()); err != nil {
		return err
	}

	if err := capVersion.SetBackendRev(bkVersion.Rev()); err != nil {
		return err
	}

	return call.Results.SetVersion(capVersion)
}
