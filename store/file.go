package store

import (
	"fmt"
	"path"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/store/wire"
	"github.com/disorganizer/brig/util/ipfsutil"
)

// File represents a single file in the repository.
// It stores all metadata about it and links to the actual data.
type File struct {
	name    string
	key     []byte
	hash    *Hash
	parent  string
	size    uint64
	modTime time.Time
	id      uint64

	fs *FS
}

func newEmptyFile(fs *FS, parent *Directory, name string) (*File, error) {
	id, err := fs.NextID()
	if err != nil {
		return nil, err
	}

	file := &File{
		name:    name,
		id:      id,
		modTime: time.Now(),
		fs:      fs,
		parent:  parent.Path(),
	}

	return file, nil
}

func (f *File) ToProto() (*wire.Node, error) {
	binModTime, err := f.modTime.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return &wire.Node{
		ID:       f.id,
		Type:     wire.NodeType_FILE,
		Name:     f.name,
		NodeSize: f.size,
		ModTime:  binModTime,
		Hash:     f.hash.Bytes(),
		File: &wire.File{
			Parent: f.parent,
			Key:    f.key,
		},
	}, nil
}

func (f *File) FromProto(pnd *wire.Node) error {
	pfi := pnd.File
	if pfi == nil {
		return fmt.Errorf("File attribute is empty. This is likely not a real file.")
	}

	modTime := time.Time{}
	if err := modTime.UnmarshalBinary(pnd.ModTime); err != nil {
		return err
	}

	f.id = pnd.ID
	f.size = pnd.NodeSize
	f.modTime = modTime
	f.hash = &Hash{pnd.Hash}
	f.parent = pfi.Parent
	f.name = pnd.Name
	f.key = pfi.Key
	return nil
}

////////////////// METADATA INTERFACE //////////////////

// Name returns the basename of the file.
func (f *File) ID() uint64         { return f.id }
func (f *File) Name() string       { return f.name }
func (f *File) Hash() *Hash        { return f.hash }
func (f *File) Size() uint64       { return f.size }
func (f *File) ModTime() time.Time { return f.modTime }

////////////////// ATTRIBUTE SETTERS //////////////////

func (f *File) SetModTime(t time.Time) { f.modTime = t }
func (f *File) SetName(n string)       { f.name = n }
func (f *File) SetKey(k []byte)        { f.key = k }
func (f *File) SetSize(s uint64) {
	f.size = s
	f.SetModTime(time.Now())
}

func (f *File) SetHash(h *Hash) {
	oldHash := f.hash
	f.hash = h
	f.fs.MemIndexSwap(f, oldHash)
}

func (f *File) Path() string {
	return prefixSlash(path.Join(f.parent, f.name))
}

////////////////// HIERARCHY INTERFACE //////////////////

// NChildren returns the number of children this file node has.
func (f *File) NChildren() int {
	return 0
}

func (f *File) Child(name string) (Node, error) {
	// A file never has a child. Sad but true.
	return nil, nil
}

// Parent returns the parent directory of File.
// If `f` is already the root, it will return itself (and never nil).
func (f *File) Parent() (Node, error) {
	return f.fs.LookupNode(f.parent)
}

func (f *File) SetParent(parent Node) error {
	if parent == nil {
		return nil
	}

	f.parent = parent.Path()
	return nil
}

func (f *File) GetType() NodeType {
	return NodeTypeFile
}

////////////////// SPECIAL METHODS //////////////////

// Stream opens a reader that yields the raw data of the file,
// already transparently decompressed and decrypted.
func (f *File) Stream(ipfs *ipfsutil.Node) (ipfsutil.Reader, error) {
	log.Debugf(
		"Stream `%s` (hash: %s) (key: %x)",
		f.Path(),
		f.hash.B58String(),
		f.key,
	)

	ipfsStream, err := ipfsutil.Cat(ipfs, f.hash.Multihash)
	if err != nil {
		return nil, err
	}

	return NewIpfsReader(f.key, ipfsStream)
}

func (f *File) Key() []byte {
	return f.key
}
