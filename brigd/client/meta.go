package client

import (
	"github.com/disorganizer/brig/brigd/capnp"
	capnplib "zombiezen.com/go/capnproto2"
)

func (cl *Client) Ping() error {
	call := cl.api.Ping(cl.ctx, func(p capnp.Meta_ping_Params) error {
		return nil
	})

	result, err := call.Struct()
	if err != nil {
		return err
	}

	_, err = result.Reply()
	return err
}

// Quit sends a quit signal to brigd.
func (cl *Client) Quit() error {
	call := cl.api.Quit(cl.ctx, func(p capnp.Meta_quit_Params) error {
		return nil
	})

	_, err := call.Struct()
	return err
}

func (cl *Client) Init(path, owner, password, backend string) error {
	call := cl.api.Init(cl.ctx, func(p capnp.Meta_init_Params) error {
		if err := p.SetOwner(owner); err != nil {
			return err
		}

		if err := p.SetPassword(password); err != nil {
			return err
		}

		if err := p.SetBasePath(path); err != nil {
			return err
		}

		return p.SetBackend(backend)
	})

	_, err := call.Struct()
	return err
}

func (cl *Client) Mount(mountPath string) error {
	call := cl.api.Mount(cl.ctx, func(p capnp.Meta_mount_Params) error {
		return p.SetMountPath(mountPath)
	})

	_, err := call.Struct()
	return err
}

func (cl *Client) Unmount(mountPath string) error {
	call := cl.api.Unmount(cl.ctx, func(p capnp.Meta_unmount_Params) error {
		return p.SetMountPath(mountPath)
	})

	_, err := call.Struct()
	return err
}

func (cl *Client) ConfigGet(key string) (string, error) {
	call := cl.api.ConfigGet(cl.ctx, func(p capnp.Meta_configGet_Params) error {
		return p.SetKey(key)
	})

	result, err := call.Struct()
	if err != nil {
		return "", err
	}

	return result.Value()
}

func (cl *Client) ConfigSet(key, value string) error {
	call := cl.api.ConfigSet(cl.ctx, func(p capnp.Meta_configSet_Params) error {
		if err := p.SetValue(value); err != nil {
			return err
		}

		return p.SetKey(key)
	})

	_, err := call.Struct()
	return err
}

func (cl *Client) ConfigAll() (map[string]string, error) {
	call := cl.api.ConfigAll(cl.ctx, func(p capnp.Meta_configAll_Params) error {
		return nil
	})

	result, err := call.Struct()
	if err != nil {
		return nil, err
	}

	pairs, err := result.All()
	if err != nil {
		return nil, err
	}

	configMap := make(map[string]string)

	for idx := 0; idx < pairs.Len(); idx++ {
		pair := pairs.At(idx)
		key, err := pair.Key()
		if err != nil {
			return nil, err
		}

		val, err := pair.Val()
		if err != nil {
			return nil, err
		}

		configMap[key] = val
	}

	return configMap, nil
}

////////////////////////
// REMOTE LIST ACCESS //
////////////////////////

type RemoteFolder struct {
	Folder string
	Perms  string
}

type Remote struct {
	Name        string         `yaml:"Name"`
	Fingerprint string         `yaml:"Fingerprint"`
	Folders     []RemoteFolder `yaml:"Folders,omitempty,flow"`
}

func capRemoteToRemote(remote capnp.Remote) (*Remote, error) {
	remoteName, err := remote.Name()
	if err != nil {
		return nil, err
	}

	remoteFp, err := remote.Fingerprint()
	if err != nil {
		return nil, err
	}

	remoteFolders, err := remote.Folders()
	if err != nil {
		return nil, err
	}

	folders := []RemoteFolder{}
	for idx := 0; idx < remoteFolders.Len(); idx++ {
		folder := remoteFolders.At(idx)
		folderName, err := folder.Folder()
		if err != nil {
			return nil, err
		}

		// TODO: Read perms here once defined.
		folders = append(folders, RemoteFolder{
			Folder: folderName,
		})
	}

	return &Remote{
		Name:        remoteName,
		Fingerprint: remoteFp,
		Folders:     folders,
	}, nil
}

func remoteToCapRemote(remote Remote, seg *capnplib.Segment) (*capnp.Remote, error) {
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

		if err := capFolder.SetPerms(folder.Perms); err != nil {
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

func (cl *Client) RemoteAdd(remote Remote) error {
	call := cl.api.RemoteAdd(cl.ctx, func(p capnp.Meta_remoteAdd_Params) error {
		capRemote, err := remoteToCapRemote(remote, p.Segment())
		if err != nil {
			return err
		}

		return p.SetRemote(*capRemote)
	})

	_, err := call.Struct()
	return err
}

func (cl *Client) RemoteRm(name string) error {
	call := cl.api.RemoteRm(cl.ctx, func(p capnp.Meta_remoteRm_Params) error {
		return p.SetName(name)
	})

	_, err := call.Struct()
	return err
}

func (cl *Client) RemoteLs() ([]Remote, error) {
	call := cl.api.RemoteLs(cl.ctx, func(p capnp.Meta_remoteLs_Params) error {
		return nil
	})

	result, err := call.Struct()
	if err != nil {
		return nil, err
	}

	capRemotes, err := result.Remotes()
	if err != nil {
		return nil, err
	}

	remotes := []Remote{}
	for idx := 0; idx < capRemotes.Len(); idx++ {
		capRemote := capRemotes.At(idx)
		remote, err := capRemoteToRemote(capRemote)
		if err != nil {
			return nil, err
		}

		remotes = append(remotes, *remote)
	}

	return remotes, nil
}

func (cl *Client) RemoteSave(remotes []Remote) error {
	call := cl.api.RemoteSave(cl.ctx, func(p capnp.Meta_remoteSave_Params) error {
		seg := p.Segment()
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

		return p.SetRemotes(capRemotes)
	})

	_, err := call.Struct()
	return err
}

func (cl *Client) RemoteLocate(who string) ([]Remote, error) {
	call := cl.api.RemoteLocate(cl.ctx, func(p capnp.Meta_remoteLocate_Params) error {
		return p.SetWho(who)
	})

	result, err := call.Struct()
	if err != nil {
		return nil, err
	}

	capRemotes, err := result.Candidates()
	if err != nil {
		return nil, err
	}

	remotes := []Remote{}
	for idx := 0; idx < capRemotes.Len(); idx++ {
		capRemote := capRemotes.At(idx)
		remote, err := capRemoteToRemote(capRemote)
		if err != nil {
			return nil, err
		}

		remotes = append(remotes, *remote)
	}

	return remotes, nil
}

func (cl *Client) RemotePing(who string) (float64, error) {
	call := cl.api.RemotePing(cl.ctx, func(p capnp.Meta_remotePing_Params) error {
		return p.SetWho(who)
	})

	result, err := call.Struct()
	if err != nil {
		return 0, err
	}

	return result.Roundtrip(), nil
}

func (cl *Client) Become(who string) error {
	call := cl.api.Become(cl.ctx, func(p capnp.Meta_become_Params) error {
		return p.SetWho(who)
	})

	_, err := call.Struct()
	return err
}

type Whoami struct {
	CurrentUser string
	Owner       string
	Fingerprint string
}

func (cl *Client) Whoami() (*Whoami, error) {
	call := cl.api.Whoami(cl.ctx, func(p capnp.Meta_whoami_Params) error {
		return nil
	})

	result, err := call.Struct()
	if err != nil {
		return nil, err
	}

	capWhoami, err := result.Whoami()
	if err != nil {
		return nil, err
	}

	whoami := &Whoami{}
	whoami.CurrentUser, err = capWhoami.CurrentUser()
	if err != nil {
		return nil, err
	}

	whoami.Fingerprint, err = capWhoami.Fingerprint()
	if err != nil {
		return nil, err
	}

	whoami.Owner, err = capWhoami.Owner()
	if err != nil {
		return nil, err
	}

	return whoami, nil
}
