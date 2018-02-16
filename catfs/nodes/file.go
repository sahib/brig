package nodes

import (
	"fmt"
	"path"
	"time"

	capnp_model "github.com/sahib/brig/catfs/nodes/capnp"
	h "github.com/sahib/brig/util/hashlib"
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

// NewEmptyFile returns a newly created file under `parent`, named `name`.
func NewEmptyFile(parent *Directory, name string, inode uint64) (*File, error) {
	file := &File{
		Base: Base{
			name:     name,
			inode:    inode,
			modTime:  time.Now().Truncate(time.Microsecond),
			nodeType: NodeTypeFile,
		},
		parent: parent.Path(),
	}

	return file, nil
}

// ToCapnp converts a file to a capnp message.
func (f *File) ToCapnp() (*capnp.Message, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, err
	}

	capnode, err := capnp_model.NewRootNode(seg)
	if err != nil {
		return nil, err
	}

	if err := f.setBaseAttrsToNode(capnode); err != nil {
		return nil, err
	}

	capfile, err := f.setFileAttrs(seg)
	if err != nil {
		return nil, err
	}

	if err := capnode.SetFile(*capfile); err != nil {
		return nil, err
	}

	return msg, nil
}

func (f *File) setFileAttrs(seg *capnp.Segment) (*capnp_model.File, error) {
	capfile, err := capnp_model.NewFile(seg)
	if err != nil {
		return nil, err
	}

	if err := capfile.SetParent(f.parent); err != nil {
		return nil, err
	}

	if err := capfile.SetKey(f.key); err != nil {
		return nil, err
	}

	capfile.SetSize(f.size)
	return &capfile, nil
}

// FromCapnp sets all state of `msg` into the file.
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

	return f.readFileAttrs(capfile)
}

func (f *File) readFileAttrs(capfile capnp_model.File) error {
	var err error

	f.parent, err = capfile.Parent()
	if err != nil {
		return err
	}

	f.nodeType = NodeTypeFile
	f.size = capfile.Size()
	f.key, err = capfile.Key()
	if err != nil {
		return err
	}

	return nil
}

////////////////// METADATA INTERFACE //////////////////

// Size returns the number of bytes in the file's content.
func (f *File) Size() uint64 { return f.size }

////////////////// ATTRIBUTE SETTERS //////////////////

// SetModTime udates the mod time of the file (i.e. "touch"es it)
func (f *File) SetModTime(t time.Time) {
	f.modTime = t.Truncate(time.Microsecond)
}

// SetName set the name of the file.
func (f *File) SetName(n string) { f.name = n }

// SetKey updates the key to a new value, taking ownership of the value.
func (f *File) SetKey(k []byte) { f.key = k }

// SetSize will update the size of the file and update it's mod time.
func (f *File) SetSize(s uint64) {
	f.size = s
	f.SetModTime(time.Now())
}

func (f *File) Copy(inode uint64) ModNode {
	return &File{
		Base:   f.Base.copyBase(inode),
		size:   f.size,
		parent: f.parent,
		key:    f.key,
	}
}

func (f *File) rehash(lkr Linker, newPath string) {
	oldHash := f.hash.Clone()
	var contentHash h.Hash
	if f.Base.content != nil {
		contentHash = f.Base.content.Clone()
	} else {
		contentHash = h.EmptyHash.Clone()
	}

	f.hash = h.Sum([]byte(fmt.Sprintf("%s|%s", newPath, contentHash)))
	lkr.MemIndexSwap(f, oldHash)
}

func (f *File) NotifyMove(lkr Linker, oldPath, newPath string) error {
	dirname, basename := path.Split(newPath)
	f.SetName(basename)
	f.parent = dirname
	f.rehash(lkr, newPath)
	return nil
}

// SetContent will update the hash of the file (and also the mod time)
func (f *File) SetContent(lkr Linker, content h.Hash) {
	f.Base.content = content
	f.rehash(lkr, f.Path())
	f.SetModTime(time.Now())
}

func (f *File) String() string {
	return fmt.Sprintf("<file %s:%s:%d>", f.Path(), f.Hash(), f.Inode())
}

// Path will return the absolute path of the file.
func (f *File) Path() string {
	return prefixSlash(path.Join(f.parent, f.name))
}

////////////////// HIERARCHY INTERFACE //////////////////

// NChildren returns the number of children this file node has.
func (f *File) NChildren(_ Linker) int {
	return 0
}

// Child will return always nil, since files don't have children.
func (f *File) Child(_ Linker, name string) (Node, error) {
	// A file never has a child. Sad but true.
	return nil, nil
}

// Parent returns the parent directory of File.
// If `f` is already the root, it will return itself (and never nil).
func (f *File) Parent(lkr Linker) (Node, error) {
	return lkr.LookupNode(f.parent)
}

// SetParent will set the parent of the file to `parent`.
func (f *File) SetParent(_ Linker, parent Node) error {
	if parent == nil {
		return nil
	}

	f.parent = parent.Path()
	return nil
}

// Key returns the current key of the file.
func (f *File) Key() []byte {
	return f.key
}

// Interface check for debugging:
var _ ModNode = &File{}
var _ Streamable = &File{}
