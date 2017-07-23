package nodes

import (
	"fmt"

	capnp_model "github.com/disorganizer/brig/model/nodes/capnp"
	h "github.com/disorganizer/brig/util/hashlib"
	capnp "zombiezen.com/go/capnproto2"
	pogs "zombiezen.com/go/capnproto2/pogs"
)

type Author struct {
	Ident string
	Hash  h.Hash
}

func (a *Author) ID() string {
	return a.Ident
}

func (a *Author) GetHash() string {
	return a.Hash.B58String()
}

func (a *Author) String() string {
	hash := "<empty hash>"
	if a.Hash != nil {
		hash = a.Hash.B58String()
	}

	return fmt.Sprintf("<author: %s (%v)>", a.Ident, hash)
}

// AuthorOfStage is the author that is displayed for the stage commit.
// Currently this is just an empty hash author that will be set later.
func AuthorOfStage() *Author {
	return &Author{
		Ident: "unknown",
		Hash:  h.EmptyHash.Clone(),
	}
}

func (a *Author) ToCapnp() (*capnp.Message, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, err
	}

	author, err := capnp_model.NewRootAuthor(seg)
	if err != nil {
		return nil, err
	}

	if err := pogs.Insert(capnp_model.Author_TypeID, author.Struct, a); err != nil {
		return nil, err
	}

	return msg, nil
}

func (a *Author) FromCapnp(msg *capnp.Message) error {
	root, err := msg.RootPtr()
	if err != nil {
		return err
	}

	if err := pogs.Extract(a, capnp_model.Author_TypeID, root.Struct()); err != nil {
		return err
	}

	return nil
}
