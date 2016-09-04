package store

import (
	"sync"
	"time"

	"github.com/disorganizer/brig/store/wire"
	"github.com/gogo/protobuf/proto"
)

type Directory struct {
	// TODO: Needed?
	sync.RWMutex

	name    string
	size    uint64
	modTime time.Time
	parent  *Hash
	hash    *Hash
	links   []*Hash

	fs *FS
}

////////////// MARSHALLING ////////////////

func (d *Directory) ToProto() (*wire.Directory, error) {
	binModTime, err := d.modTime.MarshalBinary()
	if err != nil {
		return nil, err
	}

	binLinks := [][]byte{}
	for _, link := range d.links {
		binLinks = append(binLinks, link.Bytes())
	}

	return &wire.Directory{
		FileSize: proto.Uint64(d.size),
		ModTime:  binModTime,
		Hash:     d.hash.Bytes(),
		Parent:   d.parent.Bytes(),
		Links:    binLinks,
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

	d.links = []*Hash{}
	for _, link := range pbd.GetLinks() {
		d.links = append(d.links, &Hash{link})
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
	return len(d.links)
}

func (d *Directory) Child(name string) (Node, error) {
	// TODO: efficient lookup?
	return nil, nil
}

func (d *Directory) Parent() (Node, error) {
	return d.fs.DirectoryByHash(d.parent)
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

/////////////////////////////////

// TODO: move to test
func (d *Directory) Equal(o *Directory) bool {
	return d.name == o.name && d.hash.Equal(o.hash) && d.size == o.size && d.modTime.Equal(o.modTime)
}

func (d *Directory) Insert(nd Node) error {
	d.links = append(d.links, nd.Hash())

	nodeSize := nd.Size()
	nodeHash := nd.Hash().Bytes()

	err := d.Up(func(parent *Directory) error {
		parent.size += nodeSize
		return parent.hash.MixIn(nodeHash)
	})

	if err != nil {
		return err
	}

	return nil
}

func (d *Directory) Remove(name string) {
	// Update hash sizes
}
