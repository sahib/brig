package fuse

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/store"
)

// This is very similar (and indeed mostly copied) code from:
// https://github.com/bazil/fuse/blob/master/fs/fstestutil/mounted.go
// Since that's "only" test module, api might change, so better have this
// code here (also we might do a few things differently).

type Mount struct {
	Dir   string
	FS    *FS
	Store *store.Store

	closed bool
	done   chan struct{}
	errors chan error

	Conn   *fuse.Conn
	Server *fs.Server
}

func NewMount(store *store.Store, mountpoint string) (*Mount, error) {
	conn, err := fuse.Mount(mountpoint)
	if err != nil {
		return nil, err
	}

	filesys := &FS{Store: store}

	mnt := &Mount{
		Conn:   conn,
		Server: fs.New(conn, nil),
		FS:     filesys,
		Dir:    mountpoint,
		Store:  store,
		done:   make(chan struct{}),
		errors: make(chan error),
	}

	go func() {
		defer close(mnt.done)
		log.Debugf("Serving FUSE at %v", mountpoint)
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

	for tries := 0; tries < 1000; tries++ {
		if err := fuse.Unmount(m.Dir); err != nil {
			log.Printf("unmount error: %v", err)
			time.Sleep(10 * time.Millisecond)
			continue
		}

		break
	}

	if err := m.Conn.Close(); err != nil {
		return err
	}
	log.Info("closing con")

	// Wait for serve to return:
	log.Info("Waitin for done..")
	<-m.done
	log.Info("Waitin for done.. done done")

	return nil
}

type MountTable struct {
	sync.Mutex
	m     map[string]*Mount
	Store *store.Store
}

func NewMountTable(store *store.Store) *MountTable {
	return &MountTable{
		m:     make(map[string]*Mount),
		Store: store,
	}
}

func (t *MountTable) AddMount(path string) (*Mount, error) {
	t.Lock()
	defer t.Unlock()

	m, ok := t.m[path]
	if ok {
		return m, nil
	}

	m, err := NewMount(t.Store, path)
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
