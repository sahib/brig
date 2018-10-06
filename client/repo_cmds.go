package client

import (
	"github.com/sahib/brig/server/capnp"
	h "github.com/sahib/brig/util/hashlib"
	capnplib "zombiezen.com/go/capnproto2"
)

// Quit sends a quit signal to brigd.
func (cl *Client) Quit() error {
	call := cl.api.Quit(cl.ctx, func(p capnp.Repo_quit_Params) error {
		return nil
	})

	_, err := call.Struct()
	return err
}

func (cl *Client) Ping() error {
	call := cl.api.Ping(cl.ctx, func(p capnp.Repo_ping_Params) error {
		return nil
	})

	result, err := call.Struct()
	if err != nil {
		return err
	}

	_, err = result.Reply()
	return err
}

func (cl *Client) Init(path, owner, password, backend string) error {
	call := cl.api.Init(cl.ctx, func(p capnp.Repo_init_Params) error {
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

type MountOptions struct {
	ReadOnly bool
	RootPath string
}

func mountOptionsToCapnp(opts MountOptions, seg *capnplib.Segment) (*capnp.MountOptions, error) {
	capOpts, err := capnp.NewMountOptions(seg)
	if err != nil {
		return nil, err
	}

	capOpts.SetReadOnly(opts.ReadOnly)
	if err := capOpts.SetRootPath(opts.RootPath); err != nil {
		return nil, err
	}

	return &capOpts, nil
}

func (cl *Client) Mount(mountPath string, opts MountOptions) error {
	call := cl.api.Mount(cl.ctx, func(p capnp.Repo_mount_Params) error {
		capOpts, err := mountOptionsToCapnp(opts, p.Segment())
		if err != nil {
			return err
		}

		if err := p.SetOptions(*capOpts); err != nil {
			return err
		}

		return p.SetMountPath(mountPath)
	})

	_, err := call.Struct()
	return err
}

func (cl *Client) Unmount(mountPath string) error {
	call := cl.api.Unmount(cl.ctx, func(p capnp.Repo_unmount_Params) error {
		return p.SetMountPath(mountPath)
	})

	_, err := call.Struct()
	return err
}

func (cl *Client) ConfigGet(key string) (string, error) {
	call := cl.api.ConfigGet(cl.ctx, func(p capnp.Repo_configGet_Params) error {
		return p.SetKey(key)
	})

	result, err := call.Struct()
	if err != nil {
		return "", err
	}

	return result.Value()
}

func (cl *Client) ConfigSet(key, value string) error {
	call := cl.api.ConfigSet(cl.ctx, func(p capnp.Repo_configSet_Params) error {
		if err := p.SetValue(value); err != nil {
			return err
		}

		return p.SetKey(key)
	})

	_, err := call.Struct()
	return err
}

type ConfigEntry struct {
	Key          string
	Val          string
	Doc          string
	Default      string
	NeedsRestart bool
}

func configEntryFromCapnp(capEntry capnp.ConfigEntry) (*ConfigEntry, error) {
	key, err := capEntry.Key()
	if err != nil {
		return nil, err
	}

	val, err := capEntry.Val()
	if err != nil {
		return nil, err
	}

	doc, err := capEntry.Doc()
	if err != nil {
		return nil, err
	}

	def, err := capEntry.Default()
	if err != nil {
		return nil, err
	}

	return &ConfigEntry{
		Default:      def,
		Key:          key,
		Val:          val,
		Doc:          doc,
		NeedsRestart: capEntry.NeedsRestart(),
	}, nil
}

func (cl *Client) ConfigAll() ([]ConfigEntry, error) {
	call := cl.api.ConfigAll(cl.ctx, func(p capnp.Repo_configAll_Params) error {
		return nil
	})

	result, err := call.Struct()
	if err != nil {
		return nil, err
	}

	capPairs, err := result.All()
	if err != nil {
		return nil, err
	}

	entries := []ConfigEntry{}
	for idx := 0; idx < capPairs.Len(); idx++ {
		capEntry := capPairs.At(idx)
		entry, err := configEntryFromCapnp(capEntry)
		if err != nil {
			return nil, err
		}

		entries = append(entries, *entry)
	}

	return entries, nil
}

func (cl *Client) ConfigDoc(key string) (ConfigEntry, error) {
	call := cl.api.ConfigDoc(cl.ctx, func(p capnp.Repo_configDoc_Params) error {
		return p.SetKey(key)
	})

	result, err := call.Struct()
	if err != nil {
		return ConfigEntry{}, err
	}

	capEntry, err := result.Desc()
	if err != nil {
		return ConfigEntry{}, err
	}

	entry, err := configEntryFromCapnp(capEntry)
	if err != nil {
		return ConfigEntry{}, err
	}

	return *entry, nil
}

type VersionInfo struct {
	ServerSemVer  string
	ServerRev     string
	BackendSemVer string
	BackendRev    string
}

func (cl *Client) Version() (*VersionInfo, error) {
	call := cl.api.Version(cl.ctx, func(p capnp.Repo_version_Params) error {
		return nil
	})

	result, err := call.Struct()
	if err != nil {
		return nil, err
	}

	capVersion, err := result.Version()
	if err != nil {
		return nil, err
	}

	version := &VersionInfo{}
	version.ServerSemVer, err = capVersion.ServerVersion()
	if err != nil {
		return nil, err
	}

	version.ServerRev, err = capVersion.ServerRev()
	if err != nil {
		return nil, err
	}

	version.BackendSemVer, err = capVersion.BackendVersion()
	if err != nil {
		return nil, err
	}

	version.BackendRev, err = capVersion.BackendRev()
	if err != nil {
		return nil, err
	}

	return version, nil
}

func (ctl *Client) FstabAdd(mountName, mountPath string, opts MountOptions) error {
	call := ctl.api.FstabAdd(ctl.ctx, func(p capnp.Repo_fstabAdd_Params) error {
		if err := p.SetMountName(mountName); err != nil {
			return err
		}

		if err := p.SetMountPath(mountPath); err != nil {
			return err
		}

		capOpts, err := mountOptionsToCapnp(opts, p.Segment())
		if err != nil {
			return err
		}

		return p.SetOptions(*capOpts)
	})

	_, err := call.Struct()
	return err
}

func (ctl *Client) FstabRemove(mountName string) error {
	call := ctl.api.FstabRemove(ctl.ctx, func(p capnp.Repo_fstabRemove_Params) error {
		return p.SetMountName(mountName)
	})

	_, err := call.Struct()
	return err
}

func (ctl *Client) FstabApply() error {
	call := ctl.api.FstabApply(ctl.ctx, func(p capnp.Repo_fstabApply_Params) error {
		return nil
	})

	_, err := call.Struct()
	return err
}

func (ctl *Client) FstabUnmountAll() error {
	call := ctl.api.FstabUnmountAll(ctl.ctx, func(p capnp.Repo_fstabUnmountAll_Params) error {
		return nil
	})

	_, err := call.Struct()
	return err
}

type FsTabEntry struct {
	Name     string
	Path     string
	Root     string
	Active   bool
	ReadOnly bool
}

func capMountToMount(capEntry capnp.FsTabEntry) (*FsTabEntry, error) {
	name, err := capEntry.Name()
	if err != nil {
		return nil, err
	}

	root, err := capEntry.Root()
	if err != nil {
		return nil, err
	}

	path, err := capEntry.Path()
	if err != nil {
		return nil, err
	}

	return &FsTabEntry{
		Path:     path,
		Name:     name,
		Root:     root,
		Active:   capEntry.Active(),
		ReadOnly: capEntry.ReadOnly(),
	}, nil
}

func (ctl *Client) FsTabList() ([]FsTabEntry, error) {
	call := ctl.api.FstabList(ctl.ctx, func(p capnp.Repo_fstabList_Params) error {
		return nil
	})

	result, err := call.Struct()
	if err != nil {
		return nil, err
	}

	capMounts, err := result.Mounts()
	if err != nil {
		return nil, err
	}

	mounts := []FsTabEntry{}
	for idx := 0; idx < capMounts.Len(); idx++ {
		capMount := capMounts.At(idx)
		mount, err := capMountToMount(capMount)
		if err != nil {
			return nil, err
		}

		mounts = append(mounts, *mount)
	}

	return mounts, nil
}

type GarbageItem struct {
	Path    string
	Owner   string
	Content h.Hash
}

func (cl *Client) GarbageCollect(aggressive bool) ([]*GarbageItem, error) {
	call := cl.api.GarbageCollect(cl.ctx, func(p capnp.FS_garbageCollect_Params) error {
		p.SetAggressive(aggressive)
		return nil
	})

	result, err := call.Struct()
	if err != nil {
		return nil, err
	}

	freed := []*GarbageItem{}

	capFreed, err := result.Freed()
	if err != nil {
		return nil, err
	}

	for idx := 0; idx < capFreed.Len(); idx++ {
		capGcItem := capFreed.At(idx)
		gcItem := &GarbageItem{}

		gcItem.Owner, err = capGcItem.Owner()
		if err != nil {
			return nil, err
		}

		gcItem.Path, err = capGcItem.Path()
		if err != nil {
			return nil, err
		}

		content, err := capGcItem.Content()
		if err != nil {
			return nil, err
		}

		gcItem.Content, err = h.Cast(content)
		if err != nil {
			return nil, err
		}

		freed = append(freed, gcItem)
	}

	return freed, nil
}

func (cl *Client) Become(who string) error {
	call := cl.api.Become(cl.ctx, func(p capnp.Repo_become_Params) error {
		return p.SetWho(who)
	})

	_, err := call.Struct()
	return err
}

func (cl *Client) WaitForInit() error {
	call := cl.api.WaitForInit(cl.ctx, func(p capnp.Repo_waitForInit_Params) error {
		return nil
	})

	_, err := call.Struct()
	return err
}
