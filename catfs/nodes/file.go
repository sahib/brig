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

	size       uint64
	cachedSize uint64 // MaxUint64 indicates that it is unkown
	parent     string
	key        []byte
}

// NewEmptyFile returns a newly created file under `parent`, named `name`.
func NewEmptyFile(parent *Directory, name string, user string, inode uint64) *File {
	return &File{
		Base: Base{
			name:     name,
			user:     user,
			inode:    inode,
			modTime:  time.Now().Truncate(time.Microsecond),
			nodeType: NodeTypeFile,
		},
		parent: parent.Path(),
	}
}

// ToCapnp converts a file to a capnp message.
func (f *File) ToCapnp() (*capnp.Message, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, err
	}

	capNd, err := capnp_model.NewRootNode(seg)
	if err != nil {
		return nil, err
	}

	return msg, f.ToCapnpNode(seg, capNd)
}

// ToCapnpNode converts this node to a serializable capnp proto node.
func (f *File) ToCapnpNode(seg *capnp.Segment, capNd capnp_model.Node) error {
	if err := f.setBaseAttrsToNode(capNd); err != nil {
		return err
	}

	capFile, err := f.setFileAttrs(seg)
	if err != nil {
		return err
	}

	return capNd.SetFile(*capFile)
}

func (f *File) setFileAttrs(seg *capnp.Segment) (*capnp_model.File, error) {
	capFile, err := capnp_model.NewFile(seg)
	if err != nil {
		return nil, err
	}

	if err := capFile.SetParent(f.parent); err != nil {
		return nil, err
	}

	if err := capFile.SetKey(f.key); err != nil {
		return nil, err
	}

	capFile.SetSize(f.size)
	capFile.SetCachedSize(f.cachedSize)
	return &capFile, nil
}

// FromCapnp sets all state of `msg` into the file.
func (f *File) FromCapnp(msg *capnp.Message) error {
	capNd, err := capnp_model.ReadRootNode(msg)
	if err != nil {
		return err
	}

	return f.FromCapnpNode(capNd)
}

// FromCapnpNode converts a serialized node to a normal node.
func (f *File) FromCapnpNode(capNd capnp_model.Node) error {
	if err := f.parseBaseAttrsFromNode(capNd); err != nil {
		return err
	}

	capFile, err := capNd.File()
	if err != nil {
		return err
	}

	return f.readFileAttrs(capFile)
}

func (f *File) readFileAttrs(capFile capnp_model.File) error {
	var err error

	f.parent, err = capFile.Parent()
	if err != nil {
		return err
	}

	f.nodeType = NodeTypeFile
	f.size = capFile.Size()
	f.cachedSize = capFile.CachedSize()
	f.key, err = capFile.Key()
	return err
}

////////////////// METADATA INTERFACE //////////////////

// Size returns the number of bytes in the file's content.
func (f *File) Size() uint64 { return f.size }

// Size returns the number of bytes in the file's backend storage.
func (f *File) CachedSize() uint64 { return f.cachedSize }

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

// SetSize will update the size of the file and update it's mod time.
func (f *File) SetCachedSize(s uint64) {
	f.cachedSize = s
	f.SetModTime(time.Now())
}


// Copy copies the contents of the file, except `inode`.
func (f *File) Copy(inode uint64) ModNode {
	if f == nil {
		return nil
	}

	var copyKey []byte
	if f.key != nil {
		copyKey = make([]byte, len(f.key))
		copy(copyKey, f.key)
	}

	return &File{
		Base:   f.Base.copyBase(inode),
		size:   f.size,
		cachedSize:   f.cachedSize,
		parent: f.parent,
		key:    copyKey,
	}
}

func (f *File) rehash(lkr Linker, newPath string) {
	oldHash := f.tree.Clone()
	var contentHash h.Hash
	if f.Base.content != nil {
		contentHash = f.Base.content.Clone()
	} else {
		contentHash = h.EmptyInternalHash.Clone()
	}

	f.tree = h.Sum([]byte(fmt.Sprintf("%s|%s", newPath, contentHash)))
	lkr.MemIndexSwap(f, oldHash, true)
}

// NotifyMove should be called when the node moved parents.
func (f *File) NotifyMove(lkr Linker, newParent *Directory, newPath string) error {
	dirname, basename := path.Split(newPath)
	f.SetName(basename)
	f.parent = dirname
	f.rehash(lkr, newPath)

	if newParent != nil {
		if err := newParent.Add(lkr, f); err != nil {
			return err
		}

		newParent.rebuildOrderCache()
	}

	return nil
}

// SetContent will update the hash of the file (and also the mod time)
func (f *File) SetContent(lkr Linker, content h.Hash) {
	f.Base.content = content
	f.rehash(lkr, f.Path())
	f.SetModTime(time.Now())
}

// SetBackend will update the hash of the file (and also the mod time)
func (f *File) SetBackend(lkr Linker, backend h.Hash) {
	f.Base.backend = backend
	f.SetModTime(time.Now())
}

func (f *File) String() string {
	return fmt.Sprintf("<file %s:%s:%d>", f.Path(), f.TreeHash(), f.Inode())
}

// Path will return the absolute path of the file.
func (f *File) Path() string {
	return prefixSlash(path.Join(f.parent, f.name))
}

////////////////// HIERARCHY INTERFACE //////////////////

// NChildren returns the number of children this file node has.
func (f *File) NChildren() int {
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

// SetUser sets the user that last modified the file.
func (f *File) SetUser(user string) {
	f.Base.user = user
}

// Interface check for debugging:
var _ ModNode = &File{}
var _ Streamable = &File{}
