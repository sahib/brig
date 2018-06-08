using Go = import "/go.capnp";

@0xea883e7d5248d81b;
$Go.package("capnp");
$Go.import("github.com/sahib/brig/server/capnp");

struct StatInfo $Go.doc("StatInfo is a stat-like description of any node") {
    path        @0  :Text;
    treeHash    @1  :Data;
    size        @2  :UInt64;
    inode       @3  :UInt64;
    isDir       @4  :Bool;
    depth       @5  :Int32;
    modTime     @6  :Text;
    isPinned    @7  :Bool;
    isExplicit  @8  :Bool;
    contentHash @9  :Data;
    user        @10 :Text;
    backendHash @11 :Data;
}

struct Commit $Go.doc("Single log entry") {
    hash @0 :Data;
    msg  @1 :Text;
    tags @2 :List(Text);
    date @3 :Text;
}

struct ConfigEntry $Go.doc("A config entry (including meta info)") {
    key          @0 :Text;
    val          @1 :Text;
    doc          @2 :Text;
    default      @3 :Text;
    needsRestart @4 :Bool;
}

struct Change $Go.doc("One History entry for a file") {
    path    @0 :Text;
    change  @1 :Text;
    head    @2 :Commit;
    next    @3 :Commit;
    referTo @4 :Text;
}

struct DiffPair $Go.doc("Represent two differing files") {
    src @0 :StatInfo;
    dst @1 :StatInfo;
}

struct Diff $Go.doc("Difference between two commits") {
    added   @0 :List(StatInfo);
    removed @1 :List(StatInfo);
    ignored @2 :List(StatInfo);
    missing @3 :List(StatInfo);

    moved    @4 :List(DiffPair);
    merged   @5 :List(DiffPair);
    conflict @6 :List(DiffPair);
}

struct RemoteFolder $Go.doc("A folder that a remote is allowed to access") {
    folder @0 :Text;
    perms  @1 :Text;
}

struct Remote $Go.doc("Info a remote peer we might sync with") {
    name        @0 :Text;
    fingerprint @1 :Text;
    folders     @2 :List(RemoteFolder);
}

# This is similar to a remote:
struct LocateResult {
    name        @0 :Text;
    addr        @1 :Text;
    mask        @2 :Text;
    fingerprint @3 :Text;
}

struct Identity $Go.doc("Info about our current user state") {
    currentUser @0 :Text;
    owner       @1 :Text;
    fingerprint @2 :Text;
    isOnline   @3  :Bool;
}

struct MountOptions {
    # For now empty, but there are some mount options
    # in planning.
}

struct PeerStatus $Go.doc("net status of a peer") {
    name        @0 :Text;
    addr        @1 :Text;
    lastSeen    @2 :Text;
    roundtripMs @3 :Int32;
    error       @4 :Text;
}

struct GarbageItem $Go.doc("A single item that was killed by the gc") {
    path    @0 :Text;
    content @1 :Data;
    owner   @2 :Text;
}

struct Version {
    serverVersion  @0 :Text;
    serverRev      @1 :Text;
    backendVersion @2 :Text;
    backendRev     @3 :Text;
}

struct ExplicitPin {
    path   @0 :Text;
    commit @1 :Text;
}

interface FS {
    stage             @0   (localPath :Text, repoPath :Text);
    list              @1   (root :Text, maxDepth :Int32) -> (entries :List(StatInfo));
    cat               @2   (path :Text) -> (port :Int32);
    mkdir             @3   (path :Text, createParents :Bool);
    remove            @4   (path :Text);
    move              @5   (srcPath :Text, dstPath :Text);
    copy              @6   (srcPath :Text, dstPath :Text);
    pin               @7   (path :Text);
    unpin             @8   (path :Text);
    stat              @9   (path :Text) -> (info :StatInfo);
    garbageCollect    @10  (aggressive :Bool) -> (freed :List(GarbageItem));
    touch             @11  (path :Text);
    exists            @12  (path :Text) -> (exists :Bool);
    listExplicitPins  @13  (prefix :Text, from :Text, to :Text) -> (pins :List(ExplicitPin));
    clearExplicitPins @14  (prefix :Text, from :Text, to :Text) -> (count :Int32);
    setExplicitPins   @15  (prefix :Text, from :Text, to :Text) -> (count :Int32);
}

interface VCS {
    log      @0 () -> (entries :List(Commit));
    commit   @1 (msg :Text);
    tag      @2 (rev :Text, tagName :Text);
    untag    @3 (tagName :Text);
    reset    @4 (path :Text, rev :Text, force :Bool);
    history  @5 (path :Text) -> (history :List(Change));
    makeDiff @6 (localOwner :Text, remoteOwner :Text, localRev :Text, remoteRev :Text, needFetch :Bool) -> (diff :Diff);
    sync     @7 (withWhom :Text, needFetch :Bool);
    fetch    @8 (who :Text);
}

interface Meta {
    quit    @0 ();
    ping    @1 () -> (reply :Text);
    init    @2 (basePath :Text, owner :Text, backend :Text, password :Text);
    mount   @3 (mountPath :Text, options :MountOptions);
    unmount @4 (mountPath :Text);

    configGet @5 (key :Text) -> (value :Text);
    configSet @6 (key :Text, value :Text);
    configAll @7 () -> (all :List(ConfigPair));

    remoteAdd    @8  (remote :Remote);
    remoteRm     @9  (name :Text);
    remoteLs     @10 () -> (remotes :List(Remote));
    remoteSave   @11 (remotes :List(Remote));
    remotePing   @12 (who :Text) -> (roundtrip :Float64);
    remoteClear  @13 ();

    netLocate     @14 (who :Text, timeoutSec :Float64, locateMask :Text) -> (ticket :UInt64);
    netLocateNext @15 (ticket :UInt64) -> (result :LocateResult);

    # the combined command of both is "whathaveibecome":
    whoami      @16  () -> (whoami :Identity);
    become      @17 (who :Text);

    connect     @18 ();
    disconnect  @19 ();
    onlinePeers @20 () -> (infos :List(PeerStatus));
    version     @21 () -> (version :Version);
}

# Group all interfaces together in one API object,
# because apparently we have this limitation what one interface
# more or less equals one connection.
interface API extends(FS, VCS, Meta) {
}
