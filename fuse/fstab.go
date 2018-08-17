package fuse

import (
	"fmt"
	"os"
	"strings"

	"github.com/sahib/brig/util"
	"github.com/sahib/config"
)

func FsTabAdd(cfg *config.Config, name, path string, readOnly bool) error {
	for _, key := range cfg.Keys() {
		if strings.HasSuffix(key, ".path") {
			fmt.Println("KEY", key)
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

	if err := cfg.SetBool(name+".read_only", readOnly); err != nil {
		return err
	}

	return nil
}

func FsTabRemove(cfg *config.Config, name string) error {
	return cfg.Reset(name)
}

func FsTabApply(cfg *config.Config, mounts *MountTable) error {
	mounts.mu.Lock()
	defer mounts.mu.Unlock()

	mountPaths := make(map[string]bool)
	for _, key := range cfg.Keys() {
		if strings.HasSuffix(key, ".path") {
			readOnlyKey := key[:len(key)-len(".path")] + ".read_only"
			mountPaths[cfg.String(key)] = cfg.Bool(readOnlyKey)
		}
	}

	errors := util.Errors{}
	for path := range mounts.m {
		_, isConfigured := mountPaths[path]

		if isConfigured {
			// TODO: Check if the options are the same.
			delete(mountPaths, path)
			continue
		}

		if err := mounts.unmount(path); err != nil {
			errors = append(errors, err)
		}
	}

	// mounts that were not actively used, but are configured:
	for mountPath := range mountPaths {
		if err := os.MkdirAll(mountPath, 0700); err != nil {
			return err
		}

		if _, err := mounts.addMount(mountPath); err != nil {
			errors = append(errors, err)
		}
	}

	return errors.ToErr()
}
