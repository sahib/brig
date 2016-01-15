package fuse

import (
	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/disorganizer/brig/util/trie"
)

func mount(mountpoint string) error {
	c, err := fuse.Mount(mountpoint)
	if err != nil {
		return err
	}
	defer c.Close()

	filesys := &FS{}
	if err := fs.Serve(c, filesys); err != nil {
		return err
	}

	// check if the mount process has an error to report
	<-c.Ready
	if err := c.MountError; err != nil {
		return err
	}

	return nil
}

func Mount(mountpoint string) error {
	trie := trie.NewTrie()
	trie.Insert("/home/sahib/test")

	return mount(mountpoint)
}
