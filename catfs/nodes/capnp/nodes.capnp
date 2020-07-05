using Go = import "/go.capnp";

@0x9195d073cb5c5953;

$Go.package("capnp");
$Go.import("github.com/sahib/brig/catfs/nodes/capnp");

struct Commit $Go.doc("Commit is a set of changes to nodes") {
    # Following attributes will be part of the hash:
    message @0 :Text;
    author  @1 :Text;
    parent  @2 :Data;     # Hash to parent.
    root    @3 :Data;     # Hash to root directory.
    index   @4 :Int64;    # Total number of commits.

    # Attributes not being part of the hash:
    merge :group {
        with    @5 :Text;
        head    @6 :Data;
    }
}

struct DirEntry $Go.doc("A single directory entry") {
    name @0 :Text;
    hash @1 :Data;
}

struct Directory $Go.doc("Directory contains one or more directories or files") {
    size       @0 :UInt64;
    cachedSize @1 :UInt64;
    parent     @2 :Text;
    children   @3 :List(DirEntry);
    contents   @4 :List(DirEntry);
}

struct File $Go.doc("A leaf node in the MDAG") {
    size       @0 :UInt64;
    cachedSize @1 :UInt64;
    parent     @2 :Text;
    key        @3 :Data;
}

struct Ghost $Go.doc("Ghost indicates that a certain node was at this path once") {
    ghostInode @0 :UInt64;
    ghostPath  @1 :Text;

    union {
        commit    @2 :Commit;
        directory @3 :Directory;
        file      @4 :File;
    }
}

struct Node $Go.doc("Node is a node in the merkle dag of brig") {
    name        @0 :Text;
    treeHash    @1 :Data;
    modTime     @2 :Text;     # Time as ISO8601
    inode       @3 :UInt64;
    contentHash @4 :Data;
    user        @5 :Text;

    union {
        commit    @6 :Commit;
        directory @7 :Directory;
        file      @8 :File;
        ghost     @9 :Ghost;
    }

    backendHash @10 :Data;
}
