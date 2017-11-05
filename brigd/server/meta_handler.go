package server

import (
	"github.com/disorganizer/brig/brigd/capnp"
	"github.com/disorganizer/brig/repo"
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

// TODO: Use later for remotes:
// func capRemoteToRemote(remote capnp.Remote) (*repo.Remote, error) {
// 	remoteName, err := remote.Name()
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	remoteFolders, err := remote.Folders()
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	folders := []repo.Folder{}
// 	for idx := 0; idx < remoteFolders.Len(); idx++ {
// 		folder, err := remoteFolders.At(idx)
// 		if err != nil {
// 			return nil, err
// 		}
//
// 		// TODO: Read perms here once defined.
// 		folders = append(folders, repo.Folder{
// 			Folder: folder,
// 		})
// 	}
//
// 	return &repo.Remote{
// 		Name:    remoteName,
// 		Folders: folders,
// 	}, nil
// }
//
// func (mh *metaHandler) RemoteAdd(call capnp.Meta_remoteAdd) error {
// 	repo, err := mh.base.Repo()
// 	if err != nil {
// 		return err
// 	}
//
// 	capRemote, err := call.Params.Remote()
// 	if err != nil {
// 		return err
// 	}
//
// 	remote, err := capRemoteToRemote(capRemote)
// 	if err != nil {
// 		return err
// 	}
//
// 	return repo.Remotes.AddRemote(*remote)
// }
