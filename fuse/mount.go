package fuse

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	log "github.com/Sirupsen/logrus"
	"github.com/sahib/brig/catfs"
	"github.com/sahib/brig/util"
)

// This is very similar (and indeed mostly copied) code from:
// https://github.com/bazil/fuse/blob/master/fs/fstestutil/mounted.go
// Since that's "only" test module, api might change, so better have this
// code here (also we might do a few things differently).

// Mount represents a fuse endpoint on the filesystem.
// It is used as top-level API to control a brigfs fuse mount.
type Mount struct {
	Dir string

	filesys *Filesystem
	closed  bool
	done    chan util.Empty
	errors  chan error
	conn    *fuse.Conn
	server  *fs.Server
}

// NewMount mounts a fuse endpoint at `mountpoint` retrieving data from `store`.
func NewMount(cfs *catfs.FS, mountpoint string) (*Mount, error) {
	conn, err := fuse.Mount(
		mountpoint,
		fuse.FSName("brigfs"),
		fuse.Subtype("brig"),
		fuse.AllowNonEmptyMount(),
	)

	if err != nil {
		return nil, err
	}

	filesys := &Filesystem{cfs: cfs}
	mnt := &Mount{
		conn:    conn,
		server:  fs.New(conn, nil),
		filesys: filesys,
		Dir:     mountpoint,
		done:    make(chan util.Empty),
		errors:  make(chan error),
	}

	go func() {
		defer close(mnt.done)
		log.Debugf("Serving FUSE at %v", mountpoint)
		mnt.errors <- mnt.server.Serve(filesys)
		mnt.done <- util.Empty{}
		log.Debugf("Stopped serving FUSE at %v", mountpoint)
	}()

	select {
	case <-mnt.conn.Ready:
		if err := mnt.conn.MountError; err != nil {
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

	log.Info("Unmounting fuse layer...")

	for tries := 0; tries < 20; tries++ {
		if err := fuse.Unmount(m.Dir); err != nil {
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

	if err := m.conn.Close(); err != nil {
		return err
	}

	return nil
}

// MountTable is a mapping from the mountpoint to the respective
// `Mount` struct. It's given as convenient way to maintain several mounts.
// All operations on the table are safe to call from several goroutines.
type MountTable struct {
	mu sync.Mutex
	m  map[string]*Mount
	fs *catfs.FS
}

// NewMountTable returns an empty mount table.
func NewMountTable(fs *catfs.FS) *MountTable {
	return &MountTable{
		m:  make(map[string]*Mount),
		fs: fs,
	}
}

// AddMount calls NewMount and adds it to the table at `path`.
func (t *MountTable) AddMount(path string) (*Mount, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	m, ok := t.m[path]
	if ok {
		return m, nil
	}

	m, err := NewMount(t.fs, path)
	if err == nil {
		t.m[path] = m
	}

	return m, err
}

// Unmount closes the mount at `path` and deletes it from the table.
func (t *MountTable) Unmount(path string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	m, ok := t.m[path]
	if !ok {
		return fmt.Errorf("No mount at `%v`.", path)
	}

	delete(t.m, path)
	return m.Close()
}

// Close unmounts all leftover mounts and clears the table.
func (t *MountTable) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	var err error

	for _, mount := range t.m {
		if closeErr := mount.Close(); closeErr != nil {
			err = closeErr
		}
	}

	t.m = make(map[string]*Mount)
	return err
}
