package server

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/sahib/brig/backend"
	p2pnet "github.com/sahib/brig/net"
	"github.com/sahib/brig/net/peer"
	"github.com/sahib/brig/repo"
	"github.com/sahib/brig/server/capnp"
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
	err = repo.Init(initFolder, owner, password, backendName)
	if err != nil {
		return err
	}

	rp, err := mh.base.Repo()
	if err != nil {
		return err
	}

	backendPath := rp.BackendPath(backendName)
	return backend.InitByName(backendName, backendPath)
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

	value := repo.Config.GetString(key)
	return call.Results.SetValue(value)
}

func (mh *metaHandler) ConfigSet(call capnp.Meta_configSet) error {
	repo, err := mh.base.Repo()
	if err != nil {
		return err
	}

	key, err := call.Params.Key()
	if err != nil {
		return err
	}

	val, err := call.Params.Value()
	if err != nil {
		return err
	}

	repo.Config.Set(key, val)
	return nil
}

func (mh *metaHandler) ConfigAll(call capnp.Meta_configAll) error {
	repo, err := mh.base.Repo()
	if err != nil {
		return err
	}

	all := repo.Config.AllKeys()
	seg := call.Results.Segment()

	lst, err := capnp.NewConfigPair_List(seg, int32(len(all)))
	if err != nil {
		return err
	}

	for idx, key := range all {
		pair, err := capnp.NewConfigPair(seg)
		if err != nil {
			return err
		}

		if err := pair.SetKey(key); err != nil {
			return err
		}

		if err := pair.SetVal(repo.Config.GetString(key)); err != nil {
			return err
		}

		if err := lst.Set(idx, pair); err != nil {
			return err
		}
	}

	return call.Results.SetAll(lst)
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

		if err := capFolder.SetPerms(folder.Perms.String()); err != nil {
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

func (mh *metaHandler) NetLocate(call capnp.Meta_netLocate) error {
	timeoutSec := call.Params.TimeoutSec()
	who, err := call.Params.Who()
	if err != nil {
		return err
	}

	psrv, err := mh.base.PeerServer()
	if err != nil {
		return err
	}

	log.Debugf("Trying to locate %v", who)
	foundPeers, err := psrv.Locate(
		peer.Name(who),
		int(timeoutSec),
		p2pnet.LocateAll,
	)
	if err != nil {
		return err
	}

	seg := call.Results.Segment()
	results := []capnp.LocateResult{}

	// For the client side we do not differentiate between peers and remotes.
	// Also, the pubkey/network addr is combined into a single "fingerprint".
	for mask, peers := range foundPeers {
		for _, peer := range peers {
			// This can be probably optimized furhter in case of several peeks:
			fingerprint, err := psrv.PeekFingerprint(mh.base.ctx, peer.Addr)
			if err != nil {
				log.Warningf("No fingerprint for %v (query: %s): %v", peer.Addr, who, err)
			}

			result, err := capnp.NewLocateResult(seg)
			if err != nil {
				return err
			}

			if err := result.SetAddr(peer.Addr); err != nil {
				return err
			}

			if err := result.SetMask(mask.String()); err != nil {
				return err
			}

			if err := result.SetFingerprint(string(fingerprint)); err != nil {
				return err
			}

			results = append(results, result)
		}
	}

	capResults, err := capnp.NewLocateResult_List(seg, int32(len(results)))
	if err != nil {
		return err
	}

	for resultIdx, result := range results {
		if err := capResults.Set(resultIdx, result); err != nil {
			return err
		}
	}

	return call.Results.SetCandidates(capResults)
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

		addr := remote.Fingerprint.Addr()
		if err := status.SetAddr(addr); err != nil {
			return err
		}

		if err := status.SetName(remote.Name); err != nil {
			return err
		}

		pinger, err := psrv.PingMap().For(addr)
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
