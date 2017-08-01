using Go = import "/go.capnp";

@0x9195d073cb5c5953;

$Go.package("capnp");
$Go.import("github.com/disorganizer/brig/model/nodes/capnp");


struct Person  $Go.doc("Person might be any brig user") {
    ident @0 :Text;
    hash  @1 :Data;
}

struct Commit $Go.doc("Commit is a set of changes to nodes") {
    # Following attributes will be part of the hash:
    message @0 :Text;
    # author  @1 :Person;
    parent  @1 :Data;     # Hash to parent.
    root    @2 :Data;     # Hash to root directory.

    # Attributes not being part of the hash:
    merge :group {
        isMerge @3 :Bool;
        with    @4 :Person;
        hash    @5 :Data;
    }
}

struct Ghost $Go.doc("Ghost indicates that a certain node was at this path once") {
    nodeType @0 :Uint8;
}

struct Directory $Go.doc("Directory contains one or more directories or files") {
    size     @0 :Uint64;
    parent   @1 :Text;
    children @2 :List(Data);
}

struct Node $Go.doc("Node is a node in the merkle dag of brig") {
    name    @0 :Text;
    hash    @1 :Data;
    modTime @2 :Text;     # Time as ISO8601

    union {
        commit @3 :Commit;
        ghost  @4 :Ghost;
    }
}
