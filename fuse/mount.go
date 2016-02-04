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
	"github.com/disorganizer/brig/util"
)

// This is very similar (and indeed mostly copied) code from:
// https://github.com/bazil/fuse/blob/master/fs/fstestutil/mounted.go
// Since that's "only" test module, api might change, so better have this
// code here (also we might do a few things differently).

// Mount represents a fuse endpoint on the filesystem.
// It is used as top-level API to control a brigfs fuse mount.
type Mount struct {
	Dir   string
	FS    *FS
	Store *store.Store

	closed bool
	done   chan util.Empty
	errors chan error

	Conn   *fuse.Conn
	Server *fs.Server
}

// NewMount mounts a fuse endpoint at `mountpoint` retrieving data from `store`.
func NewMount(store *store.Store, mountpoint string) (*Mount, error) {
	conn, err := fuse.Mount(
		mountpoint,
		fuse.FSName("brigfs"),
		fuse.Subtype("brig"),
	)

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
		done:   make(chan util.Empty),
		errors: make(chan error),
	}

	go func() {
		defer close(mnt.done)
		log.Debugf("Serving FUSE at %v", mountpoint)
		mnt.errors <- mnt.Server.Serve(filesys)
		mnt.done <- util.Empty{}
		log.Debugf("Stopped serving FUSE at %v", mountpoint)
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

// Close will wait until all I/O operations are done and unmount the fuse
// mount again.
func (m *Mount) Close() error {
	if m.closed {
		return nil
	}
	m.closed = true

	log.Info("Umount fuse layer...")

	for tries := 0; tries < 20; tries++ {
		if err := fuse.Unmount(m.Dir); err != nil {
			// log.Printf("unmount error: %v", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}

		break
	}

	// Be sure to drain the error channel:
	select {
	case err := <-m.errors:
		// Serve() had some error after some time:
		if err != nil {
			log.Warningf("fuse returned an error: %v", err)
		}
	}

	// Be sure to pull the item from the channel:
	<-m.done

	if err := m.Conn.Close(); err != nil {
		return err
	}
	return nil
}

// MountTable is a mapping from the mountpoint to the respective
// `Mount` struct. It's given as convenient way to maintain several mounts.
// All operations on the table are safe to call from several goroutines.
type MountTable struct {
	sync.Mutex
	m     map[string]*Mount
	Store *store.Store
}

// NewMountTable returns an empty mount table.
func NewMountTable(store *store.Store) *MountTable {
	return &MountTable{
		m:     make(map[string]*Mount),
		Store: store,
	}
}

// AddMount calls NewMount and adds it to the table at `path`.
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

// Unmount closes the mount at `path` and deletes it from the table.
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

// Close unmounts all leftover mounts and clears the table.
func (t *MountTable) Close() error {
	t.Lock()
	defer t.Unlock()

	var err error

	for _, mount := range t.m {
		if closeErr := mount.Close(); closeErr != nil {
			err = closeErr
		}
	}

	t.m = make(map[string]*Mount)
	return err
}
