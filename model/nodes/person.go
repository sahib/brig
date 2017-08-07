package nodes

import (
	"fmt"

	capnp_model "github.com/disorganizer/brig/model/nodes/capnp"
	h "github.com/disorganizer/brig/util/hashlib"
	capnp "zombiezen.com/go/capnproto2"
)

type Person struct {
	Ident string
	Hash  h.Hash
}

func (p *Person) ID() string {
	return p.Ident
}

func (p *Person) GetHash() string {
	return p.Hash.B58String()
}

func (p *Person) String() string {
	hash := "<empty hash>"
	if p.Hash != nil {
		hash = p.Hash.B58String()
	}

	return fmt.Sprintf("<Person: %s (%v)>", p.Ident, hash)
}

// AuthorOfStage is the Person that is displayed for the stage commit.
// Currently this is just an empty hash Person that will be set later.
func AuthorOfStage() *Person {
	return &Person{
		Ident: "unknown",
		Hash:  h.EmptyHash.Clone(),
	}
}

func (p *Person) Equal(o *Person) bool {
	return p.Ident == o.Ident && o.Hash.Equal(p.Hash)
}

func (p *Person) ToCapnpPerson(seg *capnp.Segment) (*capnp_model.Person, error) {
	person, err := capnp_model.NewPerson(seg)
	if err != nil {
		return nil, err
	}

	person.SetIdent(p.Ident)
	person.SetHash(p.Hash)
	return &person, nil
}

func (p *Person) FromCapnpPerson(capnpers capnp_model.Person) error {
	ident, err := capnpers.Ident()
	if err != nil {
		return err
	}

	hash, err := capnpers.Hash()
	if err != nil {
		return err
	}

	p.Ident = ident
	p.Hash = hash
	return nil
}
