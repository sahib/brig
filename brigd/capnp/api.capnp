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

struct LogEntry $Go.doc("Single log entry") {
    hash @0 :Data;
    msg  @1 :Text;
    tags @2 :List(Text);
    date @3 :Text;
}

struct ConfigPair $Go.doc("Key/Value pair in the config") {
    key @0 :Text;
    val @1 :Text;
}

struct HistoryEntry $Go.doc("One History entry for a file") {
    path   @0 :Text;
    change @1 :Text;
    ref    @2 :Data;
}

struct DiffPair {
    src @0 :StatInfo;
    dst @1 :StatInfo;
}

struct Diff {
    added   @0 :List(StatInfo);
    removed @1 :List(StatInfo);
    ignored @2 :List(StatInfo);

    merged   @3 :List(DiffPair);
    conflict @4 :List(DiffPair);
}

struct RemoteFolder {
    folder @0 :Text;
    perms  @1 :Text;
}

struct Remote {
    name        @0 :Text;
    fingerprint @1 :Text;
    folders     @2 :List(RemoteFolder);
}

interface FS {
    stage    @0 (localPath :Text, repoPath :Text);
    list     @1 (root :Text, maxDepth :Int32) -> (entries :List(StatInfo));
    cat      @2 (path :Text) -> (port :Int32);
    mkdir    @3 (path :Text, createParents :Bool);
    remove   @4 (path :Text);
    move     @5 (srcPath :Text, dstPath :Text);
    pin      @6 (path :Text);
    unpin    @7 (path :Text);
    isPinned @8 (path :Text) -> (isPinned :Bool);
}

interface VCS {
    log      @0 () -> (entries :List(LogEntry));
    commit   @1 (msg :Text);
    tag      @2 (rev :Text, tagName :Text);
    untag    @3 (tagName :Text);
    reset    @4 (path :Text, rev :Text);
    checkout @5 (rev :Text, force :Bool);
    history  @6 (path :Text) -> (history :List(HistoryEntry));
    makeDiff @7 (remoteOwner :Text, headRevOwn :Text, headRevRemote :Text) -> (diff :Diff);
}

interface Meta {
    quit    @0 ();
    ping    @1 () -> (reply :Text);
    init    @2 (basePath :Text, owner :Text, backend :Text, password :Text);
    mount   @3 (mountPath :Text);
    unmount @4 (mountPath :Text);

    configGet @5 (key :Text) -> (value :Text);
    configSet @6 (key :Text, value :Text);
    configAll @7 () -> (all :List(ConfigPair));

    remoteAdd    @8  (remote :Remote);
    remoteRm     @9  (name :Text);
    remoteLs     @10 () -> (remotes :List(Remote));
    remoteSave   @11 (remotes :List(Remote));
    remoteLocate @12 (who :Text) -> (candidates :List(Remote));
    remoteSelf   @13 () -> (self :Remote);
}

# Group all interfaces together in one API object,
# because apparently we have this limitation what one interface
# more or less equals one connection.
interface API extends(FS, VCS, Meta) {
    version @0 () -> (version :Int32);
}
