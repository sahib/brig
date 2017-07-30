package nodes

import (
	"fmt"

	capnp_model "github.com/disorganizer/brig/model/nodes/capnp"
	h "github.com/disorganizer/brig/util/hashlib"
	capnp "zombiezen.com/go/capnproto2"
	pogs "zombiezen.com/go/capnproto2/pogs"
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

func (p *Person) ToCapnp() (*capnp.Message, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, err
	}

	person, err := capnp_model.NewRootPerson(seg)
	if err != nil {
		return nil, err
	}

	if err := pogs.Insert(capnp_model.Person_TypeID, person.Struct, p); err != nil {
		return nil, err
	}

	return msg, nil
}

func (p *Person) FromCapnp(msg *capnp.Message) error {
	root, err := msg.RootPtr()
	if err != nil {
		return err
	}

	if err := pogs.Extract(p, capnp_model.Person_TypeID, root.Struct()); err != nil {
		return err
	}

	return nil
}
