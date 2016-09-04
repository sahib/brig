package store

import (
	"fmt"
	"sync"
	"time"

	"github.com/disorganizer/brig/store/wire"
	"github.com/gogo/protobuf/proto"
)

type Directory struct {
	// TODO: Needed?
	sync.RWMutex

	name     string
	size     uint64
	modTime  time.Time
	parent   *Hash
	hash     *Hash
	children map[string]*Hash

	fs *FS
}

////////////// MARSHALLING ////////////////

func (d *Directory) ToProto() (*wire.Directory, error) {
	binModTime, err := d.modTime.MarshalBinary()
	if err != nil {
		return nil, err
	}

	binLinks := [][]byte{}
	binNames := []string{}

	for name, link := range d.children {
		binLinks = append(binLinks, link.Bytes())
		binNames = append(binNames, name)
	}

	return &wire.Directory{
		FileSize: proto.Uint64(d.size),
		ModTime:  binModTime,
		Hash:     d.hash.Bytes(),
		Parent:   d.parent.Bytes(),
		Links:    binLinks,
		Names:    binNames,
		Name:     proto.String(d.name),
	}, nil
}

func (d *Directory) Marshal() ([]byte, error) {
	pbd, err := d.ToProto()
	if err != nil {
		return nil, err
	}

	return proto.Marshal(pbd)
}

func (d *Directory) FromProto(pbd *wire.Directory) error {
	modTime := time.Time{}
	if err := modTime.UnmarshalBinary(pbd.GetModTime()); err != nil {
		return err
	}

	d.modTime = modTime
	d.parent = &Hash{pbd.GetParent()}
	d.size = pbd.GetFileSize()
	d.hash = &Hash{pbd.GetHash()}
	d.name = pbd.GetName()
	d.children = make(map[string]*Hash)

	// Find our place in the world:
	links := pbd.GetLinks()
	for idx, name := range pbd.GetNames() {
		// Be cautious, input might come from everywhere:
		if idx >= 0 && idx < len(links) {
			return fmt.Errorf("Malformed input: More or less names than links in `%s`", d.name)
		}

		d.children[name] = &Hash{links[idx]}
	}

	return nil
}

func (d *Directory) Unmarshal(data []byte) error {
	pbd := &wire.Directory{}
	if err := proto.Unmarshal(data, pbd); err != nil {
		return err
	}

	return d.FromProto(pbd)
}

////////////// NODE INTERFACE /////////////////

func (d *Directory) Name() string {
	return d.name
}

func (d *Directory) Hash() *Hash {
	return d.hash
}

func (d *Directory) Size() uint64 {
	return d.size
}

func (d *Directory) ModTime() time.Time {
	return d.modTime
}

func (d *Directory) NChildren() int {
	return len(d.children)
}

func (d *Directory) Child(name string) (Node, error) {
	// TODO: efficient lookup?
	return nil, nil
}

func (d *Directory) Parent() (Node, error) {
	return d.fs.DirectoryByHash(d.parent)
}

func (d *Directory) SetParent(nd Node) error {
	if nd == nil {
		d.parent = EmptyHash
	} else {
		d.parent = nd.Hash()
	}

	// TODO: error needed?
	return nil
}

////////////// TREE MOVEMENT /////////////////

func (d *Directory) Up(visit func(par *Directory) error) error {
	var err error

	for curr := d; curr.parent != nil; {
		if err := visit(curr); err != nil {
			return err
		}

		curr, err = d.fs.DirectoryByHash(curr.parent)
		if err != nil {
			return err
		}
	}

	return nil
}

//////////// STATE ALTERING METHODS //////////////

func (d *Directory) Add(nd Node) error {
	if err := nd.SetParent(d); err != nil {
		return err
	}

	d.children[nd.Name()] = nd.Hash()
	nodeSize := nd.Size()
	nodeHash := nd.Hash().Bytes()

	return d.Up(func(parent *Directory) error {
		parent.size += nodeSize
		return parent.hash.MixIn(nodeHash)
	})
}

// RemoveChild removes the child named `name` from it's children.
//
// Note that there is no general Remove() function that works on itself.
// It is therefore not possible (or a good idea) to remove the root node.
func (d *Directory) RemoveChild(nd Node) error {
	name := nd.Name()
	if _, ok := d.children[name]; !ok {
		return NoSuchFile(name)
	}

	// Unset parent from child:
	if err := nd.SetParent(nil); err != nil {
		return err
	}

	delete(d.children, name)

	nodeSize := nd.Size()
	nodeHash := nd.Hash().Bytes()

	return d.Up(func(parent *Directory) error {
		parent.size -= nodeSize
		return parent.hash.MixIn(nodeHash)
	})
}
