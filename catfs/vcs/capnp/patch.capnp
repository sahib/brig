using Go = import "/go.capnp";
using Nodes = import "../../nodes/capnp/nodes.capnp";

@0xb943b54bf1683782;

$Go.package("capnp");
$Go.import("github.com/sahib/brig/catfs/vcs/capnp");

struct Change $Go.doc("Change describes a single change") {
    mask            @0 :UInt64;
    head            @1 :Nodes.Node;
    next            @2 :Nodes.Node;
    curr            @3 :Nodes.Node;
    movedTo         @4 :Text;
    wasPreviouslyAt @5 :Text;
}

struct Patch $Go.doc("Patch contains a single change") {
    fromIndex @0 :Int64;
    currIndex @1 :Int64;
    changes   @2 :List(Change);
}
