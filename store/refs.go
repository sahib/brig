package store

import (
	"fmt"

	"github.com/disorganizer/brig/store/wire"
	"github.com/gogo/protobuf/proto"
	"github.com/jbenet/go-multihash"
)

// Ref is a named reference to a commit
type Ref struct {
	Name string
	Hash *Hash
}

func (r *Ref) String() string {
	return fmt.Sprintf(
		"%s => %s",
		r.Name,
		r.Hash.B58String(),
	)
}

func (r *Ref) ToProto() (*wire.Ref, error) {
	return &wire.Ref{
		Name: proto.String(r.Name),
		Hash: r.Hash.Bytes(),
	}, nil
}

func (r *Ref) FromProto(pr *wire.Ref) error {
	mhash, err := multihash.Cast(pr.GetHash())
	if err != nil {
		return err
	}

	r.Name = *(pr.Name)
	r.Hash = &Hash{mhash}
	return nil
}

func (r *Ref) Marshal() ([]byte, error) {
	pr, err := r.ToProto()
	if err != nil {
		return nil, err
	}

	return proto.Marshal(pr)
}

func (r *Ref) Unmarshal(data []byte) error {
	pr := &wire.Ref{}
	if err := proto.Unmarshal(data, pr); err != nil {
		return err
	}

	return r.FromProto(pr)
}
