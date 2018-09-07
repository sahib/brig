package server

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	e "github.com/pkg/errors"
	"github.com/sahib/brig/backend"
	"github.com/sahib/brig/fuse"
	"github.com/sahib/brig/repo"
	"github.com/sahib/brig/server/capnp"
	"github.com/sahib/brig/version"
	capnplib "zombiezen.com/go/capnproto2"
	"zombiezen.com/go/capnproto2/server"
)

type repoHandler struct {
	base *base
}

func (rh *repoHandler) Quit(call capnp.Repo_quit) error {
	rh.base.quitCh <- struct{}{}
	return nil
}

func (rh *repoHandler) Ping(call capnp.Repo_ping) error {
	server.Ack(call.Options)
	return call.Results.SetReply("PONG")
}

func (rh *repoHandler) Init(call capnp.Repo_init) error {
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
	rh.base.password = password
	rh.base.basePath = initFolder

	err = repo.Init(initFolder, owner, password, backendName)
	if err != nil {
		return e.Wrapf(err, "repo-init")
	}

	rp, err := rh.base.Repo()
	if err != nil {
		return err
	}

	backendPath := rp.BackendPath(backendName)

	err = backend.InitByName(backendName, backendPath)
	return e.Wrapf(err, "backend-init")
}

func (rh *repoHandler) Mount(call capnp.Repo_mount) error {
	server.Ack(call.Options)

	mountPath, err := call.Params.MountPath()
	if err != nil {
		return err
	}

	capOpts, err := call.Params.Options()
	if err != nil {
		return err
	}

	mountOptions, err := capMountOptionsToMountOptions(capOpts)
	if err != nil {
		return err
	}

	mounts, err := rh.base.Mounts()
	if err != nil {
		return err
	}

	_, err = mounts.AddMount(mountPath, mountOptions)
	return err
}

func (rh *repoHandler) Unmount(call capnp.Repo_unmount) error {
	server.Ack(call.Options)

	mountPath, err := call.Params.MountPath()
	if err != nil {
		return err
	}

	mounts, err := rh.base.Mounts()
	if err != nil {
		return err
	}

	return mounts.Unmount(mountPath)
}

func capMountOptionsToMountOptions(capOpts capnp.MountOptions) (fuse.MountOptions, error) {
	readOnly := capOpts.ReadOnly()
	rootPath, err := capOpts.RootPath()
	if err != nil {
		return fuse.MountOptions{}, err
	}

	return fuse.MountOptions{
		ReadOnly: readOnly,
		Root:     rootPath,
	}, nil
}

func (rh *repoHandler) FstabAdd(call capnp.Repo_fstabAdd) error {
	server.Ack(call.Options)

	rp, err := rh.base.Repo()
	if err != nil {
		return err
	}

	mountName, err := call.Params.MountName()
	if err != nil {
		return err
	}

	mountPath, err := call.Params.MountPath()
	if err != nil {
		return err
	}

	options, err := call.Params.Options()
	if err != nil {
		return err
	}

	mountOptions, err := capMountOptionsToMountOptions(options)
	if err != nil {
		return err
	}

	return fuse.FsTabAdd(rp.Config.Section("mounts"), mountName, mountPath, mountOptions)
}

func (rh *repoHandler) FstabRemove(call capnp.Repo_fstabRemove) error {
	server.Ack(call.Options)

	rp, err := rh.base.Repo()
	if err != nil {
		return err
	}

	mountName, err := call.Params.MountName()
	if err != nil {
		return err
	}

	if err := fuse.FsTabRemove(rp.Config.Section("mounts"), mountName); err != nil {
		return err
	}

	return rp.SaveConfig()
}

func (rh *repoHandler) FstabApply(call capnp.Repo_fstabApply) error {
	server.Ack(call.Options)

	rp, err := rh.base.Repo()
	if err != nil {
		return err
	}

	mounts, err := rh.base.Mounts()
	if err != nil {
		return err
	}

	if err := fuse.FsTabApply(rp.Config.Section("mounts"), mounts); err != nil {
		return err
	}

	return rp.SaveConfig()
}

func (rh *repoHandler) FstabUnmountAll(call capnp.Repo_fstabUnmountAll) error {
	server.Ack(call.Options)

	rp, err := rh.base.Repo()
	if err != nil {
		return err
	}

	mounts, err := rh.base.Mounts()
	if err != nil {
		return err
	}

	return fuse.FsTabUnmountAll(rp.Config, mounts)
}

func fsTabEntryToCap(entry fuse.FsTabEntry, seg *capnplib.Segment) (*capnp.FsTabEntry, error) {
	capEntry, err := capnp.NewFsTabEntry(seg)
	if err != nil {
		return nil, err
	}

	capEntry.SetReadOnly(entry.ReadOnly)
	capEntry.SetActive(entry.Active)

	if err := capEntry.SetPath(entry.Path); err != nil {
		return nil, err
	}
	if err := capEntry.SetRoot(entry.Root); err != nil {
		return nil, err
	}
	if err := capEntry.SetName(entry.Name); err != nil {
		return nil, err
	}

	return &capEntry, nil
}

func (rh *repoHandler) FstabList(call capnp.Repo_fstabList) error {
	server.Ack(call.Options)

	rp, err := rh.base.Repo()
	if err != nil {
		return err
	}

	mounts, err := rh.base.Mounts()
	if err != nil {
		return err
	}

	entries, err := fuse.FsTabList(rp.Config, mounts)
	if err != nil {
		return err
	}

	seg := call.Results.Segment()
	capEntries, err := capnp.NewFsTabEntry_List(seg, int32(len(entries)))
	if err != nil {
		return err
	}

	for idx, entry := range entries {
		capEntry, err := fsTabEntryToCap(entry, seg)
		if err != nil {
			return err
		}

		if err := capEntries.Set(idx, *capEntry); err != nil {
			return err
		}
	}

	return call.Results.SetMounts(capEntries)
}

func (rh *repoHandler) ConfigGet(call capnp.Repo_configGet) error {
	repo, err := rh.base.Repo()
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

func (rh *repoHandler) ConfigDoc(call capnp.Repo_configDoc) error {
	repo, err := rh.base.Repo()
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
	capPair, err := rh.configDefaultEntryToCapnp(seg, key)
	if err != nil {
		return err
	}

	return call.Results.SetDesc(*capPair)
}

func (rh *repoHandler) ConfigSet(call capnp.Repo_configSet) error {
	rp, err := rh.base.Repo()
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

func (rh *repoHandler) configDefaultEntryToCapnp(seg *capnplib.Segment, key string) (*capnp.ConfigEntry, error) {
	pair, err := capnp.NewConfigEntry(seg)
	if err != nil {
		return nil, err
	}

	if err := pair.SetKey(key); err != nil {
		return nil, err
	}

	repo, err := rh.base.Repo()
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

func (rh *repoHandler) ConfigAll(call capnp.Repo_configAll) error {
	repo, err := rh.base.Repo()
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
		capPair, err := rh.configDefaultEntryToCapnp(seg, key)
		if err != nil {
			return err
		}

		if err := capLst.Set(idx, *capPair); err != nil {
			return err
		}
	}

	return call.Results.SetAll(capLst)
}

func (rh *repoHandler) Become(call capnp.Repo_become) error {
	who, err := call.Params.Who()
	if err != nil {
		return err
	}

	rp, err := rh.base.Repo()
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

func (rh *repoHandler) Version(call capnp.Repo_version) error {
	rp, err := rh.base.Repo()
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
