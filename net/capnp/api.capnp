using Go = import "/go.capnp";

@0x9bcb07fb35756ee6;
$Go.package("capnp");
$Go.import("github.com/sahib/brig/net/capnp");

interface Sync {
    fetchStore @0 () -> (data :Data);
}

interface Meta {
    ping    @0 () -> (reply :Text);
}

# Group all interfaces together in one API object,
# because apparently we have this limitation what one interface
# more or less equals one connection.
interface API extends(Sync, Meta) {
    version @0 () -> (version :Int32);
}
