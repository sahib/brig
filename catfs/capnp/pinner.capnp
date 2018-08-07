using Go = import "/go.capnp";

@0xba762188b0a6e4cf;

$Go.package("capnp");
$Go.import("github.com/sahib/brig/catfs/capnp");


struct Pin {
    inode    @0 :UInt64;
    isPinned @1 :Bool;
}

struct PinEntry $Go.doc("A single entry for a certain content node") {
    # Following attributes will be part of the hash:
    pins   @0 :List(Pin);
}
