using Go = import "/go.capnp";

@0xea883e7d5248d81b;
$Go.package("capnp");
$Go.import("github.com/disorganizer/brig/brigd/capnp");

struct Error $Go.doc("Error representation") {
    success @0 :Bool;
    what @1 :Text;
}

interface FS {
    stage @0 (abs_path :Text, repo_path :Text) -> (error :Error);
}

interface VCS {
}

interface Meta {
    quit @0 ();
    ping @1 () -> (reply :Text);
}

# Group all interfaces together in one API object,
# because apparently we have this limitation what one interface
# more or less equals one connection.
interface API extends(FS, VCS, Meta) {
    version @0 () -> (version :Int32);
}
