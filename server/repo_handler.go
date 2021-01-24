package server

import (
	"fmt"
	"strings"

	"github.com/sahib/brig/backend"
	"github.com/sahib/brig/fuse"
	gwdb "github.com/sahib/brig/gateway/db"
	gwcapnp "github.com/sahib/brig/gateway/db/capnp"
	"github.com/sahib/brig/repo/hints"
	"github.com/sahib/brig/server/capnp"
	"github.com/sahib/brig/version"
	log "github.com/sirupsen/logrus"
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

	_, err = rh.base.mounts.AddMount(mountPath, mountOptions)
	return err
}

func (rh *repoHandler) Unmount(call capnp.Repo_unmount) error {
	server.Ack(call.Options)

	mountPath, err := call.Params.MountPath()
	if err != nil {
		return err
	}

	return rh.base.mounts.Unmount(mountPath)
}

func capMountOptionsToMountOptions(capOpts capnp.MountOptions) (fuse.MountOptions, error) {
	readOnly := capOpts.ReadOnly()
	offline := capOpts.Offline()
	rootPath, err := capOpts.RootPath()
	if err != nil {
		return fuse.MountOptions{}, err
	}

	return fuse.MountOptions{
		ReadOnly: readOnly,
		Root:     rootPath,
		Offline:  offline,
	}, nil
}

func (rh *repoHandler) FstabAdd(call capnp.Repo_fstabAdd) error {
	server.Ack(call.Options)

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

	mountsCfg := rh.base.repo.Config.Section("mounts")
	return fuse.FsTabAdd(mountsCfg, mountName, mountPath, mountOptions)
}

func (rh *repoHandler) FstabRemove(call capnp.Repo_fstabRemove) error {
	server.Ack(call.Options)

	mountName, err := call.Params.MountName()
	if err != nil {
		return err
	}

	mountsCfg := rh.base.repo.Config.Section("mounts")
	if err := fuse.FsTabRemove(mountsCfg, mountName); err != nil {
		return err
	}

	return rh.base.repo.SaveConfig()
}

func (rh *repoHandler) FstabApply(call capnp.Repo_fstabApply) error {
	server.Ack(call.Options)

	mountsCfg := rh.base.repo.Config.Section("mounts")
	if err := fuse.FsTabApply(mountsCfg, rh.base.mounts); err != nil {
		return err
	}

	return rh.base.repo.SaveConfig()
}

func (rh *repoHandler) FstabUnmountAll(call capnp.Repo_fstabUnmountAll) error {
	server.Ack(call.Options)
	return fuse.FsTabUnmountAll(rh.base.repo.Config, rh.base.mounts)
}

func fsTabEntryToCap(entry fuse.FsTabEntry, seg *capnplib.Segment) (*capnp.FsTabEntry, error) {
	capEntry, err := capnp.NewFsTabEntry(seg)
	if err != nil {
		return nil, err
	}

	capEntry.SetReadOnly(entry.ReadOnly)
	capEntry.SetOffline(entry.Offline)
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

	rp := rh.base.repo
	mounts := rh.base.mounts
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
	key, err := call.Params.Key()
	if err != nil {
		return err
	}

	rp := rh.base.repo
	if !rp.Config.IsValidKey(key) {
		return fmt.Errorf("invalid key: %v", key)
	}

	value := rp.Config.Uncast(key)
	return call.Results.SetValue(value)
}

func (rh *repoHandler) ConfigDoc(call capnp.Repo_configDoc) error {
	key, err := call.Params.Key()
	if err != nil {
		return err
	}

	rp := rh.base.repo
	if !rp.Config.IsValidKey(key) {
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
	key, err := call.Params.Key()
	if err != nil {
		return err
	}

	rp := rh.base.repo
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

	rp := rh.base.repo
	clientVal := fmt.Sprintf("%v", rp.Config.Get(key))
	if err := pair.SetVal(clientVal); err != nil {
		return nil, err
	}

	entry := rp.Config.GetDefault(key)
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
	rp := rh.base.repo
	all := rp.Config.Keys()
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

	// We can only be users that are present in the remote list (and owner)
	// (This is not a technical limitation)
	rp := rh.base.repo
	if who != rp.Immutables.Owner() {
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
	server.Ack(call.Options)

	rp := rh.base.repo
	name := rp.Immutables.Backend()
	ipfsPath := rp.Config.String("daemon.ipfs_path")
	bkVersion := backend.Version(name, ipfsPath)
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

func (rh *repoHandler) GatewayUserAdd(call capnp.Repo_gatewayUserAdd) error {
	server.Ack(call.Options)
	name, err := call.Params.Name()
	if err != nil {
		return err
	}

	password, err := call.Params.Password()
	if err != nil {
		return err
	}

	folders := []string{}
	capFolders, err := call.Params.Folders()
	if err != nil {
		return err
	}

	for idx := 0; idx < capFolders.Len(); idx++ {
		folder, err := capFolders.At(idx)
		if err != nil {
			return err
		}

		if !strings.HasPrefix(folder, "/") {
			folder = "/" + folder
		}

		folders = append(folders, folder)
	}

	rights := []string{}
	capRights, err := call.Params.Rights()
	if err != nil {
		return err
	}

	for idx := 0; idx < capRights.Len(); idx++ {
		right, err := capRights.At(idx)
		if err != nil {
			return err
		}

		rights = append(rights, right)
	}

	gwDb := rh.base.gateway.UserDatabase()
	return gwDb.Add(name, password, folders, rights)
}

func (rh *repoHandler) GatewayUserRm(call capnp.Repo_gatewayUserRm) error {
	server.Ack(call.Options)

	name, err := call.Params.Name()
	if err != nil {
		return err
	}

	gwDb := rh.base.gateway.UserDatabase()
	return gwDb.Remove(name)
}

func (rh *repoHandler) GatewayUserList(call capnp.Repo_gatewayUserList) error {
	server.Ack(call.Options)

	gwDb := rh.base.gateway.UserDatabase()
	users, err := gwDb.List()
	if err != nil {
		return err
	}

	seg := call.Results.Segment()
	capUsers, err := gwcapnp.NewUser_List(seg, int32(len(users)))
	if err != nil {
		return err
	}

	for idx, user := range users {
		capUser, err := gwdb.UserToCapnp(&user, seg)
		if err != nil {
			return err
		}

		if err := capUsers.Set(idx, *capUser); err != nil {
			return err
		}
	}

	return call.Results.SetUsers(capUsers)
}

func (rh *repoHandler) DebugProfilePort(call capnp.Repo_debugProfilePort) error {
	server.Ack(call.Options)
	call.Results.SetPort(int32(rh.base.pprofPort))
	return nil
}

func (rh *repoHandler) HintSet(call capnp.Repo_hintSet) error {
	server.Ack(call.Options)

	capHint, err := call.Params.Hint()
	if err != nil {
		return err
	}

	compressionAlgo, err := capHint.CompressionAlgo()
	if err != nil {
		return err
	}

	encryptionAlgo, err := capHint.EncryptionAlgo()
	if err != nil {
		return err
	}

	path, err := capHint.Path()
	if err != nil {
		return err
	}

	if err := rh.base.repo.Hints.Set(path, hints.Hint{
		CompressionAlgo: hints.CompressionHint(compressionAlgo),
		EncryptionAlgo:  hints.EncryptionHint(encryptionAlgo),
	}); err != nil {
		return err
	}

	// Make sure the hints are immediately written to disk.
	// At time of writing this is the place where hints are changed.
	return rh.base.repo.SaveHints()
}

func (rh *repoHandler) HintRemove(call capnp.Repo_hintRemove) error {
	server.Ack(call.Options)

	path, err := call.Params.Path()
	if err != nil {
		return err
	}

	return rh.base.repo.Hints.Remove(path)
}

func (rh *repoHandler) HintList(call capnp.Repo_hintList) error {
	server.Ack(call.Options)

	hints := rh.base.repo.Hints.List()

	seg := call.Results.Segment()
	capnpHints, err := capnp.NewHint_List(seg, int32(len(hints)))
	if err != nil {
		return err
	}

	capIdx := 0

	for path, hint := range hints {
		capHint, err := capnp.NewHint(seg)
		if err != nil {
			return err
		}

		if err := capHint.SetPath(path); err != nil {
			return err
		}

		if err := capHint.SetCompressionAlgo(string(hint.CompressionAlgo)); err != nil {
			return err
		}

		if err := capHint.SetEncryptionAlgo(string(hint.EncryptionAlgo)); err != nil {
			return err
		}

		if err := capnpHints.Set(capIdx, capHint); err != nil {
			return err
		}

		capIdx++
	}

	return call.Results.SetHints(capnpHints)
}
