package nodes

import (
	"path"
	"time"

	capnp_model "github.com/disorganizer/brig/model/nodes/capnp"
	h "github.com/disorganizer/brig/util/hashlib"
	capnp "zombiezen.com/go/capnproto2"
)

// File represents a single file in the repository.
// It stores all metadata about it and links to the actual data.
type File struct {
	Base

	size   uint64
	parent string
	key    []byte
}

func NewEmptyFile(lkr Linker, parent *Directory, name string) (*File, error) {
	file := &File{
		Base: Base{
			name:     name,
			uid:      lkr.NextUID(),
			modTime:  time.Now(),
			nodeType: NodeTypeFile,
		},
		parent: parent.Path(),
	}

	return file, nil
}

func (f *File) ToCapnp() (*capnp.Message, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, err
	}

	node, err := capnp_model.NewRootNode(seg)
	if err != nil {
		return nil, err
	}

	if err := f.setBaseAttrsToNode(node); err != nil {
		return nil, err
	}

	capfile, err := capnp_model.NewFile(seg)
	if err != nil {
		return nil, err
	}

	capfile.SetParent(f.parent)
	capfile.SetKey(f.key)
	capfile.SetSize(f.size)
	node.SetFile(capfile)

	return msg, nil
}

func (f *File) FromCapnp(msg *capnp.Message) error {
	capnode, err := capnp_model.ReadRootNode(msg)
	if err != nil {
		return err
	}

	if err := f.parseBaseAttrsFromNode(capnode); err != nil {
		return err
	}

	capfile, err := capnode.File()
	if err != nil {
		return err
	}

	f.parent, err = capfile.Parent()
	if err != nil {
		return err
	}

	f.size = capfile.Size()
	f.key, err = capfile.Key()
	if err != nil {
		return err
	}

	return nil
}

////////////////// METADATA INTERFACE //////////////////

// Name returns the basename of the file.
func (f *File) Size() uint64 { return f.size }

////////////////// ATTRIBUTE SETTERS //////////////////

func (f *File) SetModTime(t time.Time) { f.modTime = t }
func (f *File) SetName(n string)       { f.name = n }
func (f *File) SetKey(k []byte)        { f.key = k }
func (f *File) SetSize(s uint64) {
	f.size = s
	f.SetModTime(time.Now())
}

func (f *File) SetHash(lkr Linker, h h.Hash) {
	oldHash := f.hash
	f.hash = h
	lkr.MemIndexSwap(f, oldHash)
	f.SetModTime(time.Now())
}

func (f *File) Path() string {
	return prefixSlash(path.Join(f.parent, f.name))
}

////////////////// HIERARCHY INTERFACE //////////////////

// NChildren returns the number of children this file node has.
func (f *File) NChildren(_ Linker) int {
	return 0
}

func (f *File) Child(_ Linker, name string) (Node, error) {
	// A file never has a child. Sad but true.
	return nil, nil
}

// Parent returns the parent directory of File.
// If `f` is already the root, it will return itself (and never nil).
func (f *File) Parent(lkr Linker) (Node, error) {
	return lkr.LookupNode(f.parent)
}

func (f *File) SetParent(_ Linker, parent Node) error {
	if parent == nil {
		return nil
	}

	f.parent = parent.Path()
	return nil
}

func (f *File) Key() []byte {
	return f.key
}

// Interface check for debugging:
var _ SettableNode = &File{}
var _ Streamable = &File{}
