package store

import (
	"crypto/rand"
	"fmt"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/store/proto"
	"github.com/disorganizer/brig/util/ipfsutil"
	"github.com/disorganizer/brig/util/trie"
	protobuf "github.com/gogo/protobuf/proto"
	"github.com/jbenet/go-multihash"
)

type Metadata struct {
	Size    FileSize
	ModTime time.Time
}

// File represents a single file in the repository.
// It stores all metadata about it and links to the actual data.
type File struct {
	*Metadata

	// Mutex protecting access to the trie.
	// Note that only one mutex exists per trie.
	*sync.RWMutex

	store *Store
	node  *trie.Node

	IsFile bool

	Hash multihash.Multihash
	Key  []byte
}

func (f *File) insert(root *File, path string) {
	f.node = root.node.InsertWithData(path, f)
}

func (f *File) Sync() {
	f.Lock()
	defer f.Unlock()

	f.sync()
}

func (f *File) UpdateSize(size uint64) {
	f.Lock()
	defer f.Unlock()

	f.Size = FileSize(size)
	f.ModTime = time.Now()
	f.sync()
}

func (f *File) UpdateModTime(modTime time.Time) {
	f.Lock()
	defer f.Unlock()

	f.ModTime = modTime
	f.sync()
}

func (f *File) sync() {
	// TODO: Save to bolt.
	// Create intermediate directories on the way up,
	// also fix size and mtime accordingly.
	f.node.Up(func(parent *trie.Node) {
		if parent.Data == f {
			return
		}

		var parentDir *File
		if parent.Data == nil {
			newDir := &File{
				store:    f.store,
				RWMutex:  f.store.Root.RWMutex,
				Metadata: &Metadata{},
			}

			parentDir = newDir
		} else {
			parentDir = parent.Data.(*File)
		}

		parentDir.Size += f.Size
		parentDir.ModTime = f.ModTime
	})
}

// New returns a file inside a repo.
// Path is relative to the repo root.
func NewFile(store *Store, path string) (*File, error) {
	key := make([]byte, 32)
	n, err := rand.Reader.Read(key)
	if err != nil {
		return nil, err
	}

	if n != 32 {
		return nil, fmt.Errorf("Read less than desired key size: %v", n)
	}

	file := &File{
		store:    store,
		RWMutex:  store.Root.RWMutex,
		Metadata: &Metadata{},
		Key:      key,
		IsFile:   true,
	}

	store.Root.Lock()
	defer store.Root.Unlock()

	file.insert(store.Root, path)
	return file, nil
}

func NewDir(store *Store, path string) (*File, error) {
	store.Root.Lock()
	defer store.Root.Unlock()

	return newDirUnlocked(store, path)
}

func newDirUnlocked(store *Store, path string) (*File, error) {
	var mu *sync.RWMutex
	if store.Root == nil {
		// We're probably just called to create store.Root.
		mu = &sync.RWMutex{}
	} else {
		mu = store.Root.RWMutex
	}

	dir := &File{
		store:   store,
		RWMutex: mu,
		Metadata: &Metadata{
			ModTime: time.Now(),
		},
	}

	var root *File
	if store.Root == nil {
		root = dir
	} else {
		root = store.Root
	}

	dir.insert(root, path)
	return dir, nil
}

func (f *File) Marshal() ([]byte, error) {
	f.RLock()
	defer f.RUnlock()

	modTimeStamp, err := f.ModTime.MarshalText()
	if err != nil {
		return nil, err
	}

	dataFile := &proto.File{
		Path:     protobuf.String(f.node.Path()),
		Key:      f.Key,
		FileSize: protobuf.Int64(int64(f.Size)),
		ModTime:  modTimeStamp,
		IsFile:   protobuf.Bool(f.IsFile),
		Hash:     f.Hash,
	}

	data, err := protobuf.Marshal(dataFile)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func Unmarshal(store *Store, buf []byte) (*File, error) {
	dataFile := &proto.File{}
	if err := protobuf.Unmarshal(buf, dataFile); err != nil {
		return nil, err
	}

	modTimeStamp := &time.Time{}
	if err := modTimeStamp.UnmarshalText(dataFile.GetModTime()); err != nil {
		return nil, err
	}

	file := &File{
		store:   store,
		RWMutex: store.Root.RWMutex,
		IsFile:  dataFile.GetIsFile(),
		Hash:    dataFile.GetHash(),
		Key:     dataFile.GetKey(),
		Metadata: &Metadata{
			Size:    FileSize(dataFile.GetFileSize()),
			ModTime: *modTimeStamp,
		},
	}

	file.Lock()
	defer file.Unlock()

	path := dataFile.GetPath()
	file.insert(store.Root, path)
	file.sync()

	return file, nil
}

///////////////////
// TRIE LIKE API //
///////////////////

// The created file is empty and has a size of 0.
// TODO: That's some ugly API
func (f *File) Insert(path string, isFile bool) (*File, error) {
	child := &File{
		store:   f.store,
		IsFile:  isFile,
		RWMutex: f.RWMutex,
		Metadata: &Metadata{
			Size:    0,
			ModTime: time.Now(),
		},
	}

	f.Lock()
	defer f.Unlock()

	child.insert(f, path)
	child.sync()
	return child, nil
}

func (f *File) Root() *File {
	f.RLock()
	defer f.RUnlock()

	return f.store.Root
}

// Lookup searches for a node references by a path.
func (f *File) Lookup(path string) *File {
	f.RLock()
	defer f.RUnlock()

	node := f.node.Lookup(path)
	if node != nil {
		return node.Data.(*File)
	}

	return nil
}

// Remove removes the node at path and all of it's children.
// The parent of the removed node is returned, which might be nil.
func (f *File) Remove() {
	f.Lock()
	defer f.Unlock()

	f.node.Remove()

	// TODO: remove from bolt
}

// Len returns the current number of elements in the trie.
// This counts only explicitly inserted Nodes.
func (f *File) Len() int64 {
	f.RLock()
	defer f.RUnlock()

	return f.node.Len()
}

// // Up goes up in the hierarchy and calls `visit` on each visited node.
func (f *File) Up(visit func(*File)) {
	f.RLock()
	defer f.RUnlock()

	f.node.Up(func(parent *trie.Node) {
		file := parent.Data.(*File)
		visit(file)
	})
}

func (f *File) IsLeaf() bool {
	f.RLock()
	defer f.RUnlock()

	return f.node.IsLeaf()
}

func (f *File) Path() string {
	f.RLock()
	defer f.RUnlock()

	return f.node.Path()
}

func (f *File) Walk(dfs bool, visit func(*File)) {
	f.RLock()
	defer f.RUnlock()

	f.node.Walk(dfs, func(n *trie.Node) {
		visit(n.Data.(*File))
	})
}

func (f *File) Children() []*File {
	f.RLock()
	defer f.RUnlock()

	children := make([]*File, 0, len(f.node.Children))
	for _, child := range f.node.Children {
		if child.Data != nil {
			children = append(children, child.Data.(*File))
		}
	}

	return children
}

func (f *File) Child(name string) *File {
	f.RLock()
	defer f.RUnlock()

	if f.node.Children == nil {
		return nil
	}

	child, ok := f.node.Children[name]
	if ok {
		return child.Data.(*File)
	}

	return nil
}

func (f *File) Name() string {
	f.RLock()
	defer f.RUnlock()

	return f.node.Name
}

func (f *File) Stream() (ipfsutil.Reader, error) {
	f.RLock()
	defer f.RUnlock()

	log.Debugf("Stream `%s` (hash: %s) (key: %x)", f.node.Path(), f.Hash.B58String(), f.Key)

	ipfsStream, err := ipfsutil.Cat(f.store.IpfsNode, f.Hash)
	if err != nil {
		return nil, err
	}

	return NewIpfsReader(f.Key, ipfsStream)
}
