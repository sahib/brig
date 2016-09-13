package store

import (
	"crypto/rand"
	"fmt"
	"io"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/store/wire"
	"github.com/disorganizer/brig/util/ipfsutil"
	"github.com/gogo/protobuf/proto"
)

// File represents a single file in the repository.
// It stores all metadata about it and links to the actual data.
type File struct {
	name    string
	key     []byte
	hash    *Hash
	parent  *Hash
	size    uint64
	modTime time.Time
	id      uint64

	fs *FS
}

func newEmptyFile(fs *FS, name string) (*File, error) {
	id, err := fs.NextID()
	if err != nil {
		return nil, err
	}

	// TODO: Make this configurable?
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}

	return &File{
		name:    name,
		id:      id,
		modTime: time.Now(),
		fs:      fs,
	}, nil
}

func (f *File) ToProto() (*wire.Node, error) {
	binModTime, err := f.modTime.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return &wire.Node{
		ID:   proto.Uint64(f.id),
		Type: wire.NodeType_FILE.Enum(),
		File: &wire.File{
			Name:     proto.String(f.name),
			Key:      f.key,
			FileSize: proto.Uint64(f.size),
			ModTime:  binModTime,
			Hash:     f.hash.Bytes(),
			Parent:   f.parent.Bytes(),
		},
	}, nil
}

func (f *File) FromProto(pnd *wire.Node) error {
	pfi := pnd.GetFile()
	if pfi == nil {
		return fmt.Errorf("File attribute is empty. This is likely not a real file.")
	}

	modTime := time.Time{}
	if err := modTime.UnmarshalBinary(pfi.GetModTime()); err != nil {
		return err
	}

	f.id = pnd.GetID()
	f.size = pfi.GetFileSize()
	f.modTime = modTime
	f.hash = &Hash{pfi.GetHash()}
	f.parent = &Hash{pfi.GetParent()}
	f.key = pfi.GetKey()
	f.name = pfi.GetName()
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

func (f *File) SetSize(s uint64)       { f.size = s }
func (f *File) SetModTime(t time.Time) { f.modTime = t }
func (f *File) SetHash(h *Hash)        { f.hash = h }
func (f *File) SetName(n string)       { f.name = n }
func (f *File) SetKey(k []byte)        { f.key = k }

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
	return f.fs.FileByHash(f.parent)
}

// Parent returns the parent directory of File.
// If `f` is already the root, it will return itself (and never nil).
func (f *File) SetParent(parent Node) error {
	if parent == nil {
		return nil
	}

	f.parent = parent.Hash()
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
		nodePath(f),
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

// func (f *File) Print() {
// 	f.node.Walk(false, func(n *trie.Node) bool {
// 		child := n.Data.(*File)
// 		space := strings.Repeat(" ", int(n.Depth)*4)
// 		fmt.Printf("%s%s [%p]\n", space, child.node.Path(), child)
// 		return true
// 	})
// }
