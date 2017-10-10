using Go = import "/go.capnp";

@0xea883e7d5248d81b;
$Go.package("capnp");
$Go.import("github.com/disorganizer/brig/brigd/capnp");

struct StatInfo $Go.doc("StatInfo is a stat-like description of any node") {
    path    @0 :Text;
    hash    @1 :Data;
    size    @2 :UInt64;
    inode   @3 :UInt64;
    isDir   @4 :Bool;
    depth   @5 :Int32;
    modTime @6 :Text;
}

struct LogEntry $Go.doc("") {
    hash @0 :Data;
    msg  @1 :Text;
    tags @2 :List(Text);
    date @3 :Text;
}

interface FS {
    stage  @0 (localPath :Text, repoPath :Text);
    list   @1 (root :Text, maxDepth :Int32) -> (entries :List(StatInfo));
    cat    @2 (path :Text) -> (fifoPath :Text);
    mkdir  @3 (path :Text, createParents :Bool);
    remove @4 (path :Text);
    move   @5 (srcPath :Text, dstPath :Text);
}

interface VCS {
    log    @0 () -> (entries :List(LogEntry));
    commit @1 (msg :Text);
    tag    @2 (rev :Text, tagName :Text);
    untag  @3 (tagName :Text);
}

struct ConfigPair {
    key @0 :Text;
    val @1 :Text;
}

interface Meta {
    quit @0 ();
    ping @1 () -> (reply :Text);
    init @2 (basePath :Text, owner :Text, backend :Text);

    configGet @3 (key :Text) -> (value :Text);
    configSet @4 (key :Text, value :Text);
    configAll @5 () -> (all :List(ConfigPair));
}

# Group all interfaces together in one API object,
# because apparently we have this limitation what one interface
# more or less equals one connection.
interface API extends(FS, VCS, Meta) {
    version @0 () -> (version :Int32);
}
