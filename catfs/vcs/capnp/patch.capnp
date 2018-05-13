using Go = import "/go.capnp";
# using Nodes = import "../../nodes/capnp/nodes.capnp";

@0xb943b54bf1683782;

$Go.package("capnp");
$Go.import("github.com/sahib/brig/catfs/vcs/capnp");

struct Change $Go.doc("Change describes a single change") {
    mask        @0 :UInt64;
    head        @1 :Data;
    next        @2 :Data;
    curr        @3 :Data;
    referToPath @4 :Text;
}

struct Patch $Go.doc("Patch contains a single change") {
    from    @0 :Data;
    changes @1 :List(Change);
}
