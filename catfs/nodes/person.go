package nodes

import (
	"fmt"

	capnp_model "github.com/disorganizer/brig/cafs/nodes/capnp"
	h "github.com/disorganizer/brig/util/hashlib"
	capnp "zombiezen.com/go/capnproto2"
)

// Person is a link between a readable name and it's hash identifier.
// It might for example be used as the commit author.
type Person struct {
	ident string
	hash  h.Hash
}

// ID returns the person's identifier.
func (p *Person) ID() string {
	return p.ident
}

// Hash returns the hash identifier of this person.
func (p *Person) Hash() h.Hash {
	return p.hash
}

func (p *Person) String() string {
	hashStr := "<empty hash>"
	if p.hash != nil {
		hashStr = p.Hash().B58String()
	}

	return fmt.Sprintf("<Person: %s (%v)>", p.ident, hashStr)
}

// AuthorOfStage is the Person that is displayed for the stage commit.
// Currently this is just an empty hash Person that will be set later.
func AuthorOfStage() *Person {
	return &Person{
		ident: "unknown",
		hash:  h.EmptyHash.Clone(),
	}
}

// Equal checks if both person structs are equal (same display name and identifier)
// Neither Person may be nil.
func (p *Person) Equal(o *Person) bool {
	return p.ident == o.ident && o.hash.Equal(p.hash)
}

// ToCapnpPerson converts a person to a capnp-Person.
func (p *Person) ToCapnpPerson(seg *capnp.Segment) (*capnp_model.Person, error) {
	person, err := capnp_model.NewPerson(seg)
	if err != nil {
		return nil, err
	}

	if err := person.SetIdent(p.ident); err != nil {
		return nil, err
	}

	if err := person.SetHash(p.hash); err != nil {
		return nil, err
	}

	return &person, nil
}

func (p *Person) ToBytes() ([]byte, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, err
	}

	if _, err := p.ToCapnpPerson(seg); err != nil {
		return nil, err
	}

	return msg.Marshal()
}

func PersonFromBytes(data []byte) (*Person, error) {
	msg, err := capnp.Unmarshal(data)
	if err != nil {
		return nil, err
	}

	capperson, err := capnp_model.ReadRootPerson(msg)
	if err != nil {
		return nil, err
	}

	person := &Person{}
	if err := person.FromCapnpPerson(capperson); err != nil {
		return nil, err
	}

	return person, nil
}

// FromCapnpPerson converts a capnp-Person to a person.
func (p *Person) FromCapnpPerson(capnpers capnp_model.Person) error {
	ident, err := capnpers.Ident()
	if err != nil {
		return err
	}

	hash, err := capnpers.Hash()
	if err != nil {
		return err
	}

	p.ident = ident
	p.hash = hash
	return nil
}
