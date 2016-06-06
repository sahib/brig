package store

import (
	"fmt"

	"github.com/disorganizer/brig/store/wire"
	"github.com/gogo/protobuf/proto"
	"github.com/jbenet/go-multihash"
)

const (
	RefTypeInvalid = iota

	// RefTypeTag is a fixed, non-moving version
	RefTypeTag
	RefTypeBranch
)

type RefType int

func (rt RefType) String() string {
	switch rt {
	case RefTypeTag:
		return "tag"
	case RefTypeBranch:
		return "branch"
	default:
		return "invalid"
	}
}

// Ref is a named reference to a commit
type Ref struct {
	Name string
	Hash *Hash
	Type RefType
}

func (r *Ref) String() string {
	return fmt.Sprintf(
		"%s %s (%s)",
		r.Hash.B58String(),
		r.Name,
		r.Type.String(),
	)
}

func (r *Ref) ToProto() (*wire.Ref, error) {
	return &wire.Ref{
		Name: proto.String(r.Name),
		Hash: r.Hash.Bytes(),
		Type: proto.Int32(int32(r.Type)),
	}, nil
}

func (r *Ref) FromProto(pr *wire.Ref) error {
	mhash, err := multihash.Cast(pr.GetHash())
	if err != nil {
		return err
	}

	r.Name = *(pr.Name)
	r.Hash = &Hash{mhash}
	r.Type = RefType(*(pr.Type))
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
