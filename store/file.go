package store

import (
	"crypto/rand"
	"fmt"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
	"github.com/disorganizer/brig/store/proto"
	"github.com/disorganizer/brig/util/ipfsutil"
	"github.com/disorganizer/brig/util/trie"
	protobuf "github.com/gogo/protobuf/proto"
	"github.com/jbenet/go-multihash"
)

// Metadata captures metadata that might be changed by the user.
type Metadata struct {
	// Size is the file size in bytes.
	size int64
	// ModTime is the time when the file or it's metadata was last changed.
	modTime time.Time
	hash    *Hash
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

	isFile bool

	Key []byte
}

func (f *File) insert(root *File, path string) {
	f.node = root.node.InsertWithData(path, f)
}

// Sync writes an up-to-date version of the file metadata to bolt.
// You probably do not need to call that yourself.
func (f *File) Sync() {
	f.Lock()
	defer f.Unlock()

	f.sync()
}

// UpdateSize updates the size (and therefore also the ModTime) of the file.
// The change is written to bolt.
func (f *File) UpdateSize(size int64) {
	f.Lock()
	defer f.Unlock()

	f.size = size
	f.modTime = time.Now()
	f.sync()
}

// Size returns the current size in a threadsafe manner.
func (f *File) Size() int64 {
	f.RLock()
	defer f.RUnlock()

	return int64(f.size)
}

// ModTime returns the current mtime in a threadsafe manner.
func (f *File) ModTime() time.Time {
	f.RLock()
	defer f.RUnlock()

	return f.modTime
}

// UpdateModTime safely updates the ModTime field of the file.
// The change is written to bolt.
func (f *File) UpdateModTime(modTime time.Time) {
	f.Lock()
	defer f.Unlock()

	f.modTime = modTime
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

		parentDir.size += f.size
		parentDir.modTime = f.modTime
	})

	path := f.node.Path()
	log.Debugf("store-sync: %s", path)

	f.store.db.Update(withBucket("index", func(tx *bolt.Tx, bucket *bolt.Bucket) error {
		data, err := f.marshal()
		if err != nil {
			return err
		}

		if err := bucket.Put([]byte(path), data); err != nil {
			return err
		}

		return nil
	}))

}

// NewFile returns a file inside a repo.
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
		isFile:   true,
	}

	store.Root.Lock()
	defer store.Root.Unlock()

	file.insert(store.Root, path)
	return file, nil
}

// NewDir returns a new empty directory File.
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
			modTime: time.Now(),
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

// Marshal converts a file to a protobuf-byte representation.
func (f *File) Marshal() ([]byte, error) {
	f.RLock()
	defer f.RUnlock()

	return f.marshal()
}

func (f *File) marshal() ([]byte, error) {
	modTimeStamp, err := f.modTime.MarshalText()
	if err != nil {
		return nil, err
	}

	dataFile := &proto.File{
		Path:     protobuf.String(f.node.Path()),
		Key:      f.Key,
		FileSize: protobuf.Int64(f.size),
		ModTime:  modTimeStamp,
		IsFile:   protobuf.Bool(f.isFile),
		Hash:     f.hashUnlocked().Multihash,
	}

	data, err := protobuf.Marshal(dataFile)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// Unmarshal decodes the data in `buf` and inserts the unmarshaled file
// into `store`.
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
		isFile:  dataFile.GetIsFile(),
		Key:     dataFile.GetKey(),
		Metadata: &Metadata{
			size:    dataFile.GetFileSize(),
			modTime: *modTimeStamp,
			hash:    &Hash{dataFile.GetHash()},
		},
	}

	file.Lock()
	path := dataFile.GetPath()
	file.insert(store.Root, path)
	file.Unlock()

	return file, nil
}

///////////////////
// TRIE LIKE API //
///////////////////

// Root returns the uppermost node reachable from the receiver.
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

	// Remove from trie:
	f.node.Remove()
}

// Len returns the current number of elements in the trie.
// This counts only explicitly inserted Nodes.
func (f *File) Len() int64 {
	f.RLock()
	defer f.RUnlock()

	return f.node.Len()
}

// Up goes up in the hierarchy and calls `visit` on each visited node.
func (f *File) Up(visit func(*File)) {
	f.RLock()
	defer f.RUnlock()

	f.node.Up(func(parent *trie.Node) {
		file := parent.Data.(*File)
		visit(file)
	})
}

// IsLeaf returns true if the file is a leaf node.
// TODO: needed?
func (f *File) IsLeaf() bool {
	f.RLock()
	defer f.RUnlock()

	return f.isFile
}

// Path returns the absolute path of the file inside the repository, starting with /.
func (f *File) Path() string {
	f.RLock()
	defer f.RUnlock()

	return f.node.Path()
}

func (f *File) path() string {
	return f.node.Path()
}

// Walk recursively calls `visit` on each child and f itself.
// If `dfs` is true, the order will be depth-first, otherwise breadth-first.
func (f *File) Walk(dfs bool, visit func(*File)) {
	f.RLock()
	defer f.RUnlock()

	f.node.Walk(dfs, func(n *trie.Node) {
		visit(n.Data.(*File))
	})
}

var emptyChildren []*File

// Children returns a list of children of the
func (f *File) Children() []*File {
	f.RLock()
	defer f.RUnlock()

	// Optimisation: Return the same empty slice for leaf nodes.
	n := len(f.node.Children)
	if n == 0 {
		return emptyChildren
	}

	children := make([]*File, 0, n)
	for _, child := range f.node.Children {
		if child.Data != nil {
			children = append(children, child.Data.(*File))
		}
	}

	return children
}

// Child returns the direct child of the receiver called `name` or nil
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

// Name returns the basename of the file.
func (f *File) Name() string {
	f.RLock()
	defer f.RUnlock()

	return f.node.Name
}

// Stream opens a reader that yields the raw data of the file,
// already transparently decompressed and decrypted.
func (f *File) Stream() (ipfsutil.Reader, error) {
	f.RLock()
	defer f.RUnlock()

	log.Debugf("Stream `%s` (hash: %s) (key: %x)", f.node.Path(), f.hash.B58String(), f.Key)

	ipfsStream, err := ipfsutil.Cat(f.store.IpfsNode, f.hash.Multihash)
	if err != nil {
		return nil, err
	}

	return NewIpfsReader(f.Key, ipfsStream)
}

// Parent returns the parent directory of File.
// If `f` is already the root, it will return itself (and never nil).
func (f *File) Parent() *File {
	f.RLock()
	defer f.RUnlock()

	parent := f.node.Parent
	if parent != nil {
		return parent.Data.(*File)
	}

	return f
}

// Hash returns the hash of a file. If it is leaf file,
// the hash is returned directly; directory hashes
// are computed by combining the child hashes.
func (f *File) Hash() *Hash {
	f.RLock()
	defer f.RUnlock()

	return f.hashUnlocked()
}

func (f *File) hashUnlocked() *Hash {
	if f.isFile {
		if !f.hash.Valid() {
			log.Warningf("file-hash: BUG: File with no hash: %v", f.node.Path())
		}

		return f.hash
	}

	if f.hash.Valid() {
		// Directory with pre-computed hash:
		return f.hash
	}

	// Compute hash by XOR'ing all child hashes:
	// (we need XOR because order must be irrelevant)
	// TODO: Get actual hash algorithm from config (or something)
	hash := make([]byte, multihash.DefaultLengths[multihash.SHA1])
	for _, childNode := range f.node.Children {
		child := childNode.Data.(*File)
		if !child.hash.Valid() {
			// Force computation:
			child.Hash()
		}

		digest, err := multihash.Decode(child.hash.Multihash)
		if err != nil {
			log.Warningf("file-hash: Invalid cksum: %v: %v", child.hash, err)
			log.Warningf("file-hash: Resulting hashsum might be incorrect.")
			continue
		}

		if len(digest.Digest) != len(hash) {
			log.Warningf("file-hash: different cksum lengths: %d <->", len(digest.Digest), len(hash))
			continue
		}

		for i := 0; i < len(hash); i++ {
			hash[i] ^= digest.Digest[i]
		}
	}

	mhash, err := multihash.Encode(hash, multihash.SHA1)
	if err != nil {
		log.Errorf("Unable to decode `%v` as multihash: %v", hash, err)
	}

	f.hash = &Hash{mhash}
	return f.hash
}
