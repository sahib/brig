// +build linux

package fuse

import (
	"fmt"
	"os"
	"sort"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/sahib/brig/util"
	"github.com/sahib/config"
)

func FsTabAdd(cfg *config.Config, name, path string, opts MountOptions) error {
	for _, key := range cfg.Keys() {
		if strings.HasSuffix(key, ".path") {
			if cfg.String(key) == path {
				return fmt.Errorf("mount `%s` at path `%s` already exists", name, path)
			}
		}
	}

	if cfg.String(name+".path") != "" {
		return fmt.Errorf("mount `%s` already exists", name)
	}

	if err := cfg.SetString(name+".path", path); err != nil {
		return err
	}

	if err := cfg.SetBool(name+".read_only", opts.ReadOnly); err != nil {
		return err
	}

	if opts.Root == "" {
		opts.Root = "/"
	}

	if err := cfg.SetString(name+".root", opts.Root); err != nil {
		return err
	}

	return nil
}

// FsTabRemove removes a mount. It does not directly unmount it,
// call FsTabApply for this.
func FsTabRemove(cfg *config.Config, name string) error {
	return cfg.Reset(name)
}

func FsTabUnmountAll(cfg *config.Config, mounts *MountTable) error {
	mounts.mu.Lock()
	defer mounts.mu.Unlock()

	errors := util.Errors{}
	for _, key := range cfg.Keys() {
		if strings.HasSuffix(key, ".path") {
			mountPath := cfg.String(key)
			log.Debugf("Unmount key %s %s", key, mountPath)
			if mountPath == "" {
				continue
			}

			if err := mounts.unmount(mountPath); err != nil {
				errors = append(errors, err)
			}
		}
	}

	return errors.ToErr()
}

// FsTabApply takes all configured mounts and makes sure that all of them are opened.
func FsTabApply(cfg *config.Config, mounts *MountTable) error {
	mounts.mu.Lock()
	defer mounts.mu.Unlock()

	mountPaths := make(map[string]*MountOptions)
	for _, key := range cfg.Keys() {
		if strings.HasSuffix(key, ".path") {
			mountPath := cfg.String(key)

			entry := &MountOptions{}
			mountPaths[mountPath] = entry

			readOnlyKey := key[:len(key)-len(".path")] + ".read_only"
			entry.ReadOnly = cfg.Bool(readOnlyKey)

			rootPathKey := key[:len(key)-len(".path")] + ".root"
			entry.Root = cfg.String(rootPathKey)
			if entry.Root == "" {
				entry.Root = "/"
			}
		}
	}

	errors := util.Errors{}
	for path, mount := range mounts.m {
		// Do not do anything when the path / options did not change.
		opts, isConfigured := mountPaths[path]
		if isConfigured && mount.EqualOptions(*opts) {
			delete(mountPaths, path)
			continue
		}

		if err := mounts.unmount(path); err != nil {
			errors = append(errors, err)
		}
	}

	for mountPath, options := range mountPaths {
		if err := os.MkdirAll(mountPath, 0700); err != nil {
			return err
		}

		if _, err := mounts.addMount(mountPath, *options); err != nil {
			errors = append(errors, err)
		}
	}

	return errors.ToErr()
}

type FsTabEntry struct {
	Name     string
	Path     string
	Root     string
	Active   bool
	ReadOnly bool
}

func FsTabList(cfg *config.Config, mounts *MountTable) ([]FsTabEntry, error) {
	mounts.mu.Lock()
	defer mounts.mu.Unlock()

	mountMap := make(map[string]*FsTabEntry)
	for _, key := range cfg.Keys() {
		split := strings.Split(key, ".")
		if len(split) < 3 || split[0] != "mounts" {
			continue
		}

		mountName := split[1]
		if _, ok := mountMap[mountName]; !ok {
			mountMap[mountName] = &FsTabEntry{}
		}

		switch split[2] {
		case "path":
			path := cfg.String(key)
			mountMap[mountName].Path = path

			_, isActive := mounts.m[path]
			mountMap[mountName].Active = isActive
		case "read_only":
			mountMap[mountName].ReadOnly = cfg.Bool(key)
		case "root":
			mountMap[mountName].Root = cfg.String(key)
		}
	}

	sortedMounts := []FsTabEntry{}
	for name, entry := range mountMap {
		entry.Name = name
		sortedMounts = append(sortedMounts, *entry)
	}

	sort.Slice(sortedMounts, func(i, j int) bool {
		return sortedMounts[i].Name < sortedMounts[j].Name
	})

	return sortedMounts, nil
}
