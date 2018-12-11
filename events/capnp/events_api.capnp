using Go = import "/go.capnp";

@0xfc8938b535319bfe;
$Go.package("capnp");
$Go.import("github.com/sahib/brig/events/capnp");

struct Event $Go.doc("") {
    type @0 :Text;
}
