using Go = import "/go.capnp";

@0x9195d073cb5c5953;

$Go.package("capnp");
$Go.import("github.com/disorganizer/brig/model/nodes/capnp");


struct Author $Go.doc("Author is a person that changed something") {
    ident @0 :Text;
    hash  @1 :Data;
}
