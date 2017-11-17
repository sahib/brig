package server

import (
	"context"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/brigd/capnp"
	peernet "github.com/disorganizer/brig/net"
	"github.com/disorganizer/brig/net/peer"
	"github.com/disorganizer/brig/repo"
	capnplib "zombiezen.com/go/capnproto2"
	"zombiezen.com/go/capnproto2/server"
)

type metaHandler struct {
	base *base
}

func (mh *metaHandler) Quit(call capnp.Meta_quit) error {
	return mh.base.Quit()
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

	// Update the in-memory password.
	mh.base.password = password
	return repo.Init(initFolder, owner, password, backendName)
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

		// TODO: Read perms here once defined.
		folders = append(folders, repo.Folder{
			Folder: folderName,
		})
	}

	return &repo.Remote{
		Name:    remoteName,
		Folders: folders,
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

func (mh *metaHandler) RemoteAdd(call capnp.Meta_remoteAdd) error {
	repo, err := mh.base.Repo()
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

	return repo.Remotes.AddRemote(*remote)
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

	return repo.Remotes.RmRemote(name)
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

	repo, err := mh.base.Repo()
	if err != nil {
		return err
	}

	return repo.Remotes.SaveList(remotes)
}

func (mh *metaHandler) RemoteLocate(call capnp.Meta_remoteLocate) error {
	who, err := call.Params.Who()
	if err != nil {
		return err
	}

	psrv, err := mh.base.PeerServer()
	if err != nil {
		return err
	}

	log.Debugf("Trying to locate %v", who)
	foundPeers, err := psrv.Locate(peer.Name(who))
	if err != nil {
		return err
	}

	bk, err := mh.base.Backend()
	if err != nil {
		return err
	}

	seg := call.Results.Segment()
	capRemotes, err := capnp.NewRemote_List(seg, int32(len(foundPeers)))
	if err != nil {
		return err
	}

	// For the client side we do not differentiate between peers and remotes.
	// Also, the pubkey/network addr is combined into a single "fingerprint".
	for idx, foundPeer := range foundPeers {
		fingerprint := peer.Fingerprint("")

		// Query the remotes pubkey and use it to build the remotes' fingerprint.
		// If not available we just send an empty string back to the client.
		subCtx, cancel := context.WithTimeout(mh.base.ctx, 1*time.Minute)
		defer cancel()

		ctl, err := peernet.Dial(who, subCtx, bk)
		if err != nil {
			remotePubKey, err := ctl.PubKeyData()
			if err != nil {
				return err
			}

			fingerprint = peer.BuildFingerprint(foundPeer.Addr, remotePubKey)
		}

		remote := repo.Remote{
			Name:        string(foundPeer.Name),
			Fingerprint: fingerprint,
		}

		capRemote, err := remoteToCapRemote(remote, seg)
		if err != nil {
			return err
		}

		capRemotes.Set(idx, *capRemote)
	}

	return call.Results.SetCandidates(capRemotes)
}

func (mh *metaHandler) RemoteSelf(call capnp.Meta_remoteSelf) error {
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
	capRemote, err := capnp.NewRemote(call.Results.Segment())
	if err != nil {
		return err
	}

	if err := capRemote.SetName(string(self.Name)); err != nil {
		return err
	}

	if err := capRemote.SetFingerprint(string(finger)); err != nil {
		return err
	}

	return call.Results.SetSelf(capRemote)
}
