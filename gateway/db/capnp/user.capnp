using Go = import "/go.capnp";

@0xa0b1c18bd0f965c4;

$Go.package("capnp");
$Go.import("github.com/sahib/brig/gateway/db/capnp");

struct User {
	name         @0 :Text;
	passwordHash @1 :Text;
	salt         @2 :Text;
	folders      @3 :List(Text);
}
