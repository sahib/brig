package fuse

import (
	"errors"
	"fmt"
	"sync"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/util/trie"
)

// This is very similar (and indeed mostly copied) code from:
// https://github.com/bazil/fuse/blob/master/fs/fstestutil/mounted.go
// Since that's "only" test module, api might change, so better have this
// code here (also we might do a few things differently).

type Mount struct {
	Dir string
	FS  *FS

	closed bool
	done   chan struct{}
	errors chan error

	Conn   *fuse.Conn
	Server *fs.Server
}

func NewMount(mountpoint string) (*Mount, error) {
	conn, err := fuse.Mount(mountpoint)
	if err != nil {
		return nil, err
	}

	trie := trie.NewTrie()
	trie.Insert("/home/sahib/test") // TODO
	filesys := &FS{Trie: trie}

	mnt := &Mount{
		Conn:   conn,
		Server: fs.New(conn, nil),
		FS:     filesys,
		Dir:    mountpoint,
		done:   make(chan struct{}),
		errors: make(chan error),
	}

	go func() {
		defer close(mnt.done)
		log.Debug("Serving FUSE at %v", mountpoint)
		mnt.errors <- mnt.Server.Serve(filesys)
		log.Debug("Stopped serving FUSE at %v", mountpoint)
	}()

	select {
	case <-mnt.Conn.Ready:
		if err := mnt.Conn.MountError; err != nil {
			return nil, err
		}
	case err = <-mnt.errors:
		// Serve quit early
		if err != nil {
			return nil, err
		}
		return nil, errors.New("Serve exited early")
	}

	return mnt, nil
}

func (m *Mount) Close() error {
	if m.closed {
		return nil
	}
	m.closed = true

	log.Info("Fuse umount")

	if err := fuse.Unmount(m.Dir); err != nil {
		return err
	}

	// Wait for serve to return:
	log.Info("Waitin for done..")
	<-m.done
	log.Info("Waitin for done.. done done")

	if err := m.Conn.Close(); err != nil {
		return err
	}
	log.Info("closing cone")

	return nil
}

type MountTable struct {
	sync.Mutex
	m map[string]*Mount
}

func NewMountTable() *MountTable {
	return &MountTable{
		m: make(map[string]*Mount),
	}
}

func (t *MountTable) AddMount(path string) (*Mount, error) {
	t.Lock()
	defer t.Unlock()

	m, ok := t.m[path]
	if ok {
		return m, nil
	}

	m, err := NewMount(path)
	if err == nil {
		t.m[path] = m
	}

	return m, err
}

func (t *MountTable) Unmount(path string) error {
	t.Lock()
	defer t.Unlock()

	m, ok := t.m[path]
	if !ok {
		return fmt.Errorf("No mount at `%v`.", path)
	}

	delete(t.m, path)
	return m.Close()
}

func (t *MountTable) Close() error {
	t.Lock()
	defer t.Unlock()

	var err error

	for _, mount := range t.m {
		if closeErr := mount.Close(); closeErr != nil {
			err = closeErr
		}
	}

	return err
}
