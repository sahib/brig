package nodes

import (
	"errors"
	"fmt"
	"path"
	"sort"
	"strings"
	"time"

	ie "github.com/sahib/brig/catfs/errors"
	capnp_model "github.com/sahib/brig/catfs/nodes/capnp"
	h "github.com/sahib/brig/util/hashlib"
	capnp "zombiezen.com/go/capnproto2"
)

// Directory is a typical directory that may contain
// several other directories or files.
type Directory struct {
	Base

	size       uint64
	cachedSize uint64 // MaxUint64 indicates that it is unkown
	parentName string
	children   map[string]h.Hash
	contents   map[string]h.Hash
	order      []string
}

// NewEmptyDirectory creates a new empty directory that does not exist yet.
func NewEmptyDirectory(
	lkr Linker, parent *Directory, name string, user string, inode uint64,
) (*Directory, error) {
	absPath := ""
	if parent != nil {
		absPath = path.Join(parent.Path(), name)
	}

	newDir := &Directory{
		Base: Base{
			inode:    inode,
			user:     user,
			tree:     h.Sum([]byte(absPath)),
			content:  h.EmptyInternalHash.Clone(),
			backend:  h.EmptyBackendHash.Clone(),
			name:     name,
			nodeType: NodeTypeDirectory,
			modTime:  time.Now().Truncate(time.Microsecond),
		},
		children: make(map[string]h.Hash),
		contents: make(map[string]h.Hash),
		order:    []string{},
	}

	if parent != nil {
		// parentName is set by Add:
		if err := parent.Add(lkr, newDir); err != nil {
			return nil, err
		}
	}

	return newDir, nil
}

func (d *Directory) String() string {
	return fmt.Sprintf("<dir %s:%s:%d>", d.Path(), d.TreeHash(), d.Inode())
}

// ToCapnp converts the directory to an easily serializable capnp message.
func (d *Directory) ToCapnp() (*capnp.Message, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, err
	}

	capNd, err := capnp_model.NewRootNode(seg)
	if err != nil {
		return nil, err
	}

	return msg, d.ToCapnpNode(seg, capNd)
}

// ToCapnpNode converts this node to a serializable capnp proto node.
func (d *Directory) ToCapnpNode(seg *capnp.Segment, capNd capnp_model.Node) error {
	if err := d.setBaseAttrsToNode(capNd); err != nil {
		return err
	}

	capDir, err := d.setDirectoryAttrs(seg)
	if err != nil {
		return err
	}

	return capNd.SetDirectory(*capDir)
}

func (d *Directory) setDirectoryAttrs(seg *capnp.Segment) (*capnp_model.Directory, error) {
	capDir, err := capnp_model.NewDirectory(seg)
	if err != nil {
		return nil, err
	}

	children, err := capnp_model.NewDirEntry_List(seg, int32(len(d.children)))
	if err != nil {
		return nil, err
	}

	entryIdx := 0
	for name, hash := range d.children {
		entry, err := capnp_model.NewDirEntry(seg)
		if err != nil {
			return nil, err
		}

		if err := entry.SetName(name); err != nil {
			return nil, err
		}
		if err := entry.SetHash(hash); err != nil {
			return nil, err
		}
		if err := children.Set(entryIdx, entry); err != nil {
			return nil, err
		}
		entryIdx++
	}

	if err := capDir.SetChildren(children); err != nil {
		return nil, err
	}

	contents, err := capnp_model.NewDirEntry_List(seg, int32(len(d.contents)))
	if err != nil {
		return nil, err
	}

	entryIdx = 0
	for name, hash := range d.contents {
		entry, err := capnp_model.NewDirEntry(seg)
		if err != nil {
			return nil, err
		}

		if err := entry.SetName(name); err != nil {
			return nil, err
		}
		if err := entry.SetHash(hash); err != nil {
			return nil, err
		}
		if err := contents.Set(entryIdx, entry); err != nil {
			return nil, err
		}

		entryIdx++
	}

	if err := capDir.SetContents(contents); err != nil {
		return nil, err
	}

	if err := capDir.SetParent(d.parentName); err != nil {
		return nil, err
	}

	capDir.SetSize(d.size)
	capDir.SetCachedSize(d.size)
	return &capDir, nil
}

// FromCapnp will take the result of ToCapnp and set all of it's attributes.
func (d *Directory) FromCapnp(msg *capnp.Message) error {
	capNd, err := capnp_model.ReadRootNode(msg)
	if err != nil {
		return err
	}

	return d.FromCapnpNode(capNd)
}

// FromCapnpNode converts a serialized node to a normal node.
func (d *Directory) FromCapnpNode(capNd capnp_model.Node) error {
	if err := d.parseBaseAttrsFromNode(capNd); err != nil {
		return err
	}

	capDir, err := capNd.Directory()
	if err != nil {
		return err
	}

	return d.readDirectoryAttr(capDir)
}

func (d *Directory) readDirectoryAttr(capDir capnp_model.Directory) error {
	var err error

	d.size = capDir.Size()
	d.cachedSize = capDir.CachedSize()
	d.parentName, err = capDir.Parent()
	if err != nil {
		return err
	}

	childList, err := capDir.Children()
	if err != nil {
		return err
	}

	d.children = make(map[string]h.Hash)
	for i := 0; i < childList.Len(); i++ {
		entry := childList.At(i)
		name, err := entry.Name()
		if err != nil {
			return err
		}

		hash, err := entry.Hash()
		if err != nil {
			return err
		}

		d.children[name] = hash
		d.order = append(d.order, name)
	}

	contentList, err := capDir.Contents()
	if err != nil {
		return err
	}

	d.contents = make(map[string]h.Hash)
	for i := 0; i < contentList.Len(); i++ {
		entry := contentList.At(i)
		name, err := entry.Name()
		if err != nil {
			return err
		}

		hash, err := entry.Hash()
		if err != nil {
			return err
		}

		d.contents[name] = hash
	}

	sort.Strings(d.order)
	d.nodeType = NodeTypeDirectory
	return nil
}

////////////// NODE INTERFACE /////////////////

// Name returns the dirname of this directory.
func (d *Directory) Name() string {
	return d.name
}

// Size returns the accumulated size of the directory
// (i.e. the sum of a files in it, excluding ghosts)
func (d *Directory) Size() uint64 {
	return d.size
}

// CachedSize is similar to Size() above but for accumulated backends storage
func (d *Directory) CachedSize() uint64 {
	return d.cachedSize 
}


// Path returns the full path of this node.
func (d *Directory) Path() string {
	return prefixSlash(path.Join(d.parentName, d.Base.name))
}

// NChildren returns the number of children the directory has.
func (d *Directory) NChildren() int {
	return len(d.children)
}

// Child returns a specific child with `name` or nil, if it was not found.
func (d *Directory) Child(lkr Linker, name string) (Node, error) {
	childHash, ok := d.children[name]
	if !ok {
		return nil, nil
	}

	return lkr.NodeByHash(childHash)
}

// Parent will return the parent of this directory or nil,
// if this directory is already the root directory.
func (d *Directory) Parent(lkr Linker) (Node, error) {
	if d.parentName == "" {
		return nil, nil
	}

	return lkr.LookupNode(d.parentName)
}

// SetParent will set the parent of this directory to `nd`.
func (d *Directory) SetParent(lkr Linker, nd Node) error {
	if d.Path() == "/" {
		return nil
	}

	if nd == nil {
		d.parentName = ""
	} else {
		d.parentName = nd.Path()
	}

	return nil
}

// ////////////// TREE MOVEMENT /////////////////

// VisitChildren will call `fn` for each of it's direct children.
// The order of visits is lexicographical based on the child name.
func (d *Directory) VisitChildren(lkr Linker, fn func(nd Node) error) error {
	for _, name := range d.order {
		hash := d.children[name]
		child, err := lkr.NodeByHash(hash)
		if err != nil {
			return err
		}

		if child == nil {
			return fmt.Errorf("BUG: dead link in tree: %s => %s", name, hash.B58String())
		}

		if err := fn(child); err != nil {
			return err
		}
	}

	return nil
}

// ChildrenSorted returns a list of children node objects, sorted lexically by
// their path. Use this whenever you want to have a defined order of nodes,
// but do not really care what order.
func (d *Directory) ChildrenSorted(lkr Linker) ([]Node, error) {
	children := []Node{}
	err := d.VisitChildren(lkr, func(nd Node) error {
		children = append(children, nd)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return children, nil
}

// Up will call `visit` for each node onto the way top to the root node,
// including this directory.
func (d *Directory) Up(lkr Linker, visit func(par *Directory) error) error {
	root, err := lkr.Root()
	if err != nil {
		return err
	}

	elems := strings.Split(d.Path(), "/")
	dirs := []*Directory{root}
	curr := root

	for _, elem := range elems {
		if elem == "" {
			continue
		}

		childHash, ok := curr.children[elem]
		if !ok {
			// This usually means that some link is missing.
			return fmt.Errorf("bug: cannot reach self from root in up()")
		}

		childNode, err := lkr.NodeByHash(childHash)
		if err != nil {
			return err
		}

		child, ok := childNode.(*Directory)
		if !ok {
			return fmt.Errorf("bug: non-directory in up(): %v", childHash)
		}

		dirs = append(dirs, child)
		curr = child
	}

	// Visit the nodes in reverse order, self first, root last:
	for idx := len(dirs) - 1; idx >= 0; idx-- {
		if err := visit(dirs[idx]); err != nil {
			return err
		}
	}

	return nil
}

// IsRoot returns true if this directory is the root directory.
func (d *Directory) IsRoot() bool {
	return d.parentName == ""
}

// ErrSkipChild can be returned inside a Walk() closure to stop descending
// recursively into a directory.
var ErrSkipChild = errors.New("skip sub directory")

// Walk calls `visit` for each node below `node`, including `node`.
// If `dfs` is true, depth first search will be used.
// If `dfs` is false, breadth first search will be used.
// It is valid to pass a File to Walk(), then visit will be called exactly once.
//
// It is possible to return the special error value ErrSkipChild in the callback.
// In this case, the children of this node are skipped.
// For this to work, `dfs` has to be false.
func Walk(lkr Linker, node Node, dfs bool, visit func(child Node) error) error {
	if node == nil {
		return nil
	}

	if node.Type() != NodeTypeDirectory {
		err := visit(node)
		if err == ErrSkipChild {
			return nil
		}

		return err
	}

	d, ok := node.(*Directory)
	if !ok {
		return ie.ErrBadNode
	}

	if !dfs {
		if err := visit(node); err != nil {
			if err == ErrSkipChild {
				return nil
			}

			return err
		}
	}

	for _, name := range d.order {
		hash := d.children[name]
		child, err := lkr.NodeByHash(hash)
		if err != nil {
			return err
		}

		if child == nil {
			return fmt.Errorf("walk: could not resolve %s (%s)", name, hash.B58String())
		}

		if err := Walk(lkr, child, dfs, visit); err != nil {
			return err
		}
	}

	if dfs {
		if err := visit(node); err != nil {
			if err == ErrSkipChild {
				panic("bug: you cannot use dfs=true and ErrSkipChild together")
			}

			return err
		}
	}

	return nil
}

// Lookup will lookup `repoPath` relative to this directory.
func (d *Directory) Lookup(lkr Linker, repoPath string) (Node, error) {
	repoPath = prefixSlash(path.Clean(repoPath))
	elems := strings.Split(repoPath, "/")

	// Strip off the first empty field:
	elems = elems[1:]

	if len(elems) == 1 && elems[0] == "" {
		return d, nil
	}

	var curr Node = d
	var err error

	for idx, elem := range elems {
		curr, err = curr.Child(lkr, elem)
		if err != nil {
			return nil, err
		}

		if curr == nil {
			return nil, ie.NoSuchFile(repoPath)
		}

		// If the child is a ghost and we did not fully resolve the path
		// yet we stop here. If it's the ghost of a directory we could
		// resolve its children, but that would be confusing.
		if curr.Type() == NodeTypeGhost && idx != len(elems)-1 {
			return nil, ie.NoSuchFile(repoPath)
		}
	}

	return curr, nil
}

//////////// STATE ALTERING METHODS //////////////

// SetSize sets the size of this directory.
func (d *Directory) SetSize(size uint64) { d.size = size }

func (d *Directory) SetCachedSize(cachedSize uint64) { d.cachedSize = cachedSize }

// SetName will set the name of this directory.
func (d *Directory) SetName(name string) {
	d.name = name
}

// SetModTime will set a new mod time to this directory (i.e. "touch" it)
func (d *Directory) SetModTime(modTime time.Time) {
	d.Base.modTime = modTime.Truncate(time.Microsecond)
}

// Copy returns a copy of the directory with `inode` changed.
func (d *Directory) Copy(inode uint64) ModNode {
	children := make(map[string]h.Hash)
	contents := make(map[string]h.Hash)

	for name, hash := range d.children {
		children[name] = hash.Clone()
	}

	for name, hash := range d.contents {
		contents[name] = hash.Clone()
	}

	order := make([]string, len(d.order))
	copy(order, d.order)

	return &Directory{
		Base:       d.Base.copyBase(inode),
		size:       d.size,
		parentName: d.parentName,
		children:   children,
		contents:   contents,
		order:      order,
	}
}

func (d *Directory) rehash(lkr Linker, updateContentHash bool) error {
	newTreeHash := h.Sum([]byte(path.Join(d.parentName, d.name)))
	newContentHash := h.EmptyInternalHash.Clone()
	for _, name := range d.order {
		newTreeHash = newTreeHash.Mix(d.children[name])

		if childContent := d.contents[name]; updateContentHash && childContent != nil {
			// The child content might be nil in case of ghost.
			// Those should not add to the content calculation.
			newContentHash = newContentHash.Mix(childContent)
		}
	}

	oldHash := d.tree.Clone()
	d.tree = newTreeHash

	if updateContentHash {
		d.content = newContentHash
	}

	lkr.MemIndexSwap(d, oldHash, true)
	return nil
}

// Add `nd` to this directory using `lkr`.
func (d *Directory) Add(lkr Linker, nd Node) error {
	if nd == d {
		return fmt.Errorf("bug: attempting to add `%s` to itself", nd.Path())
	}

	if _, ok := d.children[nd.Name()]; ok {
		twin, err := d.Child(lkr, nd.Name())
		if err != nil {
			return ie.ErrExists
		}
		if twin.Type() != NodeTypeGhost {
			return ie.ErrExists
		}
		// Twin is a ghost. We delete it to clear space for a new (added) node.
		err = d.RemoveChild(lkr, twin)
		if err != nil {
			// the ghost twin stays and we report it as existing
			return ie.ErrExists
		}
	}

	nodeSize := nd.Size()
	nodeCachedSize := nd.CachedSize()
	nodeHash := nd.TreeHash()
	nodeContent := nd.ContentHash()

	d.children[nd.Name()] = nodeHash
	if nd.Type() != NodeTypeGhost {
		d.contents[nd.Name()] = nodeContent
	}

	nameIdx := sort.SearchStrings(d.order, nd.Name())
	suffix := append([]string{nd.Name()}, d.order[nameIdx:]...)
	d.order = append(d.order[:nameIdx], suffix...)

	var lastNd Node
	err := d.Up(lkr, func(parent *Directory) error {
		if nd.Type() != NodeTypeGhost {
			// Only add to the size if it's not a ghost.
			// They do not really count as size.
			// Same goes for the node content.
			parent.size += nodeSize
			parent.cachedSize += nodeCachedSize

		}

		if lastNd != nil {
			parent.children[lastNd.Name()] = lastNd.TreeHash()

			if nd.Type() != NodeTypeGhost {
				parent.contents[lastNd.Name()] = lastNd.ContentHash()
			}
		}

		if err := parent.rehash(lkr, true); err != nil {
			return err
		}

		lastNd = parent
		return nil
	})

	if err != nil {
		return err
	}

	// Establish the link between parent and child:
	return nd.SetParent(lkr, d)
}

// RemoveChild removes the child named `name` from it's children.
// There is no way to remove the root node.
func (d *Directory) RemoveChild(lkr Linker, nd Node) error {
	name := nd.Name()
	if _, ok := d.children[name]; !ok {
		return ie.NoSuchFile(name)
	}

	// Unset parent from child:
	if err := nd.SetParent(lkr, nil); err != nil {
		return err
	}

	// Delete it from orders and children.
	// This assumes that it definitely was part of orders before.
	delete(d.children, name)
	delete(d.contents, name)

	nameIdx := sort.SearchStrings(d.order, name)
	d.order = append(d.order[:nameIdx], d.order[nameIdx+1:]...)

	var lastNd Node
	nodeSize := nd.Size()
	nodeCachedSize := nd.CachedSize()
	return d.Up(lkr, func(parent *Directory) error {
		if nd.Type() != NodeTypeGhost {
			parent.size -= nodeSize
			parent.cachedSize -= nodeCachedSize
		}

		if lastNd != nil {
			parent.children[lastNd.Name()] = lastNd.TreeHash()

			if nd.Type() != NodeTypeGhost {
				parent.contents[lastNd.Name()] = lastNd.ContentHash()
			}
		}

		if err := parent.rehash(lkr, true); err != nil {
			return err
		}

		lastNd = parent
		return nil
	})
}

func (d *Directory) rebuildOrderCache() {
	d.order = []string{}
	for name := range d.children {
		d.order = append(d.order, name)
	}
	sort.Strings(d.order)
}

// NotifyMove should be called whenever a node is being moved.
func (d *Directory) NotifyMove(lkr Linker, newParent *Directory, newPath string) error {
	visited := map[string]Node{}
	oldRootPath := d.Path()

	err := Walk(lkr, d, true, func(child Node) error {
		oldChildPath := child.Path()
		newChildPath := path.Join(newPath, oldChildPath[len(oldRootPath):])
		visited[newChildPath] = child

		switch child.Type() {
		case NodeTypeDirectory:
			childDir, ok := child.(*Directory)
			if !ok {
				return ie.ErrBadNode
			}

			for name := range childDir.children {
				movedChildPath := path.Join(newChildPath, name)
				childDir.children[name] = visited[movedChildPath].TreeHash()
			}

			if err := childDir.rehash(lkr, false); err != nil {
				return err
			}

			dirname, basename := path.Split(newChildPath)
			childDir.parentName = dirname
			childDir.SetName(basename)
			return nil
		case NodeTypeFile:
			childFile, ok := child.(*File)
			if !ok {
				return ie.ErrBadNode
			}

			if err := childFile.NotifyMove(lkr, nil, newChildPath); err != nil {
				return err
			}
		case NodeTypeGhost:
			childGhost, ok := child.(*Ghost)
			if !ok {
				return ie.ErrBadNode
			}

			childGhost.SetGhostPath(newChildPath)
		default:
			return fmt.Errorf("bad node type in NotifyMove(): %d", child.Type())
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Fixup the links from the parents to the children:
	for nodePath, node := range visited {
		if parent, ok := visited[path.Dir(nodePath)]; ok {
			parentDir := parent.(*Directory)
			parentDir.children[path.Base(nodePath)] = node.TreeHash()
			parentDir.rebuildOrderCache()
		}
	}

	if err := newParent.Add(lkr, d); err != nil {
		return err
	}

	newParent.rebuildOrderCache()
	return nil
}

// SetUser sets the user that last modified the directory.
func (d *Directory) SetUser(user string) {
	d.Base.user = user
}

// Assert that Directory follows the Node interface:
var _ ModNode = &Directory{}
